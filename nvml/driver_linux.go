// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build linux

package nvml

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

func decode(msg string, code nvml.Return) error {
	return fmt.Errorf("%s: %s", msg, nvml.ErrorString(code))
}

// Initialize nvml library by locating nvml shared object file and calling ldopen
func (n *nvmlDriver) Initialize() error {
	if code := nvml.Init(); code != nvml.SUCCESS {
		return decode("failed to initialize", code)
	}
	return nil
}

// Shutdown stops any further interaction with nvml
func (n *nvmlDriver) Shutdown() error {
	if code := nvml.Shutdown(); code != nvml.SUCCESS {
		return decode("failed to shutdown", code)
	}
	return nil
}

// SystemDriverVersion returns installed driver version
func (n *nvmlDriver) SystemDriverVersion() (string, error) {
	version, code := nvml.SystemGetDriverVersion()
	if code != nvml.SUCCESS {
		return "", decode("failed to get system driver version", code)
	}
	return version, nil
}

// DeviceCount reports number of available GPU devices
func (n *nvmlDriver) DeviceCount() (uint, error) {
	count, code := nvml.DeviceGetCount()
	if code != nvml.SUCCESS {
		return 0, decode("failed to get device count", code)
	}
	return uint(count), nil
}

// DeviceInfoByIndex returns DeviceInfo for index GPU in system device list.
func (n *nvmlDriver) DeviceInfoByIndex(index uint) (*DeviceInfo, error) {
	device, code := nvml.DeviceGetHandleByIndex(int(index))
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device info", code)
	}

	uuid, code := nvml.DeviceGetUUID(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device uuid", code)
	}

	name, code := nvml.Device.GetName(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device name", code)
	}

	memory, code := nvml.DeviceGetMemoryInfo(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device memory info", code)
	}
	memoryTotal := memory.Total / (1 << 20)

	power, code := nvml.DeviceGetPowerUsage(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device power info", code)
	}
	powerU := uint(power) / 1000

	bar1, code := nvml.DeviceGetBAR1MemoryInfo(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device bar 1 memory info", code)
	}
	bar1total := bar1.Bar1Total / (1 << 20)

	pci, code := nvml.Device.GetPciInfo(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device pci info", code)
	}

	linkWidth, code := nvml.DeviceGetMaxPcieLinkWidth(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get pcie link width", code)
	}

	linkGeneration, code := nvml.DeviceGetMaxPcieLinkGeneration(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get pcie link generation", code)
	}

	// https://en.wikipedia.org/wiki/PCI_Express
	var bandwidth uint
	switch linkGeneration {
	case 6:
		bandwidth = uint(linkWidth) * (4 << 10)
	case 5:
		bandwidth = uint(linkWidth) * (3 << 10)
	case 4:
		bandwidth = uint(linkWidth) * (2 << 10)
	case 3:
		bandwidth = uint(linkWidth) * (1 << 10)
	}

	busID := buildID(pci.BusId)

	coreClock, code := nvml.DeviceGetClockInfo(device, nvml.CLOCK_GRAPHICS)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device core clock", code)
	}
	coreClockU := uint(coreClock)

	memClock, code := nvml.DeviceGetClockInfo(device, nvml.CLOCK_MEM)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device mem clock", code)
	}
	memClockU := uint(memClock)

	mode, code := nvml.DeviceGetDisplayMode(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device display mode", code)
	}

	persistence, code := nvml.DeviceGetPersistenceMode(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device persistence mode", code)
	}

	return &DeviceInfo{
		UUID:               uuid,
		Name:               &name,
		MemoryMiB:          &memoryTotal,
		PowerW:             &powerU,
		BAR1MiB:            &bar1total,
		PCIBandwidthMBPerS: &bandwidth,
		PCIBusID:           busID,
		CoresClockMHz:      &coreClockU,
		MemoryClockMHz:     &memClockU,
		DisplayState:       fmt.Sprintf("%v", mode),
		PersistenceMode:    fmt.Sprintf("%v", persistence),
	}, nil
}

func buildID(id [32]int8) string {
	b := make([]byte, len(id))
	for i := 0; i < len(id); i++ {
		b[i] = byte(id[i])
	}
	return string(b)
}

// DeviceInfoAndStatusByIndex returns DeviceInfo and DeviceStatus for index GPU in system device list.
func (n *nvmlDriver) DeviceInfoAndStatusByIndex(index uint) (*DeviceInfo, *DeviceStatus, error) {
	di, err := n.DeviceInfoByIndex(index)
	if err != nil {
		return nil, nil, err
	}

	device, code := nvml.DeviceGetHandleByIndex(int(index))
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device info", code)
	}

	temp, code := nvml.DeviceGetTemperature(device, nvml.TEMPERATURE_GPU)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device temperature", code)
	}
	tempU := uint(temp)

	utz, code := nvml.DeviceGetUtilizationRates(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device utilization", code)
	}
	utzGPU := uint(utz.Gpu)
	utzMem := uint(utz.Memory)

	utzEnc, _, code := nvml.DeviceGetEncoderUtilization(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device encoder utilization", code)
	}
	utzEncU := uint(utzEnc)

	utzDec, _, code := nvml.Device.GetDecoderUtilization(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device decoder utilization", code)
	}
	utzDecU := uint(utzDec)

	mem, code := nvml.DeviceGetMemoryInfo(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device memory utilization", code)
	}
	memUsedU := mem.Used / (1 << 20)

	power, code := nvml.DeviceGetPowerUsage(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device power usage", code)
	}
	powerU := uint(power)

	bar, code := nvml.DeviceGetBAR1MemoryInfo(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device bar1 memory info", code)
	}
	barUsed := bar.Bar1Used / (1 << 20)

	// note: ecc memory error stats removed; couldn't figure out the API
	return di, &DeviceStatus{
		TemperatureC:       &tempU,
		GPUUtilization:     &utzGPU,
		MemoryUtilization:  &utzMem,
		EncoderUtilization: &utzEncU,
		DecoderUtilization: &utzDecU,
		UsedMemoryMiB:      &memUsedU,
		PowerUsageW:        &powerU,
		BAR1UsedMiB:        &barUsed,
	}, nil
}
