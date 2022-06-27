package support

import (
	goctx "context"
)

// FallibleProcedure is any function that, within a context, runs a procedure
// that may fail.
//
// Although, parameter passing to and from the procedure is only possible
// through the context the procedures are intended to be self-contained and not
// meant to act as functions.
//
// The returning error indicates the procedure failure.
type FallibleProcedure = func(goctx.Context) error

// ChainedOperation is a series of procedures that must execute without failure
// in a pre-determined order for the operation to be successful.
//
// If any procedure fails the whole operation will fail.
type ChainedOperation []FallibleProcedure

// Run executes the chained operation with the given context. If any procedure
// fails it stops the execution and returns the error.
func (o ChainedOperation) Run(ctx goctx.Context) error {
	for i := 0; i < len(o); i++ {
		if err := o[i](ctx); err != nil {
			return err
		}
	}

	return nil
}
