// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build windows

package nvml

import (
	"fmt"
	"syscall"
	"unsafe"
)

// NVML Return codes
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
const (
	NVML_CLOCK_GRAPHICS uint32 = 0
	NVML_CLOCK_SM       uint32 = 1
	NVML_CLOCK_MEM      uint32 = 2
	NVML_CLOCK_VIDEO    uint32 = 3
)

// NVML temperature sensors
const (
	NVML_TEMPERATURE_GPU uint32 = 0
)

// NVML memory error types
const (
	NVML_MEMORY_ERROR_TYPE_CORRECTED   uint32 = 0
	NVML_MEMORY_ERROR_TYPE_UNCORRECTED uint32 = 1
)

// NVML ECC counter types
const (
	NVML_VOLATILE_ECC  uint32 = 0
	NVML_AGGREGATE_ECC uint32 = 1
)

// NVML MIG modes
const (
	NVML_DEVICE_MIG_DISABLE uint32 = 0
	NVML_DEVICE_MIG_ENABLE  uint32 = 1
)

// nvmlMemory_t structure
type nvmlMemory struct {
	Total uint64
	Free  uint64
	Used  uint64
}

// nvmlBAR1Memory_t structure
type nvmlBAR1Memory struct {
	Bar1Total uint64
	Bar1Free  uint64
	Bar1Used  uint64
}

// nvmlUtilization_t structure
type nvmlUtilization struct {
	Gpu    uint32
	Memory uint32
}

// nvmlPciInfo_t structure (simplified)
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
type nvmlEccErrorCounts struct {
	L1Cache      uint64
	L2Cache      uint64
	DeviceMemory uint64
	RegisterFile uint64
}

// Device handle type
type nvmlDevice uintptr

var (
	nvmlDLL *syscall.LazyDLL
	
	procInit                      *syscall.LazyProc
	procShutdown                  *syscall.LazyProc
	procSystemGetDriverVersion    *syscall.LazyProc
	procDeviceGetCount            *syscall.LazyProc
	procDeviceGetHandleByIndex    *syscall.LazyProc
	procDeviceGetHandleByUUID     *syscall.LazyProc
	procDeviceGetUUID             *syscall.LazyProc
	procDeviceGetName             *syscall.LazyProc
	procDeviceGetMemoryInfo       *syscall.LazyProc
	procDeviceGetPowerUsage       *syscall.LazyProc
	procDeviceGetBAR1MemoryInfo   *syscall.LazyProc
	procDeviceGetPciInfo          *syscall.LazyProc
	procDeviceGetMaxPcieLinkWidth *syscall.LazyProc
	procDeviceGetMaxPcieLinkGeneration *syscall.LazyProc
	procDeviceGetClockInfo        *syscall.LazyProc
	procDeviceGetDisplayMode      *syscall.LazyProc
	procDeviceGetPersistenceMode  *syscall.LazyProc
	procDeviceGetMigMode          *syscall.LazyProc
	procDeviceGetMaxMigDeviceCount *syscall.LazyProc
	procDeviceGetMigDeviceHandleByIndex *syscall.LazyProc
	procDeviceGetUtilizationRates *syscall.LazyProc
	procDeviceGetEncoderUtilization *syscall.LazyProc
	procDeviceGetDecoderUtilization *syscall.LazyProc
	procDeviceGetTemperature      *syscall.LazyProc
	procDeviceGetDetailedEccErrors *syscall.LazyProc
)

func init() {
	nvmlDLL = syscall.NewLazyDLL("nvml.dll")
	
	procInit = nvmlDLL.NewProc("nvmlInit_v2")
	procShutdown = nvmlDLL.NewProc("nvmlShutdown")
	procSystemGetDriverVersion = nvmlDLL.NewProc("nvmlSystemGetDriverVersion")
	procDeviceGetCount = nvmlDLL.NewProc("nvmlDeviceGetCount_v2")
	procDeviceGetHandleByIndex = nvmlDLL.NewProc("nvmlDeviceGetHandleByIndex_v2")
	procDeviceGetHandleByUUID = nvmlDLL.NewProc("nvmlDeviceGetHandleByUUID")
	procDeviceGetUUID = nvmlDLL.NewProc("nvmlDeviceGetUUID")
	procDeviceGetName = nvmlDLL.NewProc("nvmlDeviceGetName")
	procDeviceGetMemoryInfo = nvmlDLL.NewProc("nvmlDeviceGetMemoryInfo")
	procDeviceGetPowerUsage = nvmlDLL.NewProc("nvmlDeviceGetPowerUsage")
	procDeviceGetBAR1MemoryInfo = nvmlDLL.NewProc("nvmlDeviceGetBAR1MemoryInfo")
	procDeviceGetPciInfo = nvmlDLL.NewProc("nvmlDeviceGetPciInfo_v3")
	procDeviceGetMaxPcieLinkWidth = nvmlDLL.NewProc("nvmlDeviceGetMaxPcieLinkWidth")
	procDeviceGetMaxPcieLinkGeneration = nvmlDLL.NewProc("nvmlDeviceGetMaxPcieLinkGeneration")
	procDeviceGetClockInfo = nvmlDLL.NewProc("nvmlDeviceGetClockInfo")
	procDeviceGetDisplayMode = nvmlDLL.NewProc("nvmlDeviceGetDisplayMode")
	procDeviceGetPersistenceMode = nvmlDLL.NewProc("nvmlDeviceGetPersistenceMode")
	procDeviceGetMigMode = nvmlDLL.NewProc("nvmlDeviceGetMigMode")
	procDeviceGetMaxMigDeviceCount = nvmlDLL.NewProc("nvmlDeviceGetMaxMigDeviceCount")
	procDeviceGetMigDeviceHandleByIndex = nvmlDLL.NewProc("nvmlDeviceGetMigDeviceHandleByIndex")
	procDeviceGetUtilizationRates = nvmlDLL.NewProc("nvmlDeviceGetUtilizationRates")
	procDeviceGetEncoderUtilization = nvmlDLL.NewProc("nvmlDeviceGetEncoderUtilization")
	procDeviceGetDecoderUtilization = nvmlDLL.NewProc("nvmlDeviceGetDecoderUtilization")
	procDeviceGetTemperature = nvmlDLL.NewProc("nvmlDeviceGetTemperature")
	procDeviceGetDetailedEccErrors = nvmlDLL.NewProc("nvmlDeviceGetDetailedEccErrors")
}

