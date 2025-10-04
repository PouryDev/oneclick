package kubeclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/PouryDev/oneclick/internal/domain"
)

// KubernetesClientInterface defines the interface for Kubernetes operations
type KubernetesClientInterface interface {
	GetPodsByApp(ctx context.Context, appName, namespace string) ([]domain.Pod, error)
	GetPodDetail(ctx context.Context, podName, namespace string) (*domain.PodDetail, error)
	GetPodLogs(ctx context.Context, podName, namespace string, req domain.PodLogsRequest) (*domain.PodLogsResponse, error)
	GetPodDescribe(ctx context.Context, podName, namespace string) (*domain.PodDescribeResponse, error)
	ExecInPod(ctx context.Context, podName, namespace string, req domain.PodExecRequest, conn *websocket.Conn) error
}

// KubernetesClient wraps the Kubernetes client with additional functionality
type KubernetesClient struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	logger    *zap.Logger
}

// NewKubernetesClient creates a new Kubernetes client from kubeconfig bytes
func NewKubernetesClient(kubeconfigBytes []byte, logger *zap.Logger) (*KubernetesClient, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &KubernetesClient{
		clientset: clientset,
		config:    config,
		logger:    logger,
	}, nil
}

// PodInfo represents basic pod information
type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Restarts  int32             `json:"restarts"`
	Ready     string            `json:"ready"` // "1/2" format
	Age       string            `json:"age"`
	NodeName  string            `json:"node_name"`
	Labels    map[string]string `json:"labels"`
}

// PodDetail represents detailed pod information
type PodDetail struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Restarts   int32             `json:"restarts"`
	Ready      string            `json:"ready"`
	Age        string            `json:"age"`
	NodeName   string            `json:"node_name"`
	Labels     map[string]string `json:"labels"`
	Containers []ContainerInfo   `json:"containers"`
	Events     []EventInfo       `json:"events"`
	OwnerRefs  []OwnerReference  `json:"owner_refs"`
	CreatedAt  time.Time         `json:"created_at"`
	IP         string            `json:"ip"`
	HostIP     string            `json:"host_ip"`
	Phase      string            `json:"phase"`
	Conditions []PodCondition    `json:"conditions"`
}

