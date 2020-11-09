package command

import (
	"fmt"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

func SftpCommand(config *cfg.Config, callback func(sftpClient *sftp.Client) error) error {
	var authMethods []ssh.AuthMethod

	if config.Remote.SshPrivateKey != "" {
		key, err := ioutil.ReadFile(config.Remote.SshPrivateKey)
		if err != nil {
			return fmt.Errorf("unable to read ssh private key %s", config.Remote.SshPrivateKey)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("unable to parse ssh private key: %v", err)
		}

		authMethods = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		authMethods = []ssh.AuthMethod{
			ssh.Password(config.Remote.SshPassword),
		}
	}

	sshClientConfig := &ssh.ClientConfig{
		User:            config.Remote.SshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            authMethods,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Remote.SshHost, config.Remote.SshPort), sshClientConfig)
	if err != nil {
		return fmt.Errorf("unable to ssh connect: %v", err)
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("unable to sftp connect: %v", err)
	}
	defer sftpClient.Close()

	err = callback(sftpClient)

	if err != nil {
		return err
	}

	return nil
}
