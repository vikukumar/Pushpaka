package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// InfraHandler exposes Docker and Kubernetes management endpoints.
// All operations invoke CLI tools (docker / kubectl) available on the host.
type InfraHandler struct{}

func NewInfraHandler() *InfraHandler { return &InfraHandler{} }

// ─── Docker ──────────────────────────────────────────────────────────────────

type containerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	State   string `json:"state"`
	Ports   string `json:"ports"`
	Created string `json:"created"`
}

// ListContainers returns all Docker containers on the host.
// GET /api/v1/containers
func (h *InfraHandler) ListContainers(c *gin.Context) {
	out, err := runCmd("docker", "ps", "-a",
		"--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.State}}|{{.Ports}}|{{.CreatedAt}}")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "docker not available: " + err.Error()})
		return
	}
	var containers []containerInfo
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 7)
		if len(parts) < 7 {
			continue
		}
		containers = append(containers, containerInfo{
			ID:      parts[0],
			Name:    strings.TrimPrefix(parts[1], "/"),
			Image:   parts[2],
			Status:  parts[3],
			State:   strings.ToLower(parts[4]),
			Ports:   parts[5],
			Created: parts[6],
		})
	}
	if containers == nil {
		containers = []containerInfo{}
	}
	c.JSON(http.StatusOK, containers)
}

// StartContainer starts a stopped container.
// POST /api/v1/containers/:id/start
func (h *InfraHandler) StartContainer(c *gin.Context) {
	id := sanitizeContainerID(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container id"})
		return
	}
	if _, err := runCmd("docker", "start", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "container started"})
}

// StopContainer stops a running container.
// POST /api/v1/containers/:id/stop
func (h *InfraHandler) StopContainer(c *gin.Context) {
	id := sanitizeContainerID(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container id"})
		return
	}
	if _, err := runCmd("docker", "stop", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "container stopped"})
}

// RestartContainer restarts a container.
// POST /api/v1/containers/:id/restart
func (h *InfraHandler) RestartContainer(c *gin.Context) {
	id := sanitizeContainerID(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container id"})
		return
	}
	if _, err := runCmd("docker", "restart", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "container restarted"})
}

// ContainerLogs returns recent log lines from a container.
// GET /api/v1/containers/:id/logs?lines=200
func (h *InfraHandler) ContainerLogs(c *gin.Context) {
	id := sanitizeContainerID(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container id"})
		return
	}
	lines := "200"
	if q := c.Query("lines"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 2000 {
			lines = q
		}
	}
	out, err := runCmd("docker", "logs", "--tail", lines, "--timestamps", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": out})
}

// ─── Kubernetes ───────────────────────────────────────────────────────────────

// K8sNamespaces lists all namespaces.
// GET /api/v1/k8s/namespaces
func (h *InfraHandler) K8sNamespaces(c *gin.Context) {
	out, err := runCmd("kubectl", "get", "namespaces",
		"-o", "jsonpath={range .items[*]}{.metadata.name}\\n{end}")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "kubectl not available: " + err.Error()})
		return
	}
	var ns []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line != "" {
			ns = append(ns, line)
		}
	}
	if ns == nil {
		ns = []string{"default"}
	}
	c.JSON(http.StatusOK, ns)
}

type k8sPod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	Status    string `json:"status"`
	Restarts  int    `json:"restarts"`
	Age       string `json:"age"`
}

// K8sPods lists pods in a namespace.
// GET /api/v1/k8s/pods?namespace=default
func (h *InfraHandler) K8sPods(c *gin.Context) {
	ns := sanitizeK8sName(c.DefaultQuery("namespace", "default"))
	out, err := runCmd("kubectl", "get", "pods", "-n", ns,
		"--no-headers", "-o",
		"custom-columns=NAME:.metadata.name,READY:.status.containerStatuses[0].ready,STATUS:.status.phase,RESTARTS:.status.containerStatuses[0].restartCount,AGE:.metadata.creationTimestamp")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "kubectl error: " + err.Error()})
		return
	}
	// Alternative: use jsonpath for cleaner parsing
	out2, err2 := runCmd("kubectl", "get", "pods", "-n", ns, "--no-headers")
	if err2 == nil {
		out = out2
	}
	var pods []k8sPod
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}
		restarts, _ := strconv.Atoi(fields[3])
		pods = append(pods, k8sPod{
			Name:      fields[0],
			Namespace: ns,
			Ready:     fields[1],
			Status:    fields[2],
			Restarts:  restarts,
			Age:       fields[4],
		})
	}
	if pods == nil {
		pods = []k8sPod{}
	}
	c.JSON(http.StatusOK, pods)
}

