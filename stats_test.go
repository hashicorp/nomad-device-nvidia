// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nvidia

import (
	"errors"
	"sort"
	"testing"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-device-nvidia/nvml"
	"github.com/hashicorp/nomad/helper/pointer"
	"github.com/hashicorp/nomad/plugins/device"
	"github.com/hashicorp/nomad/plugins/shared/structs"
	"github.com/shoenig/test/must"
)

func TestFilterStatsByID(t *testing.T) {
	for _, testCase := range []struct {
		Name           string
		ProvidedStats  []*nvml.StatsData
		ProvidedIDs    map[string]struct{}
		ExpectedResult []*nvml.StatsData
	}{
		{
			Name: "All ids are in the map",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
			},
			ProvidedIDs: map[string]struct{}{
				"UUID1": {},
				"UUID2": {},
				"UUID3": {},
			},
			ExpectedResult: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
			},
		},
		{
			Name: "Odd are not provided in the map",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
			},
			ProvidedIDs: map[string]struct{}{
				"UUID2": {},
			},
			ExpectedResult: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
			},
		},
		{
			Name: "Even are not provided in the map",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
			},
			ProvidedIDs: map[string]struct{}{
				"UUID1": {},
				"UUID3": {},
			},
			ExpectedResult: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
			},
		},
		{
			Name: "No Stats were provided",
			ProvidedIDs: map[string]struct{}{
				"UUID1": {},
				"UUID2": {},
				"UUID3": {},
			},
		},
		{
			Name: "No Ids were provided",
			ProvidedStats: []*nvml.StatsData{
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID1",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        nil,
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:      PowerUsageUnit,
							Desc:      PowerUsageDesc,
							StringVal: pointer.Of(notAvailable),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     nil,
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:      PowerUsageUnit,
							Desc:      PowerUsageDesc,
							StringVal: pointer.Of(notAvailable),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     nil,
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:      GPUUtilizationUnit,
							Desc:      GPUUtilizationDesc,
							StringVal: pointer.Of(notAvailable),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  nil,
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:      MemoryUtilizationUnit,
							Desc:      MemoryUtilizationDesc,
							StringVal: pointer.Of(notAvailable),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: nil,
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:      EncoderUtilizationUnit,
							Desc:      EncoderUtilizationDesc,
							StringVal: pointer.Of(notAvailable),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: nil,
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:      DecoderUtilizationUnit,
							Desc:      DecoderUtilizationDesc,
							StringVal: pointer.Of(notAvailable),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       nil,
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:      TemperatureUnit,
							Desc:      TemperatureDesc,
							StringVal: pointer.Of(notAvailable),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      nil,
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:      MemoryStateUnit,
					Desc:      MemoryStateDesc,
					StringVal: pointer.Of(notAvailable),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:      MemoryStateUnit,
							Desc:      MemoryStateDesc,
							StringVal: pointer.Of(notAvailable),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  nil,
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:      MemoryStateUnit,
					Desc:      MemoryStateDesc,
					StringVal: pointer.Of(notAvailable),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:      MemoryStateUnit,
							Desc:      MemoryStateDesc,
							StringVal: pointer.Of(notAvailable),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        nil,
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:      BAR1StateUnit,
							Desc:      BAR1StateDesc,
							StringVal: pointer.Of(notAvailable),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    nil,
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:      BAR1StateUnit,
							Desc:      BAR1StateDesc,
							StringVal: pointer.Of(notAvailable),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   nil,
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:      ECCErrorsL1CacheUnit,
							Desc:      ECCErrorsL1CacheDesc,
							StringVal: pointer.Of(notAvailable),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   nil,
				ECCErrorsDevice:    pointer.Of(uint64(100)),
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:      ECCErrorsL2CacheUnit,
							Desc:      ECCErrorsL2CacheDesc,
							StringVal: pointer.Of(notAvailable),
						},
						ECCErrorsDeviceAttr: {
							Unit:            ECCErrorsDeviceUnit,
							Desc:            ECCErrorsDeviceDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
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
					DeviceName: pointer.Of("DeviceName1"),
					MemoryMiB:  pointer.Of(uint64(1)),
					PowerW:     pointer.Of(uint(1)),
					BAR1MiB:    pointer.Of(uint64(256)),
				},
				PowerUsageW:        pointer.Of(uint(1)),
				GPUUtilization:     pointer.Of(uint(1)),
				MemoryUtilization:  pointer.Of(uint(1)),
				EncoderUtilization: pointer.Of(uint(1)),
				DecoderUtilization: pointer.Of(uint(1)),
				TemperatureC:       pointer.Of(uint(1)),
				UsedMemoryMiB:      pointer.Of(uint64(1)),
				BAR1UsedMiB:        pointer.Of(uint64(1)),
				ECCErrorsL1Cache:   pointer.Of(uint64(100)),
				ECCErrorsL2Cache:   pointer.Of(uint64(100)),
				ECCErrorsDevice:    nil,
			},
			ExpectedResult: &device.DeviceStats{
				Summary: &structs.StatValue{
					Unit:              MemoryStateUnit,
					Desc:              MemoryStateDesc,
					IntNumeratorVal:   pointer.Of(int64(1)),
					IntDenominatorVal: pointer.Of(int64(1)),
				},
				Stats: &structs.StatObject{
					Attributes: map[string]*structs.StatValue{
						PowerUsageAttr: {
							Unit:              PowerUsageUnit,
							Desc:              PowerUsageDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						GPUUtilizationAttr: {
							Unit:            GPUUtilizationUnit,
							Desc:            GPUUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryUtilizationAttr: {
							Unit:            MemoryUtilizationUnit,
							Desc:            MemoryUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						EncoderUtilizationAttr: {
							Unit:            EncoderUtilizationUnit,
							Desc:            EncoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						DecoderUtilizationAttr: {
							Unit:            DecoderUtilizationUnit,
							Desc:            DecoderUtilizationDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						TemperatureAttr: {
							Unit:            TemperatureUnit,
							Desc:            TemperatureDesc,
							IntNumeratorVal: pointer.Of(int64(1)),
						},
						MemoryStateAttr: {
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						BAR1StateAttr: {
							Unit:              BAR1StateUnit,
							Desc:              BAR1StateDesc,
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(256)),
						},
						ECCErrorsL1CacheAttr: {
							Unit:            ECCErrorsL1CacheUnit,
							Desc:            ECCErrorsL1CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsL2CacheAttr: {
							Unit:            ECCErrorsL2CacheUnit,
							Desc:            ECCErrorsL2CacheDesc,
							IntNumeratorVal: pointer.Of(int64(100)),
						},
						ECCErrorsDeviceAttr: {
							Unit:      ECCErrorsDeviceUnit,
							Desc:      ECCErrorsDeviceDesc,
							StringVal: pointer.Of(notAvailable),
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
						DeviceName: pointer.Of("DeviceName1"),
						MemoryMiB:  pointer.Of(uint64(1)),
						PowerW:     pointer.Of(uint(1)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(1)),
					GPUUtilization:     pointer.Of(uint(1)),
					MemoryUtilization:  pointer.Of(uint(1)),
					EncoderUtilization: pointer.Of(uint(1)),
					DecoderUtilization: pointer.Of(uint(1)),
					TemperatureC:       pointer.Of(uint(1)),
					UsedMemoryMiB:      pointer.Of(uint64(1)),
					BAR1UsedMiB:        pointer.Of(uint64(1)),
					ECCErrorsL1Cache:   pointer.Of(uint64(100)),
					ECCErrorsL2Cache:   pointer.Of(uint64(100)),
					ECCErrorsDevice:    pointer.Of(uint64(100)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID2",
						DeviceName: pointer.Of("DeviceName2"),
						MemoryMiB:  pointer.Of(uint64(2)),
						PowerW:     pointer.Of(uint(2)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(2)),
					GPUUtilization:     pointer.Of(uint(2)),
					MemoryUtilization:  pointer.Of(uint(2)),
					EncoderUtilization: pointer.Of(uint(2)),
					DecoderUtilization: pointer.Of(uint(2)),
					TemperatureC:       pointer.Of(uint(2)),
					UsedMemoryMiB:      pointer.Of(uint64(2)),
					BAR1UsedMiB:        pointer.Of(uint64(2)),
					ECCErrorsL1Cache:   pointer.Of(uint64(200)),
					ECCErrorsL2Cache:   pointer.Of(uint64(200)),
					ECCErrorsDevice:    pointer.Of(uint64(200)),
				},
				{
					DeviceData: &nvml.DeviceData{
						UUID:       "UUID3",
						DeviceName: pointer.Of("DeviceName3"),
						MemoryMiB:  pointer.Of(uint64(3)),
						PowerW:     pointer.Of(uint(3)),
						BAR1MiB:    pointer.Of(uint64(256)),
					},
					PowerUsageW:        pointer.Of(uint(3)),
					GPUUtilization:     pointer.Of(uint(3)),
					MemoryUtilization:  pointer.Of(uint(3)),
					EncoderUtilization: pointer.Of(uint(3)),
					DecoderUtilization: pointer.Of(uint(3)),
					TemperatureC:       pointer.Of(uint(3)),
					UsedMemoryMiB:      pointer.Of(uint64(3)),
					BAR1UsedMiB:        pointer.Of(uint64(3)),
					ECCErrorsL1Cache:   pointer.Of(uint64(300)),
					ECCErrorsL2Cache:   pointer.Of(uint64(300)),
					ECCErrorsDevice:    pointer.Of(uint64(300)),
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
							IntNumeratorVal:   pointer.Of(int64(1)),
							IntDenominatorVal: pointer.Of(int64(1)),
						},
						Stats: &structs.StatObject{
							Attributes: map[string]*structs.StatValue{
								PowerUsageAttr: {
									Unit:              PowerUsageUnit,
									Desc:              PowerUsageDesc,
									IntNumeratorVal:   pointer.Of(int64(1)),
									IntDenominatorVal: pointer.Of(int64(1)),
								},
								GPUUtilizationAttr: {
									Unit:            GPUUtilizationUnit,
									Desc:            GPUUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(1)),
								},
								MemoryUtilizationAttr: {
									Unit:            MemoryUtilizationUnit,
									Desc:            MemoryUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(1)),
								},
								EncoderUtilizationAttr: {
									Unit:            EncoderUtilizationUnit,
									Desc:            EncoderUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(1)),
								},
								DecoderUtilizationAttr: {
									Unit:            DecoderUtilizationUnit,
									Desc:            DecoderUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(1)),
								},
								TemperatureAttr: {
									Unit:            TemperatureUnit,
									Desc:            TemperatureDesc,
									IntNumeratorVal: pointer.Of(int64(1)),
								},
								MemoryStateAttr: {
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   pointer.Of(int64(1)),
									IntDenominatorVal: pointer.Of(int64(1)),
								},
								BAR1StateAttr: {
									Unit:              BAR1StateUnit,
									Desc:              BAR1StateDesc,
									IntNumeratorVal:   pointer.Of(int64(1)),
									IntDenominatorVal: pointer.Of(int64(256)),
								},
								ECCErrorsL1CacheAttr: {
									Unit:            ECCErrorsL1CacheUnit,
									Desc:            ECCErrorsL1CacheDesc,
									IntNumeratorVal: pointer.Of(int64(100)),
								},
								ECCErrorsL2CacheAttr: {
									Unit:            ECCErrorsL2CacheUnit,
									Desc:            ECCErrorsL2CacheDesc,
									IntNumeratorVal: pointer.Of(int64(100)),
								},
								ECCErrorsDeviceAttr: {
									Unit:            ECCErrorsDeviceUnit,
									Desc:            ECCErrorsDeviceDesc,
									IntNumeratorVal: pointer.Of(int64(100)),
								},
							},
						},
						Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
					},
					"UUID2": {
						Summary: &structs.StatValue{
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(2)),
							IntDenominatorVal: pointer.Of(int64(2)),
						},
						Stats: &structs.StatObject{
							Attributes: map[string]*structs.StatValue{
								PowerUsageAttr: {
									Unit:              PowerUsageUnit,
									Desc:              PowerUsageDesc,
									IntNumeratorVal:   pointer.Of(int64(2)),
									IntDenominatorVal: pointer.Of(int64(2)),
								},
								GPUUtilizationAttr: {
									Unit:            GPUUtilizationUnit,
									Desc:            GPUUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(2)),
								},
								MemoryUtilizationAttr: {
									Unit:            MemoryUtilizationUnit,
									Desc:            MemoryUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(2)),
								},
								EncoderUtilizationAttr: {
									Unit:            EncoderUtilizationUnit,
									Desc:            EncoderUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(2)),
								},
								DecoderUtilizationAttr: {
									Unit:            DecoderUtilizationUnit,
									Desc:            DecoderUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(2)),
								},
								TemperatureAttr: {
									Unit:            TemperatureUnit,
									Desc:            TemperatureDesc,
									IntNumeratorVal: pointer.Of(int64(2)),
								},
								MemoryStateAttr: {
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   pointer.Of(int64(2)),
									IntDenominatorVal: pointer.Of(int64(2)),
								},
								BAR1StateAttr: {
									Unit:              BAR1StateUnit,
									Desc:              BAR1StateDesc,
									IntNumeratorVal:   pointer.Of(int64(2)),
									IntDenominatorVal: pointer.Of(int64(256)),
								},
								ECCErrorsL1CacheAttr: {
									Unit:            ECCErrorsL1CacheUnit,
									Desc:            ECCErrorsL1CacheDesc,
									IntNumeratorVal: pointer.Of(int64(200)),
								},
								ECCErrorsL2CacheAttr: {
									Unit:            ECCErrorsL2CacheUnit,
									Desc:            ECCErrorsL2CacheDesc,
									IntNumeratorVal: pointer.Of(int64(200)),
								},
								ECCErrorsDeviceAttr: {
									Unit:            ECCErrorsDeviceUnit,
									Desc:            ECCErrorsDeviceDesc,
									IntNumeratorVal: pointer.Of(int64(200)),
								},
							},
						},
						Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
					},
					"UUID3": {
						Summary: &structs.StatValue{
							Unit:              MemoryStateUnit,
							Desc:              MemoryStateDesc,
							IntNumeratorVal:   pointer.Of(int64(3)),
							IntDenominatorVal: pointer.Of(int64(3)),
						},
						Stats: &structs.StatObject{
							Attributes: map[string]*structs.StatValue{
								PowerUsageAttr: {
									Unit:              PowerUsageUnit,
									Desc:              PowerUsageDesc,
									IntNumeratorVal:   pointer.Of(int64(3)),
									IntDenominatorVal: pointer.Of(int64(3)),
								},
								GPUUtilizationAttr: {
									Unit:            GPUUtilizationUnit,
									Desc:            GPUUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(3)),
								},
								MemoryUtilizationAttr: {
									Unit:            MemoryUtilizationUnit,
									Desc:            MemoryUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(3)),
								},
								EncoderUtilizationAttr: {
									Unit:            EncoderUtilizationUnit,
									Desc:            EncoderUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(3)),
								},
								DecoderUtilizationAttr: {
									Unit:            DecoderUtilizationUnit,
									Desc:            DecoderUtilizationDesc,
									IntNumeratorVal: pointer.Of(int64(3)),
								},
								TemperatureAttr: {
									Unit:            TemperatureUnit,
									Desc:            TemperatureDesc,
									IntNumeratorVal: pointer.Of(int64(3)),
								},
								MemoryStateAttr: {
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   pointer.Of(int64(3)),
									IntDenominatorVal: pointer.Of(int64(3)),
								},
								BAR1StateAttr: {
									Unit:              BAR1StateUnit,
									Desc:              BAR1StateDesc,
									IntNumeratorVal:   pointer.Of(int64(3)),
									IntDenominatorVal: pointer.Of(int64(256)),
								},
								ECCErrorsL1CacheAttr: {
									Unit:            ECCErrorsL1CacheUnit,
									Desc:            ECCErrorsL1CacheDesc,
									IntNumeratorVal: pointer.Of(int64(300)),
								},
								ECCErrorsL2CacheAttr: {
									Unit:            ECCErrorsL2CacheUnit,
									Desc:            ECCErrorsL2CacheDesc,
									IntNumeratorVal: pointer.Of(int64(300)),
								},
								ECCErrorsDeviceAttr: {
									Unit:            ECCErrorsDeviceUnit,
									Desc:            ECCErrorsDeviceDesc,
									IntNumeratorVal: pointer.Of(int64(300)),
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
				devices: map[string]struct{}{
					"UUID1": {},
					"UUID2": {},
					"UUID3": {},
				},
				nvmlClient: &MockNvmlClient{
					StatsResponseReturned: []*nvml.StatsData{
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID1",
								DeviceName: pointer.Of("DeviceName1"),
								MemoryMiB:  pointer.Of(uint64(1)),
								PowerW:     pointer.Of(uint(1)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(1)),
							GPUUtilization:     pointer.Of(uint(1)),
							MemoryUtilization:  pointer.Of(uint(1)),
							EncoderUtilization: pointer.Of(uint(1)),
							DecoderUtilization: pointer.Of(uint(1)),
							TemperatureC:       pointer.Of(uint(1)),
							UsedMemoryMiB:      pointer.Of(uint64(1)),
							BAR1UsedMiB:        pointer.Of(uint64(1)),
							ECCErrorsL1Cache:   pointer.Of(uint64(100)),
							ECCErrorsL2Cache:   pointer.Of(uint64(100)),
							ECCErrorsDevice:    pointer.Of(uint64(100)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID2",
								DeviceName: pointer.Of("DeviceName2"),
								MemoryMiB:  pointer.Of(uint64(2)),
								PowerW:     pointer.Of(uint(2)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(2)),
							GPUUtilization:     pointer.Of(uint(2)),
							MemoryUtilization:  pointer.Of(uint(2)),
							EncoderUtilization: pointer.Of(uint(2)),
							DecoderUtilization: pointer.Of(uint(2)),
							TemperatureC:       pointer.Of(uint(2)),
							UsedMemoryMiB:      pointer.Of(uint64(2)),
							BAR1UsedMiB:        pointer.Of(uint64(2)),
							ECCErrorsL1Cache:   pointer.Of(uint64(200)),
							ECCErrorsL2Cache:   pointer.Of(uint64(200)),
							ECCErrorsDevice:    pointer.Of(uint64(200)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID3",
								DeviceName: pointer.Of("DeviceName3"),
								MemoryMiB:  pointer.Of(uint64(3)),
								PowerW:     pointer.Of(uint(3)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(3)),
							GPUUtilization:     pointer.Of(uint(3)),
							MemoryUtilization:  pointer.Of(uint(3)),
							EncoderUtilization: pointer.Of(uint(3)),
							DecoderUtilization: pointer.Of(uint(3)),
							TemperatureC:       pointer.Of(uint(3)),
							UsedMemoryMiB:      pointer.Of(uint64(3)),
							BAR1UsedMiB:        pointer.Of(uint64(3)),
							ECCErrorsL1Cache:   pointer.Of(uint64(300)),
							ECCErrorsL2Cache:   pointer.Of(uint64(300)),
							ECCErrorsDevice:    pointer.Of(uint64(300)),
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
									IntNumeratorVal:   pointer.Of(int64(1)),
									IntDenominatorVal: pointer.Of(int64(1)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(1)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(1)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
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
									IntNumeratorVal:   pointer.Of(int64(2)),
									IntDenominatorVal: pointer.Of(int64(2)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(2)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(2)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
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
									IntNumeratorVal:   pointer.Of(int64(3)),
									IntDenominatorVal: pointer.Of(int64(3)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(3)),
											IntDenominatorVal: pointer.Of(int64(3)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(3)),
											IntDenominatorVal: pointer.Of(int64(3)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(3)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(300)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(300)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(300)),
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
				devices: map[string]struct{}{
					"UUID1": {},
					"UUID2": {},
					"UUID3": {},
				},
				nvmlClient: &MockNvmlClient{
					StatsResponseReturned: []*nvml.StatsData{
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID1",
								DeviceName: pointer.Of("DeviceName1"),
								MemoryMiB:  pointer.Of(uint64(1)),
								PowerW:     pointer.Of(uint(1)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(1)),
							GPUUtilization:     pointer.Of(uint(1)),
							MemoryUtilization:  pointer.Of(uint(1)),
							EncoderUtilization: pointer.Of(uint(1)),
							DecoderUtilization: pointer.Of(uint(1)),
							TemperatureC:       pointer.Of(uint(1)),
							UsedMemoryMiB:      pointer.Of(uint64(1)),
							BAR1UsedMiB:        pointer.Of(uint64(1)),
							ECCErrorsL1Cache:   pointer.Of(uint64(100)),
							ECCErrorsL2Cache:   pointer.Of(uint64(100)),
							ECCErrorsDevice:    pointer.Of(uint64(100)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID2",
								DeviceName: pointer.Of("DeviceName2"),
								MemoryMiB:  pointer.Of(uint64(2)),
								PowerW:     pointer.Of(uint(2)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(2)),
							GPUUtilization:     pointer.Of(uint(2)),
							MemoryUtilization:  pointer.Of(uint(2)),
							EncoderUtilization: pointer.Of(uint(2)),
							DecoderUtilization: pointer.Of(uint(2)),
							TemperatureC:       pointer.Of(uint(2)),
							UsedMemoryMiB:      pointer.Of(uint64(2)),
							BAR1UsedMiB:        pointer.Of(uint64(2)),
							ECCErrorsL1Cache:   pointer.Of(uint64(200)),
							ECCErrorsL2Cache:   pointer.Of(uint64(200)),
							ECCErrorsDevice:    pointer.Of(uint64(200)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID3",
								DeviceName: pointer.Of("DeviceName2"),
								MemoryMiB:  pointer.Of(uint64(3)),
								PowerW:     pointer.Of(uint(3)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(3)),
							GPUUtilization:     pointer.Of(uint(3)),
							MemoryUtilization:  pointer.Of(uint(3)),
							EncoderUtilization: pointer.Of(uint(3)),
							DecoderUtilization: pointer.Of(uint(3)),
							TemperatureC:       pointer.Of(uint(3)),
							UsedMemoryMiB:      pointer.Of(uint64(3)),
							BAR1UsedMiB:        pointer.Of(uint64(3)),
							ECCErrorsL1Cache:   pointer.Of(uint64(300)),
							ECCErrorsL2Cache:   pointer.Of(uint64(300)),
							ECCErrorsDevice:    pointer.Of(uint64(300)),
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
									IntNumeratorVal:   pointer.Of(int64(1)),
									IntDenominatorVal: pointer.Of(int64(1)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(1)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(1)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
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
									IntNumeratorVal:   pointer.Of(int64(3)),
									IntDenominatorVal: pointer.Of(int64(3)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(3)),
											IntDenominatorVal: pointer.Of(int64(3)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(3)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(3)),
											IntDenominatorVal: pointer.Of(int64(3)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(3)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(300)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(300)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(300)),
										},
									},
								},
								Timestamp: time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC),
							},
							"UUID2": {
								Summary: &structs.StatValue{
									Unit:              MemoryStateUnit,
									Desc:              MemoryStateDesc,
									IntNumeratorVal:   pointer.Of(int64(2)),
									IntDenominatorVal: pointer.Of(int64(2)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(2)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(2)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
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
				devices: map[string]struct{}{
					"UUID1": {},
					"UUID2": {},
				},
				nvmlClient: &MockNvmlClient{
					StatsResponseReturned: []*nvml.StatsData{
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID1",
								DeviceName: pointer.Of("DeviceName1"),
								MemoryMiB:  pointer.Of(uint64(1)),
								PowerW:     pointer.Of(uint(1)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(1)),
							GPUUtilization:     pointer.Of(uint(1)),
							MemoryUtilization:  pointer.Of(uint(1)),
							EncoderUtilization: pointer.Of(uint(1)),
							DecoderUtilization: pointer.Of(uint(1)),
							TemperatureC:       pointer.Of(uint(1)),
							UsedMemoryMiB:      pointer.Of(uint64(1)),
							BAR1UsedMiB:        pointer.Of(uint64(1)),
							ECCErrorsL1Cache:   pointer.Of(uint64(100)),
							ECCErrorsL2Cache:   pointer.Of(uint64(100)),
							ECCErrorsDevice:    pointer.Of(uint64(100)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID2",
								DeviceName: pointer.Of("DeviceName2"),
								MemoryMiB:  pointer.Of(uint64(2)),
								PowerW:     pointer.Of(uint(2)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(2)),
							GPUUtilization:     pointer.Of(uint(2)),
							MemoryUtilization:  pointer.Of(uint(2)),
							EncoderUtilization: pointer.Of(uint(2)),
							DecoderUtilization: pointer.Of(uint(2)),
							TemperatureC:       pointer.Of(uint(2)),
							UsedMemoryMiB:      pointer.Of(uint64(2)),
							BAR1UsedMiB:        pointer.Of(uint64(2)),
							ECCErrorsL1Cache:   pointer.Of(uint64(200)),
							ECCErrorsL2Cache:   pointer.Of(uint64(200)),
							ECCErrorsDevice:    pointer.Of(uint64(200)),
						},
						{
							DeviceData: &nvml.DeviceData{
								UUID:       "UUID3",
								DeviceName: pointer.Of("DeviceName3"),
								MemoryMiB:  pointer.Of(uint64(3)),
								PowerW:     pointer.Of(uint(3)),
								BAR1MiB:    pointer.Of(uint64(256)),
							},
							PowerUsageW:        pointer.Of(uint(3)),
							GPUUtilization:     pointer.Of(uint(3)),
							MemoryUtilization:  pointer.Of(uint(3)),
							EncoderUtilization: pointer.Of(uint(3)),
							DecoderUtilization: pointer.Of(uint(3)),
							TemperatureC:       pointer.Of(uint(3)),
							UsedMemoryMiB:      pointer.Of(uint64(3)),
							BAR1UsedMiB:        pointer.Of(uint64(3)),
							ECCErrorsL1Cache:   pointer.Of(uint64(300)),
							ECCErrorsL2Cache:   pointer.Of(uint64(300)),
							ECCErrorsDevice:    pointer.Of(uint64(300)),
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
									IntNumeratorVal:   pointer.Of(int64(1)),
									IntDenominatorVal: pointer.Of(int64(1)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(1)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(1)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(1)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(1)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(100)),
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
									IntNumeratorVal:   pointer.Of(int64(2)),
									IntDenominatorVal: pointer.Of(int64(2)),
								},
								Stats: &structs.StatObject{
									Attributes: map[string]*structs.StatValue{
										PowerUsageAttr: {
											Unit:              PowerUsageUnit,
											Desc:              PowerUsageDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(2)),
										},
										GPUUtilizationAttr: {
											Unit:            GPUUtilizationUnit,
											Desc:            GPUUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										MemoryUtilizationAttr: {
											Unit:            MemoryUtilizationUnit,
											Desc:            MemoryUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										EncoderUtilizationAttr: {
											Unit:            EncoderUtilizationUnit,
											Desc:            EncoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										DecoderUtilizationAttr: {
											Unit:            DecoderUtilizationUnit,
											Desc:            DecoderUtilizationDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										TemperatureAttr: {
											Unit:            TemperatureUnit,
											Desc:            TemperatureDesc,
											IntNumeratorVal: pointer.Of(int64(2)),
										},
										MemoryStateAttr: {
											Unit:              MemoryStateUnit,
											Desc:              MemoryStateDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(2)),
										},
										BAR1StateAttr: {
											Unit:              BAR1StateUnit,
											Desc:              BAR1StateDesc,
											IntNumeratorVal:   pointer.Of(int64(2)),
											IntDenominatorVal: pointer.Of(int64(256)),
										},
										ECCErrorsL1CacheAttr: {
											Unit:            ECCErrorsL1CacheUnit,
											Desc:            ECCErrorsL1CacheDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
										},
										ECCErrorsL2CacheAttr: {
											Unit:            ECCErrorsL2CacheUnit,
											Desc:            ECCErrorsL2CacheDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
										},
										ECCErrorsDeviceAttr: {
											Unit:            ECCErrorsDeviceUnit,
											Desc:            ECCErrorsDeviceDesc,
											IntNumeratorVal: pointer.Of(int64(200)),
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
