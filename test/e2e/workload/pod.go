package workload

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

const (
	AppName         = "nginx"
	ImageRepository = "registry.k8s.io/nginx-slim"
)

var (
	generation = 0
)

func ImageTag() string {
	generation++
	if generation > 27 {
		panic("ImageTag called too many times")
	}

	DeferCleanup(func() {
		generation = 0
	})

	return fmt.Sprintf("0.%d", generation)
}

func Image() string {
	return ImageRepository + ":" + ImageTag()
}

func PodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{{
			Name:  AppName,
			Image: Image(),
		}},
	}
}

func BumpImage(obj client.Object) {
	GinkgoHelper()

	Expect(testClient.Patch(context.Background(), obj, client.RawPatch(types.JSONPatchType, []byte(`[{
"op": "replace",
"path": "/spec/template/spec/containers/0/image",
"value": "`+Image()+`"
}]`)))).To(Succeed())

	Eventually(komega.Object(obj)).Should(HaveField("Status.ObservedGeneration", obj.GetGeneration()))
}
