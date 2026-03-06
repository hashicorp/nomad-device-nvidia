// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package nvidia

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/nomad-device-nvidia/nvml"
	"github.com/hashicorp/nomad/helper/pointer"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

const (
	// Attribute names and units for reporting Fingerprint output
	MemoryAttr          = "memory"
	PowerAttr           = "power"
	BAR1Attr            = "bar1"
	DriverVersionAttr   = "driver_version"
	CoresClockAttr      = "cores_clock"
	MemoryClockAttr     = "memory_clock"
	PCIBandwidthAttr    = "pci_bandwidth"
	DisplayStateAttr    = "display_state"
	PersistenceModeAttr = "persistence_mode"
)

// fingerprint is the long running goroutine that detects hardware
func (d *NvidiaDevice) fingerprint(ctx context.Context, devices chan<- *device.FingerprintResponse) {
	defer close(devices)

	if d.initErr != nil {
		if d.initErr.Error() != nvml.ErrUnavailableLib.Error() {
			d.logger.Error("exiting fingerprinting due to problems with NVML loading", "error", d.initErr)
			devices <- device.NewFingerprintError(d.initErr)
		}

		// Just close the channel to let server know that there are no working
		// Nvidia GPU units
		return
	}

	// Create a timer that will fire immediately for the first detection
	ticker := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Reset(d.fingerprintPeriod)
		}
		d.writeFingerprintToChannel(devices)
	}
}

// writeFingerprintToChannel makes nvml call and writes response to channel
func (d *NvidiaDevice) writeFingerprintToChannel(devices chan<- *device.FingerprintResponse) {
	fingerprintData, err := d.nvmlClient.GetFingerprintData()
	if err != nil {
		d.logger.Error("failed to get fingerprint nvidia devices", "error", err)
		devices <- device.NewFingerprintError(err)
		return
	}

	// ignore devices from fingerprint output
	fingerprintDevices := ignoreFingerprintedDevices(fingerprintData.Devices, d.ignoredGPUIDs)

	for _, dev := range fingerprintDevices {
		// skip mig mode devices marked ineligible by the client
		if dev.SharingStatus != device.SharingIneligible {
			continue
		}
		// set all eligible devices at once if globalPipeDirectory is set
		if d.MpsConfig.MpsPipeDirectory != "" {
			dev.SharingStatus = d.getDeviceSharingStatus("unixpacket", d.MpsConfig.MpsPipeDirectory)
			continue
		}

		// otherwise lookup appropriate pipe directory by deviceUUID
		if len(d.MpsConfig.DeviceMpsConfig) != 0 {
			devConfig := d.MpsConfig.DeviceMpsConfig[dev.UUID]
			dev.SharingStatus = d.getDeviceSharingStatus("unixpacket", devConfig.MpsPipeDirectory)
		}
	}
	// check if any device health was updated or any device was added to host
	if !d.fingerprintChanged(fingerprintDevices) {
		return
	}

	commonAttributes := map[string]*structs.Attribute{
		DriverVersionAttr: {
			String: pointer.Of(fingerprintData.DriverVersion),
		},
	}

	// Group all FingerprintDevices by DeviceName and SharingStatus attributes
	deviceListByDeviceNameAndSharing := make(map[string][]*nvml.FingerprintDeviceData)

	for _, dev := range fingerprintDevices {
		deviceName := dev.DeviceName

		if deviceName == nil {
			// nvml driver was not able to detect device name. This kind
			// of devices are placed to single group with 'notAvailable' name
			notAvailableCopy := notAvailable
			deviceName = &notAvailableCopy
		}
		sharingName := fmt.Sprintf("%s.%s", dev.SharingStatus, dev.DeviceName)
		deviceListByDeviceNameAndSharing[sharingName] = append(deviceListByDeviceNameAndSharing[sharingName], dev)

	}

	// Build Fingerprint response with computed groups and send it over the channel
	deviceGroups := make([]*device.DeviceGroup, 0, len(deviceListByDeviceNameAndSharing))
	for groupName, devices := range deviceListByDeviceNameAndSharing {
		deviceGroups = append(deviceGroups, deviceGroupFromFingerprintData(groupName, devices, commonAttributes))
	}
	devices <- device.NewFingerprint(deviceGroups...)
}

// ignoreFingerprintedDevices excludes ignored devices from fingerprint output
func ignoreFingerprintedDevices(deviceData []*nvml.FingerprintDeviceData, ignoredGPUIDs map[string]struct{}) []*nvml.FingerprintDeviceData {
	var result []*nvml.FingerprintDeviceData
	for _, fingerprintDevice := range deviceData {
		if _, ignored := ignoredGPUIDs[fingerprintDevice.UUID]; !ignored {
			result = append(result, fingerprintDevice)
		}
	}
	return result
}

