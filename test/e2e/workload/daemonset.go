package workload

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

func CreateDaemonSet(namespace string) *appsv1.DaemonSet {
	GinkgoHelper()

	labels := map[string]string{"app": AppName}
	statefulSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AppName,
			Namespace: namespace,
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
