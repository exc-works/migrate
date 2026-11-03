package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func authentication(ctx context.Context, s Settings, prompt Prompt) (ssh.AuthMethod, io.Closer, error) {
	switch s.Auth.Type {
	case AuthPasswordMode:
		secret, err := Secret(ctx, s.Auth.PasswordEnv, "SSH password", prompt)
		if err != nil {
			return nil, nil, err
		}
		return ssh.Password(string(secret)), nil, nil
	case AuthPrivateKey:
		path, err := ExpandPath(s.Auth.KeyFile, ".")
		if err != nil {
			return nil, nil, err
		}
		key, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, errors.New("cannot read SSH private key")
		}
		signer, err := ssh.ParsePrivateKey(key)
		var missing *ssh.PassphraseMissingError
		if errors.As(err, &missing) {
			secret, secretErr := Secret(ctx, s.Auth.PassphraseEnv, "SSH key passphrase", prompt)
			if secretErr != nil {
				return nil, nil, secretErr
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, secret)
		}
		if err != nil {
			return nil, nil, errors.New("cannot parse or decrypt SSH private key")
		}
		return ssh.PublicKeys(signer), nil, nil
	case AuthAgent:
		socket := os.Getenv("SSH_AUTH_SOCK")
		if socket == "" {
			return nil, nil, errors.New("SSH_AUTH_SOCK is required for agent authentication")
		}
		d := net.Dialer{}
		agentCtx, cancel := context.WithTimeout(ctx, s.Connect)
		defer cancel()
		conn, err := d.DialContext(agentCtx, "unix", socket)
		if err != nil {
			return nil, nil, errors.New("cannot connect to SSH agent")
		}
		return ssh.PublicKeysCallback(agent.NewClient(conn).Signers), conn, nil
	default:
		return nil, nil, errors.New("unsupported SSH authentication")
	}
}
