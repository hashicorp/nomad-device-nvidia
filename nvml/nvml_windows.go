// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build windows

package nvml

import (
	"fmt"
	"golang.org/x/sys/windows"
	"sync"
	"unsafe"
)

// NVML Return codes
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlReturnValues.html
type nvmlReturn uint32

const (
	NVML_SUCCESS                   nvmlReturn = 0
	NVML_ERROR_UNINITIALIZED       nvmlReturn = 1
	NVML_ERROR_INVALID_ARGUMENT    nvmlReturn = 2
	NVML_ERROR_NOT_SUPPORTED       nvmlReturn = 3
	NVML_ERROR_NO_PERMISSION       nvmlReturn = 4
	NVML_ERROR_ALREADY_INITIALIZED nvmlReturn = 5
	NVML_ERROR_NOT_FOUND           nvmlReturn = 6
	NVML_ERROR_INSUFFICIENT_SIZE   nvmlReturn = 7
	NVML_ERROR_INSUFFICIENT_POWER  nvmlReturn = 8
	NVML_ERROR_DRIVER_NOT_LOADED   nvmlReturn = 9
	NVML_ERROR_TIMEOUT             nvmlReturn = 10
	NVML_ERROR_IRQ_ISSUE           nvmlReturn = 11
	NVML_ERROR_LIBRARY_NOT_FOUND   nvmlReturn = 12
	NVML_ERROR_FUNCTION_NOT_FOUND  nvmlReturn = 13
	NVML_ERROR_CORRUPTED_INFOROM   nvmlReturn = 14
	NVML_ERROR_GPU_IS_LOST         nvmlReturn = 15
	NVML_ERROR_RESET_REQUIRED      nvmlReturn = 16
	NVML_ERROR_OPERATING_SYSTEM    nvmlReturn = 17
	NVML_ERROR_LIB_RM_VERSION_MISMATCH nvmlReturn = 18
	NVML_ERROR_IN_USE              nvmlReturn = 19
	NVML_ERROR_MEMORY              nvmlReturn = 20
	NVML_ERROR_NO_DATA             nvmlReturn = 21
	NVML_ERROR_VGPU_ECC_NOT_SUPPORTED nvmlReturn = 22
	NVML_ERROR_UNKNOWN             nvmlReturn = 999
)

// NVML clock types
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlClockTypes.html
const (
	NVML_CLOCK_GRAPHICS uint32 = 0
	NVML_CLOCK_SM       uint32 = 1
	NVML_CLOCK_MEM      uint32 = 2
	NVML_CLOCK_VIDEO    uint32 = 3
)

// NVML temperature sensors
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlTemperatureThresholds.html
const (
	NVML_TEMPERATURE_GPU uint32 = 0
)

// NVML memory error types
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlMemoryErrorTypes.html
const (
	NVML_MEMORY_ERROR_TYPE_CORRECTED   uint32 = 0
	NVML_MEMORY_ERROR_TYPE_UNCORRECTED uint32 = 1
)

// NVML ECC counter types
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlEccCounterTypes.html
const (
	NVML_VOLATILE_ECC  uint32 = 0
	NVML_AGGREGATE_ECC uint32 = 1
)

// NVML MIG modes
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceEnumvs.html#group__nvmlDeviceEnumvs_1g08c860c1f4f23e6d0c8e9c6f5e8e0d1a
const (
	NVML_DEVICE_MIG_DISABLE uint32 = 0
	NVML_DEVICE_MIG_ENABLE  uint32 = 1
)

// nvmlMemory_t structure
// See: https://docs.nvidia.com/deploy/nvml-api/structnvmlMemory__t.html
type nvmlMemory struct {
	Total uint64
	Free  uint64
	Used  uint64
}

// nvmlBAR1Memory_t structure
// See: https://docs.nvidia.com/deploy/nvml-api/structnvmlBAR1Memory__t.html
type nvmlBAR1Memory struct {
	Bar1Total uint64
	Bar1Free  uint64
	Bar1Used  uint64
}

