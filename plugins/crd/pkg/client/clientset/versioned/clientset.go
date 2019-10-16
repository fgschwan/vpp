// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	"fmt"

	contivppv1 "github.com/contiv/vpp/plugins/crd/pkg/client/clientset/versioned/typed/contivppio/v1"
	nodeconfigv1 "github.com/contiv/vpp/plugins/crd/pkg/client/clientset/versioned/typed/nodeconfig/v1"
	telemetryv1 "github.com/contiv/vpp/plugins/crd/pkg/client/clientset/versioned/typed/telemetry/v1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ContivppV1() contivppv1.ContivppV1Interface
	NodeconfigV1() nodeconfigv1.NodeconfigV1Interface
	TelemetryV1() telemetryv1.TelemetryV1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	contivppV1   *contivppv1.ContivppV1Client
	nodeconfigV1 *nodeconfigv1.NodeconfigV1Client
	telemetryV1  *telemetryv1.TelemetryV1Client
}

// ContivppV1 retrieves the ContivppV1Client
func (c *Clientset) ContivppV1() contivppv1.ContivppV1Interface {
	return c.contivppV1
}

// NodeconfigV1 retrieves the NodeconfigV1Client
func (c *Clientset) NodeconfigV1() nodeconfigv1.NodeconfigV1Interface {
	return c.nodeconfigV1
}

// TelemetryV1 retrieves the TelemetryV1Client
func (c *Clientset) TelemetryV1() telemetryv1.TelemetryV1Interface {
	return c.telemetryV1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("Burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.contivppV1, err = contivppv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.nodeconfigV1, err = nodeconfigv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.telemetryV1, err = telemetryv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.contivppV1 = contivppv1.NewForConfigOrDie(c)
	cs.nodeconfigV1 = nodeconfigv1.NewForConfigOrDie(c)
	cs.telemetryV1 = telemetryv1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.contivppV1 = contivppv1.New(c)
	cs.nodeconfigV1 = nodeconfigv1.New(c)
	cs.telemetryV1 = telemetryv1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
