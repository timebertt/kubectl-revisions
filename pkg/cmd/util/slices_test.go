package util_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/timebertt/kubectl-revisions/pkg/cmd/util"
)

var _ = Describe("Map", func() {
	It("should correctly map all elements", func() {
		Expect(Map([]string{"1", "2", "3"}, func(e string) string {
			return e + "m"
		})).To(Equal([]string{"1m", "2m", "3m"}))
	})
})
