// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package nvidia

import (
	"context"
	"errors"
	"sort"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-device-nvidia/nvml"
	"github.com/hashicorp/nomad/helper/pointer"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
	"github.com/shoenig/test/must"
)

func TestIgnoreFingerprintedDevices(t *testing.T) {
	for _, testCase := range []struct {
		Name           string
		DeviceData     []*nvml.FingerprintDeviceData
		IgnoredGPUIds  map[string]struct{}
		ExpectedResult []*nvml.FingerprintDeviceData
	}{
		{
			Name: "Odd ignored",
			DeviceData: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName1"),
						UUID:       "UUID1",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName2"),
						UUID:       "UUID2",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName3"),
						UUID:       "UUID3",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
			},
			IgnoredGPUIds: map[string]struct{}{
				"UUID2": {},
			},
			ExpectedResult: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName1"),
						UUID:       "UUID1",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName3"),
						UUID:       "UUID3",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
			},
		},
		{
			Name: "Even ignored",
			DeviceData: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName1"),
						UUID:       "UUID1",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName2"),
						UUID:       "UUID2",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName3"),
						UUID:       "UUID3",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
			},
			IgnoredGPUIds: map[string]struct{}{
				"UUID1": {},
				"UUID3": {},
			},
			ExpectedResult: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName2"),
						UUID:       "UUID2",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
			},
		},
		{
			Name: "All ignored",
			DeviceData: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName1"),
						UUID:       "UUID1",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName2"),
						UUID:       "UUID2",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName3"),
						UUID:       "UUID3",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
			},
			IgnoredGPUIds: map[string]struct{}{
				"UUID1": {},
				"UUID2": {},
				"UUID3": {},
			},
			ExpectedResult: nil,
		},
		{
			Name: "No ignored",
			DeviceData: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName1"),
						UUID:       "UUID1",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName2"),
						UUID:       "UUID2",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName3"),
						UUID:       "UUID3",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
			},
			IgnoredGPUIds: map[string]struct{}{},
			ExpectedResult: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName1"),
						UUID:       "UUID1",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName2"),
						UUID:       "UUID2",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						DeviceName: pointer.Of("DeviceName3"),
						UUID:       "UUID3",
						MemoryMiB:  pointer.Of(uint64(1000)),
					},
				},
			},
		},
		{
			Name:       "No DeviceData provided",
			DeviceData: nil,
			IgnoredGPUIds: map[string]struct{}{
				"UUID1": {},
				"UUID2": {},
				"UUID3": {},
			},
			ExpectedResult: nil,
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			actualResult := ignoreFingerprintedDevices(testCase.DeviceData, testCase.IgnoredGPUIds)
			must.Eq(t, testCase.ExpectedResult, actualResult)
		})
	}
}

func TestCheckFingerprintUpdates(t *testing.T) {
	for _, testCase := range []struct {
		Name                     string
		Device                   *NvidiaDevice
		AllDevices               []*nvml.FingerprintDeviceData
		DeviceMapAfterMethodCall map[string]struct{}
		ExpectedResult           bool
	}{
		{
			Name: "No updates",
			Device: &NvidiaDevice{devices: map[string]struct{}{
				"1": {},
				"2": {},
				"3": {},
			}},
			AllDevices: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						UUID: "1",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "2",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "3",
					},
				},
			},
			ExpectedResult: false,
			DeviceMapAfterMethodCall: map[string]struct{}{
				"1": {},
				"2": {},
				"3": {},
			},
		},
		{
			Name: "New Device Appeared",
			Device: &NvidiaDevice{devices: map[string]struct{}{
				"1": {},
				"2": {},
				"3": {},
			}},
			AllDevices: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						UUID: "1",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "2",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "3",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "I am new",
					},
				},
			},
			ExpectedResult: true,
			DeviceMapAfterMethodCall: map[string]struct{}{
				"1":        {},
				"2":        {},
				"3":        {},
				"I am new": {},
			},
		},
		{
			Name: "Device disappeared",
			Device: &NvidiaDevice{devices: map[string]struct{}{
				"1": {},
				"2": {},
				"3": {},
			}},
			AllDevices: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						UUID: "1",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "2",
					},
				},
			},
			ExpectedResult: true,
			DeviceMapAfterMethodCall: map[string]struct{}{
				"1": {},
				"2": {},
			},
		},
		{
			Name:   "No devices in NvidiaDevice map",
			Device: &NvidiaDevice{},
			AllDevices: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						UUID: "1",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "2",
					},
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID: "3",
					},
				},
			},
			ExpectedResult: true,
			DeviceMapAfterMethodCall: map[string]struct{}{
				"1": {},
				"2": {},
				"3": {},
			},
		},
		{
			Name: "No devices detected",
			Device: &NvidiaDevice{devices: map[string]struct{}{
				"1": {},
				"2": {},
				"3": {},
			}},
			AllDevices:               nil,
			ExpectedResult:           true,
			DeviceMapAfterMethodCall: map[string]struct{}{},
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			actualResult := testCase.Device.fingerprintChanged(testCase.AllDevices)
			// check that function returns valid "updated / not updated" state
			must.Eq(t, testCase.ExpectedResult, actualResult)
			// check that function propely updates devices map
			must.Eq(t, testCase.Device.devices, testCase.DeviceMapAfterMethodCall)
		})
	}
}

