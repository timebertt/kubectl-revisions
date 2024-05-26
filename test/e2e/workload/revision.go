package workload

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RevisionObjectFor(obj client.Object) client.Object {
	switch obj.(type) {
	case *appsv1.Deployment:
		return &appsv1.ReplicaSet{}
	case *appsv1.DaemonSet, *appsv1.StatefulSet:
		return &appsv1.ControllerRevision{}
	default:
		panic(fmt.Errorf("unexpted object type %T", obj))
	}
}
