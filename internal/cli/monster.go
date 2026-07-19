// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Cryptic-Codex/reeve/monster"
)

var (
	bestiary    = mustBestiary()
	monsterRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// mustBestiary loads the compiled-in bestiary; malformed built-in data is a
// build-time bug, so a failure here is fatal.
func mustBestiary() monster.Bestiary {
	b, err := monster.Builtin()
	if err != nil {
		panic("reeve: built-in bestiary failed to load: " + err.Error())
	}
	return b
}

// doMonster handles `monster [name] [count]` for both the CLI and the menu.
// With no name, or "list", it lists the bestiary. A trailing number rolls hit
// points for that many of the monster.
func doMonster(out io.Writer, args []string) {
	args, save := popFlag(args, "--save")

	if len(args) == 0 || (len(args) == 1 && args[0] == "list") {
		listMonsters(out)
		return
	}

	count := 1
	if len(args) > 1 {
		if c, err := strconv.Atoi(args[len(args)-1]); err == nil {
			if c < 1 || c > 100 {
				fmt.Fprintln(out, "count must be a number from 1 to 100")
				return
			}
			count = c
			args = args[:len(args)-1]
		}
	}

	name := strings.Join(args, " ")
	m, ok := bestiary.Lookup(name)
	if !ok {
		fmt.Fprintf(out, "unknown monster %q — try `monster list`\n", name)
		return
	}

	hps := make([]int, count)
	for i := range hps {
		hps[i] = m.RollHP(monsterRand)
	}
	printMonster(out, m, hps)
	if save {
		saveMonster(out, *m)
	}
}

func listMonsters(out io.Writer) {
	fmt.Fprintln(out, "bestiary:")
	for _, n := range bestiary.Names() {
		m, _ := bestiary.Lookup(n)
		fmt.Fprintf(out, "  %-12s HD %s\n", n, m.HitDice)
	}
}

func printMonster(out io.Writer, m *monster.Monster, hps []int) {
	fmt.Fprintln(out, m.Name)
	fmt.Fprintf(out, "  %-11s %s\n", "Hit Dice", m.HitDice)
	fmt.Fprintf(out, "  %-11s %d\n", "Armor Class", m.ArmorClass)
	if m.Move != "" {
		fmt.Fprintf(out, "  %-11s %s\n", "Move", m.Move)
	}
	fmt.Fprintf(out, "  %-11s %d × %s\n", "Attacks", m.Attacks, m.Damage)
	if m.NumberAppear != "" {
		lair := ""
		if m.InLair > 0 {
			lair = fmt.Sprintf("  (%d%% in lair)", m.InLair)
		}
		fmt.Fprintf(out, "  %-11s %s%s\n", "No. App.", m.NumberAppear, lair)
	}
	if t := m.TreasureType; t != "" && t != "-" && t != "—" {
		fmt.Fprintf(out, "  %-11s %s\n", "Treasure", t)
	}
	if m.Alignment != "" {
		fmt.Fprintf(out, "  %-11s %s\n", "Alignment", m.Alignment)
	}
	if m.Notes != "" {
		fmt.Fprintf(out, "  %s\n", m.Notes)
	}

	if len(hps) == 1 {
		fmt.Fprintf(out, "  %-11s %d\n", "Hit Points", hps[0])
	} else {
		fmt.Fprintf(out, "  %-11s %v\n", "Hit Points", hps)
	}
}
