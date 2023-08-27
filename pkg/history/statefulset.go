package history

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ History = StatefulSetHistory{}

// StatefulSetHistory implements the History interface for StatefulSets.
type StatefulSetHistory struct {
	Client client.Client
}

func (d StatefulSetHistory) ListRevisions(ctx context.Context, key client.ObjectKey) (Revisions, error) {
	statefulSet := &appsv1.StatefulSet{}
	if err := d.Client.Get(ctx, key, statefulSet); err != nil {
		return nil, err
	}

	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("error parsing StatefulSet selector: %w", err)
	}

	controllerRevisionList := &appsv1.ControllerRevisionList{}
	if err := d.Client.List(ctx, controllerRevisionList, client.InNamespace(statefulSet.Namespace), client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, fmt.Errorf("error listing ControllerRevisions: %w", err)
	}

	var revs Revisions
	for _, controllerRevision := range controllerRevisionList.Items {
		if !metav1.IsControlledBy(&controllerRevision, statefulSet) {
			continue
		}

		revision, err := NewControllerRevisionForStatefulSet(&controllerRevision)
		if err != nil {
			return nil, fmt.Errorf("error converting ControllerRevision %s: %w", controllerRevision.Name, err)
		}

		revs = append(revs, revision)
	}

	Sort(revs)
	return revs, nil
}

// NewControllerRevisionForStatefulSet transforms the given ControllerRevision of a StatefulSet to a Revision object.
func NewControllerRevisionForStatefulSet(controllerRevision *appsv1.ControllerRevision) (*ControllerRevision, error) {
	controllerRevision = controllerRevision.DeepCopy()

	revision := &ControllerRevision{}
	revision.ControllerRevision = controllerRevision

	statefulSet := &appsv1.StatefulSet{}
	if statefulSetData, ok := revision.ControllerRevision.Data.Object.(*appsv1.StatefulSet); ok && statefulSetData != nil {
		statefulSet = statefulSetData
	} else {
		if err := runtime.DecodeInto(serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder(), controllerRevision.Data.Raw, statefulSet); err != nil {
			return nil, err
		}
		revision.ControllerRevision.Data.Object = statefulSet
		revision.ControllerRevision.Data.Raw = nil
	}

	t := statefulSet.Spec.Template.DeepCopy()
	revision.Template = &corev1.Pod{
		ObjectMeta: t.ObjectMeta,
		Spec:       t.Spec,
	}

	return revision, nil
}
