package unionfind

import (
	"fmt"

	"codeberg.org/TauCeti/mangle-go/ast"
)

type unionFindFun struct {
	ufrel UnionFind
	subst map[ast.Variable]ast.BaseTerm
}

// UnifyTypeExpr unifies two type expressions. These are terms that contain
// ApplyFn nodes.
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
	for i, x := range xs {
		y := ys[i]
		xApply, xOk := x.(ast.ApplyFn)
		yApply, yOk := y.(ast.ApplyFn)
		if !xOk && !yOk {
			if x.Equals(ast.Variable{"_"}) || y.Equals(ast.Variable{"_"}) {
				continue
			}
			u.ufrel.parent[x] = x
			u.ufrel.parent[y] = y
			unifyTermsUpdate([]ast.BaseTerm{x}, []ast.BaseTerm{y}, u.ufrel)
			continue
		}
		if yOk && !xOk {
			xApply = yApply
			y = x
			xOk, yOk = true, false
		}
		if xOk && !yOk {
			yVar, yOk := y.(ast.Variable)
			if !yOk {
				return fmt.Errorf("cannot unify %v and %v", x, y)
			}
			if yExisting := u.ufrel.find(yVar); yExisting != nil && !yExisting.Equals(xApply) {
				return fmt.Errorf("cannot unify %v and %v", x, yExisting)
			}

			u.subst[yVar] = xApply
			continue
		}
		if xApply.Function != yApply.Function {
			return fmt.Errorf("cannot unify %v and %v", xApply, yApply)
		}
		if err := u.unifyFunctional(xApply.Args, yApply.Args); err != nil {
			return err
		}
	}
	return nil
}
