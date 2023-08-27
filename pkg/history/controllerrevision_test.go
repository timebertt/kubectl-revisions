package history_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/timebertt/kubectl-history/pkg/history"
)

var _ = Describe("ControllerRevision", func() {
	var (
		controllerRevision *appsv1.ControllerRevision
		template           *corev1.Pod
		rev                *ControllerRevision
	)

	BeforeEach(func() {
		controllerRevision = &appsv1.ControllerRevision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sts-1",
				Namespace: "test",
			},
			Revision: 1,
		}

		template = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "sts",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "test",
					Image: "test",
				}},
			},
		}

		rev = &ControllerRevision{
			ControllerRevision: controllerRevision,
			Template:           template,
		}
	})

	Describe("GetObjectKind", func() {
		It("should return an empty ObjectKind if the ControllerRevision is nil", func() {
			rev = nil
			Expect(rev.GetObjectKind().GroupVersionKind().Empty()).To(BeTrue())
		})

		It("should return the correct copied ObjectKind", func() {
			objectKind := rev.GetObjectKind()
			Expect(objectKind.GroupVersionKind()).To(Equal(appsv1.SchemeGroupVersion.WithKind("ControllerRevision")))
			Expect(objectKind).NotTo(BeIdenticalTo(controllerRevision.GetObjectKind()))
		})
	})

	Describe("DeepCopyObject", func() {
		It("should return nil if the ControllerRevision is nil", func() {
			rev = nil
			Expect(rev.DeepCopyObject()).To(BeNil())
		})

		It("should return a copy of the ControllerRevision", func() {
			copied := rev.DeepCopyObject()
			Expect(copied).To(BeAssignableToTypeOf(rev))
			Expect(copied).To(Equal(rev))
			Expect(copied).NotTo(BeIdenticalTo(rev))

			copiedRev := copied.(*ControllerRevision)
			Expect(copiedRev.ControllerRevision).NotTo(BeIdenticalTo(rev.ControllerRevision))
			Expect(copiedRev.Template).NotTo(BeIdenticalTo(rev.Template))
		})
	})

	Describe("Number", func() {
		It("should return the number", func() {
			Expect(rev.Number()).To(Equal(rev.ControllerRevision.Revision))
		})
	})

	Describe("Name", func() {
		It("should return the name", func() {
			Expect(rev.Name()).To(Equal(rev.ControllerRevision.Name))
		})
	})

	Describe("Object", func() {
		It("should return the ControllerRevision object", func() {
			Expect(rev.Object()).To(Equal(rev.ControllerRevision))
		})
	})

	Describe("PodTemplate", func() {
		It("should return the template", func() {
			Expect(rev.PodTemplate()).To(Equal(rev.Template))
		})
	})
})
