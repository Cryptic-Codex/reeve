// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Cryptic-Codex/reeve/dice"

	"github.com/peterh/liner"
)

const menuHelp = `reeve — referee tools

  roll <dice>          roll dice notation, e.g. roll 2d6+1
  hex <terrain> [n]    generate wilderness hexes from current terrain
  char [race] [class]  roll a 3LBB OD&D character, e.g. char elf
  table [name] [n]     roll on a referee table (no name lists them)
  monster [name] [n]   show a monster stat block (no name lists them)
  encounter [level]    roll a wandering dungeon encounter
  dungeon [rooms] [lv] stock dungeon rooms with monsters and treasure
  treasure <type>      roll a treasure hoard, e.g. treasure C
  campaign <cmd>       manage campaigns (list, new, use, show)
  help          show this menu
  quit          leave the table (also: q, exit)

Bare dice notation works too: typing "3d6" rolls 3d6.`

// StartMenu runs the interactive menu, using line editing with up/down history
// at a real terminal and plain line reading otherwise (pipes, redirects).
func StartMenu() error {
	if stdinIsTerminal() {
		return runMenuInteractive(os.Stdout)
	}
	return RunMenu(os.Stdin, os.Stdout)
}

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// RunMenu runs the prompt loop with plain line reading, reading commands from
// in and writing results to out. It returns when the user quits or input ends.
func RunMenu(in io.Reader, out io.Writer) error {
	fmt.Fprintln(out, menuHelp)
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprint(out, "\nreeve> ")
		if !scanner.Scan() {
			fmt.Fprintln(out)
			return scanner.Err() // nil on clean EOF (ctrl-D)
		}
		if dispatchMenu(out, scanner.Text()) {
			return nil
		}
	}
}

// runMenuInteractive runs the prompt loop at a terminal, with line editing and
// up/down history via liner.
func runMenuInteractive(out io.Writer) error {
	fmt.Fprintln(out, menuHelp)

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)

	for {
		fmt.Fprintln(out)
		s, err := line.Prompt("reeve> ")
		switch err {
		case nil:
		case io.EOF, liner.ErrPromptAborted: // ctrl-D or ctrl-C
			fmt.Fprintln(out)
			return nil
		default:
			return err
		}

		if strings.TrimSpace(s) != "" {
			line.AppendHistory(s)
		}
		if dispatchMenu(out, s) {
			return nil
		}
	}
}

// dispatchMenu runs one menu command line, writing results to out. It reports
// whether the session should end (the user quit).
func dispatchMenu(out io.Writer, line string) (done bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	cmd, rest, _ := strings.Cut(line, " ")
	switch strings.ToLower(cmd) {
	case "quit", "q", "exit":
		fmt.Fprintln(out, "The session ends.")
		return true

	case "help", "?":
		fmt.Fprintln(out, menuHelp)

	case "roll":
		doRoll(out, rest)

	case "hex":
		if rest == "" {
			fmt.Fprintln(out, "usage: hex <terrain> [count], e.g. hex forest 6")
			return false
		}
		doHex(out, strings.Fields(rest))

	case "char":
		doChar(out, strings.Fields(rest))

	case "table":
		doTable(out, strings.Fields(rest))

	case "monster":
		doMonster(out, strings.Fields(rest))

	case "encounter":
		doEncounter(out, strings.Fields(rest))

	case "dungeon":
		doDungeon(out, strings.Fields(rest))

	case "treasure":
		doTreasure(out, strings.Fields(rest))

	case "campaign":
		doCampaign(out, strings.Fields(rest))

	default:
		// convenience: bare "3d6+1" is treated as a roll
		if _, err := dice.Parse(line); err == nil {
			doRoll(out, line)
		} else {
			fmt.Fprintf(out, "unknown command %q — try help\n", cmd)
		}
	}
	return false
}

func doRoll(out io.Writer, notation string) {
	notation = strings.TrimSpace(notation)
	if notation == "" {
		fmt.Fprintln(out, "usage: roll <dice>, e.g. roll 2d6+1")
		return
	}

	r, err := dice.Parse(notation)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	res := r.Roll()
	if len(res.Rolls) == 1 && r.Modifier == 0 {
		fmt.Fprintf(out, "%d\n", res.Total)
		return
	}
	fmt.Fprintf(out, "%d  %v", res.Total, res.Rolls)
	if r.Modifier != 0 {
		fmt.Fprintf(out, " %+d", r.Modifier)
	}
	fmt.Fprintln(out)
}
