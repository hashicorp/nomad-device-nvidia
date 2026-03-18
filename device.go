// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package nvidia

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-device-nvidia/nvml"
	"github.com/hashicorp/nomad-device-nvidia/version"
	"github.com/hashicorp/nomad/helper/pluginutils/loader"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
)

const (
	// pluginName is the name of the plugin
	pluginName = "nvidia-gpu"

	// vendor is the vendor providing the devices
	vendor = "nvidia"

	// deviceType is the type of device being returned
	deviceType = device.DeviceTypeGPU

	// notAvailable value is returned to nomad server in case some properties were
	// undetected by nvml driver
	notAvailable = "N/A"

	// Nvidia-container-runtime environment variable names
	NvidiaVisibleDevices = "NVIDIA_VISIBLE_DEVICES"

	// MPS runtime environment variables
	MpsPipeDirectoryKey = "MPS_PIPE_DIRECTORY"
	MpsLogDirectoryKey  = "MPS_LOG_DIRECTORY"
	CustomMpsUserKey    = "MPS_USER"

	DefaultMpsSockFileAddr       = "control"
	DefaultMpsLogDirectoryValue  = "/var/log/nvidia-mps"
	DefaultMpsPipeDirectoryValue = "/tmp/nvidia-mps"
	DefaultMpsUserValue          = "unset"
)

var (
	// PluginID is the nvidia plugin metadata registered in the plugin
	// catalog.
	PluginID = loader.PluginID{
		Name:       pluginName,
		PluginType: base.PluginTypeDevice,
	}

	// PluginConfig is the nvidia factory function registered in the
	// plugin catalog.
	PluginConfig = &loader.InternalPluginConfig{
		Factory: func(ctx context.Context, l hclog.Logger) interface{} { return NewNvidiaDevice(ctx, l) },
	}

	// pluginInfo describes the plugin
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDevice,
		PluginApiVersions: []string{device.ApiVersion010},
		PluginVersion:     version.Version,
		Name:              pluginName,
	}

	// configSpec is the specification of the plugin's configuration
	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"enabled": hclspec.NewDefault(
			hclspec.NewAttr("enabled", "bool", false),
			hclspec.NewLiteral("true"),
		),

		"ignored_gpu_ids": hclspec.NewDefault(
			hclspec.NewAttr("ignored_gpu_ids", "list(string)", false),
			hclspec.NewLiteral("[]"),
		),
		"fingerprint_period": hclspec.NewDefault(
			hclspec.NewAttr("fingerprint_period", "string", false),
			hclspec.NewLiteral("\"1m\""),
		),
		"mps": hclspec.NewBlock("mps", false,
			hclspec.NewObject(map[string]*hclspec.Spec{
				"enabled":            hclspec.NewAttr("enabled", "bool", true),
				"mps_user":           hclspec.NewAttr("mps_user", "string", false),
				"mps_pipe_directory": hclspec.NewAttr("mps_pipe_directory", "string", false),
				"mps_log_directory":  hclspec.NewAttr("mps_log_directory", "string", false),
				"mps_sock_addr":      hclspec.NewAttr("mps_sock_addr", "string", false),
				"device_specific_mps_config": hclspec.NewBlockList("device_specific_mps_config",
					hclspec.NewObject(map[string]*hclspec.Spec{
						"uuid":               hclspec.NewAttr("uuid", "string", true),
						"mps_pipe_directory": hclspec.NewAttr("mps_pipe_directory", "string", true),
						"mps_log_directory":  hclspec.NewAttr("mps_log_directory", "string", true),
					}),
				),
			}),
		),
	})
)

// Config contains configuration information for the plugin.
type Config struct {
	Enabled           bool       `codec:"enabled"`
	IgnoredGPUIDs     []string   `codec:"ignored_gpu_ids"`
	FingerprintPeriod string     `codec:"fingerprint_period"`
	MpsConfig         *MpsConfig `codec:"mps"`
	Dir               string     `codec:"dir"`
	ListPeriod        string     `codec:"list_period"`
	UnhealthyPerm     string     `codec:"unhealthy_perm"`
}

// MpsConfig contains configuration for mps sharing
type MpsConfig struct {
	MpsUser          string            `codec:"mps_user"`
	MpsSockFile      string            `codec:"mps_sock_addr"`
	MpsPipeDirectory string            `codec:"mps_pipe_directory"`
	MpsLogDirectory  string            `codec:"mps_log_directory"`
	DeviceConfig     []DeviceMpsConfig `codec:"device_specific_mps_config"`
	DeviceMpsConfig  map[string]DeviceMpsConfig
}

