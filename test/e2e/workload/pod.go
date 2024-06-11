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
	AppName = "pause"
	// ImageRepository is the image repository holding the e2e test image.
	// The repository has a copy of registry.k8s.io/pause:3.10 for linux/amd64 and linux/arm64.
	// The copied image is tagged with 0.1 through 0.10.
	// It was set up with the following commands:
	//  for arch in amd64 arm64 ; do
	//    crane copy registry.k8s.io/pause:3.10 ghcr.io/timebertt/e2e-pause:$arch --platform linux/$arch
	//  done
	//  for i in $(seq 1 10) ; do
	//    crane index append -m ghcr.io/timebertt/e2e-pause:amd64 -m ghcr.io/timebertt/e2e-pause:arm64 -t ghcr.io/timebertt/e2e-pause:0.$i
	//  done
	// This image is used in e2e tests because it is small, fast to run, and has simple and consistent tags. But most
	// importantly, it makes these e2e tests independent of external registries, which might change or rate limit pulls.
	ImageRepository = "ghcr.io/timebertt/e2e-pause"
)

var (
	generation = 0
)

func ImageTag() string {
	generation++
	if generation > 10 {
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
