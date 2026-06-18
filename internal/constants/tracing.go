package constants

const (
	// Namespaces
	TracingNamespace     = "openshift-tracing"
	TempoOperatorNS      = "openshift-tempo-operator"
	OTelOperatorNS       = "openshift-opentelemetry-operator"

	// Tempo Operator
	TempoOperatorName    = "tempo-product"
	TempoPackageName     = "tempo-product"
	TempoCatalogSource   = "redhat-operators"
	DefaultTempoChannel  = "stable"
	TempoOperatorGroup   = "openshift-tempo-operator"

	// OpenTelemetry Operator
	OTelOperatorName     = "opentelemetry-product"
	OTelPackageName      = "opentelemetry-product"
	OTelCatalogSource    = "redhat-operators"
	DefaultOTelChannel   = "stable"
	OTelOperatorGroup    = "openshift-opentelemetry-operator"

	// TempoStack
	TempoStackName       = "platform"
	TempoStackSize       = "1Gi"
	TempoGatewayService  = "tempo-platform-gateway"

	// Tenants
	PlatformTenantName   = "platform"
	PlatformTenantID     = "1610b0c3-c509-4592-a256-a1871353dbfa"
	UserTenantName       = "user"
	UserTenantID         = "1610b0c3-c509-4592-a256-a1871353dbfb"

	// OpenTelemetry Collectors
	PlatformCollectorName = "platform-collector"
	UserCollectorName     = "user-collector"

	// Tracing UIPlugin
	TracingUIPluginName   = "tracing"
)
