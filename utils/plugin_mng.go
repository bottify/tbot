package utils

type PluginManager struct{}
type Plugin interface {
	Setup(cfg *Config) error
}

func (*PluginManager) Register(p *Plugin) {

}