// errorString converts NVML return code to string
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
func nvmlInit() nvmlReturn {
	ret, _, _ := procInit.Call()
	return nvmlReturn(ret)
}

// nvmlShutdown shuts down NVML library
func nvmlShutdown() nvmlReturn {
	ret, _, _ := procShutdown.Call()
	return nvmlReturn(ret)
}

// nvmlSystemGetDriverVersion gets driver version string
func nvmlSystemGetDriverVersion() (string, nvmlReturn) {
	buf := make([]byte, 80)
	ret, _, _ := procSystemGetDriverVersion.Call(
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
func nvmlDeviceGetCount() (uint32, nvmlReturn) {
	var count uint32
	ret, _, _ := procDeviceGetCount.Call(uintptr(unsafe.Pointer(&count)))
	return count, nvmlReturn(ret)
}

// nvmlDeviceGetHandleByIndex gets device handle by index
func nvmlDeviceGetHandleByIndex(index int) (nvmlDevice, nvmlReturn) {
	var device nvmlDevice
	ret, _, _ := procDeviceGetHandleByIndex.Call(
		uintptr(index),
		uintptr(unsafe.Pointer(&device)),
	)
	return device, nvmlReturn(ret)
}

// nvmlDeviceGetHandleByUUID gets device handle by UUID
func nvmlDeviceGetHandleByUUID(uuid string) (nvmlDevice, nvmlReturn) {
	uuidBytes := append([]byte(uuid), 0) // null-terminated
	var device nvmlDevice
	ret, _, _ := procDeviceGetHandleByUUID.Call(
		uintptr(unsafe.Pointer(&uuidBytes[0])),
		uintptr(unsafe.Pointer(&device)),
	)
	return device, nvmlReturn(ret)
}

// nvmlDeviceGetUUID gets device UUID string
func nvmlDeviceGetUUID(device nvmlDevice) (string, nvmlReturn) {
	buf := make([]byte, 80)
	ret, _, _ := procDeviceGetUUID.Call(
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
func nvmlDeviceGetName(device nvmlDevice) (string, nvmlReturn) {
	buf := make([]byte, 96)
	ret, _, _ := procDeviceGetName.Call(
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
func nvmlDeviceGetMemoryInfo(device nvmlDevice) (nvmlMemory, nvmlReturn) {
	var memory nvmlMemory
	ret, _, _ := procDeviceGetMemoryInfo.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&memory)),
	)
	return memory, nvmlReturn(ret)
}

// nvmlDeviceGetPowerUsage gets device power usage in milliwatts
func nvmlDeviceGetPowerUsage(device nvmlDevice) (uint32, nvmlReturn) {
	var power uint32
	ret, _, _ := procDeviceGetPowerUsage.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&power)),
	)
	return power, nvmlReturn(ret)
}

// nvmlDeviceGetBAR1MemoryInfo gets BAR1 memory info
func nvmlDeviceGetBAR1MemoryInfo(device nvmlDevice) (nvmlBAR1Memory, nvmlReturn) {
	var bar1 nvmlBAR1Memory
	ret, _, _ := procDeviceGetBAR1MemoryInfo.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&bar1)),
	)
	return bar1, nvmlReturn(ret)
}

// nvmlDeviceGetPciInfo gets device PCI info
func nvmlDeviceGetPciInfo(device nvmlDevice) (nvmlPciInfo, nvmlReturn) {
	var pci nvmlPciInfo
	ret, _, _ := procDeviceGetPciInfo.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&pci)),
	)
	return pci, nvmlReturn(ret)
}

// nvmlDeviceGetMaxPcieLinkWidth gets max PCIe link width
func nvmlDeviceGetMaxPcieLinkWidth(device nvmlDevice) (uint32, nvmlReturn) {
	var width uint32
	ret, _, _ := procDeviceGetMaxPcieLinkWidth.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&width)),
	)
	return width, nvmlReturn(ret)
}

