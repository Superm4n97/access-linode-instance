package main

import (
	"fmt"
	"github.com/Superm4n97/access-linode-instance/linode"
	"k8s.io/klog/v2"
)

func main() {
	instance, err := linode.CreateLinodeInstance()
	if err != nil {
		klog.Infof("failed to create linode instance: %s", err.Error())
		return
	}
	fmt.Println("instance ID: ", instance.ID)
}
