// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Cryptic-Codex/reeve/dice"
)

const menuHelp = `reeve — referee tools

  roll <dice>   roll dice notation, e.g. roll 2d6+1
  help          show this menu
  quit          leave the table (also: q, exit)

Bare dice notation works too: typing "3d6" rolls 3d6.`

// RunMenu runs the interactive prompt loop, reading commands from in and
// writing results to out. It returns when the user quits or input ends.
func RunMenu(in io.Reader, out io.Writer) error {
	fmt.Fprintln(out, menuHelp)
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprint(out, "\nreeve> ")
		if !scanner.Scan() {
			fmt.Fprintln(out)
			return scanner.Err() // nil on clean EOF (ctrl-D)
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		cmd, rest, _ := strings.Cut(line, " ")
		cmd = strings.ToLower(cmd)

		switch cmd {
		case "quit", "q", "exit":
			fmt.Fprintln(out, "The session ends.")
			return nil

		case "help", "?":
			fmt.Fprintln(out, menuHelp)

		case "roll":
			doRoll(out, rest)

		default:
			// convenience: bare "3d6+1" is treated as a roll
			if _, err := dice.Parse(line); err == nil {
				doRoll(out, line)
			} else {
				fmt.Fprintf(out, "unknown command %q — try help\n", cmd)
			}
		}
	}
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