// ContainerInfo represents container information
type ContainerInfo struct {
	Name         string     `json:"name"`
	Image        string     `json:"image"`
	Ready        bool       `json:"ready"`
	RestartCount int32      `json:"restart_count"`
	State        string     `json:"state"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
}

// EventInfo represents Kubernetes event information
type EventInfo struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// OwnerReference represents owner reference information
type OwnerReference struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	APIVersion string `json:"api_version"`
	Controller bool   `json:"controller"`
}

// PodCondition represents pod condition information
type PodCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// GetPodsByApp returns pods for a specific application
func (k *KubernetesClient) GetPodsByApp(ctx context.Context, appName, namespace string) ([]domain.Pod, error) {
	labelSelector := labels.Set{"app": appName}.AsSelector()

	pods, err := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var podInfos []domain.Pod
	for _, pod := range pods.Items {
		podInfo := domain.Pod{
			ID:        uuid.New(),
			AppID:     uuid.New(), // This should be passed from the service layer
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Restarts:  getTotalRestartCountFromStatuses(pod.Status.ContainerStatuses),
			Ready:     getReadyContainersFromStatuses(pod.Status.ContainerStatuses),
			Age:       getAge(pod.CreationTimestamp.Time),
			NodeName:  pod.Spec.NodeName,
			Labels:    pod.Labels,
			CreatedAt: pod.CreationTimestamp.Time,
			UpdatedAt: pod.CreationTimestamp.Time,
		}
		podInfos = append(podInfos, podInfo)
	}

	return podInfos, nil
}

// GetPodDetail returns detailed information about a specific pod
func (k *KubernetesClient) GetPodDetail(ctx context.Context, podName, namespace string) (*domain.PodDetail, error) {
	pod, err := k.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Get pod events
	events, err := k.getPodEvents(ctx, namespace, podName)
	if err != nil {
		k.logger.Warn("Failed to get pod events", zap.Error(err))
		events = []EventInfo{}
	}

	// Convert to domain model
	podDetail := k.convertPodToDetail(pod, events)

	return podDetail, nil
}

// GetPodLogs returns pod logs with optional streaming
func (k *KubernetesClient) GetPodLogs(ctx context.Context, podName, namespace string, req domain.PodLogsRequest) (*domain.PodLogsResponse, error) {
	logReq := k.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: req.Container,
		Follow:    req.Follow,
		TailLines: &req.TailLines,
	})

	stream, err := logReq.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer stream.Close()

	// Read logs
	logs, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	return &domain.PodLogsResponse{
		PodName:   podName,
		Namespace: namespace,
		Container: req.Container,
		Logs:      string(logs),
		Follow:    req.Follow,
	}, nil
}

// ExecInPod executes a command in a pod and returns a websocket connection
func (k *KubernetesClient) ExecInPod(ctx context.Context, podName, namespace string, req domain.PodExecRequest, conn *websocket.Conn) error {
	execReq := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: req.Container,
			Command:   req.Command,
			Stdin:     req.Stdin,
			Stdout:    true,
			Stderr:    true,
			TTY:       req.TTY,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(k.config, http.MethodPost, execReq.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	streamOptions := remotecommand.StreamOptions{
		Stdin:  &WebSocketReader{conn: conn},
		Stdout: &WebSocketWriter{conn: conn},
		Stderr: &WebSocketWriter{conn: conn},
		Tty:    req.TTY,
	}

	return executor.StreamWithContext(ctx, streamOptions)
}

// GetPodDescribe returns detailed information about a pod in kubectl describe format
func (k *KubernetesClient) GetPodDescribe(ctx context.Context, podName, namespace string) (*domain.PodDescribeResponse, error) {
	pod, err := k.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Get pod events
	events, err := k.getPodEvents(ctx, namespace, podName)
	if err != nil {
		k.logger.Warn("Failed to get pod events", zap.Error(err))
		events = []EventInfo{}
	}

	// Convert to domain model
	podDetail := k.convertPodToDetail(pod, events)

	return &domain.PodDescribeResponse{
		PodDetail: *podDetail,
	}, nil
}

// convertPodToDetail converts a Kubernetes pod to domain PodDetail
func (k *KubernetesClient) convertPodToDetail(pod *corev1.Pod, events []EventInfo) *domain.PodDetail {
	// Convert containers
	var containers []domain.ContainerInfo
	for _, container := range pod.Spec.Containers {
		containerStatus := findContainerStatus(pod.Status.ContainerStatuses, container.Name)
		containers = append(containers, domain.ContainerInfo{
			Name:         container.Name,
			Image:        container.Image,
			Ready:        containerStatus != nil && containerStatus.Ready,
			RestartCount: getRestartCount(containerStatus),
			State:        getContainerState(containerStatus),
			StartedAt:    getContainerStartedAtPtr(containerStatus),
		})
	}

	// Convert events
	var podEvents []domain.EventInfo
	for _, event := range events {
		podEvents = append(podEvents, domain.EventInfo{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Count:     event.Count,
			FirstSeen: event.FirstSeen,
			LastSeen:  event.LastSeen,
		})
	}

	// Convert owner references
	var ownerRefs []domain.OwnerReference
	for _, ref := range pod.OwnerReferences {
		ownerRefs = append(ownerRefs, domain.OwnerReference{
			Kind:       ref.Kind,
			Name:       ref.Name,
			APIVersion: ref.APIVersion,
			Controller: ref.Controller != nil && *ref.Controller,
		})
	}

	// Convert conditions
	var conditions []domain.PodCondition
	for _, condition := range pod.Status.Conditions {
		conditions = append(conditions, domain.PodCondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}

	return &domain.PodDetail{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Status:     string(pod.Status.Phase),
		Restarts:   getTotalRestartCountFromStatuses(pod.Status.ContainerStatuses),
		Ready:      getReadyContainersFromStatuses(pod.Status.ContainerStatuses),
		Age:        getAge(pod.CreationTimestamp.Time),
		NodeName:   pod.Spec.NodeName,
		Labels:     pod.Labels,
		Containers: containers,
		Events:     podEvents,
		OwnerRefs:  ownerRefs,
		CreatedAt:  pod.CreationTimestamp.Time,
		IP:         pod.Status.PodIP,
		HostIP:     pod.Status.HostIP,
		Phase:      string(pod.Status.Phase),
		Conditions: conditions,
	}
}

// getPodEvents retrieves events for a specific pod
func (k *KubernetesClient) getPodEvents(ctx context.Context, namespace, podName string) ([]EventInfo, error) {
	fieldSelector := fields.Set{"involvedObject.name": podName}.AsSelector()

	events, err := k.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}

	var eventInfos []EventInfo
	for _, event := range events.Items {
		eventInfo := EventInfo{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Count:     event.Count,
			FirstSeen: event.FirstTimestamp.Time,
			LastSeen:  event.LastTimestamp.Time,
		}
		eventInfos = append(eventInfos, eventInfo)
	}

	return eventInfos, nil
}

// Helper functions

func getTotalRestartCount(pod corev1.Pod) int32 {
	var total int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		total += containerStatus.RestartCount
	}
	return total
}

func getReadyContainers(pod corev1.Pod) string {
	ready := 0
	total := len(pod.Status.ContainerStatuses)

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			ready++
		}
	}

	return fmt.Sprintf("%d/%d", ready, total)
}

func getAge(createdAt time.Time) string {
	duration := time.Since(createdAt)

	if duration.Hours() >= 24 {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	} else if duration.Hours() >= 1 {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh", hours)
	} else if duration.Minutes() >= 1 {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm", minutes)
	} else {
		seconds := int(duration.Seconds())
		return fmt.Sprintf("%ds", seconds)
	}
}

func getOwnerReferences(refs []metav1.OwnerReference) []OwnerReference {
	var ownerRefs []OwnerReference
	for _, ref := range refs {
		ownerRef := OwnerReference{
			Kind:       ref.Kind,
			Name:       ref.Name,
			APIVersion: ref.APIVersion,
			Controller: ref.Controller != nil && *ref.Controller,
		}
		ownerRefs = append(ownerRefs, ownerRef)
	}
	return ownerRefs
}

func getContainersInfo(pod *corev1.Pod) []ContainerInfo {
	var containers []ContainerInfo
	for _, container := range pod.Spec.Containers {
		containerInfo := ContainerInfo{
			Name:  container.Name,
			Image: container.Image,
		}

		// Find container status
		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == container.Name {
				containerInfo.Ready = status.Ready
				containerInfo.RestartCount = status.RestartCount

				if status.State.Running != nil {
					containerInfo.State = "Running"
					containerInfo.StartedAt = &status.State.Running.StartedAt.Time
				} else if status.State.Waiting != nil {
					containerInfo.State = "Waiting"
				} else if status.State.Terminated != nil {
					containerInfo.State = "Terminated"
				}
				break
			}
		}

		containers = append(containers, containerInfo)
	}
	return containers
}

func getPodConditions(conditions []corev1.PodCondition) []PodCondition {
	var podConditions []PodCondition
	for _, condition := range conditions {
		podCondition := PodCondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		}
		podConditions = append(podConditions, podCondition)
	}
	return podConditions
}

// WebSocketReader implements io.Reader for websocket connections
type WebSocketReader struct {
	conn *websocket.Conn
}

func (r *WebSocketReader) Read(p []byte) (n int, err error) {
	_, message, err := r.conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	n = copy(p, message)
	return n, nil
}

// WebSocketWriter implements io.Writer for websocket connections
type WebSocketWriter struct {
	conn *websocket.Conn
}

func (w *WebSocketWriter) Write(p []byte) (n int, err error) {
	err = w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Helper functions for pod conversion
func findContainerStatus(statuses []corev1.ContainerStatus, name string) *corev1.ContainerStatus {
	for _, status := range statuses {
		if status.Name == name {
			return &status
		}
	}
	return nil
}

func getRestartCount(status *corev1.ContainerStatus) int32 {
	if status == nil {
		return 0
	}
	return status.RestartCount
}

func getContainerState(status *corev1.ContainerStatus) string {
	if status == nil {
		return "Unknown"
	}
	if status.State.Running != nil {
		return "Running"
	}
	if status.State.Waiting != nil {
		return "Waiting"
	}
	if status.State.Terminated != nil {
		return "Terminated"
	}
	return "Unknown"
}

func getContainerStartedAt(status *corev1.ContainerStatus) time.Time {
	if status == nil || status.State.Running == nil {
		return time.Time{}
	}
	return status.State.Running.StartedAt.Time
}

func getContainerStartedAtPtr(status *corev1.ContainerStatus) *time.Time {
	if status == nil || status.State.Running == nil {
		return nil
	}
	t := status.State.Running.StartedAt.Time
	return &t
}

func getTotalRestartCountFromStatuses(statuses []corev1.ContainerStatus) int32 {
	var total int32
	for _, status := range statuses {
		total += status.RestartCount
	}
	return total
}

func getReadyContainersFromStatuses(statuses []corev1.ContainerStatus) string {
	var ready int32
	var total int32 = int32(len(statuses))

	for _, status := range statuses {
		if status.Ready {
			ready++
		}
	}

	return fmt.Sprintf("%d/%d", ready, total)
}
