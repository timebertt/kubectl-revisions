package history

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Revision = &ReplicaSet{}

// ReplicaSet is a Revision of a Deployment.
type ReplicaSet struct {
	number int64

	ReplicaSet *appsv1.ReplicaSet
}

// NewReplicaSet transforms the given ReplicaSet to a Revision object.
func NewReplicaSet(replicaSet *appsv1.ReplicaSet) (*ReplicaSet, error) {
	replicaSet = replicaSet.DeepCopy()

	revision := &ReplicaSet{}
	revision.ReplicaSet = replicaSet

	var err error
	revision.number, err = deploymentutil.Revision(replicaSet)
	if err != nil {
		return nil, fmt.Errorf("error parsing revision: %w", err)
	}

	return revision, nil
}

// GetObjectKind implements runtime.Object.
func (r *ReplicaSet) GetObjectKind() schema.ObjectKind {
	if r == nil {
		return &metav1.TypeMeta{}
	}
	return &metav1.TypeMeta{
		APIVersion: appsv1.SchemeGroupVersion.String(),
		Kind:       "ReplicaSet",
	}
}

// DeepCopyObject implements runtime.Object.
func (r *ReplicaSet) DeepCopyObject() runtime.Object {
	if r == nil {
		return nil
	}

	out := new(ReplicaSet)
	*out = *r
	out.ReplicaSet = r.ReplicaSet.DeepCopy()
	return out
}

func (r *ReplicaSet) Number() int64 {
	return r.number
}

func (r *ReplicaSet) Name() string {
	return r.ReplicaSet.Name
}

func (r *ReplicaSet) Object() client.Object {
	return r.ReplicaSet
}

func (r *ReplicaSet) PodTemplate() *corev1.Pod {
	t := r.ReplicaSet.Spec.Template.DeepCopy()
	delete(t.Labels, appsv1.DefaultDeploymentUniqueLabelKey)
	return &corev1.Pod{
		ObjectMeta: t.ObjectMeta,
		Spec:       t.Spec,
	}
}

func (r *ReplicaSet) CurrentReplicas() int32 {
	return r.ReplicaSet.Status.Replicas
}

func (r *ReplicaSet) ReadyReplicas() int32 {
	return r.ReplicaSet.Status.ReadyReplicas
}
