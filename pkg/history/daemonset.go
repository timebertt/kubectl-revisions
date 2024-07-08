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

var _ History = DaemonSetHistory{}

// DaemonSetHistory implements the History interface for DaemonSets.
type DaemonSetHistory struct {
	Client client.Reader
}

func (d DaemonSetHistory) ListRevisions(ctx context.Context, obj client.Object) (Revisions, error) {
	daemonSet, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return nil, fmt.Errorf("expected *appsv1.DaemonSet, got %T", obj)
	}

	controllerRevisionList, podList, err := ListControllerRevisionsAndPods(ctx, d.Client, daemonSet.Namespace, daemonSet.Spec.Selector)
	if err != nil {
		return nil, err
	}

	var revs Revisions
	for _, controllerRevision := range controllerRevisionList.Items {
		if !metav1.IsControlledBy(&controllerRevision, daemonSet) {
			continue
		}

		revision, err := NewControllerRevisionForDaemonSet(&controllerRevision)
		if err != nil {
			return nil, fmt.Errorf("error converting ControllerRevision %s: %w", controllerRevision.Name, err)
		}

		revision.Replicas = CountReplicas(podList, PodBelongsToDaemonSetRevision(&controllerRevision))

		revs = append(revs, revision)
	}

	Sort(revs)
	return revs, nil
}

// NewControllerRevisionForDaemonSet transforms the given ControllerRevision of a DaemonSet to a Revision object.
func NewControllerRevisionForDaemonSet(controllerRevision *appsv1.ControllerRevision) (*ControllerRevision, error) {
	controllerRevision = controllerRevision.DeepCopy()

	revision := &ControllerRevision{}
	revision.ControllerRevision = controllerRevision

	daemonSet := &appsv1.DaemonSet{}
	if daemonSetData, ok := revision.ControllerRevision.Data.Object.(*appsv1.DaemonSet); ok && daemonSetData != nil {
		daemonSet = daemonSetData
	} else {
		if err := runtime.DecodeInto(serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder(), controllerRevision.Data.Raw, daemonSet); err != nil {
			return nil, err
		}
		revision.ControllerRevision.Data.Object = daemonSet
		revision.ControllerRevision.Data.Raw = nil
	}

	t := daemonSet.Spec.Template.DeepCopy()
	revision.Template = &corev1.Pod{
		ObjectMeta: t.ObjectMeta,
		Spec:       t.Spec,
	}

	return revision, nil
}

func PodBelongsToDaemonSetRevision(revision *appsv1.ControllerRevision) PodPredicate {
	return func(pod *corev1.Pod) bool {
		return pod.Labels[appsv1.DefaultDaemonSetUniqueLabelKey] == revision.Labels[appsv1.DefaultDaemonSetUniqueLabelKey]
	}
}
