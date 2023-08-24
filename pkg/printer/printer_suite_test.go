package printer_test

import (
	"fmt"
	"io"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
	"k8s.io/utils/pointer"
)

func TestPrinter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Printer Suite")
}

type fakePrinter struct {
	printed runtime.Object
}

func (f *fakePrinter) PrintObj(obj runtime.Object, _ io.Writer) error {
	if f.printed != nil {
		Fail("fakePrinter.PrintObj called more than once")
	}
	f.printed = obj
	return nil
}

func replicaSet(revision int64) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("foo-%d", revision),
			Annotations: map[string]string{
				deploymentutil.RevisionAnnotation: strconv.FormatInt(revision, 10),
			},
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: pointer.Int32(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "test",
						Image: fmt.Sprintf("test:%d", revision),
					}},
				},
			},
		},
	}
}
