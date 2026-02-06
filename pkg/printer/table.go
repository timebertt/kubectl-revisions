package printer

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/timebertt/kubectl-revisions/pkg/history"
)

var _ printers.ResourcePrinter = RevisionsToTablePrinter{}

// RevisionsToTablePrinter transforms revision objects to a metav1.Table and passes it on to the delegate (table)
// printer.
type RevisionsToTablePrinter struct {
	Delegate printers.ResourcePrinter

	// Columns is the list of columns that should be printed.
	Columns []TableColumn
}

// TableColumn represents a single column with a header and logic for extracting a revision's cell value.
type TableColumn struct {
	metav1.TableColumnDefinition
	Extract func(rev history.Revision) any
}

// DefaultTableColumns is the list of default column definitions.
var DefaultTableColumns = []TableColumn{
	{
		TableColumnDefinition: metav1.TableColumnDefinition{
			Name:   "Name",
			Type:   "string",
			Format: "name",
		},
		Extract: func(rev history.Revision) any { return rev.Name() },
	},
	{
		TableColumnDefinition: metav1.TableColumnDefinition{
			Name: "Revision",
			Type: "integer",
		},
		Extract: func(rev history.Revision) any { return rev.Number() },
	},
	{
		TableColumnDefinition: metav1.TableColumnDefinition{
			Name: "Ready",
			Type: "string",
		},
		Extract: func(rev history.Revision) any {
			return fmt.Sprintf("%d/%d", rev.ReadyReplicas(), rev.CurrentReplicas())
		},
	},
	{
		TableColumnDefinition: metav1.TableColumnDefinition{
			Name: "Age",
			Type: "string",
		},
		Extract: func(rev history.Revision) any {
			return table.ConvertToHumanReadableDateType(rev.Object().GetCreationTimestamp())
		},
	},
	{
		TableColumnDefinition: metav1.TableColumnDefinition{
			Name:     "Containers",
			Type:     "string",
			Priority: 1,
		},
		Extract: func(rev history.Revision) any {
			var (
				containers = rev.PodTemplate().Spec.Containers
				names      = make([]string, 0, len(containers))
			)
			for _, container := range containers {
				names = append(names, container.Name)
			}
			return strings.Join(names, ",")
		},
	},
	{
		TableColumnDefinition: metav1.TableColumnDefinition{
			Name:     "Images",
			Type:     "string",
			Priority: 1,
		},
		Extract: func(rev history.Revision) any {
			var (
				containers = rev.PodTemplate().Spec.Containers
				images     = make([]string, 0, len(containers))
			)
			for _, container := range containers {
				images = append(images, container.Image)
			}
			return strings.Join(images, ",")
		},
	},
}

func (p RevisionsToTablePrinter) PrintObj(obj runtime.Object, w io.Writer) error {
	var revs history.Revisions
	switch r := obj.(type) {
	case history.Revision:
		revs = history.Revisions{r}
	case history.Revisions:
		revs = r
	default:
		return p.Delegate.PrintObj(obj, w)
	}

	t := &metav1.Table{}

	// build column definitions
	for _, column := range p.Columns {
		t.ColumnDefinitions = append(t.ColumnDefinitions, *column.DeepCopy())
	}

	// build rows
	for _, rev := range revs {
		var cells []any

		for _, column := range p.Columns {
			cells = append(cells, column.Extract(rev))
		}

		t.Rows = append(t.Rows, metav1.TableRow{
			Cells:  cells,
			Object: runtime.RawExtension{Object: rev.Object()},
		})
	}

	return p.Delegate.PrintObj(t, w)
}
