// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package nvidia

import (
	"errors"
	"sort"
	"testing"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-device-nvidia/nvml"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
	"github.com/shoenig/test/must"
)

func TestFilterStatsByID(t *testing.T) {
	for _, testCase := range []struct {
		Name           string
		ProvidedStats  []*nvml.StatsData
		ProvidedIDs    map[string]device.Shared
		ExpectedResult []*nvml.StatsData
	}{
		{
			Name: "All ids are in the map",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
			},
			ProvidedIDs: map[string]device.Shared{
				"UUID1": "",
				"UUID2": "",
				"UUID3": "",
			},
			ExpectedResult: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
			},
		},
		{
			Name: "Odd are not provided in the map",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
			},
			ProvidedIDs: map[string]device.Shared{
				"UUID2": "",
			},
			ExpectedResult: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
			},
		},
		{
			Name: "Even are not provided in the map",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
			},
			ProvidedIDs: map[string]device.Shared{
				"UUID1": "",
				"UUID3": "",
			},
			ExpectedResult: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
			},
		},
		{
			Name: "No Stats were provided",
			ProvidedIDs: map[string]device.Shared{
				"UUID1": "",
				"UUID2": "",
				"UUID3": "",
			},
		},
		{
			Name: "No Ids were provided",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
			},
		},
	} {
		actualResult := filterStatsByID(testCase.ProvidedStats, testCase.ProvidedIDs)
		must.Eq(t, testCase.ExpectedResult, actualResult)
	}
}

