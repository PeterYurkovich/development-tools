package config

type InstallMethod string

const (
	InstallMethodFBC         InstallMethod = "fbc"
	InstallMethodBundle      InstallMethod = "bundle"
	InstallMethodMarketplace InstallMethod = "marketplace"
	InstallMethodStage       InstallMethod = "stage"
)

type Config struct {
	COO     COOConfig
	Storage StorageConfig
}

type COOConfig struct {
	Image         string
	InstallMethod InstallMethod
	Plugins       PluginsConfig
}

type PluginsConfig struct {
	Logging              PluginConfig
	Tracing              PluginConfig
	Dashboards           PluginConfig
	Monitoring           PluginConfig
	TroubleshootingPanel PluginConfig
}

type PluginConfig struct {
	Image string
}

type StorageConfig struct {
	Class      string
	UseDefault bool
}
