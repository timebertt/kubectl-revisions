package workload

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	"github.com/timebertt/kubectl-revisions/pkg/maps"
)

func CreateDaemonSet(namespace, name string) client.Object {
	GinkgoHelper()

	labels := maps.Merge(CommonLabels(), map[string]string{"app": name})
	statefulSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: PodSpec(),
			},
		},
	}

	Expect(testClient.Create(context.Background(), statefulSet)).To(Succeed())

	Eventually(komega.Object(statefulSet)).Should(HaveField("Status.ObservedGeneration", statefulSet.GetGeneration()))

	return statefulSet
}
