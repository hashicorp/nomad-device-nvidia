// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad-device-nvidia"
	"github.com/hashicorp/nomad/plugins"
)

func main() {
	// Serve the plugin
	plugins.ServeCtx(factory)
}

// factory returns a new instance of the Nvidia GPU plugin
func factory(ctx context.Context, log hclog.Logger) interface{} {
	return nvidia.NewNvidiaDevice(ctx, log)
}
