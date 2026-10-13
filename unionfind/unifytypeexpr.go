package unionfind

import (
	"fmt"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
)

type unionFindFun struct {
	ufrel UnionFind
	subst map[ast.Variable]ast.BaseTerm
}

// UnifyTypeExpr unifies two type expressions. These are terms that contain
// ApplyFn nodes.
//
// The first list holds the actual types, the second the declared types.
// The distinction matters for union type expressions: an actual type
// unifies with a declared union if it unifies with one of its
// alternatives, and an actual union unifies with a declared type if all
// of its alternatives do.
func UnifyTypeExpr(xs []ast.BaseTerm, ys []ast.BaseTerm) (map[ast.Variable]ast.BaseTerm, error) {
	if len(xs) != len(ys) {
		return nil, fmt.Errorf("not of equal size")
	}
	u := unionFindFun{UnionFind{make(map[ast.BaseTerm]ast.BaseTerm)}, make(map[ast.Variable]ast.BaseTerm)}
	if err := u.unifyFunctional(xs, ys); err != nil {
		return nil, err
	}
	for t := range u.ufrel.parent {
		v, ok := t.(ast.Variable)
		if !ok {
			continue
		}
		u.subst[v] = u.ufrel.find(v)
	}
	return u.subst, nil
}

func (u *unionFindFun) unifyFunctional(xs []ast.BaseTerm, ys []ast.BaseTerm) error {
	if len(xs) != len(ys) {
		return fmt.Errorf("cannot unify %v and %v: different number of arguments", xs, ys)
	}
	for i, x := range xs {
		if err := u.unifyOne(x, ys[i]); err != nil {
			return err
		}
	}
	return nil
}

// unifyOne unifies a single pair of type expressions. The first is
// treated as the actual type, the second as the declared type.
func (u *unionFindFun) unifyOne(x, y ast.BaseTerm) error {
	xApply, xOk := x.(ast.ApplyFn)
	yApply, yOk := y.(ast.ApplyFn)

	// Neither side is a type constructor application: constants and
	// variables. Mismatching constants are a unification error.
	if !xOk && !yOk {
		if x.Equals(ast.Variable{"_"}) || y.Equals(ast.Variable{"_"}) {
			return nil
		}
		// Register terms that are not known yet; resetting existing
		// entries would silently drop earlier bindings of x or y.
		if u.ufrel.find(x) == nil {
			u.ufrel.parent[x] = x
		}
		if u.ufrel.find(y) == nil {
			u.ufrel.parent[y] = y
		}
		if isWildcardBound(x) || isWildcardBound(y) {
			// /any and /bot are compatible with every other bound.
			// Union the two, so that variables bind to the wildcard.
			u.ufrel.union(x, y)
			return nil
		}
		return unifyTermsUpdate([]ast.BaseTerm{x}, []ast.BaseTerm{y}, u.ufrel)
	}

	// A variable binds to the other side wholesale, so that type
	// variables can also be bound to (union) type constructor
	// applications.
	if v, ok := x.(ast.Variable); ok && yOk {
		return u.bindVarTo(v, yApply)
	}
	if v, ok := y.(ast.Variable); ok && xOk {
		return u.bindVarTo(v, xApply)
	}

	// A wildcard bound on one side and a type constructor application
	// on the other are compatible.
	if isWildcardBound(x) || isWildcardBound(y) {
		return nil
	}

	// Union type expressions unify with their alternatives.
	if xOk && xApply.Function == symbols.UnionType || yOk && yApply.Function == symbols.UnionType {
		return u.unifyUnion(x, y)
	}

	if !xOk || !yOk {
		// One side is a type constructor application, the other a constant.
		return fmt.Errorf("cannot unify %v and %v", x, y)
	}
	if xApply.Function != yApply.Function {
		return fmt.Errorf("cannot unify %v and %v", xApply, yApply)
	}
	return u.unifyFunctional(xApply.Args, yApply.Args)
}

// bindVarTo binds a type variable to a type constructor application.
func (u *unionFindFun) bindVarTo(v ast.Variable, tpe ast.ApplyFn) error {
	if bound, ok := u.subst[v]; ok && !bound.Equals(tpe) {
		return fmt.Errorf("cannot unify %v and %v", tpe, bound)
	}
	if existing := u.ufrel.find(v); existing != nil && !existing.Equals(tpe) {
		return fmt.Errorf("cannot unify %v and %v", tpe, existing)
	}
	u.subst[v] = tpe
	return nil
}

// isWildcardBound returns true for the /any bound (the top of the type
// lattice, every type conforms to it) and the /bot bound (the bottom,
// to which no value belongs).
func isWildcardBound(tpe ast.BaseTerm) bool {
	return tpe.Equals(ast.AnyBound) || tpe.Equals(ast.BotBound)
}

// unifyUnion unifies an actual and a declared type expression, where at
// least one of them is a union type expression. An actual type unifies
// with a declared union if it unifies with one of its alternatives; an
// actual union unifies with a declared type if all of its alternatives
// do.
func (u *unionFindFun) unifyUnion(actual, declared ast.BaseTerm) error {
	actualApply, actualOk := actual.(ast.ApplyFn)
	declaredApply, declaredOk := declared.(ast.ApplyFn)
	actualIsUnion := actualOk && actualApply.Function == symbols.UnionType
	declaredIsUnion := declaredOk && declaredApply.Function == symbols.UnionType

	switch {
	case actualIsUnion && declaredIsUnion:
		for _, alt := range actualApply.Args {
			if err := u.unifyWithSomeAlternative(alt, declaredApply.Args); err != nil {
				return err
			}
		}
		return nil
	case actualIsUnion:
		for _, alt := range actualApply.Args {
			if err := u.unifyOne(alt, declared); err != nil {
				return err
			}
		}
		return nil
	default:
		return u.unifyWithSomeAlternative(actual, declaredApply.Args)
	}
}

// unifyWithSomeAlternative unifies x with one of the given alternatives,
// restoring the unification state if an alternative fails.
func (u *unionFindFun) unifyWithSomeAlternative(x ast.BaseTerm, alternatives []ast.BaseTerm) error {
	if len(alternatives) == 0 {
		return fmt.Errorf("cannot unify %v with empty union", x)
	}
	var err error
	for _, alt := range alternatives {
		snapshot := u.snapshot()
		if err = u.unifyOne(x, alt); err == nil {
			return nil
		}
		u.restore(snapshot)
	}
	return err
}

func (u *unionFindFun) snapshot() unionFindFun {
	subst := make(map[ast.Variable]ast.BaseTerm, len(u.subst))
	for k, v := range u.subst {
		subst[k] = v
	}
	return unionFindFun{ufrel: u.ufrel.Clone(), subst: subst}
}

func (u *unionFindFun) restore(snap unionFindFun) {
	u.ufrel = snap.ufrel
	u.subst = snap.subst
}