type k8sDeployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	UpToDate  int    `json:"up_to_date"`
	Available int    `json:"available"`
	Age       string `json:"age"`
}

// K8sDeployments lists deployments in a namespace.
// GET /api/v1/k8s/deployments?namespace=default
func (h *InfraHandler) K8sDeployments(c *gin.Context) {
	ns := sanitizeK8sName(c.DefaultQuery("namespace", "default"))
	out, err := runCmd("kubectl", "get", "deployments", "-n", ns, "--no-headers")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "kubectl error: " + err.Error()})
		return
	}
	var deployments []k8sDeployment
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}
		upToDate, _ := strconv.Atoi(fields[2])
		available, _ := strconv.Atoi(fields[3])
		deployments = append(deployments, k8sDeployment{
			Name:      fields[0],
			Namespace: ns,
			Ready:     fields[1],
			UpToDate:  upToDate,
			Available: available,
			Age:       fields[4],
		})
	}
	if deployments == nil {
		deployments = []k8sDeployment{}
	}
	c.JSON(http.StatusOK, deployments)
}

type k8sService struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Type       string `json:"type"`
	ClusterIP  string `json:"cluster_ip"`
	ExternalIP string `json:"external_ip"`
	Ports      string `json:"ports"`
	Age        string `json:"age"`
}

// K8sServices lists services in a namespace.
// GET /api/v1/k8s/services?namespace=default
func (h *InfraHandler) K8sServices(c *gin.Context) {
	ns := sanitizeK8sName(c.DefaultQuery("namespace", "default"))
	out, err := runCmd("kubectl", "get", "services", "-n", ns, "--no-headers")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "kubectl error: " + err.Error()})
		return
	}
	var services []k8sService
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 {
			continue
		}
		services = append(services, k8sService{
			Name:       fields[0],
			Namespace:  ns,
			Type:       fields[1],
			ClusterIP:  fields[2],
			ExternalIP: fields[3],
			Ports:      fields[4],
			Age:        fields[5],
		})
	}
	if services == nil {
		services = []k8sService{}
	}
	c.JSON(http.StatusOK, services)
}

// K8sRollout triggers a rollout restart of a deployment (ArgoCD-like refresh).
// POST /api/v1/k8s/deployments/:namespace/:name/rollout
func (h *InfraHandler) K8sRollout(c *gin.Context) {
	ns := sanitizeK8sName(c.Param("namespace"))
	name := sanitizeK8sName(c.Param("name"))
	if ns == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid namespace or deployment name"})
		return
	}
	// Annotate with current timestamp to force a rolling restart
	annotation := fmt.Sprintf("kubectl.kubernetes.io/restartedAt=%s", time.Now().UTC().Format(time.RFC3339))
	_, err := runCmd("kubectl", "patch", "deployment", name,
		"-n", ns,
		"--patch", `{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"`+time.Now().UTC().Format(time.RFC3339)+`"}}}}}`)
	if err != nil {
		// Fallback: rollout restart
		_, err = runCmd("kubectl", "rollout", "restart", "deployment/"+name, "-n", ns)
		if err != nil {
			_ = annotation
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "rollout restart triggered for " + name})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

var safeContainerIDRe = regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)
var safeK8sNameRe = regexp.MustCompile(`^[a-zA-Z0-9_.\-]+$`)

func sanitizeContainerID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) > 128 || !safeContainerIDRe.MatchString(id) {
		return ""
	}
	return id
}

func sanitizeK8sName(name string) string {
	name = strings.TrimSpace(name)
	if len(name) > 253 || !safeK8sNameRe.MatchString(name) {
		return "default"
	}
	return name
}

// runCmd executes an allow-listed command, combining stdout+stderr.
func runCmd(name string, args ...string) (string, error) {
	// Only allow known safe executables
	switch name {
	case "docker", "kubectl":
		// allowed
	default:
		return "", fmt.Errorf("command %q not allowed", name)
	}
	cmd := exec.Command(name, args...) //nolint:gosec // args are sanitized above
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %w", buf.String(), err)
	}
	return buf.String(), nil
}
