package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/exc-works/migrate/internal/tunnel"
	"go.yaml.in/yaml/v3"
)

type TemplateOptions struct {
	Multi             bool
	SSH               bool
	SSHHost           string
	SSHPort           int
	SSHUser           string
	SSHAuth           string
	SSHKeyFile        string
	SSHPassphraseEnv  string
	SSHPasswordEnv    string
	SSHKnownHostsFile string
	ProxyType         string
	ProxyAddress      string
	ProxyUsername     string
	ProxyPasswordEnv  string
}

// TemplateDestination infers format without reading an existing configuration.
func TemplateDestination(name, format string) (string, string, error) {
	if format != "" && format != FormatYAML && format != FormatJSON {
		return "", "", errors.New("config format must be yaml or json")
	}
	if name == "" {
		if format == FormatJSON {
			return DefaultJSONFilename, FormatJSON, nil
		}
		return DefaultYAMLFilename, FormatYAML, nil
	}
	inferred := FormatJSON
	if ext := strings.ToLower(filepath.Ext(name)); ext == ExtensionYML || ext == ExtensionYAML {
		inferred = FormatYAML
	}
	if format != "" && format != inferred {
		return "", "", errors.New("config format conflicts with the filename extension")
	}
	return name, inferred, nil
}

// BuildTemplate creates placeholders only; it does not resolve the environment,
// inspect keys, request credentials, or make network connections.
func BuildTemplate(options TemplateOptions) (FileConfig, error) {
	cfg := DefaultTemplate()
	cfg.DataSourceName = "${DATABASE_DSN}"
	if options.Multi {
		cfg.DataSourceName = ""
		cfg.DataSources = map[string]DataSourceConfig{
			"tenant_a": {
				DataSourceName: "${TENANT_A_DSN}",
			},
			"tenant_b": {
				DataSourceName: "${TENANT_B_DSN}",
			},
		}
	}
	if !options.SSH {
		if options != (TemplateOptions{Multi: options.Multi}) {
			return cfg, errors.New("SSH and proxy options require --ssh-tunnel")
		}
		return cfg, nil
	}
	if options.SSHAuth == "" {
		options.SSHAuth = tunnel.AuthAgent
	}
	ssh := &tunnel.Options{
		Host:           templateValue(options.SSHHost, "${SSH_HOST}"),
		Port:           options.SSHPort,
		User:           templateValue(options.SSHUser, "${SSH_USER}"),
		KnownHostsFile: options.SSHKnownHostsFile,
		Auth: tunnel.AuthOptions{
			Type:          options.SSHAuth,
			KeyFile:       options.SSHKeyFile,
			PassphraseEnv: options.SSHPassphraseEnv,
			PasswordEnv:   options.SSHPasswordEnv,
		},
	}
	if options.SSHAuth == tunnel.AuthPrivateKey {
		ssh.Auth.KeyFile = templateValue(ssh.Auth.KeyFile, "${SSH_KEY_FILE}")
	}
	if options.SSHAuth == tunnel.AuthPasswordMode {
		ssh.Auth.PasswordEnv = templateValue(ssh.Auth.PasswordEnv, "SSH_PASSWORD")
	}
	if options.ProxyType != "" {
		ssh.Proxy = &tunnel.ProxyOptions{
			Type:        options.ProxyType,
			Address:     templateValue(options.ProxyAddress, "${SSH_PROXY_ADDRESS}"),
			Username:    options.ProxyUsername,
			PasswordEnv: options.ProxyPasswordEnv,
		}
	} else if options.ProxyAddress != "" || options.ProxyUsername != "" || options.ProxyPasswordEnv != "" {
		return cfg, errors.New("proxy options require --proxy")
	}
	// Validate structure without trying to expand placeholder addresses or paths.
	validation := cloneTunnel(ssh)
	if validation.Proxy != nil && validation.Proxy.Address == "${SSH_PROXY_ADDRESS}" {
		validation.Proxy.Address = "proxy.example:1080"
	}
	settings, err := validation.Normalize()
	if err != nil {
		return cfg, err
	}
	ssh.Port = settings.Port
	ssh.KnownHostsFile = settings.KnownHostsFile
	cfg.SSHTunnel = ssh
	return cfg, nil
}

func templateValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func RenderTemplate(cfg FileConfig, format string) ([]byte, error) {
	if format == FormatJSON {
		data, err := json.MarshalIndent(cfg, "", "  ")
		return append(data, '\n'), err
	}
	if format != FormatYAML {
		return nil, errors.New("config format must be yaml or json")
	}
	var root yaml.Node
	if err := root.Encode(cfg); err != nil {
		return nil, err
	}
	root.HeadComment = "Migration configuration. Set referenced environment variables before running migrations.\nInitialize variants: migrate new config --multi --ssh-tunnel --proxy socks5\nUse --ssh-auth private_key or --ssh-auth password for other authentication modes."
	root.FootComment = "Script defaults: script_timeout: 5m, script_max_rows: 1000, script_max_bytes: 16777216.\nSSH defaults: connect_timeout: 10s, keepalive_interval: 15s, keepalive_timeout: 5s."
	annotateTemplate(&root)
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func annotateTemplate(node *yaml.Node) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	comments := map[string]string{
		"data_sources":         "All databases use the shared migration settings. Each database keeps its own history.",
		"data_source_name":     "Driver DSN. Set the referenced environment variable; do not store credentials here.",
		"ssh_tunnel":           "Shared SSH defaults. A data-source SSH object replaces this whole object.",
		"schema_name":          "Migration history table name (default: migration_schema).",
		"dialect":              "Database dialect (default: postgres).",
		"migration_source":     "Shared migration directory (default: migrations), relative to working_directory.",
		"working_directory":    "Empty means the current directory; --working-dir overrides this setting.",
		"migrate_out_of_order": "Allow pending versions older than the latest applied version (default: false).",
		"logger_level":         "Log level (default: info).",
		"host":                 "SSH server host; port defaults to 22.",
		"user":                 "SSH login user.",
		"auth":                 "SSH auth: agent, private_key, or password. Only the selected mode's fields are generated.",
		"key_file":             "Private key path; relative paths are resolved from this configuration file.",
		"passphrase_env":       "Environment variable NAME containing the private-key passphrase.",
		"password_env":         "Environment variable NAME containing the password; never put the password itself here.",
		"known_hosts_file":     "Verify SSH host keys. Unknown keys require explicit interactive approval; noninteractive mode rejects them.",
		"proxy":                "Optional proxy used to reach the SSH server: socks5 or http (CONNECT). No direct fallback.",
		"address":              "Proxy host:port; username and password_env must be supplied together when authentication is needed.",
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		key.HeadComment = comments[key.Value]
		annotateTemplate(node.Content[i+1])
	}
}
