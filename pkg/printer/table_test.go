package printer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/timebertt/kubectl-revisions/pkg/history"
	. "github.com/timebertt/kubectl-revisions/pkg/printer"
)

var _ = Describe("RevisionsToTablePrinter", func() {
	var (
		p        *RevisionsToTablePrinter
		delegate *fakePrinter
	)

	BeforeEach(func() {
		delegate = &fakePrinter{}
		p = &RevisionsToTablePrinter{
			Delegate: delegate,

			Columns: []TableColumn{{
				TableColumnDefinition: metav1.TableColumnDefinition{
					Name: "Name",
				},
				Extract: func(rev history.Revision) any {
					return rev.Name()
				},
			}},
		}
	})

	It("should delegate unhandled objects", func() {
		obj := &corev1.ConfigMap{}
		Expect(p.PrintObj(obj, nil)).To(Succeed())

		Expect(delegate.printed).To(Equal(obj))
	})

	It("should correctly transform a single revision", func() {
		rev, err := history.NewReplicaSet(replicaSet(1))
		Expect(err).NotTo(HaveOccurred())

		Expect(p.PrintObj(rev, nil)).To(Succeed())

		Expect(delegate.printed).To(BeAssignableToTypeOf(&metav1.Table{}))
		table := delegate.printed.(*metav1.Table)

		Expect(table.ColumnDefinitions).To(HaveExactElements(p.Columns[0].TableColumnDefinition))

		Expect(table.Rows).To(HaveExactElements(
			metav1.TableRow{
				Cells:  []any{rev.Name()},
				Object: runtime.RawExtension{Object: rev.Object()},
			},
		))
	})

	It("should correctly transform multiple revisions", func() {
		rev1, err := history.NewReplicaSet(replicaSet(1))
		Expect(err).NotTo(HaveOccurred())
		rev2, err := history.NewReplicaSet(replicaSet(2))
		Expect(err).NotTo(HaveOccurred())

		Expect(p.PrintObj(history.Revisions{rev1, rev2}, nil)).To(Succeed())

		Expect(delegate.printed).To(BeAssignableToTypeOf(&metav1.Table{}))
		table := delegate.printed.(*metav1.Table)

		Expect(table.ColumnDefinitions).To(HaveExactElements(p.Columns[0].TableColumnDefinition))

		Expect(table.Rows).To(HaveExactElements(
			metav1.TableRow{
				Cells:  []any{rev1.Name()},
				Object: runtime.RawExtension{Object: rev1.Object()},
			},
			metav1.TableRow{
				Cells:  []any{rev2.Name()},
				Object: runtime.RawExtension{Object: rev2.Object()},
			},
		))
	})
})