func TestAttributesFromFingerprintDeviceData(t *testing.T) {
	for _, testCase := range []struct {
		Name                  string
		FingerprintDeviceData *nvml.FingerprintDeviceData
		ExpectedResult        map[string]*structs.Attribute
	}{
		{
			Name: "All attributes are not nil",
			FingerprintDeviceData: &nvml.FingerprintDeviceData{
				DeviceData: &nvml.DeviceData{
					UUID:       "1",
					DeviceName: pointer.Of("Type1"),
					MemoryMiB:  pointer.Of(uint64(256)),
					PowerW:     pointer.Of(uint(2)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PCIBusID:           "pciBusID1",
				PCIBandwidthMBPerS: pointer.Of(uint(1)),
				CoresClockMHz:      pointer.Of(uint(1)),
				MemoryClockMHz:     pointer.Of(uint(1)),
				DisplayState:       "Enabled",
				PersistenceMode:    "Enabled",
			},
			ExpectedResult: map[string]*structs.Attribute{
				MemoryAttr: {
					Int:  pointer.Of(int64(256)),
					Unit: structs.UnitMiB,
				},
				PowerAttr: {
					Int:  pointer.Of(int64(2)),
					Unit: structs.UnitW,
				},
				BAR1Attr: {
					Int:  pointer.Of(int64(256)),
					Unit: structs.UnitMiB,
				},
				PCIBandwidthAttr: {
					Int:  pointer.Of(int64(1)),
					Unit: structs.UnitMBPerS,
				},
				CoresClockAttr: {
					Int:  pointer.Of(int64(1)),
					Unit: structs.UnitMHz,
				},
				MemoryClockAttr: {
					Int:  pointer.Of(int64(1)),
					Unit: structs.UnitMHz,
				},
				DisplayStateAttr: {
					String: pointer.Of("Enabled"),
				},
				PersistenceModeAttr: {
					String: pointer.Of("Enabled"),
				},
			},
		},
		{
			Name: "nil values are omitted",
			FingerprintDeviceData: &nvml.FingerprintDeviceData{
				DeviceData: &nvml.DeviceData{
					UUID:       "1",
					DeviceName: pointer.Of("Type1"),
					MemoryMiB:  nil,
					PowerW:     pointer.Of(uint(2)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PCIBusID:        "pciBusID1",
				DisplayState:    "Enabled",
				PersistenceMode: "Enabled",
			},
			ExpectedResult: map[string]*structs.Attribute{
				PowerAttr: {
					Int:  pointer.Of(int64(2)),
					Unit: structs.UnitW,
				},
				BAR1Attr: {
					Int:  pointer.Of(int64(256)),
					Unit: structs.UnitMiB,
				},
				DisplayStateAttr: {
					String: pointer.Of("Enabled"),
				},
				PersistenceModeAttr: {
					String: pointer.Of("Enabled"),
				},
			},
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			actualResult := attributesFromFingerprintDeviceData(testCase.FingerprintDeviceData)
			must.Eq(t, testCase.ExpectedResult, actualResult)
		})
	}
}

func TestDeviceGroupFromFingerprintData(t *testing.T) {
	for _, testCase := range []struct {
		Name             string
		GroupName        string
		Devices          []*nvml.FingerprintDeviceData
		CommonAttributes map[string]*structs.Attribute
		ExpectedResult   *device.DeviceGroup
	}{
		{
			Name:      "Devices are provided",
			GroupName: "Type1",
			Devices: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "1",
						DeviceName: pointer.Of("Type1"),
						MemoryMiB:  pointer.Of(uint64(100)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PCIBusID:           "pciBusID1",
					PCIBandwidthMBPerS: pointer.Of(uint(1)),
					CoresClockMHz:      pointer.Of(uint(1)),
					MemoryClockMHz:     pointer.Of(uint(1)),
					DisplayState:       "Enabled",
					PersistenceMode:    "Enabled",
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "2",
						DeviceName: pointer.Of("Type1"),
						MemoryMiB:  pointer.Of(uint64(100)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PCIBusID:           "pciBusID2",
					PCIBandwidthMBPerS: pointer.Of(uint(1)),
					CoresClockMHz:      pointer.Of(uint(1)),
					MemoryClockMHz:     pointer.Of(uint(1)),
					DisplayState:       "Enabled",
					PersistenceMode:    "Enabled",
				},
			},
			ExpectedResult: &device.DeviceGroup{
				Vendor: vendor,
				Type:   deviceType,
				Name:   "Type1",
				Devices: []*device.Device{
					{
						ID:      "1",
						Healthy: true,
						HwLocality: &device.DeviceLocality{
							PciBusID: "pciBusID1",
						},
					},
					{
						ID:      "2",
						Healthy: true,
						HwLocality: &device.DeviceLocality{
							PciBusID: "pciBusID2",
						},
					},
				},
				Attributes: map[string]*structs.Attribute{
					MemoryAttr: {
						Int:  pointer.Of(int64(100)),
						Unit: structs.UnitMiB,
					},
					PowerAttr: {
						Int:  pointer.Of(int64(2)),
						Unit: structs.UnitW,
					},
					BAR1Attr: {
						Int:  pointer.Of(int64(256)),
						Unit: structs.UnitMiB,
					},
					PCIBandwidthAttr: {
						Int:  pointer.Of(int64(1)),
						Unit: structs.UnitMBPerS,
					},
					CoresClockAttr: {
						Int:  pointer.Of(int64(1)),
						Unit: structs.UnitMHz,
					},
					MemoryClockAttr: {
						Int:  pointer.Of(int64(1)),
						Unit: structs.UnitMHz,
					},
					DisplayStateAttr: {
						String: pointer.Of("Enabled"),
					},
					PersistenceModeAttr: {
						String: pointer.Of("Enabled"),
					},
				},
			},
		},
		{
			Name:      "Devices and common attributes are provided",
			GroupName: "Type1",
			Devices: []*nvml.FingerprintDeviceData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "1",
						DeviceName: pointer.Of("Type1"),
						MemoryMiB:  pointer.Of(uint64(100)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PCIBusID:           "pciBusID1",
					PCIBandwidthMBPerS: pointer.Of(uint(1)),
					CoresClockMHz:      pointer.Of(uint(1)),
					MemoryClockMHz:     pointer.Of(uint(1)),
					DisplayState:       "Enabled",
					PersistenceMode:    "Enabled",
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "2",
						DeviceName: pointer.Of("Type1"),
						MemoryMiB:  pointer.Of(uint64(100)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PCIBusID:           "pciBusID2",
					PCIBandwidthMBPerS: pointer.Of(uint(1)),
					CoresClockMHz:      pointer.Of(uint(1)),
					MemoryClockMHz:     pointer.Of(uint(1)),
					DisplayState:       "Enabled",
					PersistenceMode:    "Enabled",
				},
			},
			CommonAttributes: map[string]*structs.Attribute{
				DriverVersionAttr: {
					String: pointer.Of("1"),
				},
			},
			ExpectedResult: &device.DeviceGroup{
				Vendor: vendor,
				Type:   deviceType,
				Name:   "Type1",
				Devices: []*device.Device{
					{
						ID:      "1",
						Healthy: true,
						HwLocality: &device.DeviceLocality{
							PciBusID: "pciBusID1",
						},
					},
					{
						ID:      "2",
						Healthy: true,
						HwLocality: &device.DeviceLocality{
							PciBusID: "pciBusID2",
						},
					},
				},
				Attributes: map[string]*structs.Attribute{
					MemoryAttr: {
						Int:  pointer.Of(int64(100)),
						Unit: structs.UnitMiB,
					},
					PowerAttr: {
						Int:  pointer.Of(int64(2)),
						Unit: structs.UnitW,
					},
					BAR1Attr: {
						Int:  pointer.Of(int64(256)),
						Unit: structs.UnitMiB,
					},
					PCIBandwidthAttr: {
						Int:  pointer.Of(int64(1)),
						Unit: structs.UnitMBPerS,
					},
					CoresClockAttr: {
						Int:  pointer.Of(int64(1)),
						Unit: structs.UnitMHz,
					},
					MemoryClockAttr: {
						Int:  pointer.Of(int64(1)),
						Unit: structs.UnitMHz,
					},
					DisplayStateAttr: {
						String: pointer.Of("Enabled"),
					},
					PersistenceModeAttr: {
						String: pointer.Of("Enabled"),
					},
					DriverVersionAttr: {
						String: pointer.Of("1"),
					},
				},
			},
		},
		{
			Name:      "Devices are not provided",
			GroupName: "Type1",
			CommonAttributes: map[string]*structs.Attribute{
				DriverVersionAttr: {
					String: pointer.Of("1"),
				},
			},
			Devices:        nil,
			ExpectedResult: nil,
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			actualResult := deviceGroupFromFingerprintData(testCase.GroupName, testCase.Devices, testCase.CommonAttributes)
			must.Eq(t, testCase.ExpectedResult, actualResult)
		})
	}
}

func TestWriteFingerprintToChannel(t *testing.T) {
	for _, testCase := range []struct {
		Name                   string
		Device                 *NvidiaDevice
		ExpectedWriteToChannel *device.FingerprintResponse
	}{
		{
			Name: "Check that FingerprintError is handled properly",
			Device: &NvidiaDevice{
				nvmlClient: &MockNvmlClient{
					FingerprintError: errors.New(""),
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.FingerprintResponse{
				Error: errors.New(""),
			},
		},
		{
			Name: "Check ignore devices works correctly",
			Device: &NvidiaDevice{
				nvmlClient: &MockNvmlClient{
					FingerprintResponseReturned: &nvml.FingerprintData{
						DriverVersion: "1",
						Devices: []*nvml.FingerprintDeviceData{
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "1",
									DeviceName: pointer.Of("Name"),
									MemoryMiB:  pointer.Of(uint64(10)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID1",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "2",
									DeviceName: pointer.Of("Name"),
									MemoryMiB:  pointer.Of(uint64(10)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID2",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
						},
					},
				},
				ignoredGPUIDs: map[string]struct{}{
					"1": {},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.FingerprintResponse{
				Devices: []*device.DeviceGroup{
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "Name",
						Devices: []*device.Device{
							{
								ID:      "2",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID2",
								},
							},
						},
						Attributes: map[string]*structs.Attribute{
							MemoryAttr: {
								Int:  pointer.Of(int64(10)),
								Unit: structs.UnitMiB,
							},
							PowerAttr: {
								Int:  pointer.Of(int64(100)),
								Unit: structs.UnitW,
							},
							BAR1Attr: {
								Int:  pointer.Of(int64(256)),
								Unit: structs.UnitMiB,
							},
							PCIBandwidthAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMBPerS,
							},
							CoresClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							MemoryClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							DisplayStateAttr: {
								String: pointer.Of("Enabled"),
							},
							PersistenceModeAttr: {
								String: pointer.Of("Enabled"),
							},
							DriverVersionAttr: {
								String: pointer.Of("1"),
							},
						},
					},
				},
			},
		},
		{
			Name: "Check devices are split to multiple device groups 1",
			Device: &NvidiaDevice{
				nvmlClient: &MockNvmlClient{
					FingerprintResponseReturned: &nvml.FingerprintData{
						DriverVersion: "1",
						Devices: []*nvml.FingerprintDeviceData{
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "1",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID1",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "2",
									DeviceName: pointer.Of("Name2"),
									MemoryMiB:  pointer.Of(uint64(11)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID2",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "3",
									DeviceName: pointer.Of("Name3"),
									MemoryMiB:  pointer.Of(uint64(12)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID3",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
						},
					},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.FingerprintResponse{
				Devices: []*device.DeviceGroup{
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "Name1",
						Devices: []*device.Device{
							{
								ID:      "1",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID1",
								},
							},
						},
						Attributes: map[string]*structs.Attribute{
							MemoryAttr: {
								Int:  pointer.Of(int64(10)),
								Unit: structs.UnitMiB,
							},
							PowerAttr: {
								Int:  pointer.Of(int64(100)),
								Unit: structs.UnitW,
							},
							BAR1Attr: {
								Int:  pointer.Of(int64(256)),
								Unit: structs.UnitMiB,
							},
							PCIBandwidthAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMBPerS,
							},
							CoresClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							MemoryClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							DisplayStateAttr: {
								String: pointer.Of("Enabled"),
							},
							PersistenceModeAttr: {
								String: pointer.Of("Enabled"),
							},
							DriverVersionAttr: {
								String: pointer.Of("1"),
							},
						},
					},
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "Name2",
						Devices: []*device.Device{
							{
								ID:      "2",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID2",
								},
							},
						},
						Attributes: map[string]*structs.Attribute{
							MemoryAttr: {
								Int:  pointer.Of(int64(11)),
								Unit: structs.UnitMiB,
							},
							PowerAttr: {
								Int:  pointer.Of(int64(100)),
								Unit: structs.UnitW,
							},
							BAR1Attr: {
								Int:  pointer.Of(int64(256)),
								Unit: structs.UnitMiB,
							},
							PCIBandwidthAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMBPerS,
							},
							CoresClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							MemoryClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							DisplayStateAttr: {
								String: pointer.Of("Enabled"),
							},
							PersistenceModeAttr: {
								String: pointer.Of("Enabled"),
							},
							DriverVersionAttr: {
								String: pointer.Of("1"),
							},
						},
					},
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "Name3",
						Devices: []*device.Device{
							{
								ID:      "3",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID3",
								},
							},
						},
						Attributes: map[string]*structs.Attribute{
							MemoryAttr: {
								Int:  pointer.Of(int64(12)),
								Unit: structs.UnitMiB,
							},
							PowerAttr: {
								Int:  pointer.Of(int64(100)),
								Unit: structs.UnitW,
							},
							BAR1Attr: {
								Int:  pointer.Of(int64(256)),
								Unit: structs.UnitMiB,
							},
							PCIBandwidthAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMBPerS,
							},
							CoresClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							MemoryClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							DisplayStateAttr: {
								String: pointer.Of("Enabled"),
							},
							PersistenceModeAttr: {
								String: pointer.Of("Enabled"),
							},
							DriverVersionAttr: {
								String: pointer.Of("1"),
							},
						},
					},
				},
			},
		},
		{
			Name: "Check devices are split to multiple device groups 2",
			Device: &NvidiaDevice{
				nvmlClient: &MockNvmlClient{
					FingerprintResponseReturned: &nvml.FingerprintData{
						DriverVersion: "1",
						Devices: []*nvml.FingerprintDeviceData{
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "1",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID1",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "2",
									DeviceName: pointer.Of("Name2"),
									MemoryMiB:  pointer.Of(uint64(11)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID2",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "3",
									DeviceName: pointer.Of("Name2"),
									MemoryMiB:  pointer.Of(uint64(12)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID3",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
						},
					},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.FingerprintResponse{
				Devices: []*device.DeviceGroup{
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "Name1",
						Devices: []*device.Device{
							{
								ID:      "1",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID1",
								},
							},
						},
						Attributes: map[string]*structs.Attribute{
							MemoryAttr: {
								Int:  pointer.Of(int64(10)),
								Unit: structs.UnitMiB,
							},
							PowerAttr: {
								Int:  pointer.Of(int64(100)),
								Unit: structs.UnitW,
							},
							BAR1Attr: {
								Int:  pointer.Of(int64(256)),
								Unit: structs.UnitMiB,
							},
							PCIBandwidthAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMBPerS,
							},
							CoresClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							MemoryClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							DisplayStateAttr: {
								String: pointer.Of("Enabled"),
							},
							PersistenceModeAttr: {
								String: pointer.Of("Enabled"),
							},
							DriverVersionAttr: {
								String: pointer.Of("1"),
							},
						},
					},
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "Name2",
						Devices: []*device.Device{
							{
								ID:      "2",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID2",
								},
							},
							{
								ID:      "3",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID3",
								},
							},
						},
						Attributes: map[string]*structs.Attribute{
							MemoryAttr: {
								Int:  pointer.Of(int64(11)),
								Unit: structs.UnitMiB,
							},
							PowerAttr: {
								Int:  pointer.Of(int64(100)),
								Unit: structs.UnitW,
							},
							BAR1Attr: {
								Int:  pointer.Of(int64(256)),
								Unit: structs.UnitMiB,
							},
							PCIBandwidthAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMBPerS,
							},
							CoresClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							MemoryClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							DisplayStateAttr: {
								String: pointer.Of("Enabled"),
							},
							PersistenceModeAttr: {
								String: pointer.Of("Enabled"),
							},
							DriverVersionAttr: {
								String: pointer.Of("1"),
							},
						},
					},
				},
			},
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			channel := make(chan *device.FingerprintResponse, 1)
			testCase.Device.writeFingerprintToChannel(channel)
			actualResult := <-channel
			// writeFingerprintToChannel iterates over map keys
			// and insterts results to an array, so order of elements in output array
			// may be different
			// actualResult, expectedResult arrays has to be sorted firsted
			sort.Slice(actualResult.Devices, func(i, j int) bool {
				return actualResult.Devices[i].Name < actualResult.Devices[j].Name
			})
			sort.Slice(testCase.ExpectedWriteToChannel.Devices, func(i, j int) bool {
				return testCase.ExpectedWriteToChannel.Devices[i].Name < testCase.ExpectedWriteToChannel.Devices[j].Name
			})
			must.Eq(t, testCase.ExpectedWriteToChannel, actualResult)
		})
	}
}

