package analysis

import "codeberg.org/TauCeti/mangle-go/ast"

// eliminateVariableEqualities rewrites clause so that every equality between
// two variables disappears: one of the two names is replaced by the other
// throughout the clause. An equality M = X gives neither variable a value on
// its own; after the substitution there is a single variable, so whatever
// binds it binds both, and a negation or inequality that waits for it keeps
// waiting. The renaming is an alpha-renaming, so it preserves the meaning of
// the clause for the bound and unbound sides alike.
func eliminateVariableEqualities(clause ast.Clause) ast.Clause {
	headVars := map[ast.Variable]bool{}
	ast.AddVars(clause.Head, headVars)

	// renames maps an eliminated variable to the name it is replaced by.
	renames := map[ast.Variable]ast.Variable{}
	dropped := map[int]bool{}
	for i, p := range clause.Premises {
		eq, ok := p.(ast.Eq)
		if !ok {
			continue
		}
		l, lok := eq.Left.(ast.Variable)
		r, rok := eq.Right.(ast.Variable)
		// Wildcards stand for distinct fresh variables and are never renamed.
		if !lok || !rok || l.Symbol == "_" || r.Symbol == "_" {
			continue
		}
		l, r = resolveVariable(l, renames), resolveVariable(r, renames)
		if l == r {
			dropped[i] = true // X = X is always true.
			continue
		}
		dropped[i] = true
		// Keep a name that occurs in the head, so head variables keep
		// their names where possible.
		if headVars[r] && !headVars[l] {
			renames[l] = r
		} else {
			renames[r] = l
		}
	}
	if len(renames) == 0 && len(dropped) == 0 {
		return clause
	}
	rename := func(v ast.Variable) ast.Variable { return resolveVariable(v, renames) }

	premises := make([]ast.Term, 0, len(clause.Premises))
	for i, p := range clause.Premises {
		if !dropped[i] {
			premises = append(premises, ast.RenameVariables(p, rename))
		}
	}
	return ast.Clause{
		Head:      ast.RenameVariables(clause.Head, rename).(ast.Atom),
		HeadTime:  ast.RenameVariablesInInterval(clause.HeadTime, rename),
		Premises:  premises,
		Transform: renameTransform(clause.Transform, rename),
	}
}

// resolveVariable follows renames to the surviving name.
func resolveVariable(v ast.Variable, renames map[ast.Variable]ast.Variable) ast.Variable {
	for {
		next, ok := renames[v]
		if !ok {
			return v
		}
		v = next
	}
}

func renameTransform(t *ast.Transform, rename func(ast.Variable) ast.Variable) *ast.Transform {
	if t == nil {
		return nil
	}
	statements := make([]ast.TransformStmt, len(t.Statements))
	for i, stmt := range t.Statements {
		if stmt.Var != nil {
			v := rename(*stmt.Var)
			statements[i].Var = &v
		}
		args := make([]ast.BaseTerm, len(stmt.Fn.Args))
		for j, arg := range stmt.Fn.Args {
			args[j] = ast.RenameVariablesBase(arg, rename)
		}
		statements[i].Fn = ast.ApplyFn{stmt.Fn.Function, args}
	}
	next := renameTransform(t.Next, rename)
	return &ast.Transform{Statements: statements, Next: next}
}