// nvmlUtilization_t structure
// See: https://docs.nvidia.com/deploy/nvml-api/structnvmlUtilization__t.html
type nvmlUtilization struct {
	Gpu    uint32
	Memory uint32
}

// nvmlPciInfo_t structure
// See: https://docs.nvidia.com/deploy/nvml-api/structnvmlPciInfo__t.html
type nvmlPciInfo struct {
	BusIdLegacy      [16]byte
	Domain           uint32
	Bus              uint32
	Device           uint32
	PciDeviceId      uint32
	PciSubSystemId   uint32
	BusId            [32]byte
}

// nvmlEccErrorCounts_t structure
// See: https://docs.nvidia.com/deploy/nvml-api/structnvmlEccErrorCounts__t.html
type nvmlEccErrorCounts struct {
	L1Cache      uint64
	L2Cache      uint64
	DeviceMemory uint64
	RegisterFile uint64
}

// Device handle type
type nvmlDevice uintptr

// NVML library handle and function pointers
var (
	nvmlLibOnce sync.Once
	nvmlLib     windows.Handle
	nvmlLibErr  error

	// Function pointers
	procInit                      uintptr
	procShutdown                  uintptr
	procSystemGetDriverVersion    uintptr
	procDeviceGetCount            uintptr
	procDeviceGetHandleByIndex    uintptr
	procDeviceGetHandleByUUID     uintptr
	procDeviceGetUUID             uintptr
	procDeviceGetName             uintptr
	procDeviceGetMemoryInfo       uintptr
	procDeviceGetPowerUsage       uintptr
	procDeviceGetBAR1MemoryInfo   uintptr
	procDeviceGetPciInfo          uintptr
	procDeviceGetMaxPcieLinkWidth uintptr
	procDeviceGetMaxPcieLinkGeneration uintptr
	procDeviceGetClockInfo        uintptr
	procDeviceGetDisplayMode      uintptr
	procDeviceGetPersistenceMode  uintptr
	procDeviceGetMigMode          uintptr
	procDeviceGetMaxMigDeviceCount uintptr
	procDeviceGetMigDeviceHandleByIndex uintptr
	procDeviceGetUtilizationRates uintptr
	procDeviceGetEncoderUtilization uintptr
	procDeviceGetDecoderUtilization uintptr
	procDeviceGetTemperature      uintptr
	procDeviceGetDetailedEccErrors uintptr
	procDeviceGetDeviceHandleFromMigDeviceHandle uintptr
	procDeviceIsMigDeviceHandle uintptr
)

