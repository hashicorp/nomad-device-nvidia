// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nvml

import (
	"errors"
	"testing"

	"github.com/hashicorp/nomad/helper/pointer"
	"github.com/shoenig/test/must"
)

type MockNVMLDriver struct {
	systemDriverCallSuccessful              bool
	listDeviceUUIDsSuccessful               bool
	deviceInfoByUUIDCallSuccessful          bool
	deviceInfoAndStatusByUUIDCallSuccessful bool
	driverVersion                           string
	devices                                 []*DeviceInfo
	deviceStatus                            []*DeviceStatus
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

func (m *MockNVMLDriver) ListDeviceUUIDs() ([]string, error) {
	if !m.listDeviceUUIDsSuccessful {
		return nil, errors.New("failed to get device length")
	}

	allNvidiaGPUUUIDs := make([]string, len(m.devices))

	for i, device := range m.devices {
		allNvidiaGPUUUIDs[i] = device.UUID
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
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               pointer.Of("ModelName1"),
						MemoryMiB:          pointer.Of(uint64(16)),
						PCIBusID:           "busId",
						PowerW:             pointer.Of(uint(100)),
						BAR1MiB:            pointer.Of(uint64(100)),
						PCIBandwidthMBPerS: pointer.Of(uint(100)),
						CoresClockMHz:      pointer.Of(uint(100)),
						MemoryClockMHz:     pointer.Of(uint(100)),
					}, {
						UUID:               "UUID2",
						Name:               pointer.Of("ModelName2"),
						MemoryMiB:          pointer.Of(uint64(8)),
						PCIBusID:           "busId",
						PowerW:             pointer.Of(uint(100)),
						BAR1MiB:            pointer.Of(uint64(100)),
						PCIBandwidthMBPerS: pointer.Of(uint(100)),
						CoresClockMHz:      pointer.Of(uint(100)),
						MemoryClockMHz:     pointer.Of(uint(100)),
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
							DeviceName: pointer.Of("ModelName1"),
							UUID:       "UUID1",
							MemoryMiB:  pointer.Of(uint64(16)),
							PowerW:     pointer.Of(uint(100)),
							BAR1MiB:    pointer.Of(uint64(100)),
						},
						PCIBusID:           "busId1",
						PCIBandwidthMBPerS: pointer.Of(uint(100)),
						CoresClockMHz:      pointer.Of(uint(100)),
						MemoryClockMHz:     pointer.Of(uint(100)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					}, {
						DeviceData: &DeviceData{
							DeviceName: pointer.Of("ModelName2"),
							UUID:       "UUID2",
							MemoryMiB:  pointer.Of(uint64(8)),
							PowerW:     pointer.Of(uint(200)),
							BAR1MiB:    pointer.Of(uint64(200)),
						},
						PCIBusID:           "busId2",
						PCIBandwidthMBPerS: pointer.Of(uint(200)),
						CoresClockMHz:      pointer.Of(uint(200)),
						MemoryClockMHz:     pointer.Of(uint(200)),
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
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               pointer.Of("ModelName1"),
						MemoryMiB:          pointer.Of(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             pointer.Of(uint(100)),
						BAR1MiB:            pointer.Of(uint64(100)),
						PCIBandwidthMBPerS: pointer.Of(uint(100)),
						CoresClockMHz:      pointer.Of(uint(100)),
						MemoryClockMHz:     pointer.Of(uint(100)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					}, {
						UUID:               "UUID2",
						Name:               pointer.Of("ModelName2"),
						MemoryMiB:          pointer.Of(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             pointer.Of(uint(200)),
						BAR1MiB:            pointer.Of(uint64(200)),
						PCIBandwidthMBPerS: pointer.Of(uint(200)),
						CoresClockMHz:      pointer.Of(uint(200)),
						MemoryClockMHz:     pointer.Of(uint(200)),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					},
				},
			},
		},
	} {
		cli := nvmlClient{driver: testCase.DriverConfiguration}
		fingerprintData, err := cli.GetFingerprintData()
		if testCase.ExpectedError {
			must.Error(t, err)
		}
		if !testCase.ExpectedError && err != nil {
			must.NoError(t, err)
		}
		must.Eq(t, testCase.ExpectedResult, fingerprintData)
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
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               pointer.Of("ModelName1"),
						MemoryMiB:          pointer.Of(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             pointer.Of(uint(100)),
						BAR1MiB:            pointer.Of(uint64(100)),
						PCIBandwidthMBPerS: pointer.Of(uint(100)),
						CoresClockMHz:      pointer.Of(uint(100)),
						MemoryClockMHz:     pointer.Of(uint(100)),
					}, {
						UUID:               "UUID2",
						Name:               pointer.Of("ModelName2"),
						MemoryMiB:          pointer.Of(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             pointer.Of(uint(200)),
						BAR1MiB:            pointer.Of(uint64(200)),
						PCIBandwidthMBPerS: pointer.Of(uint(200)),
						CoresClockMHz:      pointer.Of(uint(200)),
						MemoryClockMHz:     pointer.Of(uint(200)),
					},
				},
				deviceStatus: []*DeviceStatus{
					{
						TemperatureC:       pointer.Of(uint(1)),
						GPUUtilization:     pointer.Of(uint(1)),
						MemoryUtilization:  pointer.Of(uint(1)),
						EncoderUtilization: pointer.Of(uint(1)),
						DecoderUtilization: pointer.Of(uint(1)),
						UsedMemoryMiB:      pointer.Of(uint64(1)),
						ECCErrorsL1Cache:   pointer.Of(uint64(1)),
						ECCErrorsL2Cache:   pointer.Of(uint64(1)),
						ECCErrorsDevice:    pointer.Of(uint64(1)),
						PowerUsageW:        pointer.Of(uint(1)),
						BAR1UsedMiB:        pointer.Of(uint64(1)),
					},
					{
						TemperatureC:       pointer.Of(uint(2)),
						GPUUtilization:     pointer.Of(uint(2)),
						MemoryUtilization:  pointer.Of(uint(2)),
						EncoderUtilization: pointer.Of(uint(2)),
						DecoderUtilization: pointer.Of(uint(2)),
						UsedMemoryMiB:      pointer.Of(uint64(2)),
						ECCErrorsL1Cache:   pointer.Of(uint64(2)),
						ECCErrorsL2Cache:   pointer.Of(uint64(2)),
						ECCErrorsDevice:    pointer.Of(uint64(2)),
						PowerUsageW:        pointer.Of(uint(2)),
						BAR1UsedMiB:        pointer.Of(uint64(2)),
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
						DeviceName: pointer.Of("ModelName1"),
						UUID:       "UUID1",
						MemoryMiB:  pointer.Of(uint64(16)),
						PowerW:     pointer.Of(uint(100)),
						BAR1MiB:    pointer.Of(uint64(100)),
					},
					TemperatureC:       pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(1)),
					ECCErrorsL2Cache:   pointer.Of(uint64(1)),
					ECCErrorsDevice:    pointer.Of(uint64(1)),
					PowerUsageW:        pointer.Of(uint(1)),
					BAR1UsedMiB:        pointer.Of(uint64(1)),
				},
				{
					DeviceData: &DeviceData{
						DeviceName: pointer.Of("ModelName2"),
						UUID:       "UUID2",
						MemoryMiB:  pointer.Of(uint64(8)),
						PowerW:     pointer.Of(uint(200)),
						BAR1MiB:    pointer.Of(uint64(200)),
					},
					TemperatureC:       pointer.Of(uint(2)),
					GPUUtilization:     pointer.Of(uint(2)),
					MemoryUtilization:  pointer.Of(uint(2)),
					EncoderUtilization: pointer.Of(uint(2)),
					DecoderUtilization: pointer.Of(uint(2)),
					UsedMemoryMiB:      pointer.Of(uint64(2)),
					ECCErrorsL1Cache:   pointer.Of(uint64(2)),
					ECCErrorsL2Cache:   pointer.Of(uint64(2)),
					ECCErrorsDevice:    pointer.Of(uint64(2)),
					PowerUsageW:        pointer.Of(uint(2)),
					BAR1UsedMiB:        pointer.Of(uint64(2)),
				},
			},
			DriverConfiguration: &MockNVMLDriver{
				listDeviceUUIDsSuccessful:               true,
				deviceInfoByUUIDCallSuccessful:          true,
				deviceInfoAndStatusByUUIDCallSuccessful: true,
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               pointer.Of("ModelName1"),
						MemoryMiB:          pointer.Of(uint64(16)),
						PCIBusID:           "busId1",
						PowerW:             pointer.Of(uint(100)),
						BAR1MiB:            pointer.Of(uint64(100)),
						PCIBandwidthMBPerS: pointer.Of(uint(100)),
						CoresClockMHz:      pointer.Of(uint(100)),
						MemoryClockMHz:     pointer.Of(uint(100)),
					}, {
						UUID:               "UUID2",
						Name:               pointer.Of("ModelName2"),
						MemoryMiB:          pointer.Of(uint64(8)),
						PCIBusID:           "busId2",
						PowerW:             pointer.Of(uint(200)),
						BAR1MiB:            pointer.Of(uint64(200)),
						PCIBandwidthMBPerS: pointer.Of(uint(200)),
						CoresClockMHz:      pointer.Of(uint(200)),
						MemoryClockMHz:     pointer.Of(uint(200)),
					},
				},
				deviceStatus: []*DeviceStatus{
					{
						TemperatureC:       pointer.Of(uint(1)),
						GPUUtilization:     pointer.Of(uint(1)),
						MemoryUtilization:  pointer.Of(uint(1)),
						EncoderUtilization: pointer.Of(uint(1)),
						DecoderUtilization: pointer.Of(uint(1)),
						UsedMemoryMiB:      pointer.Of(uint64(1)),
						ECCErrorsL1Cache:   pointer.Of(uint64(1)),
						ECCErrorsL2Cache:   pointer.Of(uint64(1)),
						ECCErrorsDevice:    pointer.Of(uint64(1)),
						PowerUsageW:        pointer.Of(uint(1)),
						BAR1UsedMiB:        pointer.Of(uint64(1)),
					},
					{
						TemperatureC:       pointer.Of(uint(2)),
						GPUUtilization:     pointer.Of(uint(2)),
						MemoryUtilization:  pointer.Of(uint(2)),
						EncoderUtilization: pointer.Of(uint(2)),
						DecoderUtilization: pointer.Of(uint(2)),
						UsedMemoryMiB:      pointer.Of(uint64(2)),
						ECCErrorsL1Cache:   pointer.Of(uint64(2)),
						ECCErrorsL2Cache:   pointer.Of(uint64(2)),
						ECCErrorsDevice:    pointer.Of(uint64(2)),
						PowerUsageW:        pointer.Of(uint(2)),
						BAR1UsedMiB:        pointer.Of(uint64(2)),
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
