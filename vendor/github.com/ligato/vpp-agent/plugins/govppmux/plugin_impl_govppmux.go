// Copyright (c) 2017 Cisco and/or its affiliates.
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

package govppmux

import (
	"context"
	"errors"
	"sync"
	"time"

	"git.fd.io/govpp.git/adapter"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"

	govpp "git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

func init() {
	govpp.SetControlPingMessages(&vpe.ControlPing{}, &vpe.ControlPingReply{})
}

// GOVPPPlugin implements the govppmux plugin interface.
type GOVPPPlugin struct {
	Deps // Inject.

	vppConn    *govpp.Connection
	vppAdapter adapter.VppAdapter
	vppConChan chan govpp.ConnectionEvent

	replyTimeout    time.Duration
	reconnectResync bool
	lastConnErr     error

	// Cancel can be used to cancel all goroutines and their jobs inside of the plugin.
	cancel context.CancelFunc
	// Wait group allows to wait until all goroutines of the plugin have finished.
	wg sync.WaitGroup
}

// Deps groups injected dependencies of plugin
// so that they do not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps // inject
	Resync                *resync.Plugin
}

// Config groups the configurable parameter of GoVpp.
type Config struct {
	HealthCheckProbeInterval time.Duration `json:"health-check-probe-interval"`
	HealthCheckReplyTimeout  time.Duration `json:"health-check-reply-timeout"`
	HealthCheckThreshold     int           `json:"health-check-threshold"`
	ReplyTimeout             time.Duration `json:"reply-timeout"`
	// The prefix prepended to the name used for shared memory (SHM) segments. If not set,
	// shared memory segments are created directly in the SHM directory /dev/shm.
	ShmPrefix       string `json:"shm-prefix"`
	ReconnectResync bool   `json:"resync-after-reconnect"`
}

func defaultConfig() Config {
	return Config{
		HealthCheckProbeInterval: time.Second,
		HealthCheckReplyTimeout:  100 * time.Millisecond,
		HealthCheckThreshold:     1,
		ReplyTimeout:             time.Second,
	}
}

// FromExistingAdapter is used mainly for testing purposes.
func FromExistingAdapter(vppAdapter adapter.VppAdapter) *GOVPPPlugin {
	ret := &GOVPPPlugin{
		vppAdapter: vppAdapter,
	}
	return ret
}

// Init is the entry point called by Agent Core. A single binary-API connection to VPP is established.
func (plugin *GOVPPPlugin) Init() error {
	var err error

	govppLogger := plugin.Deps.Log.NewLogger("GoVpp")
	if govppLogger, ok := govppLogger.(*logrus.Logger); ok {
		govppLogger.SetLevel(logging.InfoLevel)
		govpp.SetLogger(govppLogger.StandardLogger())
	}

	plugin.PluginName = plugin.Deps.PluginName

	cfg := defaultConfig()
	found, err := plugin.PluginConfig.GetValue(&cfg)
	if err != nil {
		return err
	}
	var shmPrefix string
	if found {
		govpp.SetHealthCheckProbeInterval(cfg.HealthCheckProbeInterval)
		govpp.SetHealthCheckReplyTimeout(cfg.HealthCheckReplyTimeout)
		govpp.SetHealthCheckThreshold(cfg.HealthCheckThreshold)
		plugin.replyTimeout = cfg.ReplyTimeout
		plugin.reconnectResync = cfg.ReconnectResync
		shmPrefix = cfg.ShmPrefix
		plugin.Log.Debug("Setting govpp parameters", cfg)
	}

	if plugin.vppAdapter == nil {
		plugin.vppAdapter = NewVppAdapter(shmPrefix)
	} else {
		plugin.Log.Info("Reusing existing vppAdapter") //this is used for testing purposes
	}

	startTime := time.Now()
	plugin.vppConn, plugin.vppConChan, err = govpp.AsyncConnect(plugin.vppAdapter)
	if err != nil {
		return err
	}

	// TODO: Async connect & automatic reconnect support is not yet implemented in the agent,
	// so synchronously wait until connected to VPP.
	status := <-plugin.vppConChan
	if status.State != govpp.Connected {
		return errors.New("unable to connect to VPP")
	}
	vppConnectTime := time.Since(startTime)
	plugin.Log.WithField("durationInNs", vppConnectTime.Nanoseconds()).Info("Connecting to VPP took ", vppConnectTime)
	plugin.retrieveVersion()

	// Register providing status reports (push mode)
	plugin.StatusCheck.Register(plugin.PluginName, nil)
	plugin.StatusCheck.ReportStateChange(plugin.PluginName, statuscheck.OK, nil)
	plugin.Log.Debug("govpp connect success ", plugin.vppConn)

	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())
	go plugin.handleVPPConnectionEvents(ctx)

	return nil
}

