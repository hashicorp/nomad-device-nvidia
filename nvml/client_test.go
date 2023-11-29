// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nvml

import (
	"errors"
	"github.com/hashicorp/nomad/helper/pointer"
	"testing"

	"github.com/shoenig/test/must"
)

type MockNVMLDriver struct {
	systemDriverCallSuccessful               bool
	deviceCountCallSuccessful                bool
	deviceInfoByIndexCallSuccessful          bool
	deviceInfoAndStatusByIndexCallSuccessful bool
	driverVersion                            string
	devices                                  []*DeviceInfo
	deviceStatus                             []*DeviceStatus
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

func (m *MockNVMLDriver) DeviceCount() (uint, error) {
	if !m.deviceCountCallSuccessful {
		return 0, errors.New("failed to get device length")
	}
	return uint(len(m.devices)), nil
}

func (m *MockNVMLDriver) DeviceInfoByIndex(index uint) (*DeviceInfo, error) {
	if index >= uint(len(m.devices)) {
		return nil, errors.New("index is out of range")
	}
	if !m.deviceInfoByIndexCallSuccessful {
		return nil, errors.New("failed to get device info by index")
	}
	return m.devices[index], nil
}

func (m *MockNVMLDriver) DeviceInfoAndStatusByIndex(index uint) (*DeviceInfo, *DeviceStatus, error) {
	if index >= uint(len(m.devices)) || index >= uint(len(m.deviceStatus)) {
		return nil, nil, errors.New("index is out of range")
	}
	if !m.deviceInfoAndStatusByIndexCallSuccessful {
		return nil, nil, errors.New("failed to get device info and status by index")
	}
	return m.devices[index], m.deviceStatus[index], nil
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
				systemDriverCallSuccessful:      false,
				deviceCountCallSuccessful:       true,
				deviceInfoByIndexCallSuccessful: true,
			},
		},
		{
			Name:           "fail on deviceCountCallSuccessful",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:      true,
				deviceCountCallSuccessful:       false,
				deviceInfoByIndexCallSuccessful: true,
			},
		},
		{
			Name:           "fail on deviceInfoByIndexCall",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:      true,
				deviceCountCallSuccessful:       true,
				deviceInfoByIndexCallSuccessful: false,
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
				systemDriverCallSuccessful:      true,
				deviceCountCallSuccessful:       true,
				deviceInfoByIndexCallSuccessful: true,
				driverVersion:                   "driverVersion",
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
			Name:           "fail on deviceCountCallSuccessful",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:               true,
				deviceCountCallSuccessful:                false,
				deviceInfoByIndexCallSuccessful:          true,
				deviceInfoAndStatusByIndexCallSuccessful: true,
			},
		},
		{
			Name:           "fail on DeviceInfoAndStatusByIndex call",
			ExpectedError:  true,
			ExpectedResult: nil,
			DriverConfiguration: &MockNVMLDriver{
				systemDriverCallSuccessful:               true,
				deviceCountCallSuccessful:                true,
				deviceInfoAndStatusByIndexCallSuccessful: false,
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
				deviceCountCallSuccessful:                true,
				deviceInfoByIndexCallSuccessful:          true,
				deviceInfoAndStatusByIndexCallSuccessful: true,
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
