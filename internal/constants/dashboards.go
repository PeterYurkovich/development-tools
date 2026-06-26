package constants

const (
	PersesDefaultNamespace = "perses-dev"

	ThanosDatasourceName = "thanos-querier-datasource"
	LokiDatasourceName   = "loki-datasource"
	TempoDatasourceName  = "tempo-platform"

	ThanosQuerierURL = "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091"
	LokiGatewayURL   = "https://logging-loki-gateway-http.openshift-logging.svc.cluster.local:8080/api/logs/v1/application"
	TempoGatewayURL  = "https://tempo-platform-gateway.openshift-tracing.svc.cluster.local:8080/api/traces/v1/platform/tempo"

	ThanosDatasourceSecretName = "thanos-querier-datasource-secret"
	LokiDatasourceSecretName   = "loki-datasource-secret"
	TempoDatasourceSecretName  = "tempo-platform-secret"
)
