// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package nvml

import (
	"cmp"
	"fmt"
	"slices"
)

// DeviceData represents common fields for Nvidia device
type DeviceData struct {
	UUID       string
	DeviceName *string
	MemoryMiB  *uint64
	PowerW     *uint
	BAR1MiB    *uint64
}

// FingerprintDeviceData is a superset of DeviceData
// it describes device specific fields returned from
// nvml queries during fingerprinting call
type FingerprintDeviceData struct {
	*DeviceData
	PCIBandwidthMBPerS *uint
	CoresClockMHz      *uint
	MemoryClockMHz     *uint
	DisplayState       string
	PersistenceMode    string
	PCIBusID           string
}

// FingerprintData represets attributes of driver/devices
type FingerprintData struct {
	Devices       []*FingerprintDeviceData
	DriverVersion string
}

// StatsData is a superset of DeviceData
// it represents statistics data returned for every Nvidia device
type StatsData struct {
	*DeviceData
	PowerUsageW        *uint
	GPUUtilization     *uint
	MemoryUtilization  *uint
	EncoderUtilization *uint
	DecoderUtilization *uint
	TemperatureC       *uint
	UsedMemoryMiB      *uint64
	BAR1UsedMiB        *uint64
	ECCErrorsL1Cache   *uint64
	ECCErrorsL2Cache   *uint64
	ECCErrorsDevice    *uint64
}

// NvmlClient describes how users would use nvml library
type NvmlClient interface {
	GetFingerprintData() (*FingerprintData, error)
	GetStatsData() ([]*StatsData, error)
}

// nvmlClient implements NvmlClient
// Users of this lib are expected to use this struct via NewNvmlClient func
type nvmlClient struct {
	driver NvmlDriver
}

// NewNvmlClient function creates new nvmlClient with real
// NvmlDriver implementation. Also, this func initializes NvmlDriver
func NewNvmlClient() (*nvmlClient, error) {
	driver := &nvmlDriver{}
	err := driver.Initialize()
	if err != nil {
		return nil, err
	}
	return &nvmlClient{
		driver: driver,
	}, nil
}

// GetFingerprintData returns FingerprintData for available Nvidia devices
func (c *nvmlClient) GetFingerprintData() (*FingerprintData, error) {
	/*
		nvml fields to be fingerprinted # nvml_library_call
		1  - Driver Version             # nvmlSystemGetDriverVersion
		2  - Product Name               # nvmlDeviceGetName
		3  - GPU UUID                   # nvmlDeviceGetUUID
		4  - Total Memory               # nvmlDeviceGetMemoryInfo
		5  - Power                      # nvmlDeviceGetPowerManagementLimit
		6  - PCIBusID                   # nvmlDeviceGetPciInfo
		7  - BAR1 Memory                # nvmlDeviceGetBAR1MemoryInfo(
		8  - PCI Bandwidth
		9  - Memory, Cores Clock        # nvmlDeviceGetMaxClockInfo
		10 - Display Mode               # nvmlDeviceGetDisplayMode
		11 - Persistence Mode           # nvmlDeviceGetPersistenceMode
	*/

	// Assumed that this method is called with receiver retrieved from
	// NewNvmlClient because this method handles initialization of NVML library

	driverVersion, err := c.driver.SystemDriverVersion()
	if err != nil {
		return nil, fmt.Errorf("nvidia nvml SystemDriverVersion() error: %w", err)
	}

	deviceUUIDs, err := c.driver.ListDeviceUUIDs()
	if err != nil {
		return nil, fmt.Errorf("nvidia nvml ListDeviceUUIDs() error: %w", err)
	}

	allNvidiaGPUResources := make([]*FingerprintDeviceData, 0, len(deviceUUIDs))

	for uuid, mode := range deviceUUIDs {
		// do not care about phsyical parents of MIGs
		if mode == parent {
			continue
		}

		deviceInfo, err := c.driver.DeviceInfoByUUID(uuid)
		if err != nil {
			return nil, fmt.Errorf("nvidia nvml DeviceInfoByUUID() error: %w", err)
		}

		allNvidiaGPUResources = append(allNvidiaGPUResources, &FingerprintDeviceData{
			DeviceData: &DeviceData{
				DeviceName: deviceInfo.Name,
				UUID:       deviceInfo.UUID,
				MemoryMiB:  deviceInfo.MemoryMiB,
				PowerW:     deviceInfo.PowerW,
				BAR1MiB:    deviceInfo.BAR1MiB,
			},
			PCIBandwidthMBPerS: deviceInfo.PCIBandwidthMBPerS,
			CoresClockMHz:      deviceInfo.CoresClockMHz,
			MemoryClockMHz:     deviceInfo.MemoryClockMHz,
			DisplayState:       deviceInfo.DisplayState,
			PersistenceMode:    deviceInfo.PersistenceMode,
			PCIBusID:           deviceInfo.PCIBusID,
		})

		slices.SortFunc(allNvidiaGPUResources, func(a, b *FingerprintDeviceData) int {
			return cmp.Compare(a.UUID, b.UUID)
		})
	}

	return &FingerprintData{
		Devices:       allNvidiaGPUResources,
		DriverVersion: driverVersion,
	}, nil
}

