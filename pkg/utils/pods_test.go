package utils

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGetPodErrorMessage(t *testing.T) {
	t.Run("nil pod", func(t *testing.T) {
		if got := GetPodErrorMessage(nil); got != "Pod is nil" {
			t.Fatalf("GetPodErrorMessage(nil) = %q, want %q", got, "Pod is nil")
		}
	})

	t.Run("waiting message has priority", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{Message: "waiting"},
						},
					},
					{
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{Message: "terminated"},
						},
					},
				},
			},
		}

		if got := GetPodErrorMessage(pod); got != "waiting" {
			t.Fatalf("GetPodErrorMessage() = %q, want %q", got, "waiting")
		}
	})

	t.Run("terminated message is used when waiting is absent", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{Message: "terminated"},
						},
					},
				},
			},
		}

		if got := GetPodErrorMessage(pod); got != "terminated" {
			t.Fatalf("GetPodErrorMessage() = %q, want %q", got, "terminated")
		}
	})
}

func TestPodStatusHelpers(t *testing.T) {
	t.Run("is pod ready", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}

		if !IsPodReady(pod) {
			t.Fatal("IsPodReady() = false, want true")
		}
	})

	t.Run("is pod not ready when phase is not running", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}

		if IsPodReady(pod) {
			t.Fatal("IsPodReady() = true, want false")
		}
	})

	t.Run("is pod error or success", func(t *testing.T) {
		for _, phase := range []corev1.PodPhase{corev1.PodFailed, corev1.PodSucceeded} {
			pod := &corev1.Pod{Status: corev1.PodStatus{Phase: phase}}
			if !IsPodErrorOrSuccess(pod) {
				t.Fatalf("IsPodErrorOrSuccess(%s) = false, want true", phase)
			}
		}
	})

	t.Run("is pod neither error nor success", func(t *testing.T) {
		pod := &corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodRunning}}
		if IsPodErrorOrSuccess(pod) {
			t.Fatal("IsPodErrorOrSuccess() = true, want false")
		}
	})
}
