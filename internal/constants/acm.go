package constants

const (
	ACMNamespace              = "open-cluster-management"
	ACMObservabilityNamespace = "open-cluster-management-observability"

	ACMOperatorName  = "advanced-cluster-management"
	ACMOperatorGroup = "og-global"
	ACMCatalogSource = "redhat-operators"
	ACMDefaultChannel = "release-2.17"

	ACMMultiClusterHubName = "multiclusterhub"
	ACMMCOName             = "observability"

	ACMMinIOImage      = "quay.io/minio/minio:RELEASE.2021-08-25T00-41-18Z"
	ACMMinIOBucket     = "thanos"
	ACMMinIOSecretName = "thanos-object-storage"
	ACMMinIOSecretKey  = "thanos.yaml"
)
