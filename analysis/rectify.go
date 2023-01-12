package analysis

import "github.com/google/mangle/ast"

// RectifyAtom ensures all arguments of an atom are variables.
// It returns a tuple (rectified atom, extra terms, fresh variables, defined variables).
//
// An atom p(t_1, ..., t_n) is rectified if each t_i is a variable and all are distinct.
//
// We go a bit further and also ensure that t_i are distinct from a set variables that
// may have been defined previously ("used"). On the other hand, we do not generate fresh
// variables for wildcards "_".
// If the given atom is already rectified in
// this stronger sense, than extra terms and fresh variables will be empty and boundVars
// contains atom arguments in left-to-right order.
func RectifyAtom(atom ast.Atom, usedVars VarList) (ast.Atom, []ast.Term, []ast.Variable, []ast.Variable) {
	used := usedVars.AsMap()
	pred := atom.Predicate
	newArgs := make([]ast.BaseTerm, pred.Arity)

	var freshVars, boundVars []ast.Variable
	var fml []ast.Term
	makeFresh := func(i int, arg ast.BaseTerm) {
		fresh := ast.FreshVariable(used)
		used[fresh] = true
		freshVars = append(freshVars, fresh)
		newArgs[i] = fresh
		fml = append(fml, ast.Eq{fresh, arg})
	}
	for i, arg := range atom.Args {
		switch a := arg.(type) {
		case ast.ApplyFn:
			makeFresh(i, arg)

		case ast.Constant:
			makeFresh(i, arg)

		case ast.Variable:
			if a.Symbol == "_" {
				newArgs[i] = a
				continue
			}
			if _, ok := used[a]; ok {
				makeFresh(i, arg)
				continue
			}
			used[a] = true
			newArgs[i] = a
			boundVars = append(boundVars, a)
		}
	}
	return ast.Atom{pred, newArgs}, fml, freshVars, boundVars
}
