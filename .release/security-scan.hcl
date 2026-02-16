# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: MPL-2.0

binary {
  go_modules = true
  osv        = true
  go_stdlib  = true
  oss_index  = false
  nvd        = false

  secrets {
    all = true
  }

  # Triage items that are _safe_ to ignore here. Note that this list should be
  # periodically cleaned up to remove items that are no longer found by the scanner.
  triage {
    suppress {
      vulnerabilities = [
        "GO-2025-3510", // TODO(dduzgun-security): false positive scan result, investigate why and fix
        "GO-2024-3262", // TODO(dduzgun-security): false positive scan result, investigate why and fix
      ]
    }
  }
}
