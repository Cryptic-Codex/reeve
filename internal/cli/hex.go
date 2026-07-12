// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Cryptic-Codex/reeve/terrain"
)

var hexGen = terrain.New()

// doHex handles `hex <terrain> [count]` for both the CLI and the menu.
func doHex(out io.Writer, args []string) {
	ty, err := terrain.Parse(args[0])
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	count := 1
	if len(args) > 1 {
		count, err = strconv.Atoi(args[1])
		if err != nil || count < 1 || count > 1000 {
			fmt.Fprintln(out, "count must be a number from 1 to 1000")
			return
		}
	}

	for i, h := range hexGen.Walk(ty, count) {
		if count == 1 {
			fmt.Fprintf(out, "%s  (rolled %d)\n", h, h.Roll)
		} else {
			fmt.Fprintf(out, "%2d. %-30s (rolled %d)\n", i+1, h, h.Roll)
		}
	}
}
