// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build linux

package nvml

import (
	"fmt"
	"maps"
	"syscall"

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
		nvmlDev, code := nvml.DeviceGetHandleByIndex(int(i))
		if code != nvml.SUCCESS {
			return nil, decode(fmt.Sprintf("failed to get device handle %d/%d", i, count), code)
		}

		devIDs, err := uuidsFromDevice(nvmlDev)
		if err != nil {
			return nil, err
		}
		maps.Copy(uuids, devIDs)
	}

	return uuids, nil
}

// DeviceInfoByUUID returns DeviceInfo for the given GPU's UUID.
func (n *nvmlDriver) DeviceInfoByUUID(uuid string) (*DeviceInfo, error) {
	device, code := nvml.DeviceGetHandleByUUID(uuid)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device handle", code)
	}

	info, err := deviceInfoFromDevice(device)
	if err != nil {
		return nil, err
	}
	info.UUID = uuid

	return info, nil
}

func (n *nvmlDriver) DeviceStatusByUUID(uuid string) (*DeviceStatus, error) {
	device, code := nvml.DeviceGetHandleByUUID(uuid)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device info", code)
	}
	return deviceStatusByDevice(device)
}

// DeviceInfoAndStatusByUUID returns DeviceInfo and DeviceStatus for index GPU in system device list.
func (n *nvmlDriver) DeviceInfoAndStatusByUUID(uuid string) (*DeviceInfo, *DeviceStatus, error) {
	di, err := n.DeviceInfoByUUID(uuid)
	if err != nil {
		return nil, nil, err
	}

	ds, err := n.DeviceStatusByUUID(uuid)
	if err != nil {
		return nil, nil, err
	}

	return di, ds, nil
}

// uuidsFromDevice takes a device handle and returns all the UUIDs associated with
// that device. For normal gpu's this will be a single uuid, but for MIG this
// will be severals UUID's.
func uuidsFromDevice(device nvml.Device) (map[string]mode, error) {
	uuids := make(map[string]mode)

	// Get the device MIG mode, and if MIG is not enabled
	// or the device doesn't support MIG at all (indicated
	// by error code ERROR_NOT_SUPPORTED), then add the
	// device UUID to the list and continue.
	migMode, _, code := device.GetMigMode()
	if code == nvml.ERROR_NOT_SUPPORTED || migMode == nvml.DEVICE_MIG_DISABLE {
		uuid, code := device.GetUUID()
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device %d uuid", code)
		}

		uuids[uuid] = normal
		return uuids, nil
	}
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device MIG mode", code)
	}

	migCount, code := device.GetMaxMigDeviceCount()
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device MIG device count", code)
	}

	uuid, code := device.GetUUID()
	if code == nvml.SUCCESS {
		uuids[uuid] = parent
	}

	for j := 0; j < int(migCount); j++ {
		migDevice, code := device.GetMigDeviceHandleByIndex(int(j))
		if code == nvml.ERROR_NOT_FOUND || code == nvml.ERROR_INVALID_ARGUMENT {
			continue
		}
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device MIG device handle", code)
		}

		uuid, code := migDevice.GetUUID()
		if code != nvml.SUCCESS {
			return nil, decode(fmt.Sprintf("failed to get mig device uuid %d", j), code)
		}
		uuids[uuid] = mig
	}

	return uuids, nil
}

func deviceInfoFromDevice(device nvml.Device) (*DeviceInfo, error) {
	name, code := device.GetName()
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device name", code)
	}

	memory, code := device.GetMemoryInfo()
	memoryTotal, _, _, err := determineMemoryInfo(memory, code)
	if err != nil {
		return nil, err
	}

	parentDevice, code := device.GetDeviceHandleFromMigDeviceHandle()
	if code == nvml.ERROR_NOT_FOUND || code == nvml.ERROR_INVALID_ARGUMENT {
		// Device is not a MIG device, so nothing to do.
	} else if code != nvml.SUCCESS {
		return nil, decode("failed to get device parent device handle", code)
	} else {
		// Device is a MIG device, and get the auxilary properties (such as PCIE
		// bandwidth) from the parent device.
		device = parentDevice
	}

	power, code := device.GetPowerUsage()
	if code != nvml.SUCCESS {
		if code == nvml.ERROR_NOT_SUPPORTED {
			power = 0
		} else {
			return nil, decode("failed to get device power info", code)
		}
	}
	powerU := uint(power) / 1000

	bar1, code := device.GetBAR1MemoryInfo()
	var bar1total *uint64
	switch code {
	case nvml.SUCCESS:
		b1val := bytesToMegabytes(bar1.Bar1Total)
		bar1total = &b1val
	case nvml.ERROR_NOT_SUPPORTED:
		bar1total = nil
	default:
		return nil, decode("failed to get device bar 1 memory info", code)
	}

	pci, code := device.GetPciInfo()
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device pci info", code)
	}

	linkWidth, code := device.GetMaxPcieLinkWidth()
	if code != nvml.SUCCESS {
		if code == nvml.ERROR_NOT_SUPPORTED {
			linkWidth = 0
		} else {
			return nil, decode("failed to get pcie link width", code)
		}
	}

	linkGeneration, code := device.GetMaxPcieLinkGeneration()
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

	coreClock, code := device.GetClockInfo(nvml.CLOCK_GRAPHICS)
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device core clock", code)
	}
	coreClockU := uint(coreClock)

	memClock, code := device.GetClockInfo(nvml.CLOCK_MEM)
	var memClockU *uint
	switch code {
	case nvml.SUCCESS:
		val := uint(memClock)
		memClockU = &val
	case nvml.ERROR_NOT_SUPPORTED:
		memClockU = nil
	default:
		return nil, decode("failed to get device mem clock", code)
	}

	mode, code := device.GetDisplayMode()
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device display mode", code)
	}

	persistence, code := device.GetPersistenceMode()
	if code != nvml.SUCCESS {
		return nil, decode("failed to get device persistence mode", code)
	}

	return &DeviceInfo{
		Name:               &name,
		MemoryMiB:          &memoryTotal,
		PowerW:             &powerU,
		BAR1MiB:            bar1total,
		PCIBandwidthMBPerS: &bandwidth,
		PCIBusID:           busID,
		CoresClockMHz:      &coreClockU,
		MemoryClockMHz:     memClockU,
		DisplayState:       fmt.Sprintf("%v", mode),
		PersistenceMode:    fmt.Sprintf("%v", persistence),
	}, nil
}

