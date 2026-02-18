// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build linux

package nvml

import (
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/shoenig/test/must"
)

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
