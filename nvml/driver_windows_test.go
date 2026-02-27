// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build windows

package nvml

import (
	"strings"
	"testing"
)

func TestNVMLInitShutdown(t *testing.T) {
	driver := &nvmlDriver{}
	
	err := driver.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	
	err = driver.Shutdown()
	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
}

func TestSystemDriverVersion(t *testing.T) {
	driver := &nvmlDriver{}
	
	err := driver.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer driver.Shutdown()
	
	version, err := driver.SystemDriverVersion()
	if err != nil {
		t.Fatalf("SystemDriverVersion() failed: %v", err)
	}
	
	if version == "" {
		t.Error("SystemDriverVersion() returned empty string")
	}
	
	t.Logf("Driver version: %s", version)
}

func TestListDeviceUUIDs(t *testing.T) {
	driver := &nvmlDriver{}
	
	err := driver.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer driver.Shutdown()
	
	uuids, err := driver.ListDeviceUUIDs()
	if err != nil {
		t.Fatalf("ListDeviceUUIDs() failed: %v", err)
	}
	
	// Expect 3 GPUs (RTX 3090s)
	if len(uuids) < 1 {
		t.Errorf("Expected at least 1 GPU, got %d", len(uuids))
	}
	
	t.Logf("Found %d GPU(s):", len(uuids))
	for uuid, m := range uuids {
		modeName := "normal"
		if m == parent {
			modeName = "parent"
		} else if m == mig {
			modeName = "mig"
		}
		t.Logf("  UUID: %s (mode: %s)", uuid, modeName)
	}
}

func TestDeviceInfoByUUID(t *testing.T) {
	driver := &nvmlDriver{}
	
	err := driver.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer driver.Shutdown()
	
	uuids, err := driver.ListDeviceUUIDs()
	if err != nil {
		t.Fatalf("ListDeviceUUIDs() failed: %v", err)
	}
	
	if len(uuids) == 0 {
		t.Skip("No GPUs found")
	}
	
	for uuid, m := range uuids {
		if m == parent {
			// Skip parent devices (MIG), test child devices instead
			continue
		}
		
		info, err := driver.DeviceInfoByUUID(uuid)
		if err != nil {
			t.Errorf("DeviceInfoByUUID(%s) failed: %v", uuid, err)
			continue
		}
		
		t.Logf("Device Info for %s:", uuid)
		if info.Name != nil {
			t.Logf("  Name: %s", *info.Name)
			// Check if name contains expected GPU model (RTX 3090)
			if !strings.Contains(strings.ToLower(*info.Name), "3090") && !strings.Contains(strings.ToLower(*info.Name), "rtx") {
				t.Logf("  Warning: Expected RTX 3090, got: %s", *info.Name)
			}
		}
		if info.MemoryMiB != nil {
			t.Logf("  Memory: %d MiB", *info.MemoryMiB)
			// RTX 3090 should have ~24GB (24576 MiB)
			if *info.MemoryMiB < 20000 || *info.MemoryMiB > 30000 {
				t.Logf("  Warning: Expected ~24GB memory, got: %d MiB", *info.MemoryMiB)
			}
		}
		if info.PowerW != nil {
			t.Logf("  Power: %d W", *info.PowerW)
		}
		if info.BAR1MiB != nil {
			t.Logf("  BAR1: %d MiB", *info.BAR1MiB)
		}
		t.Logf("  PCI Bus ID: %s", info.PCIBusID)
		if info.PCIBandwidthMBPerS != nil {
			t.Logf("  PCIe Bandwidth: %d MB/s", *info.PCIBandwidthMBPerS)
		}
		if info.CoresClockMHz != nil {
			t.Logf("  Core Clock: %d MHz", *info.CoresClockMHz)
		}
		if info.MemoryClockMHz != nil {
			t.Logf("  Memory Clock: %d MHz", *info.MemoryClockMHz)
		}
		t.Logf("  Display State: %s", info.DisplayState)
		t.Logf("  Persistence Mode: %s", info.PersistenceMode)
	}
}

