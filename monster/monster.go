// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

// Package monster models 3LBB OD&D monsters (Monsters & Treasure, 1974): a
// stat block, hit-dice rolling, and a bestiary loaded from a simple block
// format so a referee can add their own.
package monster

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Cryptic-Codex/reeve/dice"
)

// HD is a monster's hit dice: Dice six-sided dice plus a flat Modifier, or a
// single Half die for sub-1-HD monsters written "1/2".
type HD struct {
	Dice     int
	Modifier int
	Half     bool
}

// ParseHD parses hit dice like "1", "1+1", "2-1", or "1/2".
func ParseHD(s string) (HD, error) {
	s = strings.TrimSpace(s)
	if s == "1/2" {
		return HD{Dice: 1, Half: true}, nil
	}

	digits, mod := s, 0
	if i := strings.IndexAny(s, "+-"); i > 0 {
		n, err := strconv.Atoi(strings.TrimSpace(s[i:]))
		if err != nil {
			return HD{}, fmt.Errorf("invalid hit-dice modifier: %q", s)
		}
		digits, mod = s[:i], n
	}

	n, err := strconv.Atoi(strings.TrimSpace(digits))
	if err != nil || n < 1 {
		return HD{}, fmt.Errorf("invalid hit dice: %q", s)
	}
	return HD{Dice: n, Modifier: mod}, nil
}

func (h HD) String() string {
	switch {
	case h.Half:
		return "1/2"
	case h.Modifier > 0:
		return fmt.Sprintf("%d+%d", h.Dice, h.Modifier)
	case h.Modifier < 0:
		return fmt.Sprintf("%d%d", h.Dice, h.Modifier)
	default:
		return strconv.Itoa(h.Dice)
	}
}

// MarshalText and UnmarshalText render HD as its notation (e.g. "4+1") so it
// stays readable in saved JSON.
func (h HD) MarshalText() ([]byte, error) { return []byte(h.String()), nil }

func (h *HD) UnmarshalText(b []byte) error {
	v, err := ParseHD(string(b))
	if err != nil {
		return err
	}
	*h = v
	return nil
}

// Roll rolls hit points for these hit dice (d6 each in the 3LBB), never less
// than 1.
func (h HD) Roll(rng *rand.Rand) int {
	if h.Half {
		if hp := (rng.Intn(6) + 1) / 2; hp >= 1 {
			return hp
		}
		return 1
	}
	hp := h.Modifier
	for i := 0; i < h.Dice; i++ {
		hp += rng.Intn(6) + 1
	}
	if hp < 1 {
		return 1
	}
	return hp
}

// Monster is a 3LBB stat block. Armor Class is the descending OD&D scale
// (lower is better).
type Monster struct {
	Name         string
	HitDice      HD
	ArmorClass   int
	Move         string
	Attacks      int
	Damage       string
	NumberAppear string
	InLair       int // percent chance of being in a lair
	TreasureType string
	Alignment    string
	Notes        string
}

// RollHP rolls hit points for a single monster.
func (m *Monster) RollHP(rng *rand.Rand) int { return m.HitDice.Roll(rng) }

// RollNumber rolls the monster's number appearing.
func (m *Monster) RollNumber(rng *rand.Rand) (int, error) {
	if m.NumberAppear == "" {
		return 0, fmt.Errorf("%s has no number appearing", m.Name)
	}
	r, err := dice.Parse(m.NumberAppear)
	if err != nil {
		return 0, err
	}
	total := r.Modifier
	for i := 0; i < r.Count; i++ {
		total += rng.Intn(r.Sides) + 1
	}
	return total, nil
}
