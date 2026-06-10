// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package nvml

import (
	"errors"
	"testing"

	"github.com/hashicorp/nomad/plugins/device"
	"github.com/shoenig/test/must"
)

var _ NvmlDriver = (*MockNVMLDriver)(nil)

type MockNVMLDriver struct {
	systemDriverCallSuccessful              bool
	listDeviceUUIDsSuccessful               bool
	deviceInfoByUUIDCallSuccessful          bool
	deviceInfoAndStatusByUUIDCallSuccessful bool
	driverVersion                           string
	devices                                 []*DeviceInfo
	deviceStatus                            []*DeviceStatus
	modes                                   []mode
}

func (m *MockNVMLDriver) Initialize() error {
	return nil
}

func (m *MockNVMLDriver) Shutdown() error {
	return nil
}

func (m *MockNVMLDriver) SystemDriverVersion() (string, error) {
	if !m.systemDriverCallSuccessful {
		return "", errors.New("failed to get system driver")
	}
	return m.driverVersion, nil
}

func (m *MockNVMLDriver) ListDeviceUUIDs() (map[string]mode, error) {
	if !m.listDeviceUUIDsSuccessful {
		return nil, errors.New("failed to get device length")
	}

	allNvidiaGPUUUIDs := make(map[string]mode)

	for i, device := range m.devices {
		allNvidiaGPUUUIDs[device.UUID] = m.modes[i]
	}

	return allNvidiaGPUUUIDs, nil
}

func (m *MockNVMLDriver) DeviceInfoByUUID(uuid string) (*DeviceInfo, error) {
	if !m.deviceInfoByUUIDCallSuccessful {
		return nil, errors.New("failed to get device info by UUID")
	}

	for _, device := range m.devices {
		if uuid == device.UUID {
			return device, nil
		}
	}

	return nil, errors.New("failed to get device handle")
}

func (m *MockNVMLDriver) DeviceInfoAndStatusByUUID(uuid string) (*DeviceInfo, *DeviceStatus, error) {
	if !m.deviceInfoAndStatusByUUIDCallSuccessful {
		return nil, nil, errors.New("failed to get device info and status by index")
	}

	for i, device := range m.devices {
		if uuid == device.UUID {
			return device, m.deviceStatus[i], nil
		}
	}

	return nil, nil, errors.New("failed to get device handle")
}

