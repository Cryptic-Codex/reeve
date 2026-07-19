// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package table

import (
	_ "embed"
	"strings"
)

//go:embed data/tables.txt
var builtinData string

// Builtin returns a registry of the tables bundled with reeve.
func Builtin() (Registry, error) {
	reg := Registry{}
	if err := Load(reg, strings.NewReader(builtinData)); err != nil {
		return nil, err
	}
	if err := reg.Validate(); err != nil {
		return nil, err
	}
	return reg, nil
}
