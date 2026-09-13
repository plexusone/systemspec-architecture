package catalog

// k8sDisplayNames maps Kubernetes resource kinds (sas.Technology.Service
// when Provider is "k8s") to their human-friendly names.
var k8sDisplayNames = map[string]string{ //nolint:gosec // display-name lookup table, not a credential
	"deployment":  "Kubernetes Deployment",
	"statefulset": "Kubernetes StatefulSet",
	"daemonset":   "Kubernetes DaemonSet",
	"service":     "Kubernetes Service",
	"ingress":     "Kubernetes Ingress",
	"configmap":   "Kubernetes ConfigMap",
	"secret":      "Kubernetes Secret",
	"job":         "Kubernetes Job",
	"cronjob":     "Kubernetes CronJob",
	// Cluster topology objects. A namespace or cluster is best modeled as
	// a Boundary (kind network/account) carrying this Technology, and a
	// node/pod as a compute.instance/compute.container node — the catalog
	// only supplies the display label, never a new NodeKind.
	"cluster":         "Kubernetes Cluster",
	"namespace":       "Kubernetes Namespace",
	"node":            "Kubernetes Node",
	"pod":             "Kubernetes Pod",
	"service_account": "Kubernetes ServiceAccount",
	// RBAC objects.
	"role":               "Kubernetes Role",
	"clusterrole":        "Kubernetes ClusterRole",
	"rolebinding":        "Kubernetes RoleBinding",
	"clusterrolebinding": "Kubernetes ClusterRoleBinding",
}
