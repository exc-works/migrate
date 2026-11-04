// Package tunnel owns explicit SSH and proxy transport settings.
package tunnel

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	AuthPasswordMode = "password"
	AuthPrivateKey   = "private_key"
	AuthAgent        = "agent"
	ProxySOCKS5      = "socks5"
	ProxyHTTP        = "http"
)

type AuthOptions struct {
	Type          string `json:"type" yaml:"type"`
	PasswordEnv   string `json:"password_env,omitempty" yaml:"password_env,omitempty"`
	KeyFile       string `json:"key_file,omitempty" yaml:"key_file,omitempty"`
	PassphraseEnv string `json:"passphrase_env,omitempty" yaml:"passphrase_env,omitempty"`
}

type ProxyOptions struct {
	Type        string `json:"type" yaml:"type"`
	Address     string `json:"address" yaml:"address"`
	Username    string `json:"username,omitempty" yaml:"username,omitempty"`
	PasswordEnv string `json:"password_env,omitempty" yaml:"password_env,omitempty"`
}

type Options struct {
	Host              string        `json:"host" yaml:"host"`
	Port              int           `json:"port,omitempty" yaml:"port,omitempty"`
	User              string        `json:"user" yaml:"user"`
	Auth              AuthOptions   `json:"auth" yaml:"auth"`
	KnownHostsFile    string        `json:"known_hosts_file,omitempty" yaml:"known_hosts_file,omitempty"`
	Proxy             *ProxyOptions `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	ConnectTimeout    string        `json:"connect_timeout,omitempty" yaml:"connect_timeout,omitempty"`
	KeepaliveInterval string        `json:"keepalive_interval,omitempty" yaml:"keepalive_interval,omitempty"`
	KeepaliveTimeout  string        `json:"keepalive_timeout,omitempty" yaml:"keepalive_timeout,omitempty"`
}

type Settings struct {
	Options
	Connect       time.Duration
	Keepalive     time.Duration
	KeepaliveWait time.Duration
}

var envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (o Options) Normalize() (Settings, error) {
	if strings.TrimSpace(o.Host) == "" || strings.ContainsAny(o.Host, " \t\r\n/") || strings.TrimSpace(o.User) == "" {
		return Settings{}, errors.New("ssh_tunnel requires a valid host and user")
	}
	if o.Port == 0 {
		o.Port = 22
	}
	if o.Port < 1 || o.Port > 65535 {
		return Settings{}, errors.New("invalid ssh_tunnel port")
	}
	if o.KnownHostsFile == "" {
		o.KnownHostsFile = "~/.ssh/known_hosts"
	}
	for _, name := range []string{o.Auth.PasswordEnv, o.Auth.PassphraseEnv} {
		if name != "" && !envName.MatchString(name) {
			return Settings{}, errors.New("invalid SSH secret environment variable name")
		}
	}
	switch o.Auth.Type {
	case AuthPasswordMode:
		if o.Auth.KeyFile != "" || o.Auth.PassphraseEnv != "" {
			return Settings{}, errors.New("password auth cannot specify a private key")
		}
	case AuthPrivateKey:
		if o.Auth.KeyFile == "" || o.Auth.PasswordEnv != "" {
			return Settings{}, errors.New("private_key auth requires key_file and no password_env")
		}
	case AuthAgent:
		if o.Auth.KeyFile != "" || o.Auth.PasswordEnv != "" || o.Auth.PassphraseEnv != "" {
			return Settings{}, errors.New("agent auth cannot specify key or password fields")
		}
	default:
		return Settings{}, errors.New("ssh_tunnel auth type must be password, private_key or agent")
	}
	if p := o.Proxy; p != nil {
		if p.Type != ProxySOCKS5 && p.Type != ProxyHTTP {
			return Settings{}, errors.New("proxy type must be socks5 or http")
		}
		host, port, err := net.SplitHostPort(p.Address)
		n, parseErr := strconv.Atoi(port)
		if err != nil || host == "" || strings.ContainsAny(host, " \t\r\n/") || parseErr != nil || n < 1 || n > 65535 {
			return Settings{}, errors.New("proxy address must be host:port")
		}
		if (p.Username == "") != (p.PasswordEnv == "") || (p.PasswordEnv != "" && !envName.MatchString(p.PasswordEnv)) {
			return Settings{}, errors.New("proxy authentication requires username and a valid password_env")
		}
		if strings.ContainsAny(p.Username, ":\r\n") {
			return Settings{}, errors.New("invalid proxy username")
		}
	}
	connect, err := positiveDuration(o.ConnectTimeout, 10*time.Second)
	if err != nil {
		return Settings{}, errors.New("invalid ssh_tunnel connect_timeout")
	}
	interval, err := positiveDuration(o.KeepaliveInterval, 15*time.Second)
	if err != nil {
		return Settings{}, errors.New("invalid ssh_tunnel keepalive_interval")
	}
	wait, err := positiveDuration(o.KeepaliveTimeout, 5*time.Second)
	if err != nil {
		return Settings{}, errors.New("invalid ssh_tunnel keepalive_timeout")
	}
	return Settings{
		Options:       o,
		Connect:       connect,
		Keepalive:     interval,
		KeepaliveWait: wait,
	}, nil
}

func positiveDuration(value string, fallback time.Duration) (time.Duration, error) {
	if value == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return 0, errors.New("duration must be positive")
	}
	return d, nil
}

// ExpandPath resolves current-user home syntax and config-relative SSH paths.
func ExpandPath(value, base string) (string, error) {
	if strings.HasPrefix(value, "~") {
		if value != "~" && !strings.HasPrefix(value, "~/") {
			return "", errors.New("only current-user tilde paths are supported")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("cannot resolve home directory for SSH path")
		}
		value = filepath.Join(home, strings.TrimPrefix(strings.TrimPrefix(value, "~"), "/"))
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(base, value)
	}
	return value, nil
}

// Prompt is opt-in: library callers that supply nil never read a terminal.
type Prompt func(context.Context, string) ([]byte, error)

func Secret(ctx context.Context, variable, label string, prompt Prompt) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if variable != "" {
		if !envName.MatchString(variable) {
			return nil, errors.New("invalid secret environment variable name")
		}
		if value, ok := os.LookupEnv(variable); ok {
			if value == "" {
				return nil, errors.New("secret environment variable " + variable + " is empty")
			}
			return []byte(value), nil
		}
	}
	if prompt == nil {
		return nil, errors.New("secret unavailable; set the configured password/passphrase environment variable or enable terminal input")
	}
	value, err := prompt(ctx, label)
	if err != nil {
		return nil, errors.New("secret input canceled or unavailable")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, errors.New("secret input is empty")
	}
	return value, nil
}
