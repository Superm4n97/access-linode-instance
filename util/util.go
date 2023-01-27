package util

import (
	"fmt"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"os"
	"time"
)

const (
	maxRetries      int           = 1000
	backoffTimeSecs time.Duration = 10
)

func runSCP(addr, privateKey, username string) error {
	// Build a SSH config from username/password
	// sshConf := scp.NewSSHConfigFromPassword("username", "password")

	// Build a SSH config from private key
	privPEM, err := os.ReadFile(privateKey) // "/path/to/privateKey"
	if err != nil {
		return err
	}
	// without passphrase
	sshConf, err := scp.NewSSHConfigFromPrivateKey(username, privPEM)
	if err != nil {
		return err
	}
	// with passphrase
	// sshConf, err := scp.NewSSHConfigFromPrivateKey("username", privPEM, passphrase)

	// Dial SSH to "my.server.com:22".
	// If your SSH server does not listen on 22, simply suffix the address with port.
	// e.g: "my.server.com:1234"

	var existingSSHClient *ssh.Client
	for i := 0; i < maxRetries; i++ {
		existingSSHClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", addr), sshConf)
		if err != nil {
			fmt.Println("wait for ssh", i)
			time.Sleep(backoffTimeSecs * time.Second)
		} else {
			err = nil
			fmt.Println("connected to ssh")
			break
		}
	}
	if err != nil {
		return err
	}

	//scpClient, err := scp.NewClient(addr, sshConf, &scp.ClientOption{})
	//if err != nil {
	//	return err
	//}
	//

	// Build a SCP client based on existing "golang.org/x/crypto/ssh.Client"
	scpClient, err := scp.NewClientFromExistingSSH(existingSSHClient, &scp.ClientOption{})
	if err != nil {
		return err
	}
	defer scpClient.Close()

	_, err = ExecuteTCPCommand(existingSSHClient, "ls -l", sshConf)
	if err != nil {
		return err
	}
	// fmt.Println(out)

	//// Do the file transfer without timeout/context
	//err = scpClient.CopyFileToRemote("/path/to/local/file", "/path/at/remote", &scp.FileTransferOption{})

	// Do the file copy with timeout, context and file properties preserved.
	// Note that the context and timeout will both take effect.

	//used for specifying the timeout
	//fo := &scp.FileTransferOption{
	//	Context:      context.TODO(),
	//	Timeout:      30 * time.Second,
	//	PreserveProp: true,
	//}
	//err = scpClient.CopyFileFromRemote("/root/stackscript.log", "/tmp/stackscript.log", fo)

	return nil

	// scp: /root/success2.txt: No such file or directory
}

func ExecuteTCPCommand(conn *ssh.Client, command string, config *ssh.ClientConfig) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	session.Stdout = DefaultWriter
	session.Stderr = DefaultWriter
	session.Stdin = os.Stdin
	if config.User != "root" {
		command = fmt.Sprintf("sudo %s", command)
	}
	_ = session.Run(command)
	output := DefaultWriter.Output()
	_ = session.Close()
	return output, nil
}

var DefaultWriter = &StringWriter{
	data: make([]byte, 0),
}

type StringWriter struct {
	data []byte
}

func (s *StringWriter) Flush() {
	s.data = make([]byte, 0)
}

func (s *StringWriter) Output() string {
	return string(s.data)
}

func (s *StringWriter) Write(b []byte) (int, error) {
	klog.Infoln("$ ", string(b))
	s.data = append(s.data, b...)
	return len(b), nil
}
