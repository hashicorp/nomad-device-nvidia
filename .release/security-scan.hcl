# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

binary {
  secrets    = true
  go_modules = true
  #TODO: enable OSV scan once dependencies are updated.
  osv       = true
  oss_index = false
  nvd       = false
}
