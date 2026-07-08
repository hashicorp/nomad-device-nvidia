# Windows NVML Shim - Refactor Evidence

**Commit:** `d34a8ca refactor: use interface pattern for Windows NVML shim`
**Date:** 2026-03-23
**Branch:** `feat/windows-nvml-syscall` (rebased on PR #96)

## Reviewer Feedback Addressed

### 1. LazyDLL → windows.LoadLibrary ✅

**Before:**
```go
var nvmlDLL = syscall.NewLazyDLL("nvml.dll")
var procInit = nvmlDLL.NewProc("nvmlInit_v2")
```

**After:**
```go
var (
    nvmlLib     windows.Handle
    procInit    uintptr
)

func initNVMLLibrary() error {
    nvmlLib, err = windows.LoadLibrary("nvml.dll")
    if err != nil {
        return err
    }
    procInit, err = windows.GetProcAddress(nvmlLib, "nvmlInit_v2")
    // ...
}
```

**Evidence:** 45 occurrences of LoadLibrary/GetProcAddress/docs.nvidia.com in nvml_windows.go

### 2. NVML API Documentation Links ✅

Every struct, const, and function now has a doc link:

```go
// NVML Return codes
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlReturnValues.html
type nvmlReturn uint32

// nvmlMemory_t structure
// See: https://docs.nvidia.com/deploy/nvml-api/structnvmlMemory__t.html
type nvmlMemory struct { ... }

// nvmlDeviceGetMemoryInfo
// See: https://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html#group__nvmlDeviceQueries_1g4aee...
func nvmlDeviceGetMemoryInfo(device nvmlDevice) (nvmlMemory, nvmlReturn)
```

**Count:** 40+ doc links added

### 3. Interface Pattern from PR #96 ✅

**driver_windows.go now uses shared helpers:**
```go
func (n *nvmlDriver) ListDeviceUUIDs() (map[string]mode, error) {
    // ...
    for i := 0; i < int(count); i++ {
        handle, _ := nvmlDeviceGetHandleByIndex(i)
        device := newWinDevice(handle)  // Creates nvml.Device interface
        
        devIDs, err := uuidsFromDevice(device)  // Reuses Linux helper
        if err != nil {
            return nil, err
        }
        maps.Copy(uuids, devIDs)
    }
    return uuids, nil
}
```

**device_windows.go builds nvml.Device via mock:**
```go
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
    
    // ... 18 more methods
    
    return d
}
```

## Code Structure

| File | Lines Changed | Purpose |
|------|---------------|---------|
| `nvml/device_windows.go` | +156 | Windows shim implementing `nvml.Device` |
| `nvml/driver_windows.go` | -291 (refactored) | Uses shared helpers |
| `nvml/nvml_windows.go` | +262/-174 | LoadLibrary + docs |

## Methods Implemented

The shim implements 21 methods required by the shared helpers:

- `GetUUID`, `GetName`, `GetMemoryInfo`, `GetBAR1MemoryInfo`, `GetPciInfo`
- `GetMaxPcieLinkWidth`, `GetMaxPcieLinkGeneration`, `GetClockInfo`
- `GetDisplayMode`, `GetPersistenceMode`, `GetMigMode`
- `GetMaxMigDeviceCount`, `GetMigDeviceHandleByIndex`
- `GetUtilizationRates`, `GetEncoderUtilization`, `GetDecoderUtilization`
- `GetTemperature`, `GetPowerUsage`, `GetDetailedEccErrors`
- `GetDeviceHandleFromMigDeviceHandle`, `IsMigDeviceHandle`

## Testing Status

- **Brian (RTX 3090 cluster):** Offline since 2026-03-21 (2 days)
- **Local:** No Go toolchain (installation blocked by pending MSI)
- **CI:** No Windows runners in upstream repo

**Next steps:**
1. Wait for Brian to come online
2. Run `go build ./nvml` on Windows
3. Run `make test` to verify mock tests pass
4. Update this document with test output

## Commit History

```
d34a8ca refactor: use interface pattern for Windows NVML shim
bd23e9c feat: add Windows support via syscall NVML wrapper
16fb7fa nvml: refactor driver to inject Device interface and mock tests (PR #96)
```

## References

- PR #93: https://github.com/hashicorp/nomad-device-nvidia/pull/93
- PR #96 (interface pattern): https://github.com/hashicorp/nomad-device-nvidia/pull/96
- Reviewer: @tgross, @mismithhisler
- NVML API docs: https://docs.nvidia.com/deploy/nvml-api/
