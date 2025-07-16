// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package utils

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// GetK8sClient, provides Kubernetes client handle, for
// k8s cluster corresponding to the environment variable
// KUBECONFIG set while calling this function, or
// otherwise it will be assuming to be running as part
// of a k8s cluster and form the k8s client based on the
// role assigned to the pod
func GetK8sClient() (client.Client, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("issue with kubeconfig: %s", err)
	}
	cl, err := client.New(conf, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}
	return cl, nil
}
