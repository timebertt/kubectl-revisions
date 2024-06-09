package fake

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/timebertt/kubectl-revisions/pkg/history"
)

var _ history.Revision = &Revision{}

// Revision is a fake implementation of the history.Revision interface.
type Revision struct {
	Num int64

	Obj      client.Object
	Template *corev1.Pod

	history.Replicas
}

func (r *Revision) GetObjectKind() schema.ObjectKind {
	return r.Obj.GetObjectKind()
}

func (r *Revision) DeepCopyObject() runtime.Object {
	if r == nil {
		return nil
	}

	out := new(Revision)
	*out = *r
	if r.Obj != nil {
		out.Obj = r.Obj.DeepCopyObject().(client.Object)
	}
	return out
}

func (r *Revision) Number() int64 {
	return r.Num
}

func (r *Revision) Name() string {
	return r.Obj.GetName()
}

func (r *Revision) Object() client.Object {
	return r.Obj
}

func (r *Revision) PodTemplate() *corev1.Pod {
	return r.Template
}
