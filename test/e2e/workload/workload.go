package workload

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

func Scale(obj client.Object, replicas int32) {
	GinkgoHelper()

	Expect(testClient.Patch(
		context.Background(), obj, client.RawPatch(types.MergePatchType, []byte(fmt.Sprintf(`{"spec":{"replicas": %d}}`, replicas))),
	)).To(Succeed())

	Eventually(komega.Object(obj)).Should(HaveField("Status.ObservedGeneration", obj.GetGeneration()))
}
