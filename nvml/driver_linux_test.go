// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build linux

package nvml

import (
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/NVIDIA/go-nvml/pkg/nvml/mock"
	"github.com/shoenig/test/must"
)

func TestUuidsFromDevice(t *testing.T) {

	t.Run("mig device returns parent and child uuids", func(t *testing.T) {
		md := &mock.Device{} // mig child device
		pd := &mock.Device{} // parent device
		pd.GetMigModeFunc = func() (int, int, nvml.Return) { return nvml.DEVICE_MIG_ENABLE, 0, nvml.SUCCESS }
		pd.GetUUIDFunc = func() (string, nvml.Return) { return "123", nvml.SUCCESS }
		pd.GetMaxMigDeviceCountFunc = func() (int, nvml.Return) { return 1, nvml.SUCCESS }
		pd.GetMigDeviceHandleByIndexFunc = func(n int) (nvml.Device, nvml.Return) { return md, nvml.SUCCESS }

		md.GetUUIDFunc = func() (string, nvml.Return) { return "456", nvml.SUCCESS }

		uuids, err := uuidsFromDevice(pd)
		must.NoError(t, err)

		must.Eq(t, map[string]mode{"123": parent, "456": mig}, uuids)
	})

	t.Run("non-mig device returns only normal uuid", func(t *testing.T) {
		td := &mock.Device{} // test device
		td.GetMigModeFunc = func() (int, int, nvml.Return) { return nvml.DEVICE_MIG_DISABLE, 0, nvml.SUCCESS }
		td.GetUUIDFunc = func() (string, nvml.Return) { return "123", nvml.SUCCESS }

		uuids, err := uuidsFromDevice(td)
		must.NoError(t, err)

		must.Eq(t, map[string]mode{"123": normal}, uuids)
	})
}

func TestDetermineMemoryInfo(t *testing.T) {
	t.Run("uses device memory when supported", func(t *testing.T) {
		totalMiB, usedMiB, usingSystemMemory, err := determineMemoryInfo(nvml.Memory{
			Total: 8 * (1 << 30),
			Used:  3 * (1 << 30),
		}, nvml.SUCCESS)
		must.Eq(t, false, usingSystemMemory)
		must.NoError(t, err)
		must.Eq(t, uint64(8192), totalMiB)
		must.Eq(t, uint64(3072), usedMiB)
	})

	t.Run("falls back to system memory when device memory is not supported", func(t *testing.T) {
		totalMiB, usedMiB, usingSystemMemory, err := determineMemoryInfo(nvml.Memory{}, nvml.ERROR_NOT_SUPPORTED)
		must.Eq(t, true, usingSystemMemory)
		must.NoError(t, err)
		if totalMiB == 0 {
			t.Fatal("expected non-zero system memory total in MiB")
		}
		if usedMiB == 0 {
			t.Fatal("expected non-zero system memory used in MiB")
		}
	})

	t.Run("returns error on unexpected nvml code", func(t *testing.T) {
		_, _, _, err := determineMemoryInfo(nvml.Memory{}, nvml.ERROR_UNKNOWN)
		must.Error(t, err)
	})
}
