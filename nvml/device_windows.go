// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build windows

package nvml

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	nvmlmock "github.com/NVIDIA/go-nvml/pkg/nvml/mock"
)

// newWinDevice constructs an nvml.Device implementation backed by Windows syscalls.
// It leverages the generated mock.Device type to satisfy the full nvml.Device
// interface, wiring only the methods actually used by the helper functions in
// driver_linux.go (uuidsFromDevice, deviceInfoFromDevice, deviceStatusByDevice).
func newWinDevice(handle nvmlDevice) nvml.Device {
	d := &nvmlmock.Device{}

	d.GetUUIDFunc = func() (string, nvml.Return) {
		uuid, ret := nvmlDeviceGetUUID(handle)
		return uuid, nvml.Return(ret)
	}

	d.GetNameFunc = func() (string, nvml.Return) {
		name, ret := nvmlDeviceGetName(handle)
		return name, nvml.Return(ret)
	}

	d.GetMemoryInfoFunc = func() (nvml.Memory, nvml.Return) {
		mem, ret := nvmlDeviceGetMemoryInfo(handle)
		return nvml.Memory{
			Total: mem.Total,
			Free:  mem.Free,
			Used:  mem.Used,
		}, nvml.Return(ret)
	}

	d.GetBAR1MemoryInfoFunc = func() (nvml.BAR1Memory, nvml.Return) {
		bar, ret := nvmlDeviceGetBAR1MemoryInfo(handle)
		return nvml.BAR1Memory{
			Bar1Total: bar.Bar1Total,
			Bar1Free:  bar.Bar1Free,
			Bar1Used:  bar.Bar1Used,
		}, nvml.Return(ret)
	}

	d.GetPciInfoFunc = func() (nvml.PciInfo, nvml.Return) {
		pci, ret := nvmlDeviceGetPciInfo(handle)
		return nvml.PciInfo{
			BusIdLegacy:    pci.BusIdLegacy,
			Domain:         pci.Domain,
			Bus:            pci.Bus,
			Device:         pci.Device,
			PciDeviceId:    pci.PciDeviceId,
			PciSubSystemId: pci.PciSubSystemId,
			BusId:          pci.BusId,
		}, nvml.Return(ret)
	}

	d.GetMaxPcieLinkWidthFunc = func() (int, nvml.Return) {
		width, ret := nvmlDeviceGetMaxPcieLinkWidth(handle)
		return int(width), nvml.Return(ret)
	}

	d.GetMaxPcieLinkGenerationFunc = func() (int, nvml.Return) {
		gen, ret := nvmlDeviceGetMaxPcieLinkGeneration(handle)
		return int(gen), nvml.Return(ret)
	}

	d.GetClockInfoFunc = func(clockType nvml.ClockType) (uint32, nvml.Return) {
		clock, ret := nvmlDeviceGetClockInfo(handle, uint32(clockType))
		return clock, nvml.Return(ret)
	}

	d.GetDisplayModeFunc = func() (nvml.EnableState, nvml.Return) {
		mode, ret := nvmlDeviceGetDisplayMode(handle)
		return nvml.EnableState(mode), nvml.Return(ret)
	}

	d.GetPersistenceModeFunc = func() (nvml.EnableState, nvml.Return) {
		mode, ret := nvmlDeviceGetPersistenceMode(handle)
		return nvml.EnableState(mode), nvml.Return(ret)
	}

	d.GetMigModeFunc = func() (int, int, nvml.Return) {
		current, pending, ret := nvmlDeviceGetMigMode(handle)
		return int(current), int(pending), nvml.Return(ret)
	}

	d.GetMaxMigDeviceCountFunc = func() (int, nvml.Return) {
		count, ret := nvmlDeviceGetMaxMigDeviceCount(handle)
		return int(count), nvml.Return(ret)
	}

	d.GetMigDeviceHandleByIndexFunc = func(index int) (nvml.Device, nvml.Return) {
		migHandle, ret := nvmlDeviceGetMigDeviceHandleByIndex(handle, index)
		if ret != NVML_SUCCESS {
			return nil, nvml.Return(ret)
		}
		return newWinDevice(migHandle), nvml.Return(ret)
	}

	d.GetUtilizationRatesFunc = func() (nvml.Utilization, nvml.Return) {
		util, ret := nvmlDeviceGetUtilizationRates(handle)
		return nvml.Utilization{
			Gpu:    util.Gpu,
			Memory: util.Memory,
		}, nvml.Return(ret)
	}

	d.GetEncoderUtilizationFunc = func() (uint32, uint32, nvml.Return) {
		util, period, ret := nvmlDeviceGetEncoderUtilization(handle)
		return util, period, nvml.Return(ret)
	}

	d.GetDecoderUtilizationFunc = func() (uint32, uint32, nvml.Return) {
		util, period, ret := nvmlDeviceGetDecoderUtilization(handle)
		return util, period, nvml.Return(ret)
	}

	d.GetTemperatureFunc = func(sensor nvml.TemperatureSensors) (uint32, nvml.Return) {
		temp, ret := nvmlDeviceGetTemperature(handle, uint32(sensor))
		return temp, nvml.Return(ret)
	}

	d.GetPowerUsageFunc = func() (uint32, nvml.Return) {
		power, ret := nvmlDeviceGetPowerUsage(handle)
		return power, nvml.Return(ret)
	}

	d.GetDetailedEccErrorsFunc = func(memErr nvml.MemoryErrorType, counter nvml.EccCounterType) (nvml.EccErrorCounts, nvml.Return) {
		ecc, ret := nvmlDeviceGetDetailedEccErrors(handle, uint32(memErr), uint32(counter))
		return nvml.EccErrorCounts{
			L1Cache:      ecc.L1Cache,
			L2Cache:      ecc.L2Cache,
			DeviceMemory: ecc.DeviceMemory,
			RegisterFile: ecc.RegisterFile,
		}, nvml.Return(ret)
	}

	d.GetDeviceHandleFromMigDeviceHandleFunc = func() (nvml.Device, nvml.Return) {
		parent, ret := nvmlDeviceGetDeviceHandleFromMigDeviceHandle(handle)
		if ret != NVML_SUCCESS {
			return nil, nvml.Return(ret)
		}
		return newWinDevice(parent), nvml.Return(ret)
	}

	d.IsMigDeviceHandleFunc = func() (bool, nvml.Return) {
		isMig, ret := nvmlDeviceIsMigDeviceHandle(handle)
		return isMig, nvml.Return(ret)
	}

	return d
}
