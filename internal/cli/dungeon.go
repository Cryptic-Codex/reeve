// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Cryptic-Codex/reeve/dungeon"
)

var dungeonGen = dungeon.New(bestiary, monsterRand)

// doEncounter handles `encounter [level]`, rolling a wandering encounter.
func doEncounter(out io.Writer, args []string) {
	level := 1
	if len(args) > 0 {
		l, err := strconv.Atoi(args[0])
		if err != nil || l < 1 {
			fmt.Fprintln(out, "usage: encounter [level], e.g. encounter 2")
			return
		}
		level = l
	}

	enc, err := dungeonGen.RollEncounter(level)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	fmt.Fprintf(out, "Wandering monster (dungeon level %d):\n", level)
	printEncounter(out, enc)
}

// doDungeon handles `dungeon [count]`, stocking rooms for a level. An optional
// second argument sets the dungeon level.
func doDungeon(out io.Writer, args []string) {
	count, level := 6, 1
	if len(args) > 0 {
		c, err := strconv.Atoi(args[0])
		if err != nil || c < 1 || c > 100 {
			fmt.Fprintln(out, "usage: dungeon [rooms] [level], rooms 1 to 100")
			return
		}
		count = c
	}
	if len(args) > 1 {
		l, err := strconv.Atoi(args[1])
		if err != nil || l < 1 {
			fmt.Fprintln(out, "level must be a number of 1 or more")
			return
		}
		level = l
	}

	rooms, err := dungeonGen.Stock(level, count)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	fmt.Fprintf(out, "Dungeon level %d — %d rooms:\n", level, count)
	for _, r := range rooms {
		printRoom(out, r)
	}
}

func printEncounter(out io.Writer, enc dungeon.Encounter) {
	m := enc.Monster
	fmt.Fprintf(out, "  %s  (HD %s, AC %d)\n", plural(enc.Number, m.Name), m.HitDice, m.ArmorClass)
	fmt.Fprintf(out, "  Reaction: %s\n", enc.Reaction)
	fmt.Fprintf(out, "  Hit Points: %v\n", enc.HitPoints)
}

func printRoom(out io.Writer, r dungeon.Room) {
	fmt.Fprintf(out, "%3d. ", r.Number)
	switch {
	case r.Monster != nil:
		m := r.Monster.Monster
		fmt.Fprintf(out, "%s (HD %s)", plural(r.Monster.Number, m.Name), m.HitDice)
		if r.Treasure {
			t := m.TreasureType
			if t == "" || t == "-" {
				fmt.Fprint(out, " — treasure")
			} else {
				fmt.Fprintf(out, " — treasure (type %s)", t)
			}
		}
	case r.Treasure:
		if r.Trapped {
			fmt.Fprint(out, "Empty — trapped treasure")
		} else {
			fmt.Fprint(out, "Empty — treasure")
		}
	default:
		fmt.Fprint(out, "Empty")
	}
	fmt.Fprintln(out)
}