// fingerprintChanged checks if there are any previously unseen nvidia devices located
// or any of fingerprinted nvidia devices disappeared since the last fingerprint run.
// Also, this func updates device map on NvidiaDevice with the latest data
func (d *NvidiaDevice) fingerprintChanged(allDevices []*nvml.FingerprintDeviceData) bool {
	d.deviceLock.Lock()
	defer d.deviceLock.Unlock()

	changeDetected := false
	// check if every device in allDevices is in d.devices
	for _, dev := range allDevices {
		if status, ok := d.devices[dev.UUID]; !ok || status != dev.SharingStatus {
			changeDetected = true
		}

	}

	// check if every device in d.devices is in allDevices
	fingerprintDeviceMap := make(map[string]device.DeviceSharing)
	for _, dev := range allDevices {

		// include  sharing status in the fingerprintDeviceMap
		// that gets saved to the device
		fingerprintDeviceMap[dev.UUID] = dev.SharingStatus
	}
	for id := range d.devices {
		if _, ok := fingerprintDeviceMap[id]; !ok {
			changeDetected = true
		}
	}

	d.devices = fingerprintDeviceMap
	return changeDetected
}

// deviceGroupFromFingerprintData composes deviceGroup from FingerprintDeviceData slice
func deviceGroupFromFingerprintData(groupName string, deviceList []*nvml.FingerprintDeviceData, commonAttributes map[string]*structs.Attribute) *device.DeviceGroup {
	// deviceGroup without devices makes no sense -> return nil when no devices are provided
	if len(deviceList) == 0 {
		return nil
	}

	devices := make([]*device.Device, len(deviceList))
	for index, dev := range deviceList {
		devices[index] = &device.Device{
			ID: dev.UUID,
			// all fingerprinted devices are "healthy" for now
			// to get real health data -> dcgm bindings should be used
			Healthy: true,
			HwLocality: &device.DeviceLocality{
				PciBusID: dev.PCIBusID,
			},
		}
	}

	deviceGroup := &device.DeviceGroup{
		Vendor:  vendor,
		Type:    deviceType,
		Name:    groupName,
		Devices: devices,
		// Assumption made that devices with the same DeviceName have the same
		// attributes like amount of memory, power, bar1memory etc
		Attributes: attributesFromFingerprintDeviceData(deviceList[0]),
	}

	// Extend attribute map with common attributes
	for attributeKey, attributeValue := range commonAttributes {
		deviceGroup.Attributes[attributeKey] = attributeValue
	}

	return deviceGroup
}

// attributesFromFingerprintDeviceData converts nvml.FingerprintDeviceData
// struct to device.DeviceGroup.Attributes format (map[string]string)
// this function performs all nil checks for FingerprintDeviceData pointers
func attributesFromFingerprintDeviceData(d *nvml.FingerprintDeviceData) map[string]*structs.Attribute {
	attrs := map[string]*structs.Attribute{
		DisplayStateAttr: {
			String: pointer.Of(d.DisplayState),
		},
		PersistenceModeAttr: {
			String: pointer.Of(d.PersistenceMode),
		},
	}

	if d.MemoryMiB != nil {
		attrs[MemoryAttr] = &structs.Attribute{
			Int:  pointer.Of(int64(*d.MemoryMiB)),
			Unit: structs.UnitMiB,
		}
	}
	if d.PowerW != nil {
		attrs[PowerAttr] = &structs.Attribute{
			Int:  pointer.Of(int64(*d.PowerW)),
			Unit: structs.UnitW,
		}
	}
	if d.BAR1MiB != nil {
		attrs[BAR1Attr] = &structs.Attribute{
			Int:  pointer.Of(int64(*d.BAR1MiB)),
			Unit: structs.UnitMiB,
		}
	}
	if d.CoresClockMHz != nil {
		attrs[CoresClockAttr] = &structs.Attribute{
			Int:  pointer.Of(int64(*d.CoresClockMHz)),
			Unit: structs.UnitMHz,
		}
	}
	if d.MemoryClockMHz != nil {
		attrs[MemoryClockAttr] = &structs.Attribute{
			Int:  pointer.Of(int64(*d.MemoryClockMHz)),
			Unit: structs.UnitMHz,
		}
	}
	if d.PCIBandwidthMBPerS != nil {
		attrs[PCIBandwidthAttr] = &structs.Attribute{
			Int:  pointer.Of(int64(*d.PCIBandwidthMBPerS)),
			Unit: structs.UnitMBPerS,
		}
	}

	return attrs
}

// getDeviceSharingStatus attempts to connect to the mps-control socket
// using the configured mps_pip_directory
func (d *NvidiaDevice) getDeviceSharingStatus(dialtype string, mps_pipe_directory string) device.DeviceSharing {
	sockAddr := mps_pipe_directory + "/control"
	tries := 0
	for tries < 5 {
		_, err := net.Dial(dialtype, sockAddr)
		if err == nil {
			return device.SharingActive
		}
		tries++

		backoff := time.Duration(tries*500) * time.Microsecond
		time.Sleep(backoff)

		d.logger.Error(fmt.Sprintf("failed to reach mps daemon after %d attempts", tries), "error", err.Error())
		continue

	}
	return device.SharingInactive

}
