// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build windows

package nvml

import (
	"fmt"
)

func decode(msg string, code nvmlReturn) error {
	return fmt.Errorf("%s: %s", msg, errorString(code))
}

// Initialize nvml library by locating nvml shared object file and calling ldopen
func (n *nvmlDriver) Initialize() error {
	if code := nvmlInit(); code != NVML_SUCCESS {
		return decode("failed to initialize", code)
	}
	return nil
}

// Shutdown stops any further interaction with nvml
func (n *nvmlDriver) Shutdown() error {
	if code := nvmlShutdown(); code != NVML_SUCCESS {
		return decode("failed to shutdown", code)
	}
	return nil
}

// SystemDriverVersion returns installed driver version
func (n *nvmlDriver) SystemDriverVersion() (string, error) {
	version, code := nvmlSystemGetDriverVersion()
	if code != NVML_SUCCESS {
		return "", decode("failed to get system driver version", code)
	}
	return version, nil
}

func (n *nvmlDriver) ListDeviceUUIDs() (map[string]mode, error) {
	count, code := nvmlDeviceGetCount()
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device count", code)
	}

	uuids := make(map[string]mode)
	for i := 0; i < int(count); i++ {
		device, code := nvmlDeviceGetHandleByIndex(i)
		if code != NVML_SUCCESS {
			return nil, decode(fmt.Sprintf("failed to get device handle %d/%d", i, count), code)
		}

		migMode, _, code := nvmlDeviceGetMigMode(device)
		if code == NVML_ERROR_NOT_SUPPORTED || migMode == NVML_DEVICE_MIG_DISABLE {
			uuid, code := nvmlDeviceGetUUID(device)
			if code != NVML_SUCCESS {
				return nil, decode("failed to get device uuid", code)
			}
			uuids[uuid] = normal
			continue
		}
		if code != NVML_SUCCESS {
			return nil, decode("failed to get device MIG mode", code)
		}

		migCount, code := nvmlDeviceGetMaxMigDeviceCount(device)
		if code != NVML_SUCCESS {
			return nil, decode("failed to get device MIG device count", code)
		}

		uuid, code := nvmlDeviceGetUUID(device)
		if code == NVML_SUCCESS {
			uuids[uuid] = parent
		}

		for j := 0; j < int(migCount); j++ {
			migDevice, code := nvmlDeviceGetMigDeviceHandleByIndex(device, j)
			if code == NVML_ERROR_NOT_FOUND || code == NVML_ERROR_INVALID_ARGUMENT {
				continue
			}
			if code != NVML_SUCCESS {
				return nil, decode("failed to get device MIG device handle", code)
			}
			uuid, code := nvmlDeviceGetUUID(migDevice)
			if code != NVML_SUCCESS {
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

func (n *nvmlDriver) DeviceInfoByUUID(uuid string) (*DeviceInfo, error) {
	device, code := nvmlDeviceGetHandleByUUID(uuid)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device handle", code)
	}

	name, code := nvmlDeviceGetName(device)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device name", code)
	}

	memory, code := nvmlDeviceGetMemoryInfo(device)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device memory info", code)
	}
	memTotal := bytesToMegabytes(memory.Total)

	power, code := nvmlDeviceGetPowerUsage(device)
	if code != NVML_SUCCESS {
		if code == NVML_ERROR_NOT_SUPPORTED {
			power = 0
		} else {
			return nil, decode("failed to get device power info", code)
		}
	}
	powerU := uint(power) / 1000

	bar1, code := nvmlDeviceGetBAR1MemoryInfo(device)
	var bar1total *uint64
	switch code {
	case NVML_SUCCESS:
		b1val := bytesToMegabytes(bar1.Bar1Total)
		bar1total = &b1val
	case NVML_ERROR_NOT_SUPPORTED:
		bar1total = nil
	default:
		return nil, decode("failed to get device bar 1 memory info", code)
	}

	pci, code := nvmlDeviceGetPciInfo(device)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device pci info", code)
	}

	linkWidth, code := nvmlDeviceGetMaxPcieLinkWidth(device)
	if code != NVML_SUCCESS {
		if code == NVML_ERROR_NOT_SUPPORTED {
			linkWidth = 0
		} else {
			return nil, decode("failed to get pcie link width", code)
		}
	}

	linkGeneration, code := nvmlDeviceGetMaxPcieLinkGeneration(device)
	if code != NVML_SUCCESS {
		if code == NVML_ERROR_NOT_SUPPORTED {
			linkGeneration = 0
		} else {
			return nil, decode("failed to get pcie link generation", code)
		}
	}

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

	coreClock, code := nvmlDeviceGetClockInfo(device, NVML_CLOCK_GRAPHICS)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device core clock", code)
	}
	coreClockU := uint(coreClock)

	memClock, code := nvmlDeviceGetClockInfo(device, NVML_CLOCK_MEM)
	var memClockU *uint
	switch code {
	case NVML_SUCCESS:
		val := uint(memClock)
		memClockU = &val
	case NVML_ERROR_NOT_SUPPORTED:
		memClockU = nil
	default:
		return nil, decode("failed to get device mem clock", code)
	}

	displayMode, code := nvmlDeviceGetDisplayMode(device)
	if code != NVML_SUCCESS {
		return nil, decode("failed to get device display mode", code)
	}

	persistence, code := nvmlDeviceGetPersistenceMode(device)
	if code != NVML_SUCCESS {
		if code == NVML_ERROR_NOT_SUPPORTED {
			persistence = 0
		} else {
			return nil, decode("failed to get device persistence mode", code)
		}
	}

	return &DeviceInfo{
		UUID:               uuid,
		Name:               &name,
		MemoryMiB:          &memTotal,
		PowerW:             &powerU,
		BAR1MiB:            bar1total,
		PCIBandwidthMBPerS: &bandwidth,
		PCIBusID:           busID,
		CoresClockMHz:      &coreClockU,
		MemoryClockMHz:     memClockU,
		DisplayState:       fmt.Sprintf("%v", displayMode),
		PersistenceMode:    fmt.Sprintf("%v", persistence),
	}, nil
}

