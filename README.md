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
field (see below). Plugin sends statistics for fingerprinted devices periodically.


The plugin detects whether the GPU has [`Multi-Instance GPU (MIG)`](https://www.nvidia.com/en-us/technologies/multi-instance-gpu/) enabled.
When enabled all instances will be fingerprinted as individual GPUs that can be addressed accordingly.

## MPS Support for Nvidia GPUs
The plugin can be configured to notify Nomad of GPUs with active MPS servers running against them. The plugin will either the `global_mps_pipe_directory` or the appropriate `mps_pipe_directory` variable from the `device_specific_mps_config` block to check for an active `control` file  in the pipe directory and report if the device is intended and available for sharing in its Fingerprint.

Please be aware of Nvidia's published [Considerations](https://docs.nvidia.com/deploy/mps/when-to-use-mps.html#considerations) and consult their documentation and forums for MPS related issues.

At this time, specific MPS related limitations of this plugin include
- This plugin does not currently support the `--multiuser` flag. All MPS servers and CUDA applications
must belong to the same user
- MPS is only supported on linux runtimes in Docker or Podman containers
- MPS is not currently supported on MIG partitioned GPUs or their parents

## Config
The plugin is configured in the Nomad client's
[`plugin`](https://www.nomadproject.io/docs/configuration/plugin) block:

```hcl
plugin "nvidia" {
  config {
    ignored_gpu_ids    = ["uuid1", "uuid2"]
    fingerprint_period = "5s"
    mps_enabled = true
    mps_user = "non-root-user-who-owns-mps-server-and-container-tasks"
    global_mps_pipe_directory = "/tmp/nvidia-mps"
    global_mps_log_directory = "var/log/nvidia-mps"
  }
}
```

```hcl
plugin "nvidia" {
  config {
    ignored_gpu_ids    = ["uuid1", "uuid2"]
    fingerprint_period = "5s"
    mps_enabled = true
    device_specific_mps_config [
      {
        uuid = "GPU-1234"
        mps_pipe_directory = "/tmp/nvidia-mps1"
        mps_log_directory = "/var/log/nvidia-mps1"
      },
      {
        uuid = "GPU-5678"
        mps_pipe_directory = "/tmp/nvidia-mps2"
        mps_log_directory = "/var/log/nvidia-mps2"
      },
    ]
}
```
The valid configuration options are:

* `ignored_gpu_ids` (`list(string)`: `[]`): list of GPU UUIDs strings that
  should not be exposed to nomad
* `fingerprint_period` (`string`: `"1m"`): interval to repeat the fingerprint
  process to identify possible changes.

