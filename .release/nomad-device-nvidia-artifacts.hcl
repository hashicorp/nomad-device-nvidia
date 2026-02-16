# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: MPL-2.0

schema = 1
artifacts {
  zip = [
    "nomad-device-nvidia_${version}_linux_arm64.zip",
    "nomad-device-nvidia_${version}_linux_amd64.zip",
  ]
  rpm = [
    "nomad-device-nvidia-${version_linux}-1.aarch64.rpm",
    "nomad-device-nvidia-${version_linux}-1.x86_64.rpm",
  ]
  deb = [
    "nomad-device-nvidia_${version_linux}-1_arm64.deb",
    "nomad-device-nvidia_${version_linux}-1_amd64.deb",
  ]
}
