package constants

const (
	// Cluster Logging Operator
	LoggingNamespace      = "openshift-logging"
	LoggingOperatorName   = "cluster-logging"
	LoggingPackageName    = "cluster-logging"
	LoggingCatalogSource  = "redhat-operators"
	DefaultLoggingChannel = "stable"

	LoggingOperatorGroupName = "openshift-logging"

	// Loki Operator
	LokiNamespace             = "openshift-operators-redhat"
	LokiOperatorName          = "loki-operator"
	LokiPackageName           = "loki-operator"
	LokiCatalogSource         = "redhat-operators"
	DefaultLokiChannel        = "stable"
	LokiOperatorGroupName     = "openshift-operators-redhat"
	LokiOperatorDeployment    = "loki-operator-controller-manager"
	ClusterLoggingDeployment  = "cluster-logging-operator"

	// LokiStack
	LokiStackName              = "logging-loki"
	LokiStackSize              = "1x.demo"
	LokiGatewayDeployment      = "logging-loki-gateway"

	// ClusterLogForwarder
	ClusterLogForwarderName    = "collector"
	CollectorServiceAccount    = "collector"
	LogCollectorServiceAccount = "logcollector"

	// Logging UIPlugin
	LoggingUIPluginName      = "logging"
	LoggingUIPluginLogsLimit = 50
	LoggingUIPluginTimeout   = "30s"
	LoggingUIPluginSchema    = "select"

	// Logging signal generator
	ChatNamespace      = "chat"
	ChatDeploymentName = "chat-x"
	ChatImage          = "quay.io/libpod/alpine"
)
