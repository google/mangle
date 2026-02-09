package analysis

import (
	"fmt"

	"github.com/google/mangle/ast"
	"github.com/google/mangle/symbols"
)

type inferContext struct {
	bc *BoundsAnalyzer

	pred *ast.PredicateSym

	decl  *ast.Decl  // may be nil
	modes []ast.Mode // only if decl is not nil

	// clause is the clause for which we infer possible relation types.
	clause *ast.Clause
}

// NewInferContextNoDecl returns an inferContext for a predicate that does not have a declaration.
func newInferContextNoDecl(bc *BoundsAnalyzer, pred *ast.PredicateSym, clause *ast.Clause) *inferContext {
	return &inferContext{bc: bc, pred: pred, clause: clause}
}

// NewInferContextNoDecl returns an inferContext for a predicate that does not have a declaration.
func newInferContext(bc *BoundsAnalyzer, decl *ast.Decl, clause *ast.Clause) *inferContext {
	return &inferContext{bc: bc, pred: &decl.DeclaredAtom.Predicate, decl: decl, modes: decl.Modes(), clause: clause}
}

// inferState is state of inference while iterating over premises.
// The relation type is represented implicitly in usedVars and varTpe.
// Assigns to each var in usedVars a type (possibly union) in varTpe.
type inferState struct {
	// The index of the premise to be inspected with this state.
	index    int
	usedVars VarList
	varTpe   []ast.BaseTerm
}

func (s *inferState) String() string {
	return fmt.Sprintf("<%d; %v, %v>", s.index, s.usedVars, s.varTpe)
}

func (s *inferState) makeNext() *inferState {
	dest := make([]ast.BaseTerm, len(s.varTpe))
	for i, tpe := range s.varTpe {
		dest[i] = tpe
	}
	return &inferState{s.index + 1, s.usedVars, dest}
}

// addOrRefine either adds a binding or intersects type for an existing one.
func (s *inferState) addOrRefine(v ast.Variable, tpe ast.BaseTerm) error {
	if tpe.Equals(symbols.EmptyType) {
		return fmt.Errorf("variable %v has empty type", v)
	}
	if v.Symbol == "_" {
		return nil
	}
	i := s.usedVars.Find(v)
	if i == -1 {
		s.usedVars = s.usedVars.Extend([]ast.Variable{v})
		s.varTpe = append(s.varTpe, tpe)
		return nil
	}
	tpe = symbols.LowerBound(nil /*TODO*/, []ast.BaseTerm{s.varTpe[i], tpe})
	if tpe.Equals(symbols.EmptyType) {
		return fmt.Errorf("variable %v cannot have both %v and %v", v, s.varTpe[i], tpe)
	}
	s.varTpe[i] = tpe
	return nil
}

// refineNegative uses negative information to refine an existing binding
func (s *inferState) refineNegative(v ast.Variable, tpe ast.BaseTerm) error {
	if v.Symbol == "_" {
		return nil
	}
	i := s.usedVars.Find(v)
	if i == -1 {
		return nil
	}
	existing := s.varTpe[i]
	if existing.Equals(tpe) {
		return fmt.Errorf("variable %v bounded by %v cannot be refined with negative %v", v, s.varTpe[i], tpe)
	}
	if !symbols.IsUnionTypeExpression(existing) {
		return nil
	}
	newTpe, err := symbols.RemoveFromUnionType(tpe, existing)
	if err != nil {
		return err
	}
	if newTpe.Equals(symbols.EmptyType) {
		return fmt.Errorf("variable %v bounded by %v cannot be refined with negative %v", v, s.varTpe[i], tpe)
	}
	s.varTpe[i] = newTpe
	return nil
}

func (s *inferState) asMap() map[ast.Variable]ast.BaseTerm {
	m := make(map[ast.Variable]ast.BaseTerm, len(s.varTpe))
	for i, v := range s.usedVars.Vars {
		m[v] = s.varTpe[i]
	}
	return m
}

