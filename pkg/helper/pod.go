package helper

import (
	corev1 "k8s.io/api/core/v1"
)

// IsPodReady returns true if a pod is ready.
func IsPodReady(pod *corev1.Pod) bool {
	condition := GetPodCondition(pod.Status.Conditions, corev1.PodReady)
	return condition != nil && condition.Status == corev1.ConditionTrue
}

// GetPodCondition extracts the provided condition from the given conditions list.
// Returns nil if the condition is not present.
func GetPodCondition(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) *corev1.PodCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

// SetPodCondition sets the provided condition in the Pod to the given status or adds the condition if it is missing.
func SetPodCondition(pod *corev1.Pod, conditionType corev1.PodConditionType, status corev1.ConditionStatus) {
	condition := GetPodCondition(pod.Status.Conditions, conditionType)
	if condition != nil {
		condition.Status = status
		return
	}

	condition = &corev1.PodCondition{
		Type:   conditionType,
		Status: status,
	}
	pod.Status.Conditions = append(pod.Status.Conditions, *condition)
}
