// Copyright (c) 2019 Bell Canada, Pantheon Technologies and/or its affiliates.
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

// Package sr implements plugin that watches configuration of K8s resources
// for changes in pods and updates the Segment Routing configuration
// in the VPP accordingly.
package sr

import (
	"fmt"
	"github.com/go-errors/errors"
	"net"

	controller "github.com/contiv/vpp/plugins/controller/api"
	"github.com/contiv/vpp/plugins/ipam"
	podmodel "github.com/contiv/vpp/plugins/ksr/model/pod"
	"github.com/contiv/vpp/plugins/podmanager"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/vpp-agent/api/models/vpp/srv6"
	"github.com/ligato/vpp-agent/pkg/models"
)

// Plugin watches configuration of K8s resources for changes in pods and
// updates the Segment Routing configuration in the VPP accordingly.
type Plugin struct {
	Deps

	state State
}

// Deps groups the dependencies of the Plugin.
type Deps struct {
	infra.PluginDeps
	IPAM       ipam.API
	PodManager podmanager.API
}

// State contains all SR-related objects (LocalSIDs, Policy, ...) that will be/are created for all local pods
type State map[podmodel.ID]*PodState

// PodState contains all SR-related objects (LocalSIDs, Policy, ...) that will be/are created for one local pod
type PodState struct {
	LocalSID *vpp_srv6.LocalSID
}

// Init initialize inner state of Plugin
func (p *Plugin) Init() (err error) {
	p.state = make(State)
	return nil
}

// HandlesEvent is used by Controller to check if the event is being handled
// by this handler.
func (p *Plugin) HandlesEvent(event controller.Event) bool {
	if event.Method() != controller.Update {
		// select any resync event
		return true
	}
	if _, isAddPod := event.(*podmanager.AddPod); isAddPod {
		return true
	}
	if _, isDeletePod := event.(*podmanager.DeletePod); isDeletePod {
		return true
	}
	// unhandled event
	return false
}

// Resync is called by Controller to handle event that requires full
// re-synchronization.
// For startup resync, resyncCount is 1. Higher counter values identify
// run-time resync.
func (p *Plugin) Resync(event controller.Event, kubeStateData controller.KubeStateData, resyncCount int, txn controller.ResyncOperations) error {
	// Normally it should not be needed to resync the SR state in the run-time
	// - local pod should not be added/deleted without the agent knowing about
	// it. But if we are healing after an error, recompute the desired SR state
	// and apply it vswitch correctly.
	_, isHealingResync := event.(*controller.HealingResync)
	if !(resyncCount == 1 || isHealingResync) {
		return nil
	}

	// compute new state and reconfigure vswitch using transaction txn
	// TODO txn.Delete doesn't exist -> (i.e. in case of healing resync) is not possible to remove previously added SR configuration?
	//for _, state := range p.state {
	//	txn.Delete(models.Key(state.LocalSID))
	//}
	p.state = make(State)
	for _, pod := range p.PodManager.GetLocalPods() {
		p.handleAddPod(&pod.ID, txn)
	}
	p.Log.Debugf("SR state after resync: %+v", p.state)

	return nil
}

// Update is called by Controller to handle event that can be reacted to by
// an incremental change.
// <changeDescription> should be human-readable description of changes that
// have to be performed (via txn or internally) - can be empty.
func (p *Plugin) Update(event controller.Event, txn controller.UpdateOperations) (changeDescription string, err error) {
	if addPod, isAddPod := event.(*podmanager.AddPod); isAddPod {
		return p.handleAddPod(&addPod.Pod, txn)
	}
	if deletePod, isDeletePod := event.(*podmanager.DeletePod); isDeletePod {
		return p.handleDeletePod(deletePod, txn)
	}
	return "", nil
}

func (p *Plugin) handleAddPod(podID *podmodel.ID, txn controller.ResyncOperations) (changeDescription string, err error) {
	localSIDValue, err := p.createLocalSID(podID)
	if err != nil {
		return "", fmt.Errorf("can't create localSID for pod with id  %+v due to: %v", podID, err)
	}
	// remember changes in state
	p.state[*podID] = &PodState{
		LocalSID: localSIDValue,
	}

	// putting new LocalSID into vpp-agent transaction
	txn.Put(models.Key(localSIDValue), localSIDValue)
	return fmt.Sprintf("adding LocalSid %+v for created pod %v", localSIDValue, podID), nil
}

func (p *Plugin) handleDeletePod(deletePod *podmanager.DeletePod, txn controller.UpdateOperations) (changeDescription string, err error) {
	sid, err := p.computeSid(&deletePod.Pod)
	if err != nil {
		return "", fmt.Errorf("can't delete localSID for pod with id  %+v due to: %v", deletePod.Pod, err)
	}

	// remember changes in state
	delete(p.state, deletePod.Pod)

	// key construction and delete
	localSIDValue := &vpp_srv6.LocalSID{ // sid is enough for key construction
		Sid: sid.String(),
	}
	txn.Delete(models.Key(localSIDValue))
	return fmt.Sprintf("removing LocalSid %v for created pod %v", sid.String(), deletePod.Pod.String()), nil
}

func (p *Plugin) createLocalSID(podID *podmodel.ID) (*vpp_srv6.LocalSID, error) {
	sid, err := p.computeSid(podID)
	if err != nil {
		return nil, fmt.Errorf("can't compute SRv6 SID due to: %v", err)
	}
	// constructing LocalSID
	localSIDValue := &vpp_srv6.LocalSID{
		Sid:        sid.String(),
		FibTableId: 0,
		EndFunction: &vpp_srv6.LocalSID_BaseEndFunction{BaseEndFunction: &vpp_srv6.LocalSID_End{
			Psp: false,
		}},
	}
	return localSIDValue, nil
}

// Revert is used to remove the SR pod state for given pod for which AddPod event fails.
func (p *Plugin) Revert(event controller.Event) error {
	if addPod, isAddPod := event.(*podmanager.AddPod); isAddPod {
		delete(p.state, addPod.Pod) // delete of internal state is enough because values put into transaction in update won't get written into vpp (transaction is thrown away)
	}
	return nil
}

func (p *Plugin) computeSid(pod *podmodel.ID) (*net.IP, error) {
	// applying other netmask-part for pod IP address and using it as LocalSID's SID (segment ID)
	localSIDBaseIP := net.ParseIP("5555::").To16()
	podIPNet := p.IPAM.GetPodIP(*pod)
	podIPMask := podIPNet.Mask
	if _, bitLength := podIPMask.Size(); bitLength != 16*8 { //not ipv6
		return nil, errors.Errorf("can't get sid for pod IP %+v (pod ID %+v)", podIPNet, *pod)
	}
	podIP := podIPNet.IP.To16()

	sid := net.IP(make([]byte, 16))
	for i := range podIP {
		sid[i] = podIP[i] & ^podIPMask[i] | localSIDBaseIP[i]
	}
	return &sid, nil
}
