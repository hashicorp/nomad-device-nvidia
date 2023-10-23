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

// List all compute device UUIDs in the system, includes MIG devices
// but excludes their "parent".
func (n *nvmlDriver) ListDeviceUUIDs() ([]string, error) {
	count, code := nvml.DeviceGetCount()
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device count", code)
	}

	var uuids []string
	for i := 0; i < int(count); i++ {
		device, code := nvml.DeviceGetHandleByIndex(int(i))
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device handle", code)
		}

		// Get the device MIG mode, and if MIG is not enabled
		// or the device doesn't support MIG at all (indicated
		// by error code ERROR_NOT_SUPPORTED), then add the
		// device UUID to the list and continue.
		migMode, _, code := nvml.DeviceGetMigMode(device)
		if code == nvml.ERROR_NOT_SUPPORTED || migMode == nvml.DEVICE_MIG_DISABLE {
			uuid, code := nvml.DeviceGetUUID(device)
			if code != nvml.SUCCESS {
				return nil, decode("failed to get device %d uuid", code)
			}

			uuids = append(uuids, uuid)
			continue
		}
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device MIG mode", code)
		}

		count, code = nvml.DeviceGetMaxMigDeviceCount(device)
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device MIG device count", code)
		}

		for j := 0; j < int(count); j++ {
			migDevice, code := nvml.DeviceGetMigDeviceHandleByIndex(device, int(j))
			if code == nvml.ERROR_NOT_FOUND || code == nvml.ERROR_INVALID_ARGUMENT {
				continue
			}
			if code != nvml.SUCCESS {
				return nil, decode("failed to get device MIG device handle", code)
			}

			uuid, code := nvml.DeviceGetUUID(migDevice)
			if code != nvml.SUCCESS {
				return nil, decode(fmt.Sprintf("failed to get mig device uuid %d", j), code)
			}
			uuids = append(uuids, uuid)
		}
	}

	return uuids, nil
}

// DeviceInfoByUUID returns DeviceInfo for the given GPU's UUID.
func (n *nvmlDriver) DeviceInfoByUUID(uuid string) (*DeviceInfo, error) {
	device, code := nvml.DeviceGetHandleByUUID(uuid)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device handle", code)
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
	b := make([]byte, len(id), len(id))
	for i := 0; i < len(id); i++ {
		b[i] = byte(id[i])
	}
	return string(b)
}

// DeviceInfoAndStatusByUUID returns DeviceInfo and DeviceStatus for index GPU in system device list.
func (n *nvmlDriver) DeviceInfoAndStatusByUUID(uuid string) (*DeviceInfo, *DeviceStatus, error) {
	di, err := n.DeviceInfoByUUID(uuid)
	if err != nil {
		return nil, nil, err
	}

	device, code := nvml.DeviceGetHandleByUUID(uuid)
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
