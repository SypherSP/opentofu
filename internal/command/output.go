// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"fmt"
	"strings"

	"github.com/opentofu/opentofu/internal/command/arguments"
	"github.com/opentofu/opentofu/internal/command/views"
	"github.com/opentofu/opentofu/internal/encryption"
	"github.com/opentofu/opentofu/internal/states"
	"github.com/opentofu/opentofu/internal/tfdiags"
)

// OutputCommand is a Command implementation that reads an output
// from a OpenTofu state and prints it.
type OutputCommand struct {
	Meta
}

func (c *OutputCommand) Run(rawArgs []string) int {
	// Parse and apply global view arguments
	common, rawArgs := arguments.ParseView(rawArgs)
	c.View.Configure(common)

	// Parse and validate flags
	args, diags := arguments.ParseOutput(rawArgs)
	if diags.HasErrors() {
		c.View.Diagnostics(diags)
		c.View.HelpPrompt("output")
		return 1
	}

	view := views.NewOutput(args.ViewType, c.View)

	// Load the encryption configuration
	enc, encDiags := c.Encryption()
	diags = diags.Append(encDiags)
	if encDiags.HasErrors() {
		c.View.Diagnostics(diags)
		return 1
	}

	// Fetch data from state
	outputs, diags := c.Outputs(args.StatePath, enc)
	if diags.HasErrors() {
		view.Diagnostics(diags)
		return 1
	}

	// Render the view
	viewDiags := view.Output(args.Name, outputs)
	diags = diags.Append(viewDiags)

	view.Diagnostics(diags)

	if diags.HasErrors() {
		return 1
	}

	return 0
}

func (c *OutputCommand) Outputs(statePath string, enc encryption.Encryption) (map[string]*states.OutputValue, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	// Allow state path override
	if statePath != "" {
		c.Meta.statePath = statePath
	}

	// Load the backend
	b, backendDiags := c.Backend(nil, enc.Backend())
	diags = diags.Append(backendDiags)
	if diags.HasErrors() {
		return nil, diags
	}

	// This is a read-only command
	c.ignoreRemoteVersionConflict(b)

	env, err := c.Workspace()
	if err != nil {
		diags = diags.Append(fmt.Errorf("Error selecting workspace: %w", err))
		return nil, diags
	}

	// Get the state
	stateStore, err := b.StateMgr(env)
	if err != nil {
		diags = diags.Append(fmt.Errorf("Failed to load state: %w", err))
		return nil, diags
	}

	output, err := stateStore.GetRootOutputValues()
	if err != nil {
		return nil, diags.Append(err)
	}

	return output, diags
}

func (c *OutputCommand) Help() string {
	helpText := `
Usage: tofu [global options] output [options] [NAME]

  Reads an output variable from a OpenTofu state file and prints
  the value. With no additional arguments, output will display all
  the outputs for the root module.  If NAME is not specified, all
  outputs are printed.

Options:

  -state=path      Path to the state file to read. Defaults to
                   "terraform.tfstate". Ignored when remote 
                   state is used.

  -no-color        If specified, output won't contain any color.

  -json            If specified, machine readable output will be
                   printed in JSON format.

  -raw             For value types that can be automatically
                   converted to a string, will print the raw
                   string directly, rather than a human-oriented
                   representation of the value.
`
	return strings.TrimSpace(helpText)
}

func (c *OutputCommand) Synopsis() string {
	return "Show output values from your root module"
}
