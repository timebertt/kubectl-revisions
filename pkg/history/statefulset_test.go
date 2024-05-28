package history_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	. "github.com/timebertt/kubectl-revisions/pkg/history"
)

var _ = Describe("StatefulSetHistory", func() {
	var (
		ctx        context.Context
		fakeClient client.Client

		statefulSet *appsv1.StatefulSet

		controllerRevision1, controllerRevision3, controllerRevisionUnrelated *appsv1.ControllerRevision
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeClient = fakeclient.NewClientBuilder().Build()

		statefulSet = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sts",
				Namespace: "test",
				Labels: map[string]string{
					"app": "sts",
				},
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "sts",
					},
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: "test",
						}},
					},
				},
			},
		}

		Expect(fakeClient.Create(ctx, statefulSet)).To(Succeed())

		controllerRevision1 = controllerRevisionForStatefulSet(statefulSet, 1, fakeClient.Scheme())
		controllerRevision1.Name = "app-b"
		Expect(fakeClient.Create(ctx, controllerRevision1)).To(Succeed())

		// create a non-sorted list of ControllerRevisions to verify that ListRevisions returns a sorted list
		controllerRevision3 = controllerRevisionForStatefulSet(statefulSet, 3, fakeClient.Scheme())
		controllerRevision3.Name = "app-a"
		Expect(fakeClient.Create(ctx, controllerRevision3)).To(Succeed())

		controllerRevisionUnrelated = controllerRevisionForStatefulSet(statefulSet, 0, fakeClient.Scheme())
		controllerRevisionUnrelated.OwnerReferences[0].UID = "other"
		Expect(fakeClient.Create(ctx, controllerRevisionUnrelated)).To(Succeed())
	})

	Describe("initialization", func() {
		It("should be constructable via For", func() {
			history, err := For(fakeClient, &appsv1.StatefulSet{})
			Expect(err).NotTo(HaveOccurred())
			Expect(history).NotTo(BeNil())
			Expect(history).To(BeAssignableToTypeOf(StatefulSetHistory{}))

			h := history.(StatefulSetHistory)
			Expect(h.Client).To(Equal(fakeClient))
		})

		It("should be constructable via ForGroupKind", func() {
			history, err := ForGroupKind(fakeClient, appsv1.SchemeGroupVersion.WithKind("StatefulSet").GroupKind())
			Expect(err).NotTo(HaveOccurred())
			Expect(history).NotTo(BeNil())
			Expect(history).To(BeAssignableToTypeOf(StatefulSetHistory{}))

			h := history.(StatefulSetHistory)
			Expect(h.Client).To(Equal(fakeClient))
		})
	})

	Describe("ListRevisions", func() {
		var history StatefulSetHistory

		BeforeEach(func() {
			history = StatefulSetHistory{
				Client: fakeClient,
			}
		})

		It("should fail if the StatefulSet doesn't exist", func() {
			revs, err := history.ListRevisions(ctx, client.ObjectKey{Name: "non-existing"})
			Expect(err).To(beNotFoundError())
			Expect(revs).To(BeNil())
		})

		It("should return an empty list if there are no ControllerRevisions", func() {
			statefulSet.ResourceVersion = ""
			statefulSet.UID = ""
			statefulSet.Namespace = "other"
			Expect(fakeClient.Create(ctx, statefulSet)).To(Succeed())

			revs, err := history.ListRevisions(ctx, client.ObjectKeyFromObject(statefulSet))
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(BeEmpty())
		})

		It("should return a sorted list of the owned ControllerRevisions", func() {
			revs, err := history.ListRevisions(ctx, client.ObjectKeyFromObject(statefulSet))
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(HaveLen(2))

			Expect(revs[0].Number()).To(BeEquivalentTo(1))
			Expect(revs[0].Object()).To(Equal(controllerRevision1))

			Expect(revs[1].Number()).To(BeEquivalentTo(3))
			Expect(revs[1].Object()).To(Equal(controllerRevision3))
		})
	})

	Describe("NewControllerRevisionForStatefulSet", func() {
		It("should correctly transform the ControllerRevision", func() {
			rev, err := NewControllerRevisionForStatefulSet(controllerRevision1)
			Expect(err).NotTo(HaveOccurred())

			Expect(rev.Number()).To(BeEquivalentTo(1))
			Expect(rev.Name()).To(Equal("app-b"))
			Expect(rev.Object()).To(Equal(controllerRevision1))
		})
	})
})

func controllerRevisionForStatefulSet(statefulSet *appsv1.StatefulSet, revision int64, scheme *runtime.Scheme) *appsv1.ControllerRevision {
	labels := copyMap(statefulSet.Spec.Selector.MatchLabels)

	template := statefulSet.Spec.Template.DeepCopy()
	template.Labels = labels
	template.Spec.Containers[0].Image = fmt.Sprintf("test:%d", revision)

	statefulSetData := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: template.ObjectMeta,
				Spec:       template.Spec,
			},
		},
	}

	controllerRevision := &appsv1.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", statefulSet.Name, revision),
			Namespace: statefulSet.Namespace,
			Labels:    labels,
		},
		Revision: revision,
		Data: runtime.RawExtension{
			Object: statefulSetData,
		},
	}

	Expect(controllerutil.SetControllerReference(statefulSet, controllerRevision, scheme)).To(Succeed())

	return controllerRevision
}
