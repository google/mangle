package analysis

import (
	"testing"

	"github.com/google/mangle/ast"
)

func TestVarList(t *testing.T) {
	vs := VarList{[]ast.Variable{ast.Variable{"X"}, ast.Variable{"Y"}}}

	if i := vs.Find(ast.Variable{"X"}); i != 0 {
		t.Errorf("Find(X)=%d want 0", i)
	}
	if !vs.Contains(ast.Variable{"Y"}) {
		t.Errorf("Contains(Y)=false want true")
	}
	if i := vs.Find(ast.Variable{"Z"}); i != -1 {
		t.Errorf("Find(X)=%d want -1", i)
	}
	if vs.Contains(ast.Variable{"Z"}) {
		t.Errorf("Contains(Z)=true want false")
	}

	vs = vs.Extend([]ast.Variable{ast.Variable{"Z"}})
	if !vs.Contains(ast.Variable{"Z"}) {
		t.Errorf("Contains(Z)=false want true")
	}
	if i := vs.Find(ast.Variable{"Z"}); i != 2 {
		t.Errorf("Find(Z)=%d want 2", i)
	}
}