// inferRelTypesFromPremise is called for index \in 0..len(premises). It
// maps one state of inference to its (possibly empty) list of successors.
func (ic *inferContext) inferRelTypesFromPremise(premises []ast.Term, state *inferState) ([]*inferState, error) {
	bc := ic.bc
	var nextStates []*inferState

	// TODO: this should be piped through in inferState.
	typeCtx := map[ast.Variable]ast.BaseTerm{}

	premise := premises[state.index]
	switch t := premise.(type) {
	case ast.Atom:
		atom := t
		var (
			alternatives []ast.BaseTerm
			err          error
		)
		if atom.Predicate == symbols.MatchField {
			// The first argument *must* be bound, therefore we *must* have a type for it,
			// it must be a record type the selected field.

			scrutinee, ok := atom.Args[0].(ast.Variable)
			if !ok {
				return nil, fmt.Errorf(":match_field must be applied to a variable")
			}
			i := state.usedVars.Find(scrutinee)
			if i == -1 {
				// This cannot happen.
				return nil, fmt.Errorf(":match_field must be applied to a bound variable")
			}
			scrutineeType := state.varTpe[i]
			if scrutineeType == ast.AnyBound { // Give up.
				alternatives = []ast.BaseTerm{symbols.BuiltinRelations[symbols.MatchField]}
			} else {

				name, ok := atom.Args[1].(ast.Constant)
				if !ok || name.Type != ast.NameType {
					// This cannot happen.
					return nil, fmt.Errorf(":match_field field selector must be a constant")
				}
				projected, err := symbols.StructTypeField(scrutineeType, name)
				if err != nil {
					return nil, fmt.Errorf(":match_field struct type %v does not have field %v", scrutineeType, name)
				}

				alternatives = []ast.BaseTerm{symbols.NewRelType(scrutineeType, ast.NameBound, projected)}
			}

		} else if declared, ok := bc.relTypeMap[atom.Predicate]; ok {
			alternatives, _, err = bc.feasibleAlternatives(atom.Predicate, declared, atom.Args, state.asMap(), typeCtx)
		} else {
			alternatives, err = bc.getOrInferRelTypes(atom.Predicate, atom.Args, state.asMap(), typeCtx)
		}

		if err != nil {
			return nil, fmt.Errorf("type mismatch %v : %v ", premise, err)
		}
		for _, alternative := range alternatives {
			relTypeArgs, err := symbols.RelTypeArgs(alternative)
			if err != nil {
				return nil, err // This cannot happen.
			}
			// TODO: handle type variables.
			nextState := state.makeNext()
			for i, a := range atom.Args {
				if v, ok := a.(ast.Variable); ok {
					// No error-check needed - alternative is feasible.
					nextState.addOrRefine(v, relTypeArgs[i])
				}
			}
			nextStates = append(nextStates, nextState)
		}
		return nextStates, nil

	case ast.NegAtom:
		atom := t.Atom
		var (
			alternatives []ast.BaseTerm
			err          error
		)
		if declared, ok := bc.relTypeMap[atom.Predicate]; ok {
			alternatives, _, err = bc.feasibleAlternatives(atom.Predicate, declared, atom.Args, state.asMap(), typeCtx)
		} else {
			alternatives, err = bc.getOrInferRelTypes(atom.Predicate, atom.Args, state.asMap(), typeCtx)
		}
		if err != nil {
			return nil, fmt.Errorf("type mismatch %v : %v ", premise, err)
		}
		// For negated premise, there is never a variable bound so we never need to add
		// a binding. We can refine existing bindings by using negative information.
	nextAlternative:
		for _, alternative := range alternatives {
			relTypeArgs, err := symbols.RelTypeArgs(alternative)
			if err != nil {
				return nil, err // This cannot happen.
			}
			nextState := state.makeNext()
			if atom.Predicate == symbols.MatchPrefix {
				// Alternative is feasible, but information is "negative".
				// For :match_prefix (only), we can use this to refine the type.
				for i, a := range atom.Args {
					if v, ok := a.(ast.Variable); ok {
						if err := nextState.refineNegative(v, relTypeArgs[i]); err != nil {
							continue nextAlternative
						}
					}
				}
			}
			nextStates = append(nextStates, nextState)
		}
		return nextStates, nil

	case ast.TemporalLiteral:
		// We process the inner literal.
		// Construct a temporary state for the inner literal (index 0).
		tempState := &inferState{0, state.usedVars, state.varTpe}
		tempNextStates, err := ic.inferRelTypesFromPremise([]ast.Term{t.Literal}, tempState)
		if err != nil {
			return nil, err
		}
		// Fix up the indices in the result states to match the original sequence.
		for _, s := range tempNextStates {
			s.index = state.index + 1
			// Bind interval variables if present
			if t.Interval != nil {
				if t.Interval.Start.Type == ast.VariableBound {
					if err := s.addOrRefine(t.Interval.Start.Variable, ast.TimeBound); err != nil {
						return nil, err
					}
				}
				if t.Interval.End.Type == ast.VariableBound {
					if err := s.addOrRefine(t.Interval.End.Variable, ast.TimeBound); err != nil {
						return nil, err
					}
				}
			}
		}
		return tempNextStates, nil

	case ast.Eq:
		nextState := state.makeNext()
		varRanges := nextState.asMap()
		if leftVar, ok := t.Left.(ast.Variable); ok {
			tpe := boundOfArg(t.Right, varRanges, bc.nameTrie)
			if err := nextState.addOrRefine(leftVar, tpe); err != nil {
				return nil, err
			}
		}
		if rightVar, ok := t.Right.(ast.Variable); ok {
			tpe := boundOfArg(t.Left, varRanges, bc.nameTrie)
			if err := nextState.addOrRefine(rightVar, tpe); err != nil {
				return nil, err
			}
		}
		return []*inferState{nextState}, nil

	case ast.Ineq:
		nextState := state.makeNext()
		leftTpe := boundOfArg(t.Left, state.asMap(), bc.nameTrie)
		rightTpe := boundOfArg(t.Right, state.asMap(), bc.nameTrie)

		tpe := symbols.LowerBound(nil /* TODO */, []ast.BaseTerm{leftTpe, rightTpe})
		if tpe.Equals(symbols.EmptyType) {
			return nil, fmt.Errorf("type mismatch %v : left type %v right type %v", premise, leftTpe, rightTpe)
		}
		if leftVar, ok := t.Left.(ast.Variable); ok {
			if err := nextState.addOrRefine(leftVar, tpe); err != nil {
				return nil, err
			}
		}
		if rightVar, ok := t.Right.(ast.Variable); ok {
			if err := nextState.addOrRefine(rightVar, tpe); err != nil {
				return nil, err
			}
		}
		return []*inferState{nextState}, nil
	}
	return nil, fmt.Errorf("unexpected state %v", premise)
}

