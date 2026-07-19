// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Cryptic-Codex/reeve/table"
)

var (
	tableReg  = mustBuiltinTables()
	tableRand = table.NewRand()
)

// mustBuiltinTables loads the compiled-in tables; malformed built-in data is a
// build-time bug, so a failure here is fatal.
func mustBuiltinTables() table.Registry {
	reg, err := table.Builtin()
	if err != nil {
		panic("reeve: built-in tables failed to load: " + err.Error())
	}
	return reg
}

// doTable handles `table [name] [count]` for both the CLI and the menu.
// With no name, or "list", it lists the available tables.
func doTable(out io.Writer, args []string) {
	if len(args) == 0 || args[0] == "list" {
		listTables(out)
		return
	}

	name := args[0]
	count := 1
	if len(args) > 1 {
		c, err := strconv.Atoi(args[1])
		if err != nil || c < 1 || c > 100 {
			fmt.Fprintln(out, "count must be a number from 1 to 100")
			return
		}
		count = c
	}

	for i := 0; i < count; i++ {
		res, err := tableReg.Roll(name, tableRand)
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}
		if count == 1 {
			fmt.Fprintf(out, "%s  (rolled %d)\n", res.Text, res.Roll)
		} else {
			fmt.Fprintf(out, "%2d. %s  (rolled %d)\n", i+1, res.Text, res.Roll)
		}
	}
}

func listTables(out io.Writer) {
	fmt.Fprintln(out, "available tables:")
	for _, n := range tableReg.Names() {
		fmt.Fprintf(out, "  %-12s %s\n", n, tableReg[n].Dice())
	}
}
