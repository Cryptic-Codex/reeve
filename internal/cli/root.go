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
  reeve char [race] [class]  roll a 3LBB OD&D character, e.g. reeve char elf
  reeve table [name] [n]     roll on a referee table (no name lists them)
  reeve monster [name] [n]   show a monster stat block (no name lists them)
  reeve encounter [level]    roll a wandering dungeon encounter
  reeve dungeon [rooms] [lv] stock dungeon rooms with monsters and treasure
  reeve treasure <type>      roll a treasure hoard, e.g. reeve treasure C
  reeve campaign <cmd>       manage campaigns (list, new, use, show)
  reeve help                 show this help

Add --save to a char or monster command to store it in the current campaign.`

// Execute dispatches on os.Args. Each command joins its arguments and hands
// off to a doX handler shared with the interactive menu.
func Execute() {
	if len(os.Args) < 2 {
		if err := StartMenu(); err != nil {
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

	case "char":
		doChar(os.Stdout, os.Args[2:])

	case "table":
		doTable(os.Stdout, os.Args[2:])

	case "monster":
		doMonster(os.Stdout, os.Args[2:])

	case "encounter":
		doEncounter(os.Stdout, os.Args[2:])

	case "dungeon":
		doDungeon(os.Stdout, os.Args[2:])

	case "treasure":
		doTreasure(os.Stdout, os.Args[2:])

	case "campaign":
		doCampaign(os.Stdout, os.Args[2:])

	case "menu":
		if err := StartMenu(); err != nil {
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
