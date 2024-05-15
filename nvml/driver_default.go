// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !linux

package nvml

// Initialize nvml library by locating nvml shared object file and calling ldopen
func (n *nvmlDriver) Initialize() error {
	return UnavailableLib
}

// Shutdown stops any further interaction with nvml
func (n *nvmlDriver) Shutdown() error {
	return UnavailableLib
}

// SystemDriverVersion returns installed driver version
func (n *nvmlDriver) SystemDriverVersion() (string, error) {
	return "", UnavailableLib
}

// ListDeviceUUIDs reports number of available GPU devices
func (n *nvmlDriver) ListDeviceUUIDs() ([]string, error) {
	return nil, UnavailableLib
}

// DeviceInfoByUUID returns DeviceInfo for the GPU matching the given UUID
func (n *nvmlDriver) DeviceInfoByUUID(uuid string) (*DeviceInfo, error) {
	return nil, UnavailableLib
}

// DeviceInfoAndStatusByUUID returns DeviceInfo and DeviceStatus for the GPU matching the given UUID
func (n *nvmlDriver) DeviceInfoAndStatusByUUID(uuid string) (*DeviceInfo, *DeviceStatus, error) {
	return nil, nil, UnavailableLib
}
