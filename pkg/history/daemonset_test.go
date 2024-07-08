package history_test

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/timebertt/kubectl-revisions/pkg/helper"
	. "github.com/timebertt/kubectl-revisions/pkg/history"
	"github.com/timebertt/kubectl-revisions/pkg/maps"
)

var _ = Describe("DaemonSetHistory", func() {
	var (
		ctx        context.Context
		fakeClient client.Client

		daemonSet *appsv1.DaemonSet

		controllerRevision1, controllerRevision3, controllerRevisionUnrelated *appsv1.ControllerRevision
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeClient = fakeclient.NewClientBuilder().Build()

		daemonSet = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ds",
				Namespace: "test",
				Labels: map[string]string{
					"app": "ds",
				},
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "ds",
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

		Expect(fakeClient.Create(ctx, daemonSet)).To(Succeed())

		controllerRevision1 = controllerRevisionForDaemonSet(daemonSet, 1, fakeClient.Scheme())
		controllerRevision1.Name = "app-b"
		Expect(fakeClient.Create(ctx, controllerRevision1)).To(Succeed())

		// create a non-sorted list of ControllerRevisions to verify that ListRevisions returns a sorted list
		controllerRevision3 = controllerRevisionForDaemonSet(daemonSet, 3, fakeClient.Scheme())
		controllerRevision3.Name = "app-a"
		Expect(fakeClient.Create(ctx, controllerRevision3)).To(Succeed())

		controllerRevisionUnrelated = controllerRevisionForDaemonSet(daemonSet, 0, fakeClient.Scheme())
		controllerRevisionUnrelated.OwnerReferences[0].UID = "other"
		Expect(fakeClient.Create(ctx, controllerRevisionUnrelated)).To(Succeed())
	})

	Describe("initialization", func() {
		It("should be constructable via For", func() {
			history, err := For(fakeClient, &appsv1.DaemonSet{})
			Expect(err).NotTo(HaveOccurred())
			Expect(history).NotTo(BeNil())
			Expect(history).To(BeAssignableToTypeOf(DaemonSetHistory{}))

			h := history.(DaemonSetHistory)
			Expect(h.Client).To(Equal(fakeClient))
		})

		It("should be constructable via ForGroupKind", func() {
			history, err := ForGroupKind(fakeClient, appsv1.SchemeGroupVersion.WithKind("DaemonSet").GroupKind())
			Expect(err).NotTo(HaveOccurred())
			Expect(history).NotTo(BeNil())
			Expect(history).To(BeAssignableToTypeOf(DaemonSetHistory{}))

			h := history.(DaemonSetHistory)
			Expect(h.Client).To(Equal(fakeClient))
		})
	})

	Describe("ListRevisions", func() {
		var history DaemonSetHistory

		BeforeEach(func() {
			history = DaemonSetHistory{
				Client: fakeClient,
			}
		})

		It("should return an empty list if there are no ControllerRevisions", func() {
			daemonSet.ResourceVersion = ""
			daemonSet.UID = ""
			daemonSet.Namespace = "other"
			Expect(fakeClient.Create(ctx, daemonSet)).To(Succeed())

			revs, err := history.ListRevisions(ctx, daemonSet)
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(BeEmpty())
		})

		It("should return a sorted list of the owned ControllerRevisions", func() {
			// prepare some pods for all revisions
			pod := podForDaemonSetRevision(controllerRevision1)
			helper.SetPodCondition(pod, corev1.PodReady, corev1.ConditionTrue)
			Expect(fakeClient.Create(context.Background(), pod)).To(Succeed())
			pod = podForDaemonSetRevision(controllerRevision1)
			helper.SetPodCondition(pod, corev1.PodReady, corev1.ConditionFalse)
			Expect(fakeClient.Create(context.Background(), pod)).To(Succeed())

			pod = podForDaemonSetRevision(controllerRevision3)
			helper.SetPodCondition(pod, corev1.PodReady, corev1.ConditionTrue)
			Expect(fakeClient.Create(context.Background(), pod)).To(Succeed())

			pod = podForDaemonSetRevision(controllerRevisionUnrelated)
			helper.SetPodCondition(pod, corev1.PodReady, corev1.ConditionTrue)
			Expect(fakeClient.Create(context.Background(), pod)).To(Succeed())
			pod = podForDaemonSetRevision(controllerRevisionUnrelated)
			helper.SetPodCondition(pod, corev1.PodReady, corev1.ConditionTrue)
			Expect(fakeClient.Create(context.Background(), pod)).To(Succeed())

			revs, err := history.ListRevisions(ctx, daemonSet)
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(HaveLen(2))

			Expect(revs[0].Number()).To(BeEquivalentTo(1))
			Expect(revs[0].Object()).To(Equal(controllerRevision1))
			Expect(revs[0].CurrentReplicas()).To(BeEquivalentTo(2))
			Expect(revs[0].ReadyReplicas()).To(BeEquivalentTo(1))

			Expect(revs[1].Number()).To(BeEquivalentTo(3))
			Expect(revs[1].Object()).To(Equal(controllerRevision3))
			Expect(revs[1].CurrentReplicas()).To(BeEquivalentTo(1))
			Expect(revs[1].ReadyReplicas()).To(BeEquivalentTo(1))
		})

		It("should also work via ListRevisions shortcut", func() {
			revs, err := ListRevisions(ctx, fakeClient, daemonSet)
			Expect(err).NotTo(HaveOccurred())
			Expect(revs).To(HaveLen(2))

			Expect(revs[0].Object()).To(Equal(controllerRevision1))
			Expect(revs[1].Object()).To(Equal(controllerRevision3))
		})
	})

	Describe("NewControllerRevisionForDaemonSet", func() {
		It("should correctly transform the ControllerRevision", func() {
			rev, err := NewControllerRevisionForDaemonSet(controllerRevision1)
			Expect(err).NotTo(HaveOccurred())

			Expect(rev.Number()).To(BeEquivalentTo(1))
			Expect(rev.Name()).To(Equal("app-b"))
			Expect(rev.Object()).To(Equal(controllerRevision1))
		})
	})

	Describe("PodBelongsToDaemonSetRevision", func() {
		var related *corev1.Pod

		BeforeEach(func() {
			related = podForDaemonSetRevision(controllerRevision1)
		})

		It("should return true for a related pod", func() {
			Expect(PodBelongsToDaemonSetRevision(controllerRevision1)(related)).To(BeTrue())
		})

		It("should return true for a related pod", func() {
			unrelated := related.DeepCopy()
			unrelated.Labels["controller-revision-hash"] = "other"

			Expect(PodBelongsToDaemonSetRevision(controllerRevision1)(unrelated)).To(BeFalse())
		})
	})
})

