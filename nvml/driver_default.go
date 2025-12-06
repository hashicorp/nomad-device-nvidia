// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build !linux

package nvml

// Initialize nvml library by locating nvml shared object file and calling ldopen
func (n *nvmlDriver) Initialize() error {
	return ErrUnavailableLib
}

// Shutdown stops any further interaction with nvml
func (n *nvmlDriver) Shutdown() error {
	return ErrUnavailableLib
}

// SystemDriverVersion returns installed driver version
func (n *nvmlDriver) SystemDriverVersion() (string, error) {
	return "", ErrUnavailableLib
}

// ListDeviceUUIDs reports number of available GPU devices
func (n *nvmlDriver) ListDeviceUUIDs() (map[string]mode, error) {
	return nil, ErrUnavailableLib
}

// DeviceInfoByUUID returns DeviceInfo for the GPU matching the given UUID
func (n *nvmlDriver) DeviceInfoByUUID(uuid string) (*DeviceInfo, error) {
	return nil, ErrUnavailableLib
}

// DeviceInfoAndStatusByUUID returns DeviceInfo and DeviceStatus for the GPU matching the given UUID
func (n *nvmlDriver) DeviceInfoAndStatusByUUID(uuid string) (*DeviceInfo, *DeviceStatus, error) {
	return nil, nil, ErrUnavailableLib
}
