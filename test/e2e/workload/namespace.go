package workload

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/timebertt/kubectl-history/pkg/test/matcher"
)

func PrepareTestNamespace() string {
	GinkgoHelper()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-",
			Labels: map[string]string{
				"e2e-test": "kubectl-history",
			},
		},
	}

	Expect(testClient.Create(context.Background(), namespace)).To(Succeed())
	logf.Log.Info("Created test namespace", "namespace", namespace.Name)

	DeferCleanup(func() {
		logf.Log.Info("Deleting test namespace", "namespace", namespace.Name)
		Expect(testClient.Delete(context.Background(), namespace)).To(Or(Succeed(), BeNotFoundError()))
	})

	return namespace.Name
}
