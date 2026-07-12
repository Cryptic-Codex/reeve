// Copyright (C) 2026 COPYRIGHT_HOLDER
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"os"
	"strings"
)

const usage = `reeve — referee tools for OSR games

usage:
  reeve                      start the interactive menu
  reeve roll <dice>          roll dice notation, e.g. reeve roll 2d6+1
  reeve hex <terrain> [n]    generate wilderness hexes from current terrain
  reeve help                 show this help`

// Execute dispatches on os.Args. This is the seam cobra will later replace:
// swap this file for a cobra root command and nothing else changes.
func Execute() {
	if len(os.Args) < 2 {
		if err := RunMenu(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	switch strings.ToLower(os.Args[1]) {
	case "roll":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: reeve roll <dice>, e.g. reeve roll 2d6+1")
			os.Exit(2)
		}
		// join so both `reeve roll 2d6+1` and `reeve roll 2d6 + 1` work
		doRoll(os.Stdout, strings.Join(os.Args[2:], " "))

	case "hex":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: reeve hex <terrain> [count], e.g. reeve hex plain 6")
			os.Exit(2)
		}
		doHex(os.Stdout, os.Args[2:])

	case "menu":
		if err := RunMenu(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "help", "-h", "--help":
		fmt.Println(usage)

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s\n", os.Args[1], usage)
		os.Exit(2)
	}
}
