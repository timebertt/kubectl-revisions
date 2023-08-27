package history_test

import (
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"

	. "github.com/timebertt/kubectl-history/pkg/history"
)

var _ = Describe("ReplicaSet", func() {
	var (
		replicaSet *appsv1.ReplicaSet
		rev        *ReplicaSet
	)

	BeforeEach(func() {
		replicaSet = &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deploy-1",
				Namespace: "test",
				Annotations: map[string]string{
					deploymentutil.RevisionAnnotation: strconv.FormatInt(1, 10),
				},
			},
			Spec: appsv1.ReplicaSetSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":               "deploy",
							"pod-template-hash": "deploy-1",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "test",
							Image: "test",
						}},
					},
				},
			},
		}

		var err error
		rev, err = NewReplicaSet(replicaSet)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("NewReplicaSet", func() {
		It("should correctly transform the ReplicaSet", func() {
			Expect(rev.Number()).To(BeEquivalentTo(1))
			Expect(rev.Name()).To(Equal("deploy-1"))
			Expect(rev.Object()).To(Equal(replicaSet))
		})

		It("should fail parsing the revision number", func() {
			replicaSet.Annotations[deploymentutil.RevisionAnnotation] = "foo"

			rev, err := NewReplicaSet(replicaSet)
			Expect(err).To(MatchError(ContainSubstring("error parsing revision")))
			Expect(rev).To(BeNil())
		})
	})

	Describe("GetObjectKind", func() {
		It("should return an empty ObjectKind if the ReplicaSet is nil", func() {
			rev = nil
			Expect(rev.GetObjectKind().GroupVersionKind().Empty()).To(BeTrue())
		})

		It("should return the correct copied ObjectKind", func() {
			objectKind := rev.GetObjectKind()
			Expect(objectKind.GroupVersionKind()).To(Equal(appsv1.SchemeGroupVersion.WithKind("ReplicaSet")))
			Expect(objectKind).NotTo(BeIdenticalTo(replicaSet.GetObjectKind()))
		})
	})

	Describe("DeepCopyObject", func() {
		It("should return nil if the ReplicaSet is nil", func() {
			rev = nil
			Expect(rev.DeepCopyObject()).To(BeNil())
		})

		It("should return a copy of the ReplicaSet", func() {
			copied := rev.DeepCopyObject()
			Expect(copied).To(BeAssignableToTypeOf(rev))
			Expect(copied).To(Equal(rev))
			Expect(copied).NotTo(BeIdenticalTo(rev))

			copiedRev := copied.(*ReplicaSet)
			Expect(copiedRev.ReplicaSet).NotTo(BeIdenticalTo(rev.ReplicaSet))
		})
	})

	Describe("PodTemplate", func() {
		It("should return a copy of the template without the pod-template-hash label", func() {
			expectedTemplate := replicaSet.Spec.Template.DeepCopy()
			delete(expectedTemplate.Labels, "pod-template-hash")

			template := rev.PodTemplate()
			Expect(template.ObjectMeta).To(Equal(expectedTemplate.ObjectMeta))
			Expect(template.Spec).To(Equal(expectedTemplate.Spec))
		})
	})
})