// GetStatsData returns statistics data for all devices on this machine
func (c *nvmlClient) GetStatsData() ([]*StatsData, error) {
	/*
	   nvml fields to be reported to stats api     # nvml_library_call
	   1  - Used Memory                            # nvmlDeviceGetMemoryInfo
	   2  - Utilization of GPU                     # nvmlDeviceGetUtilizationRates
	   3  - Utilization of Memory                  # nvmlDeviceGetUtilizationRates
	   4  - Utilization of Decoder                 # nvmlDeviceGetDecoderUtilization
	   5  - Utilization of Encoder                 # nvmlDeviceGetEncoderUtilization
	   6  - Current GPU Temperature                # nvmlDeviceGetTemperature
	   7  - Power Draw                             # nvmlDeviceGetPowerUsage
	   8  - BAR1 Used memory                       # nvmlDeviceGetBAR1MemoryInfo
	   9  - ECC Errors on requesting L1Cache       # nvmlDeviceGetMemoryErrorCounter
	   10 - ECC Errors on requesting L2Cache       # nvmlDeviceGetMemoryErrorCounter
	   11 - ECC Errors on requesting Device memory # nvmlDeviceGetMemoryErrorCounter
	*/

	// Assumed that this method is called with receiver retrieved from
	// NewNvmlClient because this method handles initialization of NVML library

	deviceUUIDs, err := c.driver.ListDeviceUUIDs()
	if err != nil {
		return nil, fmt.Errorf("nvidia nvml ListDeviceUUIDs() error: %v", err)
	}

	allNvidiaGPUStats := make([]*StatsData, 0, len(deviceUUIDs))

	for uuid, mode := range deviceUUIDs {

		// A30/A100 MIG devices have no stats.
		//
		// https://docs.nvidia.com/datacenter/tesla/mig-user-guide/#telemetry
		//
		// Is this fixed on H100 or later? Maybe?
		if mode == mig || mode == parent {
			continue
		}

		deviceInfo, deviceStatus, err := c.driver.DeviceInfoAndStatusByUUID(uuid)
		if err != nil {
			return nil, fmt.Errorf("nvidia nvml DeviceInfoAndStatusByUUID() error: %v", err)
		}

		allNvidiaGPUStats = append(allNvidiaGPUStats, &StatsData{
			DeviceData: &DeviceData{
				DeviceName: deviceInfo.Name,
				UUID:       deviceInfo.UUID,
				MemoryMiB:  deviceInfo.MemoryMiB,
				PowerW:     deviceInfo.PowerW,
				BAR1MiB:    deviceInfo.BAR1MiB,
			},
			PowerUsageW:        deviceStatus.PowerUsageW,
			GPUUtilization:     deviceStatus.GPUUtilization,
			MemoryUtilization:  deviceStatus.MemoryUtilization,
			EncoderUtilization: deviceStatus.EncoderUtilization,
			DecoderUtilization: deviceStatus.DecoderUtilization,
			TemperatureC:       deviceStatus.TemperatureC,
			UsedMemoryMiB:      deviceStatus.UsedMemoryMiB,
			BAR1UsedMiB:        deviceStatus.BAR1UsedMiB,
			ECCErrorsL1Cache:   deviceStatus.ECCErrorsL1Cache,
			ECCErrorsL2Cache:   deviceStatus.ECCErrorsL2Cache,
			ECCErrorsDevice:    deviceStatus.ECCErrorsDevice,
		})

		slices.SortFunc(allNvidiaGPUStats, func(a, b *StatsData) int {
			return cmp.Compare(a.UUID, b.UUID)
		})
	}
	return allNvidiaGPUStats, nil
}
