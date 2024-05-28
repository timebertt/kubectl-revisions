package diff_test

import (
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/timebertt/kubectl-revisions/pkg/diff"
)

var _ = Describe("Files", func() {
	It("should create a temp dir with two files", func() {
		f, err := NewFiles("test", "a", "b")
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			Expect(f.TearDown()).To(Succeed())
			Expect(f.Dir).NotTo(BeADirectory())
			Expect(f.A.Name()).NotTo(BeARegularFile())
			Expect(f.B.Name()).NotTo(BeARegularFile())
		})

		Expect(f.Dir).NotTo(BeEmpty())
		Expect(f.Dir).To(BeADirectory())

		testDiffFile := func(file *os.File, name string) {
			ExpectWithOffset(1, file).NotTo(BeNil())
			ExpectWithOffset(1, file.Name()).To(BeARegularFile())
			ExpectWithOffset(1, filepath.Dir(file.Name())).To(Equal(f.Dir))
			ExpectWithOffset(1, filepath.Base(file.Name())).To(Equal(name))

			Expect(io.WriteString(file, "test "+name)).NotTo(BeZero())
		}

		testDiffFile(f.A, "a")
		testDiffFile(f.B, "b")
	})
})
