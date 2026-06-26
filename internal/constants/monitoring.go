package constants

const (
	MonitoringNamespace = "openshift-monitoring"
	CMODeployment       = "cluster-monitoring-operator"
	PluginDeployment    = "monitoring-plugin"
	MonitoringUIPluginName = "monitoring"

	ACMAlertmanagerURL  = "https://alertmanager.open-cluster-management-observability.svc:9095"
	ACMThanosQuerierURL = "https://rbac-query-proxy.open-cluster-management-observability.svc:8443"
)