func TestGetFingerprintDataFromNVML(t *testing.T) {
	for _, testCase := range []struct {
		Name                string
		DriverConfiguration *MockNVMLDriver
		ExpectedError       bool
		ExpectedResult      *FingerprintData
	}{
		{
			Name:           "fail on systemDriverCallSuccessful",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:     false,
				listDeviceUUIDsSuccessful:      true,
				deviceInfoByUUIDCallSuccessful: true,
			},
		},
		{
			Name:           "fail on deviceCountCallSuccessful",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:     true,
				listDeviceUUIDsSuccessful:      false,
				deviceInfoByUUIDCallSuccessful: true,
			},
		},
		{
			Name:           "fail on deviceInfoByUUIDCall",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:     true,
				listDeviceUUIDsSuccessful:      true,
				deviceInfoByUUIDCallSuccessful: false,
				modes:                          []mode{normal, normal},
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               new("ModelName1"),
						MemoryMiB:          new(uint64(16)),
						PCIBusID:           "busId",
						PowerW:             new(uint(100)),
						BAR1MiB:            new(uint64(100)),
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
					}, {
						UUID:               "UUID2",
						Name:               new("ModelName2"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId",
						PowerW:             new(uint(100)),
						BAR1MiB:            new(uint64(100)),
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
					},
				},
			},
		},
		{
			Name:          "successful outcome",
			ExpectedError: false,
			ExpectedResult: &FingerprintData{
				DriverVersion: "driverVersion",
				Devices: []*FingerprintDeviceData{
					{
						DeviceData: &DeviceData{
							DeviceName: new("ModelName1"),
							UUID:       "UUID1",
							MemoryMiB:  new(uint64(16)),
							PowerW:     new(uint(100)),
							BAR1MiB:    new(uint64(100)),
						},
						PCIBusID:           "busId1",
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					}, {
						DeviceData: &DeviceData{
							DeviceName: new("ModelName2"),
							UUID:       "UUID2",
							MemoryMiB:  new(uint64(8)),
							PowerW:     new(uint(200)),
							BAR1MiB:    new(uint64(200)),
						},
						PCIBusID:           "busId2",
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					},
				},
			},
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:     true,
				listDeviceUUIDsSuccessful:      true,
				deviceInfoByUUIDCallSuccessful: true,
				driverVersion:                  "driverVersion",
				modes:                          []mode{normal, normal},
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               new("ModelName1"),
						MemoryMiB:          new(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             new(uint(100)),
						BAR1MiB:            new(uint64(100)),
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					}, {
						UUID:               "UUID2",
						Name:               new("ModelName2"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					},
				},
			},
		},
		{
			Name:          "successful migs",
			ExpectedError: false,
			ExpectedResult: &FingerprintData{
				DriverVersion: "driverVersion",
				Devices: []*FingerprintDeviceData{
					{
						DeviceData: &DeviceData{
							DeviceName: new("ModelName"),
							UUID:       "UUID1",
							MemoryMiB:  new(uint64(16)),
							PowerW:     new(uint(100)),
							BAR1MiB:    new(uint64(100)),
						},
						PCIBusID:           "busId1",
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
						Shared:             device.SharingIneligible,
					},
					{
						DeviceData: &DeviceData{
							DeviceName: new("ModelName"),
							UUID:       "UUID2",
							MemoryMiB:  new(uint64(8)),
							PowerW:     new(uint(200)),
							BAR1MiB:    new(uint64(200)),
						},
						PCIBusID:           "busId2",
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
						Shared:             device.SharingIneligible,
					},
					{
						DeviceData: &DeviceData{
							DeviceName: new("ModelName"),
							UUID:       "UUID4",
							MemoryMiB:  new(uint64(8)),
							PowerW:     new(uint(200)),
							BAR1MiB:    new(uint64(200)),
						},
						PCIBusID:           "busId3",
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
						Shared:             device.SharingIneligible,
					},
				},
			},
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:     true,
				listDeviceUUIDsSuccessful:      true,
				deviceInfoByUUIDCallSuccessful: true,
				driverVersion:                  "driverVersion",
				modes:                          []mode{normal, normal, parent, mig},
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             new(uint(100)),
						BAR1MiB:            new(uint64(100)),
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					},
					{
						UUID:               "UUID2",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					},
					{
						UUID:               "UUID3",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId3",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					},
					{
						UUID:               "UUID4",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId3",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					},
				},
			},
		},
	} {

		t.Run(testCase.Name, func(t *testing.T) {
			cli := nvmlClient{driver: testCase.DriverConfiguration}
			fingerprintData, err := cli.GetFingerprintData()
			if testCase.ExpectedError {
				must.Error(t, err)
			}
			if !testCase.ExpectedError && err != nil {
				must.NoError(t, err)
			}
			must.Eq(t, testCase.ExpectedResult, fingerprintData)
		})
	}
}

