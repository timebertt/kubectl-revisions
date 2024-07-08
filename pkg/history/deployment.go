package history

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ History = DeploymentHistory{}

// DeploymentHistory implements the History interface for Deployments.
type DeploymentHistory struct {
	Client client.Reader
}

func (d DeploymentHistory) ListRevisions(ctx context.Context, obj client.Object) (Revisions, error) {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil, fmt.Errorf("expected *appsv1.Deployment, got %T", obj)
	}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("error parsing Deployment selector: %w", err)
	}

	replicaSetList := &appsv1.ReplicaSetList{}
	if err := d.Client.List(ctx, replicaSetList, client.InNamespace(deployment.Namespace), client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, fmt.Errorf("error listing ReplicaSets: %w", err)
	}

	var revs Revisions
	for _, replicaSet := range replicaSetList.Items {
		if !metav1.IsControlledBy(&replicaSet, deployment) {
			continue
		}

		revision, err := NewReplicaSet(&replicaSet)
		if err != nil {
			return nil, fmt.Errorf("error converting ReplicaSet %s: %w", replicaSet.Name, err)
		}

		revs = append(revs, revision)
	}

	Sort(revs)
	return revs, nil
}