func TestDeviceInfoAndStatusByUUID(t *testing.T) {
	driver := &nvmlDriver{}
	
	err := driver.Initialize()
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer driver.Shutdown()
	
	uuids, err := driver.ListDeviceUUIDs()
	if err != nil {
		t.Fatalf("ListDeviceUUIDs() failed: %v", err)
	}
	
	if len(uuids) == 0 {
		t.Skip("No GPUs found")
	}
	
	for uuid, m := range uuids {
		if m == parent {
			continue
		}
		
		info, status, err := driver.DeviceInfoAndStatusByUUID(uuid)
		if err != nil {
			t.Errorf("DeviceInfoAndStatusByUUID(%s) failed: %v", uuid, err)
			continue
		}
		
		t.Logf("Device Status for %s:", uuid)
		if info.Name != nil {
			t.Logf("  Name: %s", *info.Name)
		}
		if status.TemperatureC != nil {
			t.Logf("  Temperature: %d C", *status.TemperatureC)
			// Temperature should be reasonable (20-100C)
			if *status.TemperatureC < 20 || *status.TemperatureC > 100 {
				t.Logf("  Warning: Temperature seems unusual: %d C", *status.TemperatureC)
			}
		}
		if status.GPUUtilization != nil {
			t.Logf("  GPU Utilization: %d%%", *status.GPUUtilization)
		}
		if status.MemoryUtilization != nil {
			t.Logf("  Memory Utilization: %d%%", *status.MemoryUtilization)
		}
		if status.UsedMemoryMiB != nil {
			t.Logf("  Used Memory: %d MiB", *status.UsedMemoryMiB)
		}
		if status.PowerUsageW != nil {
			t.Logf("  Power Usage: %d mW", *status.PowerUsageW)
		}
		if status.EncoderUtilization != nil {
			t.Logf("  Encoder Utilization: %d%%", *status.EncoderUtilization)
		}
		if status.DecoderUtilization != nil {
			t.Logf("  Decoder Utilization: %d%%", *status.DecoderUtilization)
		}
		if status.BAR1UsedMiB != nil {
			t.Logf("  BAR1 Used: %d MiB", *status.BAR1UsedMiB)
		}
		// ECC errors (will be 0 or nil for consumer GPUs)
		t.Logf("  ECC Errors: L1=%v L2=%v Device=%v RegisterFile=%v",
			status.ECCErrorsL1Cache, status.ECCErrorsL2Cache,
			status.ECCErrorsDevice, status.ECCErrorsRegisterFile)
	}
}

func TestLowLevelNVMLFunctions(t *testing.T) {
	// Test the low-level NVML wrapper functions directly
	
	ret := nvmlInit()
	if ret != NVML_SUCCESS {
		t.Fatalf("nvmlInit() failed: %s", errorString(ret))
	}
	defer nvmlShutdown()
	
	// Test driver version
	version, ret := nvmlSystemGetDriverVersion()
	if ret != NVML_SUCCESS {
		t.Fatalf("nvmlSystemGetDriverVersion() failed: %s", errorString(ret))
	}
	t.Logf("Low-level driver version: %s", version)
	
	// Test device count
	count, ret := nvmlDeviceGetCount()
	if ret != NVML_SUCCESS {
		t.Fatalf("nvmlDeviceGetCount() failed: %s", errorString(ret))
	}
	t.Logf("Device count: %d", count)
	
	// Test each device
	for i := 0; i < int(count); i++ {
		device, ret := nvmlDeviceGetHandleByIndex(i)
		if ret != NVML_SUCCESS {
			t.Errorf("nvmlDeviceGetHandleByIndex(%d) failed: %s", i, errorString(ret))
			continue
		}
		
		name, ret := nvmlDeviceGetName(device)
		if ret != NVML_SUCCESS {
			t.Errorf("nvmlDeviceGetName() failed: %s", errorString(ret))
		} else {
			t.Logf("Device %d name: %s", i, name)
		}
		
		uuid, ret := nvmlDeviceGetUUID(device)
		if ret != NVML_SUCCESS {
			t.Errorf("nvmlDeviceGetUUID() failed: %s", errorString(ret))
		} else {
			t.Logf("Device %d UUID: %s", i, uuid)
		}
		
		memory, ret := nvmlDeviceGetMemoryInfo(device)
		if ret != NVML_SUCCESS {
			t.Errorf("nvmlDeviceGetMemoryInfo() failed: %s", errorString(ret))
		} else {
			t.Logf("Device %d memory: Total=%d MiB, Used=%d MiB, Free=%d MiB",
				i, memory.Total/(1<<20), memory.Used/(1<<20), memory.Free/(1<<20))
		}
		
		temp, ret := nvmlDeviceGetTemperature(device, NVML_TEMPERATURE_GPU)
		if ret != NVML_SUCCESS {
			t.Errorf("nvmlDeviceGetTemperature() failed: %s", errorString(ret))
		} else {
			t.Logf("Device %d temperature: %d C", i, temp)
		}
	}
}
