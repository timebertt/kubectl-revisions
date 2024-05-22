package runutil_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/timebertt/kubectl-history/pkg/runutil"
)

var _ = Describe("CaptureError", func() {
	var (
		result error
		f      func() error

		succeeds, fails func() error
		originalError   error
		functionError   error
	)

	BeforeEach(func() {
		result = nil

		originalError = fmt.Errorf("original")
		functionError = fmt.Errorf("function")

		succeeds = func() error { return nil }
		fails = func() error { return functionError }

		f = succeeds
	})

	Context("no original error set", func() {
		Context("f succeeds", func() {
			It("result should be nil", func() {
				CaptureError(&result, f)

				Expect(result).NotTo(HaveOccurred())
			})
		})

		Context("f fails", func() {
			BeforeEach(func() {
				f = fails
			})

			It("result should be error from f", func() {
				CaptureError(&result, f)

				Expect(result).To(Equal(functionError))
			})
		})
	})

	Context("original error is set", func() {
		BeforeEach(func() {
			result = originalError
		})

		Context("f succeeds", func() {
			It("result should be original error", func() {
				CaptureError(&result, f)

				Expect(result).To(Equal(originalError))
			})
		})

		Context("f fails", func() {
			BeforeEach(func() {
				f = fails
			})

			It("result should be merged error", func() {
				CaptureError(&result, f)

				Expect(result).To(MatchError(And(
					ContainSubstring(originalError.Error()),
					ContainSubstring(functionError.Error()),
				)))
			})
		})
	})
})
