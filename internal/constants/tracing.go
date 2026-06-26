package constants

const (
	// Namespaces
	TracingNamespace = "openshift-tracing"
	TempoOperatorNS  = "openshift-tempo-operator"
	OTelOperatorNS   = "openshift-opentelemetry-operator"

	// Signal app namespaces
	HotrodNamespace       = "tracing-app-hotrod"
	K6TracingNamespace    = "tracing-app-k6"
	TelemetrygenNamespace = "tracing-app-telemetrygen"

	// Tempo Operator
	TempoOperatorName   = "tempo-product"
	TempoPackageName    = "tempo-product"
	TempoCatalogSource  = "redhat-operators"
	DefaultTempoChannel = "stable"
	TempoOperatorGroup  = "openshift-tempo-operator"

	// OpenTelemetry Operator
	OTelOperatorName   = "opentelemetry-product"
	OTelPackageName    = "opentelemetry-product"
	OTelCatalogSource  = "redhat-operators"
	DefaultOTelChannel = "stable"
	OTelOperatorGroup  = "openshift-opentelemetry-operator"

	// TempoStack
	TempoStackName      = "platform"
	TempoStackSize      = "1Gi"
	TempoGatewayService = "tempo-platform-gateway"

	// Tenants
	PlatformTenantName = "platform"
	PlatformTenantID   = "1610b0c3-c509-4592-a256-a1871353dbfa"
	UserTenantName     = "user"
	UserTenantID       = "1610b0c3-c509-4592-a256-a1871353dbfb"

	// OpenTelemetry Collectors
	PlatformCollectorName       = "platform"
	UserCollectorName           = "user"
	PlatformCollectorDeployment = "platform-collector"
	UserCollectorDeployment     = "user-collector"

	// Tracing UIPlugin — name enforced by CRD validation rule
	TracingUIPluginName = "distributed-tracing"

	// Trace reader/writer RBAC
	TracesReaderPlatformRole = "traces-reader-platform"
	TracesReaderUserRole     = "traces-reader-user"
	PlatformCollectorRole    = "openshift-tracing-platform-collector"
	UserCollectorRole        = "openshift-tracing-user-collector"
	TracesWriterPlatformRole = "traces-writer-platform"
	TracesWriterUserRole     = "traces-writer-user"

	// Signal app images
	HotrodImage       = "jaegertracing/example-hotrod:1.46"
	K6TracingImage    = "ghcr.io/grafana/xk6-client-tracing:v0.0.5"
	TelemetrygenImage = "ghcr.io/open-telemetry/opentelemetry-collector-contrib/telemetrygen:v0.105.0"
)
