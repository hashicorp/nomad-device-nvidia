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

	DefaultMpsSockFileAddr = "control"
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
				"mps_user": hclspec.NewDefault(
					hclspec.NewAttr("mps_user", "string", false),
					hclspec.NewLiteral("unset"),
				),
				"mps_pipe_directory": hclspec.NewDefault(
					hclspec.NewAttr("mps_pipe_directory", "string", false),
					hclspec.NewLiteral("/tmp/nvidia-mps"),
				),
				"mps_log_directory": hclspec.NewDefault(
					hclspec.NewAttr("mps_log_directory", "string", false),
					hclspec.NewLiteral("/var/log/nvidia-mps"),
				),
				"mps_sock_addr": hclspec.NewDefault(
					hclspec.NewAttr("mps_sock_addr", "string", false),
					hclspec.NewLiteral("control"),
				),
				"device_specific_mps_config": hclspec.NewBlockList("device_specific_mps_config",
					hclspec.NewArray(
						[]*hclspec.Spec{
							hclspec.NewObject(map[string]*hclspec.Spec{
								"uuid": hclspec.NewAttr("uuid", "string", true),
								"mps_pipe_directory": hclspec.NewDefault(
									hclspec.NewAttr("mps_pipe_directory", "string", true),
									hclspec.NewLiteral("/tmp/nvidia-mps"),
								),
								"mps_log_directory": hclspec.NewDefault(
									hclspec.NewAttr("mps_log_directory", "string", true),
									hclspec.NewLiteral("/tmp/nvidia-mps"),
								),
							}),
						},
					)),
			})),
	})
)

// Config contains configuration information for the plugin.
type Config struct {
	Enabled           bool       `codec:"enabled"`
	IgnoredGPUIDs     []string   `codec:"ignored_gpu_ids"`
	FingerprintPeriod string     `codec:"fingerprint_period"`
	MpsConfig         *MpsConfig `codec:"mps"`
}

type MpsConfig struct {
	MpsUser          string                     `codec:"mps_user"`
	MpsSockFile      string                     `codec:"mps_sock_addr"`
	MpsPipeDirectory string                     `codec:"mps_pipe_directory"`
	MpsLogDirectory  string                     `codec:"mps_log_directory"`
	DeviceMpsConfig  map[string]DeviceMpsConfig `codec:"device_specific_mps_config"`
}
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
		// ensure only global or device specific config are set
		if config.MpsConfig.MpsPipeDirectory != "" || config.MpsConfig.MpsLogDirectory != "" &&
			len(config.MpsConfig.DeviceMpsConfig) != 0 {
			return errors.New("only global mps variables or device_specific_mps_config block may be set ")
		}

		// set straightforward value on device
		d.MpsConfig.MpsUser = config.MpsConfig.MpsUser
		d.MpsConfig.MpsSockFile = DefaultMpsSockFileAddr
		// overwrite sock file addr if set
		if config.MpsConfig.MpsSockFile != "" {
			d.MpsConfig.MpsSockFile = config.MpsConfig.MpsSockFile
		}
		// if present set device specific mps config, otherwise set top level config
		if len(config.MpsConfig.DeviceMpsConfig) != 0 {

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
		} else {
			// set top level mps directories if no device specific config
			// we have defaults so always use config values
			d.MpsConfig.MpsPipeDirectory = config.MpsConfig.MpsPipeDirectory
			d.MpsConfig.MpsLogDirectory = config.MpsConfig.MpsLogDirectory
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
		notExistingIDs       []string
		containerEnvs        map[string]string
		reserveDeviceBuilder strings.Builder
	)
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
		if d.MpsConfig != nil && len(d.MpsConfig.DeviceMpsConfig) == 0 {
			containerEnvs[MpsPipeDirectoryKey] = d.MpsConfig.MpsPipeDirectory
			containerEnvs[MpsLogDirectoryKey] = d.MpsConfig.MpsLogDirectory

			if i == 0 {
				reserveDeviceBuilder.WriteString(strings.Join(deviceIDs, ","))
			}
		}
		// build appropriate mounts and envs if
		// mps is limited to specific GPUS
		if c, ok := d.MpsConfig.DeviceMpsConfig[id]; ok {
			reserveDeviceBuilder.WriteString(id)
			if i < len(deviceIDs)-1 {
				reserveDeviceBuilder.WriteString(",")
			}

			// use the first deviceID to look up and set envvars
			if i == 0 {
				containerEnvs[MpsPipeDirectoryKey] = c.MpsPipeDirectory
				containerEnvs[MpsLogDirectoryKey] = c.MpsLogDirectory
			}
		}
	}

	d.deviceLock.RUnlock()
	if len(notExistingIDs) != 0 {
		return nil, &reservationError{notExistingIDs}
	}
	containerEnvs[NvidiaVisibleDevices] = reserveDeviceBuilder.String()
	if d.MpsConfig.MpsUser != "unset" {
		containerEnvs[CustomMpsUserKey] = d.MpsConfig.MpsUser
	}

	mountPipeDir := &device.Mount{
		HostPath: containerEnvs[MpsPipeDirectoryKey],
		TaskPath: containerEnvs[MpsPipeDirectoryKey],
		ReadOnly: false,
	}

	return &device.ContainerReservation{
		Envs:   containerEnvs,
		Mounts: []*device.Mount{mountPipeDir},
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
