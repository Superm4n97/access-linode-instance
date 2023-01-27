package linode

import (
	"errors"
	"fmt"
	"github.com/linode/linodego"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"time"
)

const (
	retryCopyLimit = 5
	retrySSHLimit  = 20
	backoffTime    = 5
)

type FileCopyOptions struct {
	LinodeCredentials                            *CredentialOptions
	Address, User, LocalFilePath, RemoteFilePath string
}

func getInstanceIPAddress(instance *linodego.Instance) string {
	for _, ip := range instance.IPv4 {
		if !(ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
			return ip.String()

		}
	}
	return ""
}

func createSCPClient(cpyParam *FileCopyOptions) (*scp.Client, error) {
	//build ssh config
	var sshCfg *ssh.ClientConfig

	//build ssh client config from password
	if cpyParam.LinodeCredentials.Username != "" && cpyParam.LinodeCredentials.Password != "" {
		sshCfg = scp.NewSSHConfigFromPassword(cpyParam.User, cpyParam.LinodeCredentials.Password)
	} else if cpyParam.LinodeCredentials.PrivateKey != nil {
		var err error
		sshCfg, err = scp.NewSSHConfigFromPrivateKey(cpyParam.User, cpyParam.LinodeCredentials.PrivateKey)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unable to create sshConfig from credential")
	}

	// Dial SSH to "my.server.com:22".
	// If your SSH server does not listen on 22, simply suffix the address with port.
	// e.g: "my.server.com:1234"

	var err error
	var existingSSHClient *ssh.Client
	for i := 1; i <= retrySSHLimit; i++ {
		existingSSHClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", cpyParam.Address), sshCfg)
		if err != nil {
			klog.Info("wait for ssh ", i)
			time.Sleep(backoffTime * time.Second)
		} else {
			err = nil
			klog.Infof("connected to ssh")
			break
		}
	}
	if err != nil {
		return nil, err
	}

	//create scp client
	scpClient, err := scp.NewClientFromExistingSSH(existingSSHClient, &scp.ClientOption{})
	if err != nil {
		return nil, err
	}

	return scpClient, nil
}

func copyToRemote(cpyParam *FileCopyOptions) error {
	scpClient, err := createSCPClient(cpyParam)
	if err != nil {
		return err
	}
	defer scpClient.Close()

	if err = scpClient.CopyFileToRemote(cpyParam.LocalFilePath, cpyParam.RemoteFilePath, &scp.FileTransferOption{}); err != nil {
		return err
	}
	return nil
}

func CopyFileToRemote(instance *linodego.Instance, credential *CredentialOptions, localFilePath, remoteFilePath string) error {

	//get the instance address
	addr := getInstanceIPAddress(instance)
	if addr == "" {
		return errors.New("failed to detect IP for Linode instance")
	}

	cpyOps := &FileCopyOptions{
		LinodeCredentials: credential,
		Address:           addr,
		User:              "root",
		LocalFilePath:     localFilePath,
		RemoteFilePath:    remoteFilePath,
	}

	for i := 1; i <= retryCopyLimit; i++ {
		klog.Infof("coping file to remote attempt: ", i)

		//You need to register your SSH key to linode, which will be used as private key
		if err := copyToRemote(cpyOps); err != nil {
			if i == retryCopyLimit {
				return err
			}

		}
		break
	}

	return nil
}
