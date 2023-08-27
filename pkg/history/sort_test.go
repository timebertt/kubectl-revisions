package history_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/timebertt/kubectl-history/pkg/history"
	"github.com/timebertt/kubectl-history/pkg/history/fake"
)

var _ = Describe("Sort", func() {
	var (
		revs Revisions
	)

	BeforeEach(func() {
		revs = Revisions{
			&fake.Revision{Num: 2},
			&fake.Revision{Num: 3},
			&fake.Revision{Num: 1},
		}
	})

	It("should correctly sort the revisions list", func() {
		before := revs.DeepCopyObject().(Revisions)

		Sort(revs)

		Expect(revs).To(HaveExactElements(before[2], before[0], before[1]))
	})
})
