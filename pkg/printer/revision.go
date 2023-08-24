package printer

import (
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/timebertt/kubectl-history/pkg/history"
)

var _ printers.ResourcePrinter = RevisionPrinter{}

// RevisionPrinter prints revisions' objects or pod templates using the given delegate printer.
type RevisionPrinter struct {
	Delegate     printers.ResourcePrinter
	TemplateOnly bool
}

// PrintObj prints a revision or list of revisions to the given writer using the printer's delegate.
func (p RevisionPrinter) PrintObj(obj runtime.Object, w io.Writer) error {
	switch r := obj.(type) {
	case history.Revision:
		return p.Delegate.PrintObj(Printable(r, p.TemplateOnly), w)
	case history.Revisions:
		// collect all revision objects in an unstructured list, which is properly handled by all used printers
		list := &unstructured.UnstructuredList{
			Object: map[string]interface{}{
				"kind":       "List",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"resourceVersion": "",
				},
			},
		}

		for _, rev := range r {
			var object client.Object = rev.PodTemplate()
			if !p.TemplateOnly {
				object = rev.Object()
				gvk, err := apiutil.GVKForObject(object, scheme.Scheme)
				if err != nil {
					return err
				}
				object.GetObjectKind().SetGroupVersionKind(gvk)
			}

			unstructuredContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
			if err != nil {
				return err
			}
			list.Items = append(list.Items, unstructured.Unstructured{Object: unstructuredContent})
		}

		return p.Delegate.PrintObj(list, w)
	}

	return p.Delegate.PrintObj(obj, w)
}

// Printable returns the actually printable object of a Revision based on the --template-only flag.
func Printable(rev history.Revision, templateOnly bool) client.Object {
	if templateOnly {
		return rev.PodTemplate()
	}
	return rev.Object()
}