func TestGetStatsDataFromNVML(t *testing.T) {
	for _, testCase := range []struct {
		Name                string
		DriverConfiguration *MockNVMLDriver
		ExpectedError       bool
		ExpectedResult      []*StatsData
	}{
		{
			Name:           "fail on listDeviceUUIDsCallSuccessful",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:              true,
				listDeviceUUIDsSuccessful:               false,
				deviceInfoByUUIDCallSuccessful:          true,
				deviceInfoAndStatusByUUIDCallSuccessful: true,
			},
		},
		{
			Name:           "fail on DeviceInfoAndStatusByUUID call",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:              true,
				listDeviceUUIDsSuccessful:               true,
				deviceInfoAndStatusByUUIDCallSuccessful: false,
				modes:                                   []mode{normal, normal},
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               new("ModelName1"),
						MemoryMiB:          new(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             new(uint(100)),
						BAR1MiB:            new(uint64(100)),
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
					}, {
						UUID:               "UUID2",
						Name:               new("ModelName2"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
					},
				},
				deviceStatus: []*DeviceStatus{
					{
						TemperatureC:       new(uint(1)),
						GPUUtilization:     new(uint(1)),
						MemoryUtilization:  new(uint(1)),
						EncoderUtilization: new(uint(1)),
						DecoderUtilization: new(uint(1)),
						UsedMemoryMiB:      new(uint64(1)),
						ECCErrorsL1Cache:   new(uint64(1)),
						ECCErrorsL2Cache:   new(uint64(1)),
						ECCErrorsDevice:    new(uint64(1)),
						PowerUsageW:        new(uint(1)),
						BAR1UsedMiB:        new(uint64(1)),
					},
					{
						TemperatureC:       new(uint(2)),
						GPUUtilization:     new(uint(2)),
						MemoryUtilization:  new(uint(2)),
						EncoderUtilization: new(uint(2)),
						DecoderUtilization: new(uint(2)),
						UsedMemoryMiB:      new(uint64(2)),
						ECCErrorsL1Cache:   new(uint64(2)),
						ECCErrorsL2Cache:   new(uint64(2)),
						ECCErrorsDevice:    new(uint64(2)),
						PowerUsageW:        new(uint(2)),
						BAR1UsedMiB:        new(uint64(2)),
					},
				},
			},
		},
		{
			Name:          "successful outcome",
			ExpectedError: false,
			ExpectedResult: []*StatsData{
				{
					DeviceData: &DeviceData{
						DeviceName: new("ModelName1"),
						UUID:       "UUID1",
						MemoryMiB:  new(uint64(16)),
						PowerW:     new(uint(100)),
						BAR1MiB:    new(uint64(100)),
					},
					TemperatureC:       new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(1)),
					ECCErrorsL2Cache:   new(uint64(1)),
					ECCErrorsDevice:    new(uint64(1)),
					PowerUsageW:        new(uint(1)),
					BAR1UsedMiB:        new(uint64(1)),
				},
				{
					DeviceData: &DeviceData{
						DeviceName: new("ModelName2"),
						UUID:       "UUID2",
						MemoryMiB:  new(uint64(8)),
						PowerW:     new(uint(200)),
						BAR1MiB:    new(uint64(200)),
					},
					TemperatureC:       new(uint(2)),
					GPUUtilization:     new(uint(2)),
					MemoryUtilization:  new(uint(2)),
					EncoderUtilization: new(uint(2)),
					DecoderUtilization: new(uint(2)),
					UsedMemoryMiB:      new(uint64(2)),
					ECCErrorsL1Cache:   new(uint64(2)),
					ECCErrorsL2Cache:   new(uint64(2)),
					ECCErrorsDevice:    new(uint64(2)),
					PowerUsageW:        new(uint(2)),
					BAR1UsedMiB:        new(uint64(2)),
				},
			},
			DriverConfiguration: &MockNVMLDriver{
				listDeviceUUIDsSuccessful:               true,
				deviceInfoByUUIDCallSuccessful:          true,
				deviceInfoAndStatusByUUIDCallSuccessful: true,
				modes:                                   []mode{normal, normal},
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               new("ModelName1"),
						MemoryMiB:          new(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             new(uint(100)),
						BAR1MiB:            new(uint64(100)),
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
					}, {
						UUID:               "UUID2",
						Name:               new("ModelName2"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
					},
				},
				deviceStatus: []*DeviceStatus{
					{
						TemperatureC:       new(uint(1)),
						GPUUtilization:     new(uint(1)),
						MemoryUtilization:  new(uint(1)),
						EncoderUtilization: new(uint(1)),
						DecoderUtilization: new(uint(1)),
						UsedMemoryMiB:      new(uint64(1)),
						ECCErrorsL1Cache:   new(uint64(1)),
						ECCErrorsL2Cache:   new(uint64(1)),
						ECCErrorsDevice:    new(uint64(1)),
						PowerUsageW:        new(uint(1)),
						BAR1UsedMiB:        new(uint64(1)),
					},
					{
						TemperatureC:       new(uint(2)),
						GPUUtilization:     new(uint(2)),
						MemoryUtilization:  new(uint(2)),
						EncoderUtilization: new(uint(2)),
						DecoderUtilization: new(uint(2)),
						UsedMemoryMiB:      new(uint64(2)),
						ECCErrorsL1Cache:   new(uint64(2)),
						ECCErrorsL2Cache:   new(uint64(2)),
						ECCErrorsDevice:    new(uint64(2)),
						PowerUsageW:        new(uint(2)),
						BAR1UsedMiB:        new(uint64(2)),
					},
				},
			},
		},
		{
			Name: "successful migs",
			// stats not available on migs
			ExpectedError: false,
			ExpectedResult: []*StatsData{
				{
					DeviceData: &DeviceData{
						DeviceName: new("ModelName"),
						UUID:       "UUID1",
						MemoryMiB:  new(uint64(16)),
						PowerW:     new(uint(100)),
						BAR1MiB:    new(uint64(100)),
					},
					TemperatureC:       new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(1)),
					ECCErrorsL2Cache:   new(uint64(1)),
					ECCErrorsDevice:    new(uint64(1)),
					PowerUsageW:        new(uint(1)),
					BAR1UsedMiB:        new(uint64(1)),
				},
				{
					DeviceData: &DeviceData{
						DeviceName: new("ModelName"),
						UUID:       "UUID2",
						MemoryMiB:  new(uint64(8)),
						PowerW:     new(uint(200)),
						BAR1MiB:    new(uint64(200)),
					},
					TemperatureC:       new(uint(2)),
					GPUUtilization:     new(uint(2)),
					MemoryUtilization:  new(uint(2)),
					EncoderUtilization: new(uint(2)),
					DecoderUtilization: new(uint(2)),
					UsedMemoryMiB:      new(uint64(2)),
					ECCErrorsL1Cache:   new(uint64(2)),
					ECCErrorsL2Cache:   new(uint64(2)),
					ECCErrorsDevice:    new(uint64(2)),
					PowerUsageW:        new(uint(2)),
					BAR1UsedMiB:        new(uint64(2)),
				},
			},
			DriverConfiguration: &MockNVMLDriver{
				listDeviceUUIDsSuccessful:               true,
				deviceInfoByUUIDCallSuccessful:          true,
				deviceInfoAndStatusByUUIDCallSuccessful: true,
				modes:                                   []mode{normal, normal, parent, mig},
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             new(uint(100)),
						BAR1MiB:            new(uint64(100)),
						PCIBandwidthMBPerS: new(uint(100)),
						CoresClockMHz:      new(uint(100)),
						MemoryClockMHz:     new(uint(100)),
					},
					{
						UUID:               "UUID2",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
					},
					{ // parent, no stats
						UUID:               "UUID3",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId3",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
					},
					{ // mig, no stats
						UUID:               "UUID4",
						Name:               new("ModelName"),
						MemoryMiB:          new(uint64(8)),
						PCIBusID:           "busId3",
						PowerW:             new(uint(200)),
						BAR1MiB:            new(uint64(200)),
						PCIBandwidthMBPerS: new(uint(200)),
						CoresClockMHz:      new(uint(200)),
						MemoryClockMHz:     new(uint(200)),
					},
				},
				deviceStatus: []*DeviceStatus{
					{
						TemperatureC:       new(uint(1)),
						GPUUtilization:     new(uint(1)),
						MemoryUtilization:  new(uint(1)),
						EncoderUtilization: new(uint(1)),
						DecoderUtilization: new(uint(1)),
						UsedMemoryMiB:      new(uint64(1)),
						ECCErrorsL1Cache:   new(uint64(1)),
						ECCErrorsL2Cache:   new(uint64(1)),
						ECCErrorsDevice:    new(uint64(1)),
						PowerUsageW:        new(uint(1)),
						BAR1UsedMiB:        new(uint64(1)),
					},
					{
						TemperatureC:       new(uint(2)),
						GPUUtilization:     new(uint(2)),
						MemoryUtilization:  new(uint(2)),
						EncoderUtilization: new(uint(2)),
						DecoderUtilization: new(uint(2)),
						UsedMemoryMiB:      new(uint64(2)),
						ECCErrorsL1Cache:   new(uint64(2)),
						ECCErrorsL2Cache:   new(uint64(2)),
						ECCErrorsDevice:    new(uint64(2)),
						PowerUsageW:        new(uint(2)),
						BAR1UsedMiB:        new(uint64(2)),
					},
				},
			},
		},
	} {
		cli := nvmlClient{driver: testCase.DriverConfiguration}
		statsData, err := cli.GetStatsData()

		if testCase.ExpectedError {
			must.Error(t, err)
		}
		if !testCase.ExpectedError && err != nil {
			must.NoError(t, err)
		}
		must.Eq(t, testCase.ExpectedResult, statsData)
	}
}
