package runutil

import (
	"github.com/hashicorp/go-multierror"
)

// CaptureError runs f and optionally adds the returned error to the given error. This is supposed to be used for
// handling errors in defer statements. The first argument should point to an error return parameter.
// Often, errors are discarded like this, which is error-prone as closing an open file might be necessary to flush the
// written contents:
//
//	var f *os.File
//	// ...
//	defer f.Close()
//
// Instead, use CaptureError to handle the returned error:
//
//	 func foo() (err error) {
//		   var f *os.File
//		   // ...
//		   defer runutil.CaptureError(&err, f.Close)
//		   // ...
//		 }
//
// If the error return parameter is non-nil and Close errors, CaptureError sets the error return parameter to a
// multierror capturing both errors. If only one error occurs, the return parameter is set to exactly that error.
func CaptureError(err *error, f func() error) {
	// collect an already existing return error and a new error
	errs := multierror.Append(*err, f())

	// if only one occurred, set err directly (don't wrap it in a multierror)
	if len(errs.Errors) == 1 {
		*err = errs.Errors[0]
		return
	}

	// set err to the combined error or nil
	*err = errs.ErrorOrNil()
}
