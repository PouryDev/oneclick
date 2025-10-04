package kubeclient

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MockKubernetesClient is a mock implementation of KubernetesClient
type MockKubernetesClient struct {
	mock.Mock
}

func (m *MockKubernetesClient) GetPodsByApp(ctx context.Context, namespace, appName string) ([]PodInfo, error) {
	args := m.Called(ctx, namespace, appName)
	return args.Get(0).([]PodInfo), args.Error(1)
}

func (m *MockKubernetesClient) GetPodDetail(ctx context.Context, namespace, podName string) (*PodDetail, error) {
	args := m.Called(ctx, namespace, podName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PodDetail), args.Error(1)
}

func (m *MockKubernetesClient) GetPodLogs(ctx context.Context, namespace, podName, container string, follow bool, tailLines int64) (io.ReadCloser, error) {
	args := m.Called(ctx, namespace, podName, container, follow, tailLines)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockKubernetesClient) ExecInPod(ctx context.Context, namespace, podName, container string, command []string, conn *websocket.Conn) error {
	args := m.Called(ctx, namespace, podName, container, command, conn)
	return args.Error(0)
}

func TestGetPodsByApp(t *testing.T) {
	// Create fake Kubernetes objects
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod-1",
			Namespace:         "test-namespace",
			Labels:            map[string]string{"app": "test-app"},
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
		},
		Spec: corev1.PodSpec{
			NodeName: "test-node",
			Containers: []corev1.Container{
				{
					Name:  "container-1",
					Image: "nginx:latest",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "container-1",
					Ready:        true,
					RestartCount: 0,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{Time: time.Now().Add(-30 * time.Minute)},
						},
					},
				},
			},
		},
	}

	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod-2",
			Namespace:         "test-namespace",
			Labels:            map[string]string{"app": "test-app"},
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-2 * time.Hour)},
		},
		Spec: corev1.PodSpec{
			NodeName: "test-node-2",
			Containers: []corev1.Container{
				{
					Name:  "container-2",
					Image: "redis:latest",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "container-2",
					Ready:        false,
					RestartCount: 2,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ImagePullBackOff",
						},
					},
				},
			},
		},
	}

	// Test helper functions directly since we can't easily mock the Kubernetes client
	// Test getTotalRestartCount
	totalRestarts := getTotalRestartCount(*pod1)
	assert.Equal(t, int32(0), totalRestarts)

	totalRestarts2 := getTotalRestartCount(*pod2)
	assert.Equal(t, int32(2), totalRestarts2)

	// Test getReadyContainers
	readyContainers := getReadyContainers(*pod1)
	assert.Equal(t, "1/1", readyContainers)

	readyContainers2 := getReadyContainers(*pod2)
	assert.Equal(t, "0/1", readyContainers2)

	// Test getAge
	age := getAge(pod1.CreationTimestamp.Time)
	assert.NotEmpty(t, age)
}

func TestGetPodDetail(t *testing.T) {
	// Create fake pod with events
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod",
			Namespace:         "test-namespace",
			Labels:            map[string]string{"app": "test-app"},
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "ReplicaSet",
					Name:       "test-rs",
					APIVersion: "apps/v1",
					Controller: func() *bool { b := true; return &b }(),
				},
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "test-node",
			Containers: []corev1.Container{
				{
					Name:  "container-1",
					Image: "nginx:latest",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase:  corev1.PodRunning,
			PodIP:  "10.0.0.1",
			HostIP: "192.168.1.1",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "container-1",
					Ready:        true,
					RestartCount: 0,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{Time: time.Now().Add(-30 * time.Minute)},
						},
					},
				},
			},
			Conditions: []corev1.PodCondition{
				{
					Type:               corev1.PodReady,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-30 * time.Minute)},
					Reason:             "PodReady",
					Message:            "Pod is ready",
				},
			},
		},
	}

	// Test helper functions directly
	// Test getTotalRestartCount
	totalRestarts := getTotalRestartCount(*pod)
	assert.Equal(t, int32(0), totalRestarts)

	// Test getReadyContainers
	readyContainers := getReadyContainers(*pod)
	assert.Equal(t, "1/1", readyContainers)

	// Test getAge
	age := getAge(pod.CreationTimestamp.Time)
	assert.NotEmpty(t, age)

	// Test getOwnerReferences
	ownerRefs := getOwnerReferences(pod.OwnerReferences)
	assert.Len(t, ownerRefs, 1)
	assert.Equal(t, "ReplicaSet", ownerRefs[0].Kind)
	assert.Equal(t, "test-rs", ownerRefs[0].Name)
	assert.True(t, ownerRefs[0].Controller)

	// Test getContainersInfo
	containers := getContainersInfo(pod)
	assert.Len(t, containers, 1)
	assert.Equal(t, "container-1", containers[0].Name)
	assert.Equal(t, "nginx:latest", containers[0].Image)
	assert.True(t, containers[0].Ready)
	assert.Equal(t, int32(0), containers[0].RestartCount)
	assert.Equal(t, "Running", containers[0].State)

	// Test getPodConditions
	conditions := getPodConditions(pod.Status.Conditions)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "Ready", conditions[0].Type)
	assert.Equal(t, "True", conditions[0].Status)
	assert.Equal(t, "PodReady", conditions[0].Reason)
	assert.Equal(t, "Pod is ready", conditions[0].Message)
}

func TestGetAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		created  time.Time
		expected string
	}{
		{
			name:     "1 day old",
			created:  now.Add(-25 * time.Hour),
			expected: "1d",
		},
		{
			name:     "2 hours old",
			created:  now.Add(-2 * time.Hour),
			expected: "2h",
		},
		{
			name:     "30 minutes old",
			created:  now.Add(-30 * time.Minute),
			expected: "30m",
		},
		{
			name:     "45 seconds old",
			created:  now.Add(-45 * time.Second),
			expected: "45s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAge(tt.created)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetReadyContainers(t *testing.T) {
	tests := []struct {
		name     string
		pod      corev1.Pod
		expected string
	}{
		{
			name: "all containers ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "container-1", Ready: true},
						{Name: "container-2", Ready: true},
					},
				},
			},
			expected: "2/2",
		},
		{
			name: "some containers ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "container-1", Ready: true},
						{Name: "container-2", Ready: false},
					},
				},
			},
			expected: "1/2",
		},
		{
			name: "no containers ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "container-1", Ready: false},
						{Name: "container-2", Ready: false},
					},
				},
			},
			expected: "0/2",
		},
		{
			name: "no containers",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			expected: "0/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReadyContainers(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTotalRestartCount(t *testing.T) {
	tests := []struct {
		name     string
		pod      corev1.Pod
		expected int32
	}{
		{
			name: "multiple containers with restarts",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "container-1", RestartCount: 3},
						{Name: "container-2", RestartCount: 2},
					},
				},
			},
			expected: 5,
		},
		{
			name: "no restarts",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "container-1", RestartCount: 0},
						{Name: "container-2", RestartCount: 0},
					},
				},
			},
			expected: 0,
		},
		{
			name: "no containers",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTotalRestartCount(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}
