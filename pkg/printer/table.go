package printer

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/timebertt/kubectl-history/pkg/history"
)

var _ Printer = TablePrinter{}

// TablePrinter prints revisions in a tab-delimited table.
// Call PrintHeaders first to add table headers.
type TablePrinter struct {
	// Columns is the list of columns that should be printed.
	Columns []TableColumn
}

// TableColumn represents a single column with a header and logic for extracting a revision's cell value.
type TableColumn struct {
	Header  string
	Extract func(rev history.Revision) string
}

// DefaultTableColumns is the default list of columns to print.
var DefaultTableColumns = []TableColumn{
	{Header: "Name", Extract: func(rev history.Revision) string { return rev.Name() }},
	{Header: "Revision", Extract: func(rev history.Revision) string { return strconv.FormatInt(rev.Number(), 10) }},
	{Header: "Age", Extract: func(rev history.Revision) string { return humanDurationSince(rev.Object().GetCreationTimestamp()) }},
}

func (t TablePrinter) PrintHeaders(w io.Writer) error {
	var headers []string
	for _, column := range t.Columns {
		headers = append(headers, strings.ToUpper(column.Header))
	}

	_, err := fmt.Fprintln(w, strings.Join(headers, "\t"))
	return err
}

func (t TablePrinter) Print(rev history.Revision, w io.Writer) error {
	var cells []string
	for _, column := range t.Columns {
		cells = append(cells, column.Extract(rev))
	}

	_, err := fmt.Fprintln(w, strings.Join(cells, "\t"))
	return err
}

func humanDurationSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
