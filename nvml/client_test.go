// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nvml

import (
	"errors"
	"testing"

	"github.com/hashicorp/nomad/helper"
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
						Name:               helper.StringToPtr("ModelName1"),
						MemoryMiB:          helper.Uint64ToPtr(16),
						PCIBusID:           "busId",
						PowerW:             helper.UintToPtr(100),
						BAR1MiB:            helper.Uint64ToPtr(100),
						PCIBandwidthMBPerS: helper.UintToPtr(100),
						CoresClockMHz:      helper.UintToPtr(100),
						MemoryClockMHz:     helper.UintToPtr(100),
					}, {
						UUID:               "UUID2",
						Name:               helper.StringToPtr("ModelName2"),
						MemoryMiB:          helper.Uint64ToPtr(8),
						PCIBusID:           "busId",
						PowerW:             helper.UintToPtr(100),
						BAR1MiB:            helper.Uint64ToPtr(100),
						PCIBandwidthMBPerS: helper.UintToPtr(100),
						CoresClockMHz:      helper.UintToPtr(100),
						MemoryClockMHz:     helper.UintToPtr(100),
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
							DeviceName: helper.StringToPtr("ModelName1"),
							UUID:       "UUID1",
							MemoryMiB:  helper.Uint64ToPtr(16),
							PowerW:     helper.UintToPtr(100),
							BAR1MiB:    helper.Uint64ToPtr(100),
						},
						PCIBusID:           "busId1",
						PCIBandwidthMBPerS: helper.UintToPtr(100),
						CoresClockMHz:      helper.UintToPtr(100),
						MemoryClockMHz:     helper.UintToPtr(100),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					}, {
						DeviceData: &DeviceData{
							DeviceName: helper.StringToPtr("ModelName2"),
							UUID:       "UUID2",
							MemoryMiB:  helper.Uint64ToPtr(8),
							PowerW:     helper.UintToPtr(200),
							BAR1MiB:    helper.Uint64ToPtr(200),
						},
						PCIBusID:           "busId2",
						PCIBandwidthMBPerS: helper.UintToPtr(200),
						CoresClockMHz:      helper.UintToPtr(200),
						MemoryClockMHz:     helper.UintToPtr(200),
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
						Name:               helper.StringToPtr("ModelName1"),
						MemoryMiB:          helper.Uint64ToPtr(16),
						PCIBusID:           "busId1",
						PowerW:             helper.UintToPtr(100),
						BAR1MiB:            helper.Uint64ToPtr(100),
						PCIBandwidthMBPerS: helper.UintToPtr(100),
						CoresClockMHz:      helper.UintToPtr(100),
						MemoryClockMHz:     helper.UintToPtr(100),
						DisplayState:       "Enabled",
						PersistenceMode:    "Enabled",
					}, {
						UUID:               "UUID2",
						Name:               helper.StringToPtr("ModelName2"),
						MemoryMiB:          helper.Uint64ToPtr(8),
						PCIBusID:           "busId2",
						PowerW:             helper.UintToPtr(200),
						BAR1MiB:            helper.Uint64ToPtr(200),
						PCIBandwidthMBPerS: helper.UintToPtr(200),
						CoresClockMHz:      helper.UintToPtr(200),
						MemoryClockMHz:     helper.UintToPtr(200),
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
						Name:               helper.StringToPtr("ModelName1"),
						MemoryMiB:          helper.Uint64ToPtr(16),
						PCIBusID:           "busId1",
						PowerW:             helper.UintToPtr(100),
						BAR1MiB:            helper.Uint64ToPtr(100),
						PCIBandwidthMBPerS: helper.UintToPtr(100),
						CoresClockMHz:      helper.UintToPtr(100),
						MemoryClockMHz:     helper.UintToPtr(100),
					}, {
						UUID:               "UUID2",
						Name:               helper.StringToPtr("ModelName2"),
						MemoryMiB:          helper.Uint64ToPtr(8),
						PCIBusID:           "busId2",
						PowerW:             helper.UintToPtr(200),
						BAR1MiB:            helper.Uint64ToPtr(200),
						PCIBandwidthMBPerS: helper.UintToPtr(200),
						CoresClockMHz:      helper.UintToPtr(200),
						MemoryClockMHz:     helper.UintToPtr(200),
					},
				},
				deviceStatus: []*DeviceStatus{
					{
						TemperatureC:       helper.UintToPtr(1),
						GPUUtilization:     helper.UintToPtr(1),
						MemoryUtilization:  helper.UintToPtr(1),
						EncoderUtilization: helper.UintToPtr(1),
						DecoderUtilization: helper.UintToPtr(1),
						UsedMemoryMiB:      helper.Uint64ToPtr(1),
						ECCErrorsL1Cache:   helper.Uint64ToPtr(1),
						ECCErrorsL2Cache:   helper.Uint64ToPtr(1),
						ECCErrorsDevice:    helper.Uint64ToPtr(1),
						PowerUsageW:        helper.UintToPtr(1),
						BAR1UsedMiB:        helper.Uint64ToPtr(1),
					},
					{
						TemperatureC:       helper.UintToPtr(2),
						GPUUtilization:     helper.UintToPtr(2),
						MemoryUtilization:  helper.UintToPtr(2),
						EncoderUtilization: helper.UintToPtr(2),
						DecoderUtilization: helper.UintToPtr(2),
						UsedMemoryMiB:      helper.Uint64ToPtr(2),
						ECCErrorsL1Cache:   helper.Uint64ToPtr(2),
						ECCErrorsL2Cache:   helper.Uint64ToPtr(2),
						ECCErrorsDevice:    helper.Uint64ToPtr(2),
						PowerUsageW:        helper.UintToPtr(2),
						BAR1UsedMiB:        helper.Uint64ToPtr(2),
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
						DeviceName: helper.StringToPtr("ModelName1"),
						UUID:       "UUID1",
						MemoryMiB:  helper.Uint64ToPtr(16),
						PowerW:     helper.UintToPtr(100),
						BAR1MiB:    helper.Uint64ToPtr(100),
					},
					TemperatureC:       helper.UintToPtr(1),
					GPUUtilization:     helper.UintToPtr(1),
					MemoryUtilization:  helper.UintToPtr(1),
					EncoderUtilization: helper.UintToPtr(1),
					DecoderUtilization: helper.UintToPtr(1),
					UsedMemoryMiB:      helper.Uint64ToPtr(1),
					ECCErrorsL1Cache:   helper.Uint64ToPtr(1),
					ECCErrorsL2Cache:   helper.Uint64ToPtr(1),
					ECCErrorsDevice:    helper.Uint64ToPtr(1),
					PowerUsageW:        helper.UintToPtr(1),
					BAR1UsedMiB:        helper.Uint64ToPtr(1),
				},
				{
					DeviceData: &DeviceData{
						DeviceName: helper.StringToPtr("ModelName2"),
						UUID:       "UUID2",
						MemoryMiB:  helper.Uint64ToPtr(8),
						PowerW:     helper.UintToPtr(200),
						BAR1MiB:    helper.Uint64ToPtr(200),
					},
					TemperatureC:       helper.UintToPtr(2),
					GPUUtilization:     helper.UintToPtr(2),
					MemoryUtilization:  helper.UintToPtr(2),
					EncoderUtilization: helper.UintToPtr(2),
					DecoderUtilization: helper.UintToPtr(2),
					UsedMemoryMiB:      helper.Uint64ToPtr(2),
					ECCErrorsL1Cache:   helper.Uint64ToPtr(2),
					ECCErrorsL2Cache:   helper.Uint64ToPtr(2),
					ECCErrorsDevice:    helper.Uint64ToPtr(2),
					PowerUsageW:        helper.UintToPtr(2),
					BAR1UsedMiB:        helper.Uint64ToPtr(2),
				},
			},
			DriverConfiguration: &MockNVMLDriver{
				listDeviceUUIDsSuccessful:               true,
				deviceInfoByUUIDCallSuccessful:          true,
				deviceInfoAndStatusByUUIDCallSuccessful: true,
				devices: []*DeviceInfo{
					{
						UUID:               "UUID1",
						Name:               helper.StringToPtr("ModelName1"),
						MemoryMiB:          helper.Uint64ToPtr(16),
						PCIBusID:           "busId1",
						PowerW:             helper.UintToPtr(100),
						BAR1MiB:            helper.Uint64ToPtr(100),
						PCIBandwidthMBPerS: helper.UintToPtr(100),
						CoresClockMHz:      helper.UintToPtr(100),
						MemoryClockMHz:     helper.UintToPtr(100),
					}, {
						UUID:               "UUID2",
						Name:               helper.StringToPtr("ModelName2"),
						MemoryMiB:          helper.Uint64ToPtr(8),
						PCIBusID:           "busId2",
						PowerW:             helper.UintToPtr(200),
						BAR1MiB:            helper.Uint64ToPtr(200),
						PCIBandwidthMBPerS: helper.UintToPtr(200),
						CoresClockMHz:      helper.UintToPtr(200),
						MemoryClockMHz:     helper.UintToPtr(200),
					},
				},
				deviceStatus: []*DeviceStatus{
					{
						TemperatureC:       helper.UintToPtr(1),
						GPUUtilization:     helper.UintToPtr(1),
						MemoryUtilization:  helper.UintToPtr(1),
						EncoderUtilization: helper.UintToPtr(1),
						DecoderUtilization: helper.UintToPtr(1),
						UsedMemoryMiB:      helper.Uint64ToPtr(1),
						ECCErrorsL1Cache:   helper.Uint64ToPtr(1),
						ECCErrorsL2Cache:   helper.Uint64ToPtr(1),
						ECCErrorsDevice:    helper.Uint64ToPtr(1),
						PowerUsageW:        helper.UintToPtr(1),
						BAR1UsedMiB:        helper.Uint64ToPtr(1),
					},
					{
						TemperatureC:       helper.UintToPtr(2),
						GPUUtilization:     helper.UintToPtr(2),
						MemoryUtilization:  helper.UintToPtr(2),
						EncoderUtilization: helper.UintToPtr(2),
						DecoderUtilization: helper.UintToPtr(2),
						UsedMemoryMiB:      helper.Uint64ToPtr(2),
						ECCErrorsL1Cache:   helper.Uint64ToPtr(2),
						ECCErrorsL2Cache:   helper.Uint64ToPtr(2),
						ECCErrorsDevice:    helper.Uint64ToPtr(2),
						PowerUsageW:        helper.UintToPtr(2),
						BAR1UsedMiB:        helper.Uint64ToPtr(2),
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
