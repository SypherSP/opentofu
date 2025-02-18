// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package static

import (
	"encoding/hex"
	"fmt"

	"github.com/opentofu/opentofu/internal/encryption/keyprovider"
)

// Config contains the configuration for this key provider supplied by the user. This struct must have hcl tags in order
// to function.
type Config struct {
	Key string `hcl:"key"`
}

// Build will create the usable key provider.
func (c Config) Build() (keyprovider.KeyProvider, keyprovider.KeyMeta, error) {
	decodedData, err := hex.DecodeString(c.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hex-decode the provided key (%w)", err)
	}

	return &staticKeyProvider{decodedData}, new(Metadata), nil
}
