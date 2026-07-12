// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package cli

import (
	"fmt"
	"reeve/dice"

	"github.com/spf13/cobra"
)

var rollCmd = &cobra.Command{
	Use:   "roll [dice]",
	Short: "Roll dice (e.g. 2d6, 3d6+1, d20)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := dice.Parse(args[0])
		if err != nil {
			return err
		}
		res := r.Roll()
		fmt.Println(res.Total)
		return nil
	},
}