// DeviceMpsConfig contains configuration GPU level mps sharing
type DeviceMpsConfig struct {
	UUID             string `codec:"uuid"`
	MpsPipeDirectory string `codec:"mps_pipe_directory"`
	MpsLogDirectory  string `codec:"mps_log_directory"`
}

// NvidiaDevice contains all plugin specific data
type NvidiaDevice struct {
	// enabled indicates whether the plugin should be enabled
	enabled bool

	// nvmlClient is used to get data from nvidia
	nvmlClient nvml.NvmlClient

	// initErr holds an error retrieved during
	// nvmlClient initialization
	initErr error

	// ignoredGPUIDs is a set of UUIDs that would not be exposed to nomad
	ignoredGPUIDs map[string]struct{}

	// fingerprintPeriod is how often we should call nvml to get list of devices
	fingerprintPeriod time.Duration

	//MpsConfig holds a pointer to the MPS configuration
	MpsConfig *MpsConfig

	// devices is the set of detected eligible devices
	devices    map[string]device.DeviceSharing
	deviceLock sync.RWMutex

	logger hclog.Logger
}

// NewNvidiaDevice returns a new nvidia device plugin.
func NewNvidiaDevice(_ context.Context, log hclog.Logger) *NvidiaDevice {
	nvmlClient, err := nvml.NewNvmlClient()
	logger := log.Named(pluginName)
	if err != nil && err.Error() != nvml.ErrUnavailableLib.Error() {
		logger.Error("unable to initialize Nvidia driver", "reason", err)
	}
	return &NvidiaDevice{
		logger:        logger,
		devices:       make(map[string]device.DeviceSharing),
		ignoredGPUIDs: make(map[string]struct{}),
		nvmlClient:    nvmlClient,
		initErr:       err,
	}
}

// PluginInfo returns information describing the plugin.
func (d *NvidiaDevice) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

// ConfigSchema returns the plugins configuration schema.
func (d *NvidiaDevice) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

func selectConfigOrDefault(c string, d string) string {
	if config := c; config != "" {
		return c
	}
	return d
}

// SetConfig is used to set the configuration of the plugin.
func (d *NvidiaDevice) SetConfig(cfg *base.Config) error {
	var config Config
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return err
		}
	}

	// set MPS config values
	if config.MpsConfig != nil {
		d.MpsConfig = &MpsConfig{}
		// ensure only global or device specific config are set
		if (config.MpsConfig.MpsPipeDirectory != "" || config.MpsConfig.MpsLogDirectory != "") &&
			len(config.MpsConfig.DeviceMpsConfig) != 0 {
			return errors.New("only top level mps directory variables or device_specific_mps_config block may be set ")
		}
		// set top level only mps values
		d.MpsConfig.MpsUser = selectConfigOrDefault(config.MpsConfig.MpsUser, "unset")
		d.MpsConfig.MpsSockFile = selectConfigOrDefault(config.MpsConfig.MpsSockFile, DefaultMpsSockFileAddr)

		// if no device specific config, set top level values
		// otherwise set device level config
		if len(config.MpsConfig.DeviceMpsConfig) == 0 {
			d.MpsConfig.MpsPipeDirectory = selectConfigOrDefault(config.MpsConfig.MpsPipeDirectory, DefaultMpsPipeDirectoryValue)
			d.MpsConfig.MpsLogDirectory = selectConfigOrDefault(config.MpsConfig.MpsLogDirectory, DefaultMpsLogDirectoryValue)
		} else {
			// build map of device UUIDs to config
			deviceConfigMap := make(map[string]DeviceMpsConfig, len(config.MpsConfig.DeviceMpsConfig))

			for _, devConfig := range config.MpsConfig.DeviceMpsConfig {
				deviceConfigMap[devConfig.UUID] = DeviceMpsConfig{
					UUID:             devConfig.UUID,
					MpsPipeDirectory: devConfig.MpsPipeDirectory,
					MpsLogDirectory:  devConfig.MpsLogDirectory,
				}
			}
			// set device specific mpsConfig
			d.MpsConfig.DeviceMpsConfig = deviceConfigMap
		}
	}
	for _, ignoredGPUId := range config.IgnoredGPUIDs {
		d.ignoredGPUIDs[ignoredGPUId] = struct{}{}
	}

	period, err := time.ParseDuration(config.FingerprintPeriod)
	if err != nil {
		return fmt.Errorf("failed to parse fingerprint period %q: %v", config.FingerprintPeriod, err)
	}
	d.fingerprintPeriod = period

	return nil
}

