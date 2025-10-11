package z3

// #include "go-z3.h"
import "C"
import "unsafe"

type Optimize struct {
	rawOptimize C.Z3_optimize
	rawCtx      C.Z3_context
}

func (c *Context) NewOptimizer() *Optimize {
	rawSolver := C.Z3_mk_optimize(c.Z3Value())
	C.Z3_optimize_inc_ref(c.raw, rawSolver)
	return &Optimize{
		rawOptimize: rawSolver,
		rawCtx:      c.Z3Value(),
	}
}

func (o *Optimize) Maximize(a *AST) {
	C.Z3_optimize_maximize(o.rawCtx, o.rawOptimize, a.rawAST)
}

func (o *Optimize) Close() error {
	C.Z3_optimize_dec_ref(o.rawCtx, o.rawOptimize)
	return nil
}

// Assert asserts a constraint onto the Optimizer.
//
// Maps to: Z3_optimize_assert
func (o *Optimize) Assert(a *AST) {
	C.Z3_optimize_assert(o.rawCtx, o.rawOptimize, a.rawAST)
}

// Check checks if the currently set formula is consistent.
//
// Maps to: Z3_optimize_check
func (o *Optimize) Check() LBool {
	raws := make([]C.Z3_ast, 0)
	return LBool(C.Z3_optimize_check(o.rawCtx, o.rawOptimize, 0, (*C.Z3_ast)(unsafe.Pointer(&raws))))
}

// Model returns the last model from a Check.
//
// Maps to: Z3_optimize_get_model
func (o *Optimize) Model() *Model {
	m := &Model{
		rawCtx:   o.rawCtx,
		rawModel: C.Z3_optimize_get_model(o.rawCtx, o.rawOptimize),
	}
	m.IncRef()
	return m
}
