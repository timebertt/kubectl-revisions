package history_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/timebertt/kubectl-revisions/pkg/history"
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
			Replicas: Replicas{
				Current: 2,
				Ready:   1,
			},
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

	Describe("CurrentReplicas", func() {
		It("should return the value of the Replicas.Current field", func() {
			Expect(rev.CurrentReplicas()).To(Equal(rev.Current))
		})
	})

	Describe("ReadyReplicas", func() {
		It("should return the value of the Replicas.Ready field", func() {
			Expect(rev.ReadyReplicas()).To(Equal(rev.Ready))
		})
	})
})

var _ = Describe("ListControllerRevisionsAndPods", func() {
	var (
		fakeClient client.Client

		namespace string
		selector  *metav1.LabelSelector

		revision *appsv1.ControllerRevision
		pod      *corev1.Pod
	)

	BeforeEach(func() {
		namespace = "default"
		labels := map[string]string{"app": "test"}
		selector = &metav1.LabelSelector{MatchLabels: labels}

		revision = &appsv1.ControllerRevision{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: namespace,
				Labels:    labels,
			},
		}

		revisionOtherNamespace := revision.DeepCopy()
		revisionOtherNamespace.Name += "-other"
		revisionOtherNamespace.Namespace += "-other"

		revisionUnrelated := revision.DeepCopy()
		revisionUnrelated.Name += "-not-matching"
		revisionUnrelated.Labels["app"] = "other"

		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: namespace,
				Labels:    labels,
			},
		}

		podOtherNamespace := pod.DeepCopy()
		podOtherNamespace.Name += "-other"
		podOtherNamespace.Namespace += "-other"

		podUnrelated := pod.DeepCopy()
		podUnrelated.Name += "-not-matching"
		podUnrelated.Labels["app"] = "other"

		fakeClient = fakeclient.NewFakeClient(
			revision, revisionOtherNamespace, revisionUnrelated,
			pod, podOtherNamespace, podUnrelated,
		)
	})

	It("should return matching objects in the same namespace", func() {
		controllerRevisionList, podList, err := ListControllerRevisionsAndPods(context.Background(), fakeClient, namespace, selector)
		Expect(err).NotTo(HaveOccurred())

		Expect(controllerRevisionList.Items).To(ConsistOf(*revision))
		Expect(podList.Items).To(ConsistOf(*pod))
	})
})