func TestStatsForItem(t *testing.T) {
	for _, testCase := range []struct {
		Name           string
		Timestamp      time.Time
		ItemStat       *nvml.StatsData
		ExpectedResult *device.DeviceStats
	}{
		{
			Name:      "All fields in ItemStat are not nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "Power usage is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        nil,
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:      PowerUsageUnit,
							Desc:      PowerUsageDesc,
							StringVal: new(notAvailable),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "PowerW is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     nil,
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:      PowerUsageUnit,
							Desc:      PowerUsageDesc,
							StringVal: new(notAvailable),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "GPUUtilization is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     nil,
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:      GPUUtilizationUnit,
							Desc:      GPUUtilizationDesc,
							StringVal: new(notAvailable),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "MemoryUtilization is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  nil,
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:      MemoryUtilizationUnit,
							Desc:      MemoryUtilizationDesc,
							StringVal: new(notAvailable),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "EncoderUtilization is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: nil,
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:      EncoderUtilizationUnit,
							Desc:      EncoderUtilizationDesc,
							StringVal: new(notAvailable),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "DecoderUtilization is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: nil,
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:      DecoderUtilizationUnit,
							Desc:      DecoderUtilizationDesc,
							StringVal: new(notAvailable),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "Temperature is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       nil,
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:      TemperatureUnit,
							Desc:      TemperatureDesc,
							StringVal: new(notAvailable),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "UsedMemoryMiB is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      nil,
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:      MemoryStateUnit,
					Desc:      MemoryStateDesc,
					StringVal: new(notAvailable),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:      MemoryStateUnit,
							Desc:      MemoryStateDesc,
							StringVal: new(notAvailable),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "MemoryMiB is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  nil,
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:      MemoryStateUnit,
					Desc:      MemoryStateDesc,
					StringVal: new(notAvailable),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:      MemoryStateUnit,
							Desc:      MemoryStateDesc,
							StringVal: new(notAvailable),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "BAR1UsedMiB is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        nil,
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:      BAR1StateUnit,
							Desc:      BAR1StateDesc,
							StringVal: new(notAvailable),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "BAR1MiB is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    nil,
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:      BAR1StateUnit,
							Desc:      BAR1StateDesc,
							StringVal: new(notAvailable),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "ECCErrorsL1Cache is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   nil,
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:      ECCErrorsL1CacheUnit,
							Desc:      ECCErrorsL1CacheDesc,
							StringVal: new(notAvailable),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "ECCErrorsL2Cache is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   nil,
				ECCErrorsDevice:    new(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:      ECCErrorsL2CacheUnit,
							Desc:      ECCErrorsL2CacheDesc,
							StringVal: new(notAvailable),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: new(int64(100)),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			Name:      "ECCErrorsDevice is nil",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ItemStat: &nvml.StatsData{
				DeviceData: &nvml.DeviceData{
					UUID:       "UUID1",
					DeviceName: new("DeviceName1"),
					MemoryMiB:  new(uint64(1)),
					PowerW:     new(uint(1)),
					BAR1MiB:    new(uint64(256)),
				},
				PowerUsageW:        new(uint(1)),
				GPUUtilization:     new(uint(1)),
				MemoryUtilization:  new(uint(1)),
				EncoderUtilization: new(uint(1)),
				DecoderUtilization: new(uint(1)),
				TemperatureC:       new(uint(1)),
				UsedMemoryMiB:      new(uint64(1)),
				BAR1UsedMiB:        new(uint64(1)),
				ECCErrorsL1Cache:   new(uint64(100)),
				ECCErrorsL2Cache:   new(uint64(100)),
				ECCErrorsDevice:    nil,
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   new(int64(1)),
					IntDenominatorVal: new(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: new(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: new(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: new(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:      ECCErrorsDeviceUnit,
							Desc:      ECCErrorsDeviceDesc,
							StringVal: new(notAvailable),
						},
					},
				},
				Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			},
		},
	} {
		actualResult := statsForItem(testCase.ItemStat, testCase.Timestamp)
		must.Eq(t, testCase.ExpectedResult, actualResult)
	}
}

func TestStatsForGroup(t *testing.T) {
	for _, testCase := range []struct {
		Name           string
		Timestamp      time.Time
		GroupStats     []*nvml.StatsData
		GroupName      string
		ExpectedResult *device.DeviceGroupStats
	}{
		{
			Name:      "make sure that all data is transformed correctly",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			GroupName: "DeviceName1",
			GroupStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: new("DeviceName1"),
						MemoryMiB:  new(uint64(1)),
						PowerW:     new(uint(1)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(1)),
					GPUUtilization:     new(uint(1)),
					MemoryUtilization:  new(uint(1)),
					EncoderUtilization: new(uint(1)),
					DecoderUtilization: new(uint(1)),
					TemperatureC:       new(uint(1)),
					UsedMemoryMiB:      new(uint64(1)),
					BAR1UsedMiB:        new(uint64(1)),
					ECCErrorsL1Cache:   new(uint64(100)),
					ECCErrorsL2Cache:   new(uint64(100)),
					ECCErrorsDevice:    new(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: new("DeviceName2"),
						MemoryMiB:  new(uint64(2)),
						PowerW:     new(uint(2)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(2)),
					GPUUtilization:     new(uint(2)),
					MemoryUtilization:  new(uint(2)),
					EncoderUtilization: new(uint(2)),
					DecoderUtilization: new(uint(2)),
					TemperatureC:       new(uint(2)),
					UsedMemoryMiB:      new(uint64(2)),
					BAR1UsedMiB:        new(uint64(2)),
					ECCErrorsL1Cache:   new(uint64(200)),
					ECCErrorsL2Cache:   new(uint64(200)),
					ECCErrorsDevice:    new(uint64(200)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: new("DeviceName3"),
						MemoryMiB:  new(uint64(3)),
						PowerW:     new(uint(3)),
						BAR1MiB:    new(uint64(256)),
					},
					PowerUsageW:        new(uint(3)),
					GPUUtilization:     new(uint(3)),
					MemoryUtilization:  new(uint(3)),
					EncoderUtilization: new(uint(3)),
					DecoderUtilization: new(uint(3)),
					TemperatureC:       new(uint(3)),
					UsedMemoryMiB:      new(uint64(3)),
					BAR1UsedMiB:        new(uint64(3)),
					ECCErrorsL1Cache:   new(uint64(300)),
					ECCErrorsL2Cache:   new(uint64(300)),
					ECCErrorsDevice:    new(uint64(300)),
				},
			},
			ExpectedResult: &device.DeviceGroupStats{
				Vendor: vendor,
				Type:   deviceType,
				Name:   "DeviceName1",
				InstanceStats: map[string]*device.DeviceStats{
					"UUID1": {
						Summary: &structs.StatValue{
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(1)),
							IntDenominatorVal: new(int64(1)),
						},
						Stats: &structs.StatObject{
							Attributes: map[string]*structs.StatValue{
								PowerUsageAttr: {
									Unit:              PowerUsageUnit,
									Desc:              PowerUsageDesc,
									IntNumeratorVal:   new(int64(1)),
									IntDenominatorVal: new(int64(1)),
								},
								GPUUtilizationAttr: {
									Unit:            GPUUtilizationUnit,
									Desc:            GPUUtilizationDesc,
									IntNumeratorVal: new(int64(1)),
								},
								MemoryUtilizationAttr: {
									Unit:            MemoryUtilizationUnit,
									Desc:            MemoryUtilizationDesc,
									IntNumeratorVal: new(int64(1)),
								},
								EncoderUtilizationAttr: {
									Unit:            EncoderUtilizationUnit,
									Desc:            EncoderUtilizationDesc,
									IntNumeratorVal: new(int64(1)),
								},
								DecoderUtilizationAttr: {
									Unit:            DecoderUtilizationUnit,
									Desc:            DecoderUtilizationDesc,
									IntNumeratorVal: new(int64(1)),
								},
								TemperatureAttr: {
									Unit:            TemperatureUnit,
									Desc:            TemperatureDesc,
									IntNumeratorVal: new(int64(1)),
								},
								MemoryStateAttr: {
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(1)),
									IntDenominatorVal: new(int64(1)),
								},
								BAR1StateAttr: {
									Unit:              BAR1StateUnit,
									Desc:              BAR1StateDesc,
									IntNumeratorVal:   new(int64(1)),
									IntDenominatorVal: new(int64(256)),
								},
								ECCErrorsL1CacheAttr: {
									Unit:            ECCErrorsL1CacheUnit,
									Desc:            ECCErrorsL1CacheDesc,
									IntNumeratorVal: new(int64(100)),
								},
								ECCErrorsL2CacheAttr: {
									Unit:            ECCErrorsL2CacheUnit,
									Desc:            ECCErrorsL2CacheDesc,
									IntNumeratorVal: new(int64(100)),
								},
								ECCErrorsDeviceAttr: {
									Unit:            ECCErrorsDeviceUnit,
									Desc:            ECCErrorsDeviceDesc,
									IntNumeratorVal: new(int64(100)),
								},
							},
						},
						Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
					},
					"UUID2": {
						Summary: &structs.StatValue{
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(2)),
							IntDenominatorVal: new(int64(2)),
						},
						Stats: &structs.StatObject{
							Attributes: map[string]*structs.StatValue{
								PowerUsageAttr: {
									Unit:              PowerUsageUnit,
									Desc:              PowerUsageDesc,
									IntNumeratorVal:   new(int64(2)),
									IntDenominatorVal: new(int64(2)),
								},
								GPUUtilizationAttr: {
									Unit:            GPUUtilizationUnit,
									Desc:            GPUUtilizationDesc,
									IntNumeratorVal: new(int64(2)),
								},
								MemoryUtilizationAttr: {
									Unit:            MemoryUtilizationUnit,
									Desc:            MemoryUtilizationDesc,
									IntNumeratorVal: new(int64(2)),
								},
								EncoderUtilizationAttr: {
									Unit:            EncoderUtilizationUnit,
									Desc:            EncoderUtilizationDesc,
									IntNumeratorVal: new(int64(2)),
								},
								DecoderUtilizationAttr: {
									Unit:            DecoderUtilizationUnit,
									Desc:            DecoderUtilizationDesc,
									IntNumeratorVal: new(int64(2)),
								},
								TemperatureAttr: {
									Unit:            TemperatureUnit,
									Desc:            TemperatureDesc,
									IntNumeratorVal: new(int64(2)),
								},
								MemoryStateAttr: {
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(2)),
									IntDenominatorVal: new(int64(2)),
								},
								BAR1StateAttr: {
									Unit:              BAR1StateUnit,
									Desc:              BAR1StateDesc,
									IntNumeratorVal:   new(int64(2)),
									IntDenominatorVal: new(int64(256)),
								},
								ECCErrorsL1CacheAttr: {
									Unit:            ECCErrorsL1CacheUnit,
									Desc:            ECCErrorsL1CacheDesc,
									IntNumeratorVal: new(int64(200)),
								},
								ECCErrorsL2CacheAttr: {
									Unit:            ECCErrorsL2CacheUnit,
									Desc:            ECCErrorsL2CacheDesc,
									IntNumeratorVal: new(int64(200)),
								},
								ECCErrorsDeviceAttr: {
									Unit:            ECCErrorsDeviceUnit,
									Desc:            ECCErrorsDeviceDesc,
									IntNumeratorVal: new(int64(200)),
								},
							},
						},
						Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
					},
					"UUID3": {
						Summary: &structs.StatValue{
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   new(int64(3)),
							IntDenominatorVal: new(int64(3)),
						},
						Stats: &structs.StatObject{
							Attributes: map[string]*structs.StatValue{
								PowerUsageAttr: {
									Unit:              PowerUsageUnit,
									Desc:              PowerUsageDesc,
									IntNumeratorVal:   new(int64(3)),
									IntDenominatorVal: new(int64(3)),
								},
								GPUUtilizationAttr: {
									Unit:            GPUUtilizationUnit,
									Desc:            GPUUtilizationDesc,
									IntNumeratorVal: new(int64(3)),
								},
								MemoryUtilizationAttr: {
									Unit:            MemoryUtilizationUnit,
									Desc:            MemoryUtilizationDesc,
									IntNumeratorVal: new(int64(3)),
								},
								EncoderUtilizationAttr: {
									Unit:            EncoderUtilizationUnit,
									Desc:            EncoderUtilizationDesc,
									IntNumeratorVal: new(int64(3)),
								},
								DecoderUtilizationAttr: {
									Unit:            DecoderUtilizationUnit,
									Desc:            DecoderUtilizationDesc,
									IntNumeratorVal: new(int64(3)),
								},
								TemperatureAttr: {
									Unit:            TemperatureUnit,
									Desc:            TemperatureDesc,
									IntNumeratorVal: new(int64(3)),
								},
								MemoryStateAttr: {
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(3)),
									IntDenominatorVal: new(int64(3)),
								},
								BAR1StateAttr: {
									Unit:              BAR1StateUnit,
									Desc:              BAR1StateDesc,
									IntNumeratorVal:   new(int64(3)),
									IntDenominatorVal: new(int64(256)),
								},
								ECCErrorsL1CacheAttr: {
									Unit:            ECCErrorsL1CacheUnit,
									Desc:            ECCErrorsL1CacheDesc,
									IntNumeratorVal: new(int64(300)),
								},
								ECCErrorsL2CacheAttr: {
									Unit:            ECCErrorsL2CacheUnit,
									Desc:            ECCErrorsL2CacheDesc,
									IntNumeratorVal: new(int64(300)),
								},
								ECCErrorsDeviceAttr: {
									Unit:            ECCErrorsDeviceUnit,
									Desc:            ECCErrorsDeviceDesc,
									IntNumeratorVal: new(int64(300)),
								},
							},
						},
						Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
					},
				},
			},
		},
	} {
		actualResult := statsForGroup(testCase.GroupName, testCase.GroupStats, testCase.Timestamp)
		must.Eq(t, testCase.ExpectedResult, actualResult)
	}
}

