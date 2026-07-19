// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/Cryptic-Codex/reeve/treasure"
)

// doTreasure handles `treasure <type>`, rolling a hoard. With no type, or
// "list", it lists the available treasure types.
func doTreasure(out io.Writer, args []string) {
	if len(args) == 0 || args[0] == "list" {
		fmt.Fprintf(out, "treasure types: %s\n", strings.Join(treasure.Types(), ", "))
		return
	}

	h, err := treasure.Roll(args[0], monsterRand)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	printHoard(out, h)
}

func printHoard(out io.Writer, h treasure.Hoard) {
	fmt.Fprintf(out, "Treasure type %s:\n", h.Type)

	empty := true
	coin := func(label string, n int) {
		if n > 0 {
			fmt.Fprintf(out, "  %-8s %d\n", label, n)
			empty = false
		}
	}
	coin("Copper", h.Copper)
	coin("Silver", h.Silver)
	coin("Gold", h.Gold)

	if len(h.Gems) > 0 {
		fmt.Fprintf(out, "  %-8s %d %v gp\n", "Gems", len(h.Gems), h.Gems)
		empty = false
	}
	if len(h.Jewelry) > 0 {
		fmt.Fprintf(out, "  %-8s %d %v gp\n", "Jewelry", len(h.Jewelry), h.Jewelry)
		empty = false
	}
	if h.Magic > 0 {
		fmt.Fprintf(out, "  %-8s %s (roll on the magic tables)\n", "Magic", plural(h.Magic, "item"))
		empty = false
	}

	if empty {
		fmt.Fprintln(out, "  (nothing this time)")
		return
	}
	fmt.Fprintf(out, "  %-8s %d gp\n", "Total", h.TotalGP)
}