// initNVMLLibrary loads nvml.dll and resolves function pointers
// Uses golang.org/x/sys/windows.LoadLibrary instead of syscall.LazyDLL
func initNVMLLibrary() error {
	nvmlLibOnce.Do(func() {
		// Load nvml.dll using windows.LoadLibrary
		// See: https://docs.nvidia.com/deploy/nvml-api/index.html
		nvmlLib, nvmlLibErr = windows.LoadLibrary("nvml.dll")
		if nvmlLibErr != nil {
			return
		}

		// Resolve all required function pointers using GetProcAddress
		procs := []struct {
			name string
			ptr  *uintptr
		}{
			{"nvmlInit_v2", &procInit},
			{"nvmlShutdown", &procShutdown},
			{"nvmlSystemGetDriverVersion", &procSystemGetDriverVersion},
			{"nvmlDeviceGetCount_v2", &procDeviceGetCount},
			{"nvmlDeviceGetHandleByIndex_v2", &procDeviceGetHandleByIndex},
			{"nvmlDeviceGetHandleByUUID", &procDeviceGetHandleByUUID},
			{"nvmlDeviceGetUUID", &procDeviceGetUUID},
			{"nvmlDeviceGetName", &procDeviceGetName},
			{"nvmlDeviceGetMemoryInfo", &procDeviceGetMemoryInfo},
			{"nvmlDeviceGetPowerUsage", &procDeviceGetPowerUsage},
			{"nvmlDeviceGetBAR1MemoryInfo", &procDeviceGetBAR1MemoryInfo},
			{"nvmlDeviceGetPciInfo_v3", &procDeviceGetPciInfo},
			{"nvmlDeviceGetMaxPcieLinkWidth", &procDeviceGetMaxPcieLinkWidth},
			{"nvmlDeviceGetMaxPcieLinkGeneration", &procDeviceGetMaxPcieLinkGeneration},
			{"nvmlDeviceGetClockInfo", &procDeviceGetClockInfo},
			{"nvmlDeviceGetDisplayMode", &procDeviceGetDisplayMode},
			{"nvmlDeviceGetPersistenceMode", &procDeviceGetPersistenceMode},
			{"nvmlDeviceGetMigMode", &procDeviceGetMigMode},
			{"nvmlDeviceGetMaxMigDeviceCount", &procDeviceGetMaxMigDeviceCount},
			{"nvmlDeviceGetMigDeviceHandleByIndex", &procDeviceGetMigDeviceHandleByIndex},
			{"nvmlDeviceGetUtilizationRates", &procDeviceGetUtilizationRates},
			{"nvmlDeviceGetEncoderUtilization", &procDeviceGetEncoderUtilization},
			{"nvmlDeviceGetDecoderUtilization", &procDeviceGetDecoderUtilization},
			{"nvmlDeviceGetTemperature", &procDeviceGetTemperature},
			{"nvmlDeviceGetDetailedEccErrors", &procDeviceGetDetailedEccErrors},
			{"nvmlDeviceGetDeviceHandleFromMigDeviceHandle", &procDeviceGetDeviceHandleFromMigDeviceHandle},
			{"nvmlDeviceIsMigDeviceHandle", &procDeviceIsMigDeviceHandle},
		}

		for _, p := range procs {
			*p.ptr, nvmlLibErr = windows.GetProcAddress(nvmlLib, p.name)
			if nvmlLibErr != nil {
				return
			}
		}
	})
	return nvmlLibErr
}

// errorString converts NVML return code to string
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlReturnValues.html
func errorString(ret nvmlReturn) string {
	switch ret {
	case NVML_SUCCESS:
		return "Success"
	case NVML_ERROR_UNINITIALIZED:
		return "Uninitialized"
	case NVML_ERROR_INVALID_ARGUMENT:
		return "Invalid Argument"
	case NVML_ERROR_NOT_SUPPORTED:
		return "Not Supported"
	case NVML_ERROR_NO_PERMISSION:
		return "No Permission"
	case NVML_ERROR_ALREADY_INITIALIZED:
		return "Already Initialized"
	case NVML_ERROR_NOT_FOUND:
		return "Not Found"
	case NVML_ERROR_INSUFFICIENT_SIZE:
		return "Insufficient Size"
	case NVML_ERROR_DRIVER_NOT_LOADED:
		return "Driver Not Loaded"
	case NVML_ERROR_LIBRARY_NOT_FOUND:
		return "Library Not Found"
	case NVML_ERROR_FUNCTION_NOT_FOUND:
		return "Function Not Found"
	case NVML_ERROR_GPU_IS_LOST:
		return "GPU Is Lost"
	default:
		return fmt.Sprintf("Unknown Error (%d)", ret)
	}
}

// nvmlInit initializes NVML library
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlLibraryQueries.html#group__nvmlLibraryQueries_1gb8737e5dc77a4c0b3a47ee3c9f4bf9aa
func nvmlInit() nvmlReturn {
	if err := initNVMLLibrary(); err != nil {
		return NVML_ERROR_LIBRARY_NOT_FOUND
	}
	ret, _, _ := windows.SyscallN(procInit)
	return nvmlReturn(ret)
}

// nvmlShutdown shuts down NVML library
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlLibraryQueries.html#group__nvmlLibraryQueries_1g5e8052da0c1f7e8c82fd9b3c4ce11d52
func nvmlShutdown() nvmlReturn {
	ret, _, _ := windows.SyscallN(procShutdown)
	return nvmlReturn(ret)
}

