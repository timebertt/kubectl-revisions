package helper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	. "github.com/timebertt/kubectl-revisions/pkg/helper"
)

var _ = Describe("Pod helpers", func() {
	var pod *corev1.Pod

	BeforeEach(func() {
		pod = &corev1.Pod{}
	})

	Describe("IsPodReady", func() {
		It("should return false if the Ready condition is missing", func() {
			SetPodCondition(pod, corev1.PodInitialized, corev1.ConditionTrue)
			Expect(IsPodReady(pod)).To(BeFalse())
		})

		It("should return false if the Ready condition is not true", func() {
			SetPodCondition(pod, corev1.PodReady, corev1.ConditionUnknown)
			Expect(IsPodReady(pod)).To(BeFalse())
			SetPodCondition(pod, corev1.PodReady, corev1.ConditionFalse)
			Expect(IsPodReady(pod)).To(BeFalse())
		})
	})

	Describe("GetPodCondition", func() {
		It("should return the condition if it is present", func() {
			SetPodCondition(pod, corev1.PodInitialized, corev1.ConditionTrue)
			SetPodCondition(pod, corev1.PodReady, corev1.ConditionTrue)
			SetPodCondition(pod, corev1.ContainersReady, corev1.ConditionTrue)
			Expect(GetPodCondition(pod.Status.Conditions, corev1.PodReady)).To(BeIdenticalTo(&pod.Status.Conditions[1]))
		})

		It("should return nil if the condition is missing", func() {
			Expect(GetPodCondition(pod.Status.Conditions, corev1.PodReady)).To(BeNil())

			SetPodCondition(pod, corev1.PodInitialized, corev1.ConditionTrue)
			SetPodCondition(pod, corev1.ContainersReady, corev1.ConditionTrue)
			Expect(GetPodCondition(pod.Status.Conditions, corev1.PodReady)).To(BeNil())
		})
	})

	Describe("SetPodCondition", func() {
		It("should set the condition status if it is present", func() {
			pod.Status.Conditions = []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionFalse},
				{Type: corev1.ContainersReady, Status: corev1.ConditionFalse},
			}

			SetPodCondition(pod, corev1.PodReady, corev1.ConditionTrue)

			Expect(pod.Status.Conditions).To(ConsistOf(
				corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue},
				corev1.PodCondition{Type: corev1.ContainersReady, Status: corev1.ConditionFalse},
			))
		})

		It("should add the condition status if it is present", func() {
			pod.Status.Conditions = []corev1.PodCondition{
				{Type: corev1.PodInitialized, Status: corev1.ConditionFalse},
				{Type: corev1.ContainersReady, Status: corev1.ConditionFalse},
			}

			SetPodCondition(pod, corev1.PodReady, corev1.ConditionTrue)

			Expect(pod.Status.Conditions).To(ConsistOf(
				corev1.PodCondition{Type: corev1.PodInitialized, Status: corev1.ConditionFalse},
				corev1.PodCondition{Type: corev1.ContainersReady, Status: corev1.ConditionFalse},
				corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			))
		})
	})
})