func controllerRevisionForDaemonSet(daemonSet *appsv1.DaemonSet, revision int64, scheme *runtime.Scheme) *appsv1.ControllerRevision {
	labels := copyMap(daemonSet.Spec.Selector.MatchLabels)

	template := daemonSet.Spec.Template.DeepCopy()
	template.Labels = labels
	template.Spec.Containers[0].Image = fmt.Sprintf("test:%d", revision)

	daemonSetData := &appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: template.ObjectMeta,
				Spec:       template.Spec,
			},
		},
	}

	controllerRevision := &appsv1.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", daemonSet.Name, revision),
			Namespace: daemonSet.Namespace,
			Labels:    maps.Merge(labels, map[string]string{"controller-revision-hash": strconv.FormatInt(revision, 10)}),
		},
		Revision: revision,
		Data: runtime.RawExtension{
			Object: daemonSetData,
		},
	}

	Expect(controllerutil.SetControllerReference(daemonSet, controllerRevision, scheme)).To(Succeed())

	return controllerRevision
}

func podForDaemonSetRevision(revision *appsv1.ControllerRevision) *corev1.Pod {
	daemonSet := revision.Data.Object.(*appsv1.DaemonSet)

	template := daemonSet.Spec.Template.DeepCopy()
	pod := &corev1.Pod{
		ObjectMeta: template.ObjectMeta,
		Spec:       template.Spec,
	}
	pod.Labels["controller-revision-hash"] = strconv.FormatInt(revision.Revision, 10)
	pod.Namespace = revision.Namespace
	// this is not like in the real-world case but allows to easily create multiple pod on the fake client
	pod.GenerateName = revision.Name + "-"

	return pod
}
