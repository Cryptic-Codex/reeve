// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

// Package dungeon generates dungeon encounters and stocks rooms after the 3LBB
// (The Underworld & Wilderness Adventures). It draws monsters from reeve's
// bestiary, so the same stat blocks power encounters and room contents.
package dungeon

import (
	"fmt"
	"math/rand"

	"github.com/Cryptic-Codex/reeve/monster"
)

// encounterTables maps a dungeon level to the monsters that may be met there,
// drawn from the built-in bestiary. Deeper levels field tougher foes.
var encounterTables = map[int][]string{
	1: {"Kobold", "Goblin", "Giant Rat", "Skeleton", "Orc", "Stirge"},
	2: {"Orc", "Hobgoblin", "Zombie", "Giant Rat", "Wolf", "Gnoll"},
	3: {"Hobgoblin", "Gnoll", "Wolf", "Zombie", "Ogre", "Orc"},
}

// MaxLevel is the deepest level with its own encounter table; deeper levels
// reuse it.
const MaxLevel = 3

func clampLevel(level int) int {
	switch {
	case level < 1:
		return 1
	case level > MaxLevel:
		return MaxLevel
	default:
		return level
	}
}

// Reaction is a monster's initial disposition, rolled on 2d6.
type Reaction int

const (
	Hostile Reaction = iota
	Unfriendly
	Uncertain
	Indifferent
	Friendly
)

func (r Reaction) String() string {
	switch r {
	case Hostile:
		return "Hostile — attacks at once"
	case Unfriendly:
		return "Unfriendly, likely to attack"
	case Uncertain:
		return "Uncertain, wary and watchful"
	case Indifferent:
		return "Indifferent, no immediate hostility"
	default:
		return "Friendly, offers aid"
	}
}

// Encounter is a generated monster encounter.
type Encounter struct {
	Monster   *monster.Monster
	Number    int
	HitPoints []int
	Reaction  Reaction
}

// Room is a stocked dungeon room. Monster is nil when the room holds no
// monsters; Treasure marks any hoard, and Trapped marks unguarded treasure.
type Room struct {
	Number   int
	Monster  *Encounter
	Treasure bool
	Trapped  bool
}

// Generator produces encounters and rooms. It is not safe for concurrent use.
type Generator struct {
	bestiary monster.Bestiary
	rng      *rand.Rand
}

// New returns a Generator drawing from the given bestiary and source.
func New(b monster.Bestiary, rng *rand.Rand) *Generator {
	return &Generator{bestiary: b, rng: rng}
}

// RollReaction rolls a 2d6 reaction.
func (g *Generator) RollReaction() Reaction {
	switch roll := g.rng.Intn(6) + g.rng.Intn(6) + 2; {
	case roll == 2:
		return Hostile
	case roll <= 5:
		return Unfriendly
	case roll <= 8:
		return Uncertain
	case roll <= 11:
		return Indifferent
	default:
		return Friendly
	}
}

// RollEncounter rolls a wandering encounter for a dungeon level.
func (g *Generator) RollEncounter(level int) (Encounter, error) {
	names := encounterTables[clampLevel(level)]
	name := names[g.rng.Intn(len(names))]

	m, ok := g.bestiary.Lookup(name)
	if !ok {
		return Encounter{}, fmt.Errorf("bestiary is missing %q", name)
	}

	n, err := m.RollNumber(g.rng)
	if err != nil || n < 1 {
		n = 1
	}
	hps := make([]int, n)
	for i := range hps {
		hps[i] = m.RollHP(g.rng)
	}

	return Encounter{Monster: m, Number: n, HitPoints: hps, Reaction: g.RollReaction()}, nil
}

// StockRoom stocks a single numbered room for a dungeon level, following the
// 3LBB odds: a monster on 1-2 of a d6, treasure with a monster on 1-3 of a d6,
// and otherwise unguarded (trapped) treasure on a 1 of a d6.
func (g *Generator) StockRoom(level, number int) (Room, error) {
	r := Room{Number: number}
	if g.rng.Intn(6) < 2 {
		enc, err := g.RollEncounter(level)
		if err != nil {
			return Room{}, err
		}
		r.Monster = &enc
		r.Treasure = g.rng.Intn(6) < 3
	} else if g.rng.Intn(6) == 0 {
		r.Treasure = true
		r.Trapped = true
	}
	return r, nil
}

// Stock stocks count rooms for a dungeon level.
func (g *Generator) Stock(level, count int) ([]Room, error) {
	rooms := make([]Room, count)
	for i := range rooms {
		room, err := g.StockRoom(level, i+1)
		if err != nil {
			return nil, err
		}
		rooms[i] = room
	}
	return rooms, nil
}
