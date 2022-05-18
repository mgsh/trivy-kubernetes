package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	KindPod        = "Pod"
	KindJob        = "Job"
	KindCronJob    = "CronJob"
	KindReplicaSet = "ReplicaSet"

	Deployments            = "deployments"
	ReplicaSets            = "replicasets"
	ReplicationControllers = "replicationcontrollers"
	StatefulSets           = "statefulsets"
	DaemonSets             = "daemonsets"
	CronJobs               = "cronjobs"
	Services               = "services"
	Jobs                   = "jobs"
	Pods                   = "pods"
	ConfigMaps             = "configmaps"
	Roles                  = "roles"
	RoleBindings           = "rolebindings"
	NetworkPolicys         = "networkpolicies"
	Ingresss               = "ingresses"
	ResourceQuotas         = "resourcequotas"
	LimitRanges            = "limitranges"
	ClusterRoles           = "clusterroles"
	ClusterRoleBindings    = "clusterrolebindings"
	PodSecurityPolicies    = "podsecuritypolicies"
)

// Cluster interface represents the operations needed to scan a cluster
type Cluster interface {
	// GetCurrentContext returns local kubernetes current-context
	GetCurrentContext() string
	// GetCurrentNamespace returns local kubernetes current namespace
	GetCurrentNamespace() string
	// GetDynamicClient returns a dynamic k8s client
	GetDynamicClient() dynamic.Interface
	// GetGVRs returns cluster GroupVersionResource to query kubernetes, receives
	// a boolean to determine if returns namespaced GVRs only or all GVRs
	GetGVRs(bool) ([]schema.GroupVersionResource, error)
	// GetGVR returns resource GroupVersionResource to query kubernetes, receives
	// a string with the resource kind
	GetGVR(string) (schema.GroupVersionResource, error)
}

type cluster struct {
	currentContext   string
	currentNamespace string
	dynamicClient    dynamic.Interface
	restMapper       meta.RESTMapper
}

// GetCluster returns a current configured cluster
func GetCluster() (Cluster, error) {
	cf := genericclioptions.NewConfigFlags(true)

	kubeConfig, err := cf.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// disable warnings
	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	k8sDynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	rawCfg, err := cf.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return nil, err
	}

	var namespace string
	if context, ok := rawCfg.Contexts[rawCfg.CurrentContext]; ok {
		namespace = context.Namespace
	}

	restMapper, err := cf.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	return &cluster{
		currentContext:   rawCfg.CurrentContext,
		currentNamespace: namespace,
		dynamicClient:    k8sDynamicClient,
		restMapper:       restMapper,
	}, nil
}

// GetCurrentContext returns local kubernetes current-context
func (c *cluster) GetCurrentContext() string {
	return c.currentContext
}

// GetCurrentNamespace returns local kubernetes current namespace
func (c *cluster) GetCurrentNamespace() string {
	return c.currentNamespace
}

// GetDynamicClient returns a dynamic k8s client
func (c *cluster) GetDynamicClient() dynamic.Interface {
	return c.dynamicClient
}

// GetGVRs returns cluster GroupVersionResource to query kubernetes, receives
// a boolean to determine if returns namespaced GVRs only or all GVRs
func (c *cluster) GetGVRs(namespaced bool) ([]schema.GroupVersionResource, error) {
	grvs := make([]schema.GroupVersionResource, 0)

	resources := getNamespaceResources()
	if !namespaced {
		resources = append(resources, getClusterResources()...)
	}

	for _, resource := range resources {
		list, err := c.restMapper.ResourcesFor(schema.GroupVersionResource{Resource: resource})
		if err != nil {
			return nil, err
		}

		grvs = append(grvs, list...)
	}

	return grvs, nil
}

func (c *cluster) GetGVR(kind string) (schema.GroupVersionResource, error) {
	return c.restMapper.ResourceFor(schema.GroupVersionResource{Resource: kind})
}

func IsClusterResource(gvr schema.GroupVersionResource) bool {
	for _, r := range getClusterResources() {
		if gvr.Resource == r {
			return true
		}
	}
	return false
}

func getClusterResources() []string {
	return []string{
		ClusterRoles,
		ClusterRoleBindings,
		PodSecurityPolicies,
	}
}

func getNamespaceResources() []string {
	return []string{
		Deployments,
		Pods,
		ReplicaSets,
		ReplicationControllers,
		StatefulSets,
		DaemonSets,
		CronJobs,
		Jobs,
		Services,
		ConfigMaps,
		Roles,
		RoleBindings,
		NetworkPolicys,
		Ingresss,
		ResourceQuotas,
		LimitRanges,
	}
}