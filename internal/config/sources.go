package config

import (
	"errors"
	"sort"

	"github.com/exc-works/migrate/internal/tunnel"
)

type DataSourceConfig struct {
	DataSourceName string          `json:"data_source_name" yaml:"data_source_name"`
	SSHTunnel      *tunnel.Options `json:"ssh_tunnel,omitempty" yaml:"ssh_tunnel,omitempty"`
}

type Target struct {
	Name   string
	Config FileConfig
}

// Targets expands shared options into independent, sorted configurations.
// Read resolves the effective connection settings before returning a config.
func (cfg FileConfig) Targets() []Target {
	if cfg.DataSources == nil {
		cfg.SSHTunnel = cloneTunnel(cfg.SSHTunnel)
		return []Target{{
			Name:   "default",
			Config: cfg,
		}}
	}
	names := make([]string, 0, len(cfg.DataSources))
	for name := range cfg.DataSources {
		names = append(names, name)
	}
	sort.Strings(names)
	targets := make([]Target, 0, len(names))
	for _, name := range names {
		source := cfg.DataSources[name]
		effective := cfg
		effective.DataSources = nil
		effective.DataSourceName = source.DataSourceName
		effective.SSHTunnel = source.SSHTunnel
		if effective.SSHTunnel == nil {
			effective.SSHTunnel = cfg.SSHTunnel
		}
		effective.SSHTunnel = cloneTunnel(effective.SSHTunnel)
		targets = append(targets, Target{
			Name:   name,
			Config: effective,
		})
	}
	return targets
}

func cloneTunnel(options *tunnel.Options) *tunnel.Options {
	if options == nil {
		return nil
	}
	copy := *options
	if options.Proxy != nil {
		proxy := *options.Proxy
		copy.Proxy = &proxy
	}
	return &copy
}

func resolveConnection(cfg *FileConfig, base string) error {
	dsn, err := ReplaceEnvVar(cfg.DataSourceName)
	if err != nil {
		return err
	}
	if dsn == "" {
		return errors.New("data_source_name must be set")
	}
	cfg.DataSourceName = dsn
	if cfg.SSHTunnel != nil {
		return resolveTunnelConfig(cfg.SSHTunnel, base)
	}
	return nil
}
