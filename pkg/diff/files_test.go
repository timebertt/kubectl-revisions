package diff_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/printers"

	. "github.com/timebertt/kubectl-revisions/pkg/diff"
)

var _ = Describe("diff files", func() {
	var (
		f *Files
	)

	BeforeEach(func() {
		var err error
		f, err = NewFiles("a", "b")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(f.TearDown()).To(Succeed())
	})

	It("should create and tear down two temp dirs", func() {
		Expect(f.From.Dir).To(BeADirectory())
		Expect(f.To.Dir).To(BeADirectory())

		Expect(f.TearDown()).To(Succeed())
		Expect(f.From.Dir).NotTo(BeADirectory())
		Expect(f.To.Dir).NotTo(BeADirectory())
	})

	It("should print the given object", func() {
		name := "foo"
		fileName := filepath.Join(f.From.Dir, name)

		obj := &corev1.Namespace{}
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Namespace"))

		Expect(f.From.Print(name, obj, &printers.YAMLPrinter{})).To(Succeed())
		Expect(fileName).To(BeARegularFile())

		// nolint:gosec // this is test code
		bytes, err := os.ReadFile(fileName)
		Expect(err).NotTo(HaveOccurred())
		Expect(BufferWithBytes(bytes)).To(Say("kind: Namespace"))
	})

	Describe("TearDown", func() {
		It("should handle partial creation error", func() {
			Expect(f.From.Dir).To(BeADirectory())
			Expect(f.To.Dir).To(BeADirectory())

			// simulate partial creation error: we failed to create the second version
			Expect(os.RemoveAll(f.To.Dir)).To(Succeed())
			Expect(f.To.Dir).NotTo(BeADirectory())
			f.To = nil

			// should delete all directories
			Expect(f.TearDown()).To(Succeed())
			Expect(f.From.Dir).NotTo(BeADirectory())
		})
	})
})