func TestWriteStatsToChannel(t *testing.T) {
	for _, testCase := range []struct {
		Name                   string
		ExpectedWriteToChannel *device.StatsResponse
		Timestamp              time.Time
		Device                 *NvidiaDevice
	}{
		{
			Name:      "NVML wrapper returns error",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			ExpectedWriteToChannel: &device.StatsResponse{
				Error: errors.New(""),
			},
			Device: &NvidiaDevice{
				nvmlClient: &MockNvmlClient{
					StatsError: errors.New(""),
				},
				logger: hclog.NewNullLogger(),
			},
		},
		{
			Name:      "Check that stats with multiple DeviceNames are assigned to different groups",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			Device: &NvidiaDevice{
				devices: map[string]device.Shared{
					"UUID1": "",
					"UUID2": "",
					"UUID3": "",
				},
				nvmlClient: &MockNvmlClient{
					StatsResponseReturned: []*nvml.StatsData{
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID1",
								DeviceName: new("DeviceName1"),
								MemoryMiB:  new(uint64(1)),
								PowerW:     new(uint(1)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(1)),
							GPUUtilization:     new(uint(1)),
							MemoryUtilization:  new(uint(1)),
							EncoderUtilization: new(uint(1)),
							DecoderUtilization: new(uint(1)),
							TemperatureC:       new(uint(1)),
							UsedMemoryMiB:      new(uint64(1)),
							BAR1UsedMiB:        new(uint64(1)),
							ECCErrorsL1Cache:   new(uint64(100)),
							ECCErrorsL2Cache:   new(uint64(100)),
							ECCErrorsDevice:    new(uint64(100)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID2",
								DeviceName: new("DeviceName2"),
								MemoryMiB:  new(uint64(2)),
								PowerW:     new(uint(2)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(2)),
							GPUUtilization:     new(uint(2)),
							MemoryUtilization:  new(uint(2)),
							EncoderUtilization: new(uint(2)),
							DecoderUtilization: new(uint(2)),
							TemperatureC:       new(uint(2)),
							UsedMemoryMiB:      new(uint64(2)),
							BAR1UsedMiB:        new(uint64(2)),
							ECCErrorsL1Cache:   new(uint64(200)),
							ECCErrorsL2Cache:   new(uint64(200)),
							ECCErrorsDevice:    new(uint64(200)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID3",
								DeviceName: new("DeviceName3"),
								MemoryMiB:  new(uint64(3)),
								PowerW:     new(uint(3)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(3)),
							GPUUtilization:     new(uint(3)),
							MemoryUtilization:  new(uint(3)),
							EncoderUtilization: new(uint(3)),
							DecoderUtilization: new(uint(3)),
							TemperatureC:       new(uint(3)),
							UsedMemoryMiB:      new(uint64(3)),
							BAR1UsedMiB:        new(uint64(3)),
							ECCErrorsL1Cache:   new(uint64(300)),
							ECCErrorsL2Cache:   new(uint64(300)),
							ECCErrorsDevice:    new(uint64(300)),
						},
					},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.StatsResponse{
				Groups: []*device.DeviceGroupStats{
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "DeviceName1",
						InstanceStats: map[string]*device.DeviceStats{
							"UUID1": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(1)),
									IntDenominatorVal: new(int64(1)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(1)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(1)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(1)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(100)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(100)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(100)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
						},
					},
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "DeviceName2",
						InstanceStats: map[string]*device.DeviceStats{
							"UUID2": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(2)),
									IntDenominatorVal: new(int64(2)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(2)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(2)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(2)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(200)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(200)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(200)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
						},
					},
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "DeviceName3",
						InstanceStats: map[string]*device.DeviceStats{
							"UUID3": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(3)),
									IntDenominatorVal: new(int64(3)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(3)),
											IntDenominatorVal: new(int64(3)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(3)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(3)),
											IntDenominatorVal: new(int64(3)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(3)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(300)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(300)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(300)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
						},
					},
				},
			},
		},
		{
			Name:      "Check that stats with multiple DeviceNames are assigned to different groups 2",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			Device: &NvidiaDevice{
				devices: map[string]device.Shared{
					"UUID1": "",
					"UUID2": "",
					"UUID3": "",
				},
				nvmlClient: &MockNvmlClient{
					StatsResponseReturned: []*nvml.StatsData{
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID1",
								DeviceName: new("DeviceName1"),
								MemoryMiB:  new(uint64(1)),
								PowerW:     new(uint(1)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(1)),
							GPUUtilization:     new(uint(1)),
							MemoryUtilization:  new(uint(1)),
							EncoderUtilization: new(uint(1)),
							DecoderUtilization: new(uint(1)),
							TemperatureC:       new(uint(1)),
							UsedMemoryMiB:      new(uint64(1)),
							BAR1UsedMiB:        new(uint64(1)),
							ECCErrorsL1Cache:   new(uint64(100)),
							ECCErrorsL2Cache:   new(uint64(100)),
							ECCErrorsDevice:    new(uint64(100)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID2",
								DeviceName: new("DeviceName2"),
								MemoryMiB:  new(uint64(2)),
								PowerW:     new(uint(2)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(2)),
							GPUUtilization:     new(uint(2)),
							MemoryUtilization:  new(uint(2)),
							EncoderUtilization: new(uint(2)),
							DecoderUtilization: new(uint(2)),
							TemperatureC:       new(uint(2)),
							UsedMemoryMiB:      new(uint64(2)),
							BAR1UsedMiB:        new(uint64(2)),
							ECCErrorsL1Cache:   new(uint64(200)),
							ECCErrorsL2Cache:   new(uint64(200)),
							ECCErrorsDevice:    new(uint64(200)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID3",
								DeviceName: new("DeviceName2"),
								MemoryMiB:  new(uint64(3)),
								PowerW:     new(uint(3)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(3)),
							GPUUtilization:     new(uint(3)),
							MemoryUtilization:  new(uint(3)),
							EncoderUtilization: new(uint(3)),
							DecoderUtilization: new(uint(3)),
							TemperatureC:       new(uint(3)),
							UsedMemoryMiB:      new(uint64(3)),
							BAR1UsedMiB:        new(uint64(3)),
							ECCErrorsL1Cache:   new(uint64(300)),
							ECCErrorsL2Cache:   new(uint64(300)),
							ECCErrorsDevice:    new(uint64(300)),
						},
					},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.StatsResponse{
				Groups: []*device.DeviceGroupStats{
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "DeviceName1",
						InstanceStats: map[string]*device.DeviceStats{
							"UUID1": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(1)),
									IntDenominatorVal: new(int64(1)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(1)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(1)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(1)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(100)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(100)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(100)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
						},
					},
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "DeviceName2",
						InstanceStats: map[string]*device.DeviceStats{
							"UUID3": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(3)),
									IntDenominatorVal: new(int64(3)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(3)),
											IntDenominatorVal: new(int64(3)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(3)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(3)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(3)),
											IntDenominatorVal: new(int64(3)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(3)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(300)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(300)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(300)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
							"UUID2": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(2)),
									IntDenominatorVal: new(int64(2)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(2)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(2)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(2)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(200)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(200)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(200)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
						},
					},
				},
			},
		},
		{
			Name:      "Check that only devices from NvidiaDevice.device map stats are reported",
			Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
			Device: &NvidiaDevice{
				devices: map[string]device.Shared{
					"UUID1": "",
					"UUID2": "",
				},
				nvmlClient: &MockNvmlClient{
					StatsResponseReturned: []*nvml.StatsData{
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID1",
								DeviceName: new("DeviceName1"),
								MemoryMiB:  new(uint64(1)),
								PowerW:     new(uint(1)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(1)),
							GPUUtilization:     new(uint(1)),
							MemoryUtilization:  new(uint(1)),
							EncoderUtilization: new(uint(1)),
							DecoderUtilization: new(uint(1)),
							TemperatureC:       new(uint(1)),
							UsedMemoryMiB:      new(uint64(1)),
							BAR1UsedMiB:        new(uint64(1)),
							ECCErrorsL1Cache:   new(uint64(100)),
							ECCErrorsL2Cache:   new(uint64(100)),
							ECCErrorsDevice:    new(uint64(100)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID2",
								DeviceName: new("DeviceName2"),
								MemoryMiB:  new(uint64(2)),
								PowerW:     new(uint(2)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(2)),
							GPUUtilization:     new(uint(2)),
							MemoryUtilization:  new(uint(2)),
							EncoderUtilization: new(uint(2)),
							DecoderUtilization: new(uint(2)),
							TemperatureC:       new(uint(2)),
							UsedMemoryMiB:      new(uint64(2)),
							BAR1UsedMiB:        new(uint64(2)),
							ECCErrorsL1Cache:   new(uint64(200)),
							ECCErrorsL2Cache:   new(uint64(200)),
							ECCErrorsDevice:    new(uint64(200)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID3",
								DeviceName: new("DeviceName3"),
								MemoryMiB:  new(uint64(3)),
								PowerW:     new(uint(3)),
								BAR1MiB:    new(uint64(256)),
							},
							PowerUsageW:        new(uint(3)),
							GPUUtilization:     new(uint(3)),
							MemoryUtilization:  new(uint(3)),
							EncoderUtilization: new(uint(3)),
							DecoderUtilization: new(uint(3)),
							TemperatureC:       new(uint(3)),
							UsedMemoryMiB:      new(uint64(3)),
							BAR1UsedMiB:        new(uint64(3)),
							ECCErrorsL1Cache:   new(uint64(300)),
							ECCErrorsL2Cache:   new(uint64(300)),
							ECCErrorsDevice:    new(uint64(300)),
						},
					},
				},
				logger: hclog.NewNullLogger(),
			},
			ExpectedWriteToChannel: &device.StatsResponse{
				Groups: []*device.DeviceGroupStats{
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "DeviceName1",
						InstanceStats: map[string]*device.DeviceStats{
							"UUID1": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(1)),
									IntDenominatorVal: new(int64(1)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(1)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(1)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(1)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(1)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(1)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(100)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(100)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(100)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
						},
					},
					{
						Vendor: vendor,
						Type:   deviceType,
						Name:   "DeviceName2",
						InstanceStats: map[string]*device.DeviceStats{
							"UUID2": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   new(int64(2)),
									IntDenominatorVal: new(int64(2)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(2)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: new(int64(2)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: new(int64(2)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(2)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   new(int64(2)),
											IntDenominatorVal: new(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: new(int64(200)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: new(int64(200)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: new(int64(200)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
						},
					},
				},
			},
		},
	} {
		channel := make(chan *device.StatsResponse, 1)
		testCase.Device.writeStatsToChannel(channel, testCase.Timestamp)
		actualResult := <-channel
		// writeStatsToChannel iterates over map keys
		// and insterts results to an array, so order of elements in output array
		// may be different
		// actualResult, expectedWriteToChannel arrays has to be sorted firsted
		sort.Slice(actualResult.Groups, func(i, j int) bool {
			return actualResult.Groups[i].Name < actualResult.Groups[j].Name
		})
		sort.Slice(testCase.ExpectedWriteToChannel.Groups, func(i, j int) bool {
			return testCase.ExpectedWriteToChannel.Groups[i].Name < testCase.ExpectedWriteToChannel.Groups[j].Name
		})
		must.Eq(t, testCase.ExpectedWriteToChannel, actualResult)
	}
}