// nvmlDeviceGetMaxPcieLinkGeneration gets max PCIe link generation
func nvmlDeviceGetMaxPcieLinkGeneration(device nvmlDevice) (uint32, nvmlReturn) {
	var gen uint32
	ret, _, _ := procDeviceGetMaxPcieLinkGeneration.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&gen)),
	)
	return gen, nvmlReturn(ret)
}

// nvmlDeviceGetClockInfo gets device clock info
func nvmlDeviceGetClockInfo(device nvmlDevice, clockType uint32) (uint32, nvmlReturn) {
	var clock uint32
	ret, _, _ := procDeviceGetClockInfo.Call(
		uintptr(device),
		uintptr(clockType),
		uintptr(unsafe.Pointer(&clock)),
	)
	return clock, nvmlReturn(ret)
}

// nvmlDeviceGetDisplayMode gets display mode
func nvmlDeviceGetDisplayMode(device nvmlDevice) (uint32, nvmlReturn) {
	var displayMode uint32
	ret, _, _ := procDeviceGetDisplayMode.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&displayMode)),
	)
	return displayMode, nvmlReturn(ret)
}

// nvmlDeviceGetPersistenceMode gets persistence mode
func nvmlDeviceGetPersistenceMode(device nvmlDevice) (uint32, nvmlReturn) {
	var mode uint32
	ret, _, _ := procDeviceGetPersistenceMode.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&mode)),
	)
	return mode, nvmlReturn(ret)
}

// nvmlDeviceGetMigMode gets MIG mode
func nvmlDeviceGetMigMode(device nvmlDevice) (uint32, uint32, nvmlReturn) {
	var currentMode, pendingMode uint32
	ret, _, _ := procDeviceGetMigMode.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&currentMode)),
		uintptr(unsafe.Pointer(&pendingMode)),
	)
	return currentMode, pendingMode, nvmlReturn(ret)
}

// nvmlDeviceGetMaxMigDeviceCount gets max MIG device count
func nvmlDeviceGetMaxMigDeviceCount(device nvmlDevice) (uint32, nvmlReturn) {
	var count uint32
	ret, _, _ := procDeviceGetMaxMigDeviceCount.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&count)),
	)
	return count, nvmlReturn(ret)
}

// nvmlDeviceGetMigDeviceHandleByIndex gets MIG device handle by index
func nvmlDeviceGetMigDeviceHandleByIndex(device nvmlDevice, index int) (nvmlDevice, nvmlReturn) {
	var migDevice nvmlDevice
	ret, _, _ := procDeviceGetMigDeviceHandleByIndex.Call(
		uintptr(device),
		uintptr(index),
		uintptr(unsafe.Pointer(&migDevice)),
	)
	return migDevice, nvmlReturn(ret)
}

// nvmlDeviceGetUtilizationRates gets GPU and memory utilization
func nvmlDeviceGetUtilizationRates(device nvmlDevice) (nvmlUtilization, nvmlReturn) {
	var util nvmlUtilization
	ret, _, _ := procDeviceGetUtilizationRates.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&util)),
	)
	return util, nvmlReturn(ret)
}

// nvmlDeviceGetEncoderUtilization gets encoder utilization
func nvmlDeviceGetEncoderUtilization(device nvmlDevice) (uint32, uint32, nvmlReturn) {
	var util, samplingPeriod uint32
	ret, _, _ := procDeviceGetEncoderUtilization.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&util)),
		uintptr(unsafe.Pointer(&samplingPeriod)),
	)
	return util, samplingPeriod, nvmlReturn(ret)
}

// nvmlDeviceGetDecoderUtilization gets decoder utilization
func nvmlDeviceGetDecoderUtilization(device nvmlDevice) (uint32, uint32, nvmlReturn) {
	var util, samplingPeriod uint32
	ret, _, _ := procDeviceGetDecoderUtilization.Call(
		uintptr(device),
		uintptr(unsafe.Pointer(&util)),
		uintptr(unsafe.Pointer(&samplingPeriod)),
	)
	return util, samplingPeriod, nvmlReturn(ret)
}

// nvmlDeviceGetTemperature gets device temperature
func nvmlDeviceGetTemperature(device nvmlDevice, sensorType uint32) (uint32, nvmlReturn) {
	var temp uint32
	ret, _, _ := procDeviceGetTemperature.Call(
		uintptr(device),
		uintptr(sensorType),
		uintptr(unsafe.Pointer(&temp)),
	)
	return temp, nvmlReturn(ret)
}

// nvmlDeviceGetDetailedEccErrors gets ECC error counts
func nvmlDeviceGetDetailedEccErrors(device nvmlDevice, errorType uint32, counterType uint32) (nvmlEccErrorCounts, nvmlReturn) {
	var counts nvmlEccErrorCounts
	ret, _, _ := procDeviceGetDetailedEccErrors.Call(
		uintptr(device),
		uintptr(errorType),
		uintptr(counterType),
		uintptr(unsafe.Pointer(&counts)),
	)
	return counts, nvmlReturn(ret)
}
