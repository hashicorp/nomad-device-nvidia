# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

schema = 1
artifacts {
  zip = [
    "nomad-device-nvidia_${version}_linux_amd64.zip",
  ]
  rpm = [
    "nomad-device-nvidia-${version_linux}-1.x86_64.rpm",
  ]
  deb = [
    "nomad-device-nvidia_${version_linux}-1_amd64.deb",
  ]
}
