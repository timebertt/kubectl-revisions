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
)

var _ Revision = &ControllerRevision{}

// ControllerRevision is a Revision of a StatefulSet or DaemonSet.
type ControllerRevision struct {
	ControllerRevision *appsv1.ControllerRevision
	Template           *corev1.Pod

	Replicas
}

// GetObjectKind implements runtime.Object.
func (c *ControllerRevision) GetObjectKind() schema.ObjectKind {
	if c == nil {
		return &metav1.TypeMeta{}
	}
	return &metav1.TypeMeta{
		APIVersion: appsv1.SchemeGroupVersion.String(),
		Kind:       "ControllerRevision",
	}
}

// DeepCopyObject implements runtime.Object.
func (c *ControllerRevision) DeepCopyObject() runtime.Object {
	if c == nil {
		return nil
	}

	out := new(ControllerRevision)
	*out = *c
	out.ControllerRevision = c.ControllerRevision.DeepCopy()
	out.Template = c.Template.DeepCopy()
	return out
}

func (c *ControllerRevision) Number() int64 {
	return c.ControllerRevision.Revision
}

func (c *ControllerRevision) Name() string {
	return c.ControllerRevision.Name
}

func (c *ControllerRevision) Object() client.Object {
	return c.ControllerRevision
}

func (c *ControllerRevision) PodTemplate() *corev1.Pod {
	return c.Template
}

// ListControllerRevisionsAndPods is a helper for a ControllerRevision-based History implementation that needs to find
// all ControllerRevisions and Pods belonging to a given workload object.
func ListControllerRevisionsAndPods(ctx context.Context, r client.Reader, namespace string, selector *metav1.LabelSelector) (*appsv1.ControllerRevisionList, *corev1.PodList, error) {
	listOptions := &client.ListOptions{
		Namespace: namespace,
	}

	var err error
	listOptions.LabelSelector, err = metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing selector: %w", err)
	}

	controllerRevisionList := &appsv1.ControllerRevisionList{}
	if err := r.List(ctx, controllerRevisionList, listOptions); err != nil {
		return nil, nil, fmt.Errorf("error listing ControllerRevisions: %w", err)
	}

	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, listOptions); err != nil {
		return nil, nil, fmt.Errorf("error listing Pods: %w", err)
	}

	return controllerRevisionList, podList, nil
}