func deviceStatusByDevice(device nvml.Device) (*DeviceStatus, error) {
	nvmlMemory, code := device.GetMemoryInfo()
	memTotalU, memUsedU, usingSystemMemory, err := determineMemoryInfo(nvmlMemory, code)
	if err != nil {
		return nil, err
	}

	bar, code := device.GetBAR1MemoryInfo()
	var barUsed *uint64
	switch code {
	case nvml.SUCCESS:
		val := bytesToMegabytes(bar.Bar1Used)
		barUsed = &val
	case nvml.ERROR_NOT_SUPPORTED:
		barUsed = nil
	default:
		return nil, decode("failed to get device bar1 memory info", code)
	}

	isMig, code := device.IsMigDeviceHandle()
	if code != nvml.SUCCESS {
		return nil, decode("failed to determine if device handle was mig", code)
	}

	// MIG devices don't have temperature, power usage or utilization properties
	// so just nil them out.
	utzGPU, utzMem, utzEncU, utzDecU := uint(0), uint(0), uint(0), uint(0)
	powerU, tempU := uint(0), uint(0)
	if !isMig {
		utz, code := device.GetUtilizationRates()
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device utilization", code)
		}
		utzGPU = uint(utz.Gpu)
		utzMem = uint(utz.Memory)

		// NVML memory utilization is not reported for iGPU devices, derive it from used/total memory.
		if usingSystemMemory && memTotalU > 0 {
			utzMem = uint((memUsedU * 100) / memTotalU)
		}

		utzEnc, _, code := device.GetEncoderUtilization()
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device encoder utilization", code)
		}
		utzEncU = uint(utzEnc)

		utzDec, _, code := device.GetDecoderUtilization()
		if code != nvml.SUCCESS {
			return nil, decode("failed to get device decoder utilization", code)
		}
		utzDecU = uint(utzDec)

		temp, code := device.GetTemperature(nvml.TEMPERATURE_GPU)
		if code != nvml.SUCCESS {
			if code == nvml.ERROR_NOT_SUPPORTED {
				temp = 0
			} else {
				return nil, decode("failed to get device temperature", code)
			}
		}
		tempU = uint(temp)

		power, code := device.GetPowerUsage()
		if code != nvml.SUCCESS {
			if code == nvml.ERROR_NOT_SUPPORTED {
				power = 0
			} else {
				return nil, decode("failed to get device power usage", code)
			}
		}
		powerU = uint(power)
	}

	ecc, code := device.GetDetailedEccErrors(nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.VOLATILE_ECC)
	if code != nvml.SUCCESS {
		if code == nvml.ERROR_NOT_SUPPORTED {
			ecc = nvml.EccErrorCounts{}
		} else {
			return nil, decode("failed to get device ecc error counts", code)
		}
	}
	return &DeviceStatus{
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

func determineMemoryInfo(memory nvml.Memory, memoryCode nvml.Return) (totalMiB uint64, usedMiB uint64, usingSystemMemory bool, err error) {
	switch memoryCode {
	case nvml.SUCCESS:
		return bytesToMegabytes(memory.Total), bytesToMegabytes(memory.Used), false, nil
	case nvml.ERROR_NOT_SUPPORTED:
		// iGPU systems don't support device memory queries, fall back to system memory.
		var info syscall.Sysinfo_t
		if err := syscall.Sysinfo(&info); err != nil {
			return 0, 0, true, fmt.Errorf("failed to get system memory info: %w", err)
		}

		unit := uint64(info.Unit)
		total := info.Totalram * unit

		free := info.Freeram * unit
		return bytesToMegabytes(total), bytesToMegabytes(total - free), true, nil
	default:
		return 0, 0, false, decode("failed to get device memory info", memoryCode)
	}
}

func bytesToMegabytes(size uint64) uint64 {
	return size / (1 << 20)
}

func buildID(id [32]uint8) string {
	b := make([]byte, len(id))
	for i := range len(id) {
		b[i] = byte(id[i])
	}
	return string(b)
}
