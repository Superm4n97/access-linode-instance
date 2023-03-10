package linode

import (
	"context"
	"github.com/linode/linodego"
	_ "github.com/povsister/scp"
	_ "golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	"gomodules.xyz/homedir"
	passgen "gomodules.xyz/password-generator"
	"gomodules.xyz/pointer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	RetryInterval = 10 * time.Second
	RetryTimeout  = 10 * time.Minute

	instanceName  string = "raselTest"
	linodeRegion  string = "us-central"
	linodeMachine string = "g6-standard-2"
	instanceImage string = "linode/ubuntu22.04"
)

type CredentialOptions struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey []byte `json:"privateKey"`
	CLIToken   string `json:"cliToken"`
}

func NewCredentialOptions() *CredentialOptions {
	ret := CredentialOptions{}
	ret.Username = os.Getenv("USERNAME")
	ret.Password = os.Getenv("PASSWORD")
	pk := os.Getenv("SSH_PATH")
	if pk == "" {
		pk = filepath.Join(homedir.HomeDir(), ".ssh/id_rsa")
	}
	ret.PrivateKey, _ = os.ReadFile(pk)

	return &ret
}

func NewClient() *linodego.Client {
	token := os.Getenv("LINODE_CLI_TOKEN")
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	c := linodego.NewClient(oauth2Client)
	return &c
}

func CreateLinodeInstance() (*linodego.Instance, error) {
	c := NewClient()

	sshKeys, err := c.ListSSHKeys(context.Background(), &linodego.ListOptions{})
	if err != nil {
		return nil, err
	}
	authorizedKeys := make([]string, 0, len(sshKeys))
	for _, r := range sshKeys {
		authorizedKeys = append(authorizedKeys, r.SSHKey)
	}

	rootPass := passgen.Generate(20)
	klog.Infof("password: %s", rootPass)
	createOpt := linodego.InstanceCreateOptions{
		Region:         linodeRegion,
		Type:           linodeMachine,
		Label:          instanceName,
		RootPass:       rootPass,
		AuthorizedKeys: authorizedKeys,
		Image:          instanceImage,
		BackupsEnabled: false,
		SwapSize:       pointer.IntP(0),
	}

	inst, err := c.CreateInstance(context.Background(), createOpt)
	if err != nil {
		return nil, err
	}

	if err := waitForStatus(c, inst.ID, linodego.InstanceRunning); err != nil {
		return nil, err
	}

	klog.Infof("%s machine created", instanceName)

	return inst, nil
}

func waitForStatus(c *linodego.Client, id int, status linodego.InstanceStatus) error {
	attempt := 0
	klog.Infoln("waiting for instance status", "status", status)
	return wait.PollImmediate(RetryInterval, RetryTimeout, func() (bool, error) {
		attempt++

		instance, err := c.GetInstance(context.Background(), id)
		if err != nil {
			return false, nil
		}
		if instance == nil {
			return false, nil
		}
		klog.Infoln("current instance state", "instance", instance.Label, "status", instance.Status, "attempt", attempt)
		if instance.Status == status {
			klog.Infoln("current instance status", "status", status)
			return true, nil
		}
		return false, nil
	})
}

func GetInstanceFromID(instanceID int) (*linodego.Instance, error) {
	client := NewClient()
	instance, err := client.GetInstance(context.Background(), instanceID)
	if err != nil {
		return nil, err
	}
	return instance, nil
}
