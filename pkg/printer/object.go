package printer

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/timebertt/kubectl-history/pkg/history"
)

var _ Printer = ObjectPrinter{}

// ObjectPrinter prints revisions' objects or pod templates using the given Encoder.
type ObjectPrinter struct {
	Encoder           Encoder
	TemplateOnly      bool
	ShowManagedFields bool
}

func (p ObjectPrinter) Print(rev history.Revision, w io.Writer) error {
	var printable runtime.Object = rev.PodTemplate()

	if !p.TemplateOnly {
		obj := rev.Object().DeepCopyObject().(client.Object)

		if !p.ShowManagedFields {
			// remove noisy managedFields if requested
			obj.SetManagedFields(nil)
		}

		// add back apiVersion and kind (removed by client's decoder)
		gvk, err := apiutil.GVKForObject(obj, clientgoscheme.Scheme)
		if err != nil {
			return err
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)

		printable = obj
	}

	return p.Encoder.Encode(printable, w)
}