// nvmlSystemGetDriverVersion gets driver version string
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlSystemQueries.html#group__nvmlSystemQueries_1g49a4ceee74c5b6424ba0ff6f8ec4b489
func nvmlSystemGetDriverVersion() (string, nvmlReturn) {
	buf := make([]byte, 80)
	ret, _, _ := windows.SyscallN(procSystemGetDriverVersion,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if nvmlReturn(ret) != NVML_SUCCESS {
		return "", nvmlReturn(ret)
	}
	// Find null terminator
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(buf[:n]), NVML_SUCCESS
}

// nvmlDeviceGetCount gets number of GPU devices
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g0d7da2c2dd7d5c0d4a8c4efc7d6e0b1a
func nvmlDeviceGetCount() (uint32, nvmlReturn) {
	var count uint32
	ret, _, _ := windows.SyscallN(procDeviceGetCount, uintptr(unsafe.Pointer(&count)))
	return count, nvmlReturn(ret)
}

// nvmlDeviceGetHandleByIndex gets device handle by index
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g2a0551d808d3e6c6ae4df0c78a7e3e6a
func nvmlDeviceGetHandleByIndex(index int) (nvmlDevice, nvmlReturn) {
	var device nvmlDevice
	ret, _, _ := windows.SyscallN(procDeviceGetHandleByIndex,
		uintptr(uint32(index)),
		uintptr(unsafe.Pointer(&device)),
	)
	return device, nvmlReturn(ret)
}

// nvmlDeviceGetHandleByUUID gets device handle by UUID
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g5be56d6e6f1e3f9a14c35c4e3f5c4e3f
func nvmlDeviceGetHandleByUUID(uuid string) (nvmlDevice, nvmlReturn) {
	uuidBytes := append([]byte(uuid), 0) // null-terminated
	var device nvmlDevice
	ret, _, _ := windows.SyscallN(procDeviceGetHandleByUUID,
		uintptr(unsafe.Pointer(&uuidBytes[0])),
		uintptr(unsafe.Pointer(&device)),
	)
	return device, nvmlReturn(ret)
}

// nvmlDeviceGetUUID gets device UUID string
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g8bc4a2e9c2c4e3f1a5b6c7d8e9f0a1b2
func nvmlDeviceGetUUID(device nvmlDevice) (string, nvmlReturn) {
	buf := make([]byte, 80)
	ret, _, _ := windows.SyscallN(procDeviceGetUUID,
		uintptr(device),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if nvmlReturn(ret) != NVML_SUCCESS {
		return "", nvmlReturn(ret)
	}
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(buf[:n]), NVML_SUCCESS
}

// nvmlDeviceGetName gets device name
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g4a6c8e9b7f6e5d4c3b2a190876543210
func nvmlDeviceGetName(device nvmlDevice) (string, nvmlReturn) {
	buf := make([]byte, 96)
	ret, _, _ := windows.SyscallN(procDeviceGetName,
		uintptr(device),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if nvmlReturn(ret) != NVML_SUCCESS {
		return "", nvmlReturn(ret)
	}
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(buf[:n]), NVML_SUCCESS
}

// nvmlDeviceGetMemoryInfo gets device memory info
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g1d3a4c5b7e8f9a0b1c2d3e4f5a6b7c8d
func nvmlDeviceGetMemoryInfo(device nvmlDevice) (nvmlMemory, nvmlReturn) {
	var memory nvmlMemory
	ret, _, _ := windows.SyscallN(procDeviceGetMemoryInfo,
		uintptr(device),
		uintptr(unsafe.Pointer(&memory)),
	)
	return memory, nvmlReturn(ret)
}

// nvmlDeviceGetPowerUsage gets device power usage in milliwatts
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g7f7c7e8d9a0b1c2d3e4f5a6b7c8d9e0f
func nvmlDeviceGetPowerUsage(device nvmlDevice) (uint32, nvmlReturn) {
	var power uint32
	ret, _, _ := windows.SyscallN(procDeviceGetPowerUsage,
		uintptr(device),
		uintptr(unsafe.Pointer(&power)),
	)
	return power, nvmlReturn(ret)
}

// nvmlDeviceGetBAR1MemoryInfo gets BAR1 memory info
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d
func nvmlDeviceGetBAR1MemoryInfo(device nvmlDevice) (nvmlBAR1Memory, nvmlReturn) {
	var bar1 nvmlBAR1Memory
	ret, _, _ := windows.SyscallN(procDeviceGetBAR1MemoryInfo,
		uintptr(device),
		uintptr(unsafe.Pointer(&bar1)),
	)
	return bar1, nvmlReturn(ret)
}

// nvmlDeviceGetPciInfo gets device PCI info
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e
func nvmlDeviceGetPciInfo(device nvmlDevice) (nvmlPciInfo, nvmlReturn) {
	var pci nvmlPciInfo
	ret, _, _ := windows.SyscallN(procDeviceGetPciInfo,
		uintptr(device),
		uintptr(unsafe.Pointer(&pci)),
	)
	return pci, nvmlReturn(ret)
}

// nvmlDeviceGetMaxPcieLinkWidth gets max PCIe link width
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f
func nvmlDeviceGetMaxPcieLinkWidth(device nvmlDevice) (uint32, nvmlReturn) {
	var width uint32
	ret, _, _ := windows.SyscallN(procDeviceGetMaxPcieLinkWidth,
		uintptr(device),
		uintptr(unsafe.Pointer(&width)),
	)
	return width, nvmlReturn(ret)
}

// nvmlDeviceGetMaxPcieLinkGeneration gets max PCIe link generation
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a
func nvmlDeviceGetMaxPcieLinkGeneration(device nvmlDevice) (uint32, nvmlReturn) {
	var gen uint32
	ret, _, _ := windows.SyscallN(procDeviceGetMaxPcieLinkGeneration,
		uintptr(device),
		uintptr(unsafe.Pointer(&gen)),
	)
	return gen, nvmlReturn(ret)
}

// nvmlDeviceGetClockInfo gets device clock info
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b
func nvmlDeviceGetClockInfo(device nvmlDevice, clockType uint32) (uint32, nvmlReturn) {
	var clock uint32
	ret, _, _ := windows.SyscallN(procDeviceGetClockInfo,
		uintptr(device),
		uintptr(clockType),
		uintptr(unsafe.Pointer(&clock)),
	)
	return clock, nvmlReturn(ret)
}

// nvmlDeviceGetDisplayMode gets display mode
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c
func nvmlDeviceGetDisplayMode(device nvmlDevice) (uint32, nvmlReturn) {
	var displayMode uint32
	ret, _, _ := windows.SyscallN(procDeviceGetDisplayMode,
		uintptr(device),
		uintptr(unsafe.Pointer(&displayMode)),
	)
	return displayMode, nvmlReturn(ret)
}

// nvmlDeviceGetPersistenceMode gets persistence mode
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d
func nvmlDeviceGetPersistenceMode(device nvmlDevice) (uint32, nvmlReturn) {
	var mode uint32
	ret, _, _ := windows.SyscallN(procDeviceGetPersistenceMode,
		uintptr(device),
		uintptr(unsafe.Pointer(&mode)),
	)
	return mode, nvmlReturn(ret)
}

// nvmlDeviceGetMigMode gets MIG mode
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e
func nvmlDeviceGetMigMode(device nvmlDevice) (uint32, uint32, nvmlReturn) {
	var currentMode, pendingMode uint32
	ret, _, _ := windows.SyscallN(procDeviceGetMigMode,
		uintptr(device),
		uintptr(unsafe.Pointer(&currentMode)),
		uintptr(unsafe.Pointer(&pendingMode)),
	)
	return currentMode, pendingMode, nvmlReturn(ret)
}

// nvmlDeviceGetMaxMigDeviceCount gets max MIG device count
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f
func nvmlDeviceGetMaxMigDeviceCount(device nvmlDevice) (uint32, nvmlReturn) {
	var count uint32
	ret, _, _ := windows.SyscallN(procDeviceGetMaxMigDeviceCount,
		uintptr(device),
		uintptr(unsafe.Pointer(&count)),
	)
	return count, nvmlReturn(ret)
}

// nvmlDeviceGetMigDeviceHandleByIndex gets MIG device handle by index
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a
func nvmlDeviceGetMigDeviceHandleByIndex(device nvmlDevice, index int) (nvmlDevice, nvmlReturn) {
	var migDevice nvmlDevice
	ret, _, _ := windows.SyscallN(procDeviceGetMigDeviceHandleByIndex,
		uintptr(device),
		uintptr(uint32(index)),
		uintptr(unsafe.Pointer(&migDevice)),
	)
	return migDevice, nvmlReturn(ret)
}

// nvmlDeviceGetUtilizationRates gets GPU and memory utilization
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b
func nvmlDeviceGetUtilizationRates(device nvmlDevice) (nvmlUtilization, nvmlReturn) {
	var util nvmlUtilization
	ret, _, _ := windows.SyscallN(procDeviceGetUtilizationRates,
		uintptr(device),
		uintptr(unsafe.Pointer(&util)),
	)
	return util, nvmlReturn(ret)
}

// nvmlDeviceGetEncoderUtilization gets encoder utilization
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c
func nvmlDeviceGetEncoderUtilization(device nvmlDevice) (uint32, uint32, nvmlReturn) {
	var util, samplingPeriod uint32
	ret, _, _ := windows.SyscallN(procDeviceGetEncoderUtilization,
		uintptr(device),
		uintptr(unsafe.Pointer(&util)),
		uintptr(unsafe.Pointer(&samplingPeriod)),
	)
	return util, samplingPeriod, nvmlReturn(ret)
}

// nvmlDeviceGetDecoderUtilization gets decoder utilization
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d
func nvmlDeviceGetDecoderUtilization(device nvmlDevice) (uint32, uint32, nvmlReturn) {
	var util, samplingPeriod uint32
	ret, _, _ := windows.SyscallN(procDeviceGetDecoderUtilization,
		uintptr(device),
		uintptr(unsafe.Pointer(&util)),
		uintptr(unsafe.Pointer(&samplingPeriod)),
	)
	return util, samplingPeriod, nvmlReturn(ret)
}

// nvmlDeviceGetTemperature gets device temperature
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e
func nvmlDeviceGetTemperature(device nvmlDevice, sensorType uint32) (uint32, nvmlReturn) {
	var temp uint32
	ret, _, _ := windows.SyscallN(procDeviceGetTemperature,
		uintptr(device),
		uintptr(sensorType),
		uintptr(unsafe.Pointer(&temp)),
	)
	return temp, nvmlReturn(ret)
}

// nvmlDeviceGetDetailedEccErrors gets ECC error counts
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f
func nvmlDeviceGetDetailedEccErrors(device nvmlDevice, errorType uint32, counterType uint32) (nvmlEccErrorCounts, nvmlReturn) {
	var counts nvmlEccErrorCounts
	ret, _, _ := windows.SyscallN(procDeviceGetDetailedEccErrors,
		uintptr(device),
		uintptr(errorType),
		uintptr(counterType),
		uintptr(unsafe.Pointer(&counts)),
	)
	return counts, nvmlReturn(ret)
}

// nvmlDeviceGetDeviceHandleFromMigDeviceHandle gets parent device from MIG device
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a
func nvmlDeviceGetDeviceHandleFromMigDeviceHandle(device nvmlDevice) (nvmlDevice, nvmlReturn) {
	var parentDevice nvmlDevice
	ret, _, _ := windows.SyscallN(procDeviceGetDeviceHandleFromMigDeviceHandle,
		uintptr(device),
		uintptr(unsafe.Pointer(&parentDevice)),
	)
	return parentDevice, nvmlReturn(ret)
}

// nvmlDeviceIsMigDeviceHandle checks if device is a MIG device
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b
func nvmlDeviceIsMigDeviceHandle(device nvmlDevice) (bool, nvmlReturn) {
	var isMig uint32
	ret, _, _ := windows.SyscallN(procDeviceIsMigDeviceHandle,
		uintptr(device),
		uintptr(unsafe.Pointer(&isMig)),
	)
	return isMig != 0, nvmlReturn(ret)
}
