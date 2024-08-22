# Nomad Nvidia Device Plugin

This repository provides an implementation of a Nomad
[`device`](https://www.nomadproject.io/docs/job-specification/device) plugin
for Nvidia GPUs.

## Behavior

The Nvidia device plugin uses
[NVML](https://github.com/NVIDIA/gpu-monitoring-tools) bindings to get data
regarding available Nvidia devices and will expose them via
[`Fingerprint`](https://www.nomadproject.io/docs/internals/plugins/devices#fingerprint-context-context-chan-fingerprintresponse-error)
RPC. GPUs can be excluded from fingerprinting by setting the `ignored_gpu_ids`
field (see below). Plugin sends statistics for fingerprinted devices every
`stats_period` period.

The plugin detects whether the GPU has [`Multi-Instance GPU (MIG)`](https://www.nvidia.com/en-us/technologies/multi-instance-gpu/) enabled.
When enabled all instances will be fingerprinted as individual GPUs that can be addressed accordingly.

## Config

The plugin is configured in the Nomad client's
[`plugin`](https://www.nomadproject.io/docs/configuration/plugin) block:

```hcl
plugin "nvidia" {
  config {
    ignored_gpu_ids    = ["uuid1", "uuid2"]
    fingerprint_period = "5s"
  }
}
```

The valid configuration options are:

* `ignored_gpu_ids` (`list(string)`: `[]`): list of GPU UUIDs strings that
  should not be exposed to nomad
* `fingerprint_period` (`string`: `"1m"`): interval to repeat the fingerprint
  process to identify possible changes.
