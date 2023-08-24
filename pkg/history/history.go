package history

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// History is a kind-specific client that knows how to access the revision history of objects of that kind.
// Instantiate a History with For or ForGroupKind.
type History interface {
	// ListRevisions returns a sorted revision history (ascending) of the object identified by the given key.
	ListRevisions(ctx context.Context, key client.ObjectKey) (Revisions, error)
}

// For instantiates a new History client for the given Object.
func For(c client.Client, obj client.Object) (History, error) {
	gvk, err := apiutil.GVKForObject(obj, c.Scheme())
	if err != nil {
		return nil, err
	}

	return ForGroupKind(c, gvk.GroupKind())
}

// ForGroupKind instantiates a new History client for the given GroupKind.
func ForGroupKind(c client.Client, gk schema.GroupKind) (History, error) {
	switch {
	case gk.Group == appsv1.GroupName && gk.Kind == "Deployment":
		return DeploymentHistory{Client: c}, nil
	}

	return nil, fmt.Errorf("%s is not supported", gk.String())
}

// Revisions implements runtime.Object for passing it around like a printable API object to printers.ResourcePrinter.
var _ runtime.Object = Revisions(nil)

// Revisions is a list of Revision objects.
type Revisions []Revision

// Revision represents a single revision in the history of a workload object. E.g., a ReplicaSet in a Deployment's
// history.
type Revision interface {
	// Object is embedded in Revision so that Revision objects can be passed around like a printable API object to
	// printers.ResourcePrinter.
	runtime.Object

	// Number returns the revision number identifying this Revision.
	Number() int64
	// Name returns the name of the underlying revision object.
	Name() string
	// Object returns the full revision object (e.g., *appsv1.ReplicaSet or *appsv1.ControllerRevision).
	Object() client.Object
	// PodTemplate returns the PodTemplate that was specified in this revision of the object.
	PodTemplate() *corev1.Pod
}

// GetObjectKind implements runtime.Object.
func (r Revisions) GetObjectKind() schema.ObjectKind {
	if len(r) == 0 {
		return &metav1.TypeMeta{}
	}

	return r[0].GetObjectKind()
}

// DeepCopyObject implements runtime.Object.
func (r Revisions) DeepCopyObject() runtime.Object {
	if r == nil {
		return nil
	}

	out := make(Revisions, len(r))
	for i, rev := range r {
		out[i] = rev.DeepCopyObject().(Revision)
	}

	return out
}

// ByNumber finds the Revision with the given revision number in a sorted revision list.
// -1 denotes the latest revision, -2 the previous one, etc.
func (r Revisions) ByNumber(number int64) (Revision, error) {
	if number == 0 {
		return nil, fmt.Errorf("invalid revision number %d", number)
	}

	// resolve negative revision number
	if number < 0 {
		i := len(r) + int(number)
		if i < 0 {
			return nil, fmt.Errorf("revision %d not found", number)
		}
		return r[i], nil
	}

	// find the revision by actual number (index and number don't have to relate strictly)
	for _, revision := range r {
		if revision.Number() == number {
			return revision, nil
		}
	}

	return nil, fmt.Errorf("revision %d not found", number)
}

// Predecessor finds the Revision in a sorted revision list that preceded the Revision identified by the given revision
// number. See also ByNumber.
func (r Revisions) Predecessor(number int64) (Revision, error) {
	// resolve revision number
	successor, err := r.ByNumber(number)
	if err != nil {
		return nil, err
	}

	// find index of successor by actual number (index and number don't have to relate strictly)
	var i int
	for i = range r {
		if r[i].Number() == successor.Number() {
			break
		}
	}

	if i < 1 {
		return nil, fmt.Errorf("predecessor of revision %d not found", number)
	}

	return r[i-1], nil
}