// inferRelTypesFromClause infers possible relation types for the head predicate of a single clause.
func (ic *inferContext) inferRelTypesFromClause() (ast.BaseTerm, error) {
	clause := ic.clause

	usedVars := VarList{}
	state := &inferState{0, usedVars, []ast.BaseTerm{}}

	if ic.decl != nil && len(ic.modes) == 1 {
		for i, m := range ic.modes[0] {
			v, ok := clause.Head.Args[i].(ast.Variable)
			if !ok {
				continue
			}
			if m == ast.ArgModeInput {
				// Treat as bound variable.
				state.usedVars = state.usedVars.Extend([]ast.Variable{v})
				types := []ast.BaseTerm{}
				if len(ic.decl.Bounds) > 0 {
					for _, b := range ic.decl.Bounds {
						tpe := b.Bounds[i]
						types = append(types, tpe)
					}
				} else {
					types = append(types, ast.AnyBound)
				}
				if len(types) == 1 {
					state.varTpe = append(state.varTpe, types[0])
				} else {
					state.varTpe = append(state.varTpe, symbols.NewUnionType(types...))
				}
			}
		}
	}
	levels := make([][]*inferState, len(clause.Premises)+1)
	levels[0] = []*inferState{state}
	for i := range clause.Premises {
		for _, state := range levels[i] {
			nextStates, err := ic.inferRelTypesFromPremise(clause.Premises, state)
			if err != nil {
				continue
			}
			levels[i+1] = append(levels[i+1], nextStates...)
		}
		if len(levels[i+1]) == 0 {
			return nil, fmt.Errorf("type mismatch: cannot find assignment that works for premise %v", clause.Premises[i])
		}
	}

	// Constrain variables used in HeadTime to be of type Time.
	if clause.HeadTime != nil {
		var nextStates []*inferState
		var lastErr error
		for _, s := range levels[len(clause.Premises)] {
			valid := true
			if clause.HeadTime.Start.Type == ast.VariableBound {
				if err := s.addOrRefine(clause.HeadTime.Start.Variable, ast.TimeBound); err != nil {
					valid = false
					lastErr = err
				}
			}
			if valid && clause.HeadTime.End.Type == ast.VariableBound {
				if err := s.addOrRefine(clause.HeadTime.End.Variable, ast.TimeBound); err != nil {
					valid = false
					lastErr = err
				}
			}
			if valid {
				nextStates = append(nextStates, s)
			}
		}
		if len(nextStates) == 0 {
			if lastErr != nil {
				return nil, fmt.Errorf("HeadTime variables must be of type Time: %w", lastErr)
			}
			return nil, fmt.Errorf("HeadTime variables must be of type Time")
		}
		levels[len(clause.Premises)] = nextStates
	}
	var relTypes []ast.BaseTerm
	for _, state := range levels[len(clause.Premises)] {
		s := state.makeNext()
		if clause.Transform != nil {
			for _, tr := range clause.Transform.Statements {
				if tr.Var != nil {
					s.addOrRefine(*tr.Var, typeOfFn(tr.Fn, s.asMap(), ic.bc.nameTrie))
				}
			}
		}

		headTuple := make([]ast.BaseTerm, len(clause.Head.Args))
		for i, arg := range clause.Head.Args {
			headTuple[i] = boundOfArg(arg, s.asMap(), ic.bc.nameTrie)
		}
		relTypes = append(relTypes, symbols.NewRelType(headTuple...))
	}
	return symbols.RelTypeFromAlternatives(relTypes), nil
}
