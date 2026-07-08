// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build windows

package nvml

import (
	"fmt"
	"maps"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

func decode(msg string, code nvml.Return) error {
	return fmt.Errorf("%s: %s", msg, nvml.ErrorString(code))
}

// Initialize nvml library by loading nvml.dll and calling nvmlInit
func (n *nvmlDriver) Initialize() error {
	if code := nvmlInit(); code != NVML_SUCCESS {
		return decode("failed to initialize", nvml.Return(code))
	}
	return nil
}

// Shutdown stops any further interaction with nvml
func (n *nvmlDriver) Shutdown() error {
	if code := nvmlShutdown(); code != NVML_SUCCESS {
		return decode("failed to shutdown", nvml.Return(code))
	}
	return nil
}

// SystemDriverVersion returns installed driver version
func (n *nvmlDriver) SystemDriverVersion() (string, error) {
	version, code := nvmlSystemGetDriverVersion()
	if code != NVML_SUCCESS {
		return "", decode("failed to get system driver version", nvml.Return(code))
	}
	return version, nil
}

// List all compute device UUIDs in the system.
// Includes all instances, including normal GPUs, MIGs, and their physical parents.
// Each UUID is associated with a mode indication which type it is.
func (n *nvmlDriver) ListDeviceUUIDs() (map[string]mode, error) {
	count, code := nvmlDeviceGetCount()
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device count", nvml.Return(code))
	}

	uuids := make(map[string]mode)

	for i := 0; i < int(count); i++ {
		handle, code := nvmlDeviceGetHandleByIndex(i)
		if code != NVML_SUCCESS {
			return nil, decode(fmt.Sprintf("failed to get device handle %d/%d", i, count), nvml.Return(code))
		}

		device := newWinDevice(handle)

		devIDs, err := uuidsFromDevice(device)
		if err != nil {
			return nil, err
		}
		maps.Copy(uuids, devIDs)
	}

	return uuids, nil
}

// DeviceInfoByUUID returns DeviceInfo for the given GPU's UUID.
func (n *nvmlDriver) DeviceInfoByUUID(uuid string) (*DeviceInfo, error) {
	handle, code := nvmlDeviceGetHandleByUUID(uuid)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device handle", nvml.Return(code))
	}

	device := newWinDevice(handle)

	info, err := deviceInfoFromDevice(device)
	if err != nil {
		return nil, err
	}
	info.UUID = uuid

	return info, nil
}

// DeviceStatusByUUID returns DeviceStatus for the given GPU's UUID.
func (n *nvmlDriver) DeviceStatusByUUID(uuid string) (*DeviceStatus, error) {
	handle, code := nvmlDeviceGetHandleByUUID(uuid)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device info", nvml.Return(code))
	}

	device := newWinDevice(handle)

	return deviceStatusByDevice(device)
}

// DeviceInfoAndStatusByUUID returns DeviceInfo and DeviceStatus for index GPU in system device list.
func (n *nvmlDriver) DeviceInfoAndStatusByUUID(uuid string) (*DeviceInfo, *DeviceStatus, error) {
	di, err := n.DeviceInfoByUUID(uuid)
	if err != nil {
		return nil, nil, err
	}

	ds, err := n.DeviceStatusByUUID(uuid)
	if err != nil {
		return nil, nil, err
	}

	return di, ds, nil
}
