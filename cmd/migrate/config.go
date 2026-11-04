package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/exc-works/migrate/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	var options config.TemplateOptions
	var shardCount int
	var format string
	var force bool
	command := &cobra.Command{
		Use:   "config [filename]",
		Short: "Create a commented YAML or plain JSON configuration",
		Example: "  migrate new config\n" +
			"  migrate new config --multi --shards 4\n" +
			"  migrate new config --multi --ssh-tunnel --proxy socks5\n" +
			"  migrate new config --ssh-tunnel --ssh-auth private_key\n" +
			"  migrate new config --format json",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTemplateFlags(cmd, options); err != nil {
				return err
			}
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			explicitFormat := ""
			if cmd.Flags().Changed("format") {
				explicitFormat = format
			}
			name, selectedFormat, err := config.TemplateDestination(name, explicitFormat)
			if err != nil {
				return err
			}
			templateOptions := options
			if templateOptions.Multi {
				templateOptions.ShardCount = shardCount
			}
			cfg, err := config.BuildTemplate(templateOptions)
			if err != nil {
				return err
			}
			data, err := config.RenderTemplate(cfg, selectedFormat)
			if err != nil {
				return err
			}
			flags := os.O_CREATE | os.O_WRONLY | os.O_EXCL
			if force {
				flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
			}
			file, err := os.OpenFile(name, flags, 0o644)
			if err != nil {
				if os.IsExist(err) {
					return errors.New("config file already exists (use --force to overwrite)")
				}
				return err
			}
			_, err = file.Write(data)
			return errors.Join(err, file.Close())
		},
	}
	flags := command.Flags()
	flags.BoolVar(&options.Multi, "multi", false, "generate named shards sharing migration settings")
	flags.IntVar(&shardCount, "shards", config.DefaultShardCount, "number of shards to generate (requires --multi)")
	flags.StringVar(&format, "format", config.FormatYAML, "config format: yaml|json (otherwise inferred from filename)")
	flags.BoolVar(&options.SSH, "ssh-tunnel", false, "generate shared SSH settings (default auth: agent)")
	flags.StringVar(&options.SSHHost, "ssh-host", "", "SSH server host (default: environment placeholder)")
	flags.IntVar(&options.SSHPort, "ssh-port", 0, "SSH server port (default when enabled: 22)")
	flags.StringVar(&options.SSHUser, "ssh-user", "", "SSH login user (default: environment placeholder)")
	flags.StringVar(&options.SSHAuth, "ssh-auth", "", "SSH auth: agent|private_key|password (default: agent)")
	flags.StringVar(&options.SSHKeyFile, "ssh-key-file", "", "private key path (default: environment placeholder)")
	flags.StringVar(&options.SSHPassphraseEnv, "ssh-passphrase-env", "", "private-key passphrase environment variable NAME")
	flags.StringVar(&options.SSHPasswordEnv, "ssh-password-env", "", "SSH password environment variable NAME (default: SSH_PASSWORD)")
	flags.StringVar(&options.SSHKnownHostsFile, "ssh-known-hosts-file", "", "known_hosts path (default: ~/.ssh/known_hosts)")
	flags.StringVar(&options.ProxyType, "proxy", "", "SSH upstream proxy: socks5|http (requires --ssh-tunnel)")
	flags.StringVar(&options.ProxyAddress, "proxy-address", "", "proxy host:port (default: environment placeholder)")
	flags.StringVar(&options.ProxyUsername, "proxy-username", "", "proxy username; requires --proxy-password-env")
	flags.StringVar(&options.ProxyPasswordEnv, "proxy-password-env", "", "proxy password environment variable NAME; requires --proxy-username")
	flags.BoolVar(&force, "force", false, "overwrite the selected configuration file")
	return command
}

func validateTemplateFlags(cmd *cobra.Command, options config.TemplateOptions) error {
	if cmd.Flags().Changed("shards") {
		if !options.Multi {
			return errors.New("--shards requires --multi")
		}
		count, _ := cmd.Flags().GetInt("shards")
		if count <= 0 {
			return errors.New("--shards must be a positive integer")
		}
	}
	for _, name := range []string{"ssh-host", "ssh-port", "ssh-user", "ssh-auth", "ssh-key-file", "ssh-passphrase-env", "ssh-password-env", "ssh-known-hosts-file", "proxy"} {
		if cmd.Flags().Changed(name) && !options.SSH {
			return fmt.Errorf("--%s requires --ssh-tunnel", name)
		}
	}
	for _, name := range []string{"proxy-address", "proxy-username", "proxy-password-env"} {
		if cmd.Flags().Changed(name) && options.ProxyType == "" {
			return fmt.Errorf("--%s requires --proxy", name)
		}
	}
	if cmd.Flags().Changed("ssh-port") && options.SSHPort <= 0 {
		return errors.New("ssh-port must be between 1 and 65535")
	}
	if cmd.Flags().Changed("ssh-auth") && options.SSHAuth == "" {
		return errors.New("ssh-auth must be agent, private_key, or password")
	}
	if cmd.Flags().Changed("proxy") && options.ProxyType == "" {
		return errors.New("proxy must be socks5 or http")
	}
	if cmd.Flags().Changed("format") {
		value, _ := cmd.Flags().GetString("format")
		if value == "" {
			return errors.New("config format must be yaml or json")
		}
	}
	return nil
}