// Test if nonworking driver returns empty fingerprint data
func TestFingerprint(t *testing.T) {
	for _, testCase := range []struct {
		Name                   string
		Device                 *NvidiaDevice
		ExpectedWriteToChannel *device.FingerprintResponse
	}{
		{
			Name: "Check that working driver returns valid fingeprint data",
			Device: &NvidiaDevice{
				initErr: nil,
				nvmlClient: &MockNvmlClient{
					FingerprintResponseReturned: &nvml.FingerprintData{
						DriverVersion: "1",
						Devices: []*nvml.FingerprintDeviceData{
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "1",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID1",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "2",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID2",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "3",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
									PowerW:     pointer.Of(uint(100)),
									BAR1MiB:    pointer.Of(uint64(256)),
								},
								PCIBusID:           "pciBusID3",
								PCIBandwidthMBPerS: pointer.Of(uint(1)),
								CoresClockMHz:      pointer.Of(uint(1)),
								MemoryClockMHz:     pointer.Of(uint(1)),
								DisplayState:       "Enabled",
								PersistenceMode:    "Enabled",
							},
						},
					},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.FingerprintResponse{
				Devices: []*device.DeviceGroup{
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "Name1",
						Devices: []*device.Device{
							{
								ID:      "1",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID1",
								},
							},
							{
								ID:      "2",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID2",
								},
							},
							{
								ID:      "3",
								Healthy: true,
								HwLocality: &device.DeviceLocality{
									PciBusID: "pciBusID3",
								},
							},
						},
						Attributes: map[string]*structs.Attribute{
							MemoryAttr: {
								Int:  pointer.Of(int64(10)),
								Unit: structs.UnitMiB,
							},
							PowerAttr: {
								Int:  pointer.Of(int64(100)),
								Unit: structs.UnitW,
							},
							BAR1Attr: {
								Int:  pointer.Of(int64(256)),
								Unit: structs.UnitMiB,
							},
							PCIBandwidthAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMBPerS,
							},
							CoresClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							MemoryClockAttr: {
								Int:  pointer.Of(int64(1)),
								Unit: structs.UnitMHz,
							},
							DisplayStateAttr: {
								String: pointer.Of("Enabled"),
							},
							PersistenceModeAttr: {
								String: pointer.Of("Enabled"),
							},
							DriverVersionAttr: {
								String: pointer.Of("1"),
							},
						},
					},
				},
			},
		},
		{
			Name: "Check that not working driver returns error fingeprint data",
			Device: &NvidiaDevice{
				initErr: errors.New("foo"),
				nvmlClient: &MockNvmlClient{
					FingerprintResponseReturned: &nvml.FingerprintData{
						DriverVersion: "1",
						Devices: []*nvml.FingerprintDeviceData{
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "1",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
								},
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "2",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
								},
							},
							{
								DeviceData: &nvml.DeviceData{
									UUID:       "3",
									DeviceName: pointer.Of("Name1"),
									MemoryMiB:  pointer.Of(uint64(10)),
								},
							},
						},
					},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.FingerprintResponse{
				Error: errors.New("foo"),
			},
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			outCh := make(chan *device.FingerprintResponse)
			ctx, cancel := context.WithCancel(context.Background())
			go testCase.Device.fingerprint(ctx, outCh)
			result := <-outCh
			cancel()
			must.Eq(t, result, testCase.ExpectedWriteToChannel)
		})
	}
}