func buildID(id [32]byte) string {
	n := 0
	for n < len(id) && id[n] != 0 {
		n++
	}
	return string(id[:n])
}

func (n *nvmlDriver) DeviceInfoAndStatusByUUID(uuid string) (*DeviceInfo, *DeviceStatus, error) {
	di, err := n.DeviceInfoByUUID(uuid)
	if err != nil {
		return nil, nil, err
	}

	device, code := nvmlDeviceGetHandleByUUID(uuid)
	if code != NVML_SUCCESS {
		return nil, nil, decode("failed to get device info", code)
	}

	nvmlMemory, code := nvmlDeviceGetMemoryInfo(device)
	if code != NVML_SUCCESS {
		return nil, nil, decode("failed to get device memory info", code)
	}
	memUsedU := bytesToMegabytes(nvmlMemory.Used)

	bar, code := nvmlDeviceGetBAR1MemoryInfo(device)
	var barUsed *uint64
	switch code {
	case NVML_SUCCESS:
		val := bytesToMegabytes(bar.Bar1Used)
		barUsed = &val
	case NVML_ERROR_NOT_SUPPORTED:
		barUsed = nil
	default:
		return nil, nil, decode("failed to get device bar1 memory info", code)
	}

	utz, code := nvmlDeviceGetUtilizationRates(device)
	if code != NVML_SUCCESS {
		return nil, nil, decode("failed to get device utilization", code)
	}
	utzGPU := uint(utz.Gpu)
	utzMem := uint(utz.Memory)

	utzEnc, _, code := nvmlDeviceGetEncoderUtilization(device)
	if code != NVML_SUCCESS {
		return nil, nil, decode("failed to get device encoder utilization", code)
	}
	utzEncU := uint(utzEnc)

	utzDec, _, code := nvmlDeviceGetDecoderUtilization(device)
	if code != NVML_SUCCESS {
		return nil, nil, decode("failed to get device decoder utilization", code)
	}
	utzDecU := uint(utzDec)

	temp, code := nvmlDeviceGetTemperature(device, NVML_TEMPERATURE_GPU)
	if code != NVML_SUCCESS {
		if code == NVML_ERROR_NOT_SUPPORTED {
			temp = 0
		} else {
			return nil, nil, decode("failed to get device temperature", code)
		}
	}
	tempU := uint(temp)

	power, code := nvmlDeviceGetPowerUsage(device)
	if code != NVML_SUCCESS {
		if code == NVML_ERROR_NOT_SUPPORTED {
			power = 0
		} else {
			return nil, nil, decode("failed to get device power usage", code)
		}
	}
	powerU := uint(power)

	ecc, code := nvmlDeviceGetDetailedEccErrors(device, NVML_MEMORY_ERROR_TYPE_CORRECTED, NVML_VOLATILE_ECC)
	if code != NVML_SUCCESS {
		if code == NVML_ERROR_NOT_SUPPORTED {
			ecc = nvmlEccErrorCounts{}
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
		BAR1UsedMiB:           barUsed,
		ECCErrorsDevice:       &ecc.DeviceMemory,
		ECCErrorsL1Cache:      &ecc.L1Cache,
		ECCErrorsL2Cache:      &ecc.L2Cache,
		ECCErrorsRegisterFile: &ecc.RegisterFile,
	}, nil
}