// Close cleans up the resources allocated by the govppmux plugin.
func (plugin *GOVPPPlugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	defer func() {
		if plugin.vppConn != nil {
			plugin.vppConn.Disconnect()
		}
	}()

	return nil
}

// NewAPIChannel returns a new API channel for communication with VPP via govpp core.
// It uses default buffer sizes for the request and reply Go channels.
//
// Example of binary API call from some plugin using GOVPP:
//      ch, _ := govpp_mux.NewAPIChannel()
//      ch.SendRequest(req).ReceiveReply
func (plugin *GOVPPPlugin) NewAPIChannel() (govppapi.Channel, error) {
	ch, err := plugin.vppConn.NewAPIChannel()
	if err != nil {
		return nil, err
	}
	if plugin.replyTimeout > 0 {
		ch.SetReplyTimeout(plugin.replyTimeout)
	}
	return ch, nil
}

// NewAPIChannelBuffered returns a new API channel for communication with VPP via govpp core.
// It allows to specify custom buffer sizes for the request and reply Go channels.
//
// Example of binary API call from some plugin using GOVPP:
//      ch, _ := govpp_mux.NewAPIChannelBuffered(100, 100)
//      ch.SendRequest(req).ReceiveReply
func (plugin *GOVPPPlugin) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (govppapi.Channel, error) {
	ch, err := plugin.vppConn.NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize)
	if err != nil {
		return nil, err
	}
	if plugin.replyTimeout > 0 {
		ch.SetReplyTimeout(plugin.replyTimeout)
	}
	return ch, nil
}

// handleVPPConnectionEvents handles VPP connection events.
func (plugin *GOVPPPlugin) handleVPPConnectionEvents(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case status := <-plugin.vppConChan:
			if status.State == govpp.Connected {
				plugin.retrieveVersion()
				if plugin.reconnectResync && plugin.lastConnErr != nil {
					plugin.Log.Info("Starting resync after VPP reconnect")
					if plugin.Resync != nil {
						plugin.Resync.DoResync()
						plugin.lastConnErr = nil
					} else {
						plugin.Log.Warn("Expected resync after VPP reconnect could not start because of missing Resync plugin")
					}
				}
				plugin.StatusCheck.ReportStateChange(plugin.PluginName, statuscheck.OK, nil)
			} else {
				plugin.lastConnErr = errors.New("VPP disconnected")
				plugin.StatusCheck.ReportStateChange(plugin.PluginName, statuscheck.Error, plugin.lastConnErr)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (plugin *GOVPPPlugin) retrieveVersion() {
	vppAPIChan, err := plugin.vppConn.NewAPIChannel()
	if err != nil {
		plugin.Log.Error("getting new api channel failed:", err)
		return
	}
	defer vppAPIChan.Close()

	info, err := vppcalls.GetVersionInfo(vppAPIChan)
	if err != nil {
		plugin.Log.Warn("getting version info failed:", err)
		return
	}

	plugin.Log.Debugf("version info: %+v", info)
	plugin.Log.Infof("VPP version: %v (%v)", info.Version, info.BuildDate)
}