// Fingerprint streams detected devices. If device changes are detected or the
// devices health changes, messages will be emitted.
func (d *NvidiaDevice) Fingerprint(ctx context.Context) (<-chan *device.FingerprintResponse, error) {
	if !d.enabled {
		return nil, device.ErrPluginDisabled
	}

	outCh := make(chan *device.FingerprintResponse)
	go d.fingerprint(ctx, outCh)
	return outCh, nil
}

type reservationError struct {
	notExistingIDs []string
}

func (e *reservationError) Error() string {
	return fmt.Sprintf("unknown device IDs: %s", strings.Join(e.notExistingIDs, ","))
}

// Reserve returns information on how to mount given devices.
// Assumption is made that nomad server is responsible for correctness of
// GPU allocations, handling tricky cases such as double-allocation of single GPU
func (d *NvidiaDevice) Reserve(deviceIDs []string) (*device.ContainerReservation, error) {
	if len(deviceIDs) == 0 {
		return &device.ContainerReservation{}, nil
	}
	if !d.enabled {
		return nil, device.ErrPluginDisabled
	}
	var (
		notExistingIDs []string

		reservedDeviceIDs []string
	)
	containerEnvs := make(map[string]string)
	// Due to the asynchronous nature of NvidiaPlugin, there is a possibility
	// of race condition
	//
	// Timeline:
	// 	1 - fingerprint reports that GPU with id "1" is present
	//  2 - the following events happen at the same time:
	// 		a) server decides to allocate GPU with id "1"
	//      b) fingerprint check reports that GPU with id "1" is no more present
	//
	// The latest and always valid version of fingerprinted ids are stored in
	// d.devices map. To avoid this race condition an error is returned if
	// any of provided deviceIDs is not found in d.devices map
	d.deviceLock.RLock()

	for i, id := range deviceIDs {
		if _, deviceIDExists := d.devices[id]; !deviceIDExists {
			notExistingIDs = append(notExistingIDs, id)
		}

		// if set, build mps environment variables
		if d.MpsConfig != nil {
			// check for custom user and add to envar map
			if d.MpsConfig.MpsUser != "unset" {
				containerEnvs[CustomMpsUserKey] = d.MpsConfig.MpsUser
			}
			// pass along top-level mounts and envs if mps != nil and no
			// device specific config exists
			if len(d.MpsConfig.DeviceMpsConfig) == 0 {
				containerEnvs[MpsPipeDirectoryKey] = d.MpsConfig.MpsPipeDirectory
				containerEnvs[MpsLogDirectoryKey] = d.MpsConfig.MpsLogDirectory

				reservedDeviceIDs = append(reservedDeviceIDs, id)
				continue
			}

			// build appropriate mounts and envs if device specific config exists
			// and the specific device is in the map
			if c, ok := d.MpsConfig.DeviceMpsConfig[id]; ok {
				reservedDeviceIDs = append(reservedDeviceIDs, id)

				// each task definition must target a single MPS server so
				// use the first deviceID to look up and set envvars
				if i == 0 {
					containerEnvs[MpsPipeDirectoryKey] = c.MpsPipeDirectory
					containerEnvs[MpsLogDirectoryKey] = c.MpsLogDirectory
				}
			}

		}
	}

	d.deviceLock.RUnlock()
	if len(notExistingIDs) != 0 {
		return nil, &reservationError{notExistingIDs}
	}
	// return all available devices if mps is not set
	if d.MpsConfig == nil {
		return &device.ContainerReservation{
			Envs: map[string]string{NvidiaVisibleDevices: strings.Join(deviceIDs, ",")},
		}, nil
	}

	// if mps is set return configured devices and mounts
	containerEnvs[NvidiaVisibleDevices] = strings.Join(reservedDeviceIDs, ",")
	return &device.ContainerReservation{
		Envs: containerEnvs,
		Mounts: []*device.Mount{{
			HostPath: containerEnvs[MpsPipeDirectoryKey],
			TaskPath: containerEnvs[MpsPipeDirectoryKey],
			ReadOnly: false,
		}},
	}, nil
}

// Stats streams statistics for the detected devices.
func (d *NvidiaDevice) Stats(ctx context.Context, interval time.Duration) (<-chan *device.StatsResponse, error) {
	if !d.enabled {
		return nil, device.ErrPluginDisabled
	}

	outCh := make(chan *device.StatsResponse)
	go d.stats(ctx, outCh, interval)
	return outCh, nil
}
