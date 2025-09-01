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

// List all compute device UUIDs in the system.
// Includes all instances, including normal GPUs, MIGs, and their physical parents.
// Each UUID is associated with a mode indication which type it is.
func (n *nvmlDriver) ListDeviceUUIDs() (map[string]mode, error) {
	count, code := nvml.DeviceGetCount()
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device count", code)
	}

	uuids := make(map[string]mode)

	for i := 0; i < int(count); i++ {
		device, code := nvml.DeviceGetHandleByIndex(int(i))
		if code != nvml.SUCCESS {
			return nil, decode(fmt.Sprintf("failed to get device handle %d/%d", i, count), code)
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

			uuids[uuid] = normal
			continue
		}
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device MIG mode", code)
		}

		migCount, code := nvml.DeviceGetMaxMigDeviceCount(device)
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device MIG device count", code)
		}

		uuid, code := nvml.DeviceGetUUID(device)
		if code == nvml.SUCCESS {
			uuids[uuid] = parent
		}

		for j := 0; j < int(migCount); j++ {
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
			uuids[uuid] = mig
		}
	}

	return uuids, nil
}

func bytesToMegabytes(size uint64) uint64 {
	return size / (1 << 20)
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
	memoryTotal := bytesToMegabytes(memory.Total)

	parentDevice, code := nvml.DeviceGetDeviceHandleFromMigDeviceHandle(device)
	if code == nvml.ERROR_NOT_FOUND || code == nvml.ERROR_INVALID_ARGUMENT {
		// Device is not a MIG device, so nothing to do.
	} else if code != nvml.SUCCESS {
		return nil, decode("failed to get device parent device handle", code)
	} else {
		// Device is a MIG device, and get the auxilary properties (such as PCIE
		// bandwidth) from the parent device.
		device = parentDevice
	}

	power, code := nvml.DeviceGetPowerUsage(device)
	if code != nvml.SUCCESS {
		if code == nvml.ERROR_NOT_SUPPORTED {
			power = 0
		} else {
			return nil, decode("failed to get device power info", code)
		}
	}
	powerU := uint(power) / 1000

	bar1, code := nvml.DeviceGetBAR1MemoryInfo(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device bar 1 memory info", code)
	}
	bar1total := bytesToMegabytes(bar1.Bar1Total)

	pci, code := nvml.Device.GetPciInfo(device)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device pci info", code)
	}

	linkWidth, code := nvml.DeviceGetMaxPcieLinkWidth(device)
	if code != nvml.SUCCESS {
		if code == nvml.ERROR_NOT_SUPPORTED {
			linkWidth = 0
		} else {
			return nil, decode("failed to get pcie link width", code)
		}
	}

	linkGeneration, code := nvml.DeviceGetMaxPcieLinkGeneration(device)
	if code != nvml.SUCCESS {
		if code == nvml.ERROR_NOT_SUPPORTED {
			linkGeneration = 0
		} else {
			return nil, decode("failed to get pcie link generation", code)
		}
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

func buildID(id [32]uint8) string {
	b := make([]byte, len(id))
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

	mem, code := nvml.DeviceGetMemoryInfo(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device memory utilization", code)
	}
	memUsedU := bytesToMegabytes(mem.Used)

	bar, code := nvml.DeviceGetBAR1MemoryInfo(device)
	if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device bar1 memory info", code)
	}
	barUsed := bytesToMegabytes(bar.Bar1Used)

	isMig := false
	_, code = nvml.DeviceGetDeviceHandleFromMigDeviceHandle(device)
	if code == nvml.ERROR_NOT_FOUND || code == nvml.ERROR_INVALID_ARGUMENT {
		// Device is not a MIG device.
	} else if code != nvml.SUCCESS {
		return nil, nil, decode("failed to get device parent device handle", code)
	} else {
		isMig = true
	}

	// MIG devices don't have temperature, power usage or utilization properties
	// so just nil them out.
	utzGPU, utzMem, utzEncU, utzDecU := uint(0), uint(0), uint(0), uint(0)
	powerU, tempU := uint(0), uint(0)
	if !isMig {
		utz, code := nvml.DeviceGetUtilizationRates(device)
		if code != nvml.SUCCESS {
			return nil, nil, decode("failed to get device utilization", code)
		}
		utzGPU = uint(utz.Gpu)
		utzMem = uint(utz.Memory)

		utzEnc, _, code := nvml.DeviceGetEncoderUtilization(device)
		if code != nvml.SUCCESS {
			return nil, nil, decode("failed to get device encoder utilization", code)
		}
		utzEncU = uint(utzEnc)

		utzDec, _, code := nvml.Device.GetDecoderUtilization(device)
		if code != nvml.SUCCESS {
			return nil, nil, decode("failed to get device decoder utilization", code)
		}
		utzDecU = uint(utzDec)

		temp, code := nvml.DeviceGetTemperature(device, nvml.TEMPERATURE_GPU)
		if code != nvml.SUCCESS {
			if code == nvml.ERROR_NOT_SUPPORTED {
				temp = 0
			} else {
				return nil, nil, decode("failed to get device temperature", code)
			}
		}
		tempU = uint(temp)

		power, code := nvml.DeviceGetPowerUsage(device)
		if code != nvml.SUCCESS {
			if code == nvml.ERROR_NOT_SUPPORTED {
				power = 0
			} else {
				return nil, nil, decode("failed to get device power usage", code)
			}
		}
		powerU = uint(power)
	}

	ecc, code := nvml.DeviceGetDetailedEccErrors(device, nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.VOLATILE_ECC)
	if code != nvml.SUCCESS {
		if code == nvml.ERROR_NOT_SUPPORTED {
			ecc = nvml.EccErrorCounts{}
		} else {
			return nil, nil, decode("failed to get device ecc error counts", code)
		}
	}

	return di, &DeviceStatus{
		TemperatureC:          &tempU,
		GPUUtilization:        &utzGPU,
		MemoryUtilization:     &utzMem,
		EncoderUtilization:    &utzEncU,
		DecoderUtilization:    &utzDecU,
		UsedMemoryMiB:         &memUsedU,
		PowerUsageW:           &powerU,
		BAR1UsedMiB:           &barUsed,
		ECCErrorsDevice:       &ecc.DeviceMemory,
		ECCErrorsL1Cache:      &ecc.L1Cache,
		ECCErrorsL2Cache:      &ecc.L2Cache,
		ECCErrorsRegisterFile: &ecc.RegisterFile,
	}, nil
}
