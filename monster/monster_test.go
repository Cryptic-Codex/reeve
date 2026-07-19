// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package monster

import (
	"math/rand"
	"strings"
	"testing"
)

func seeded(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

func TestParseHD(t *testing.T) {
	tests := []struct {
		in   string
		want HD
	}{
		{"1", HD{Dice: 1}},
		{"3", HD{Dice: 3}},
		{"1+1", HD{Dice: 1, Modifier: 1}},
		{"4+1", HD{Dice: 4, Modifier: 1}},
		{"2-1", HD{Dice: 2, Modifier: -1}},
		{"1/2", HD{Dice: 1, Half: true}},
	}
	for _, tt := range tests {
		got, err := ParseHD(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("ParseHD(%q) = %+v, %v; want %+v", tt.in, got, err, tt.want)
		}
		if got.String() != tt.in {
			t.Errorf("HD.String() = %q, want %q", got.String(), tt.in)
		}
	}
	for _, bad := range []string{"", "d6", "0", "x+1"} {
		if _, err := ParseHD(bad); err == nil {
			t.Errorf("ParseHD(%q) should error", bad)
		}
	}
}

func TestHDRollRanges(t *testing.T) {
	rng := seeded(1)
	tests := []struct {
		hd       string
		min, max int
	}{
		{"1", 1, 6},
		{"4+1", 5, 25},
		{"2-1", 1, 11},
		{"1/2", 1, 3},
	}
	for _, tt := range tests {
		hd, _ := ParseHD(tt.hd)
		for i := 0; i < 1000; i++ {
			hp := hd.Roll(rng)
			if hp < tt.min || hp > tt.max {
				t.Fatalf("HD %s rolled %d, want [%d,%d]", tt.hd, hp, tt.min, tt.max)
			}
		}
	}
}

func TestLoadAndLookup(t *testing.T) {
	const src = `
monster: Goblin
hd: 1-1
ac: 6
move: 6"
number: 2d4
lair: 40
treasure: C
alignment: Chaos
notes: -1 to hit in full daylight.
`
	b := Bestiary{}
	if err := Load(b, strings.NewReader(src)); err != nil {
		t.Fatalf("Load: %v", err)
	}
	m, ok := b.Lookup("goblin")
	if !ok {
		t.Fatal("goblin not found")
	}
	if m.ArmorClass != 6 || m.Attacks != 1 || m.Damage != "1d6" || m.Alignment != "Chaos" {
		t.Errorf("unexpected fields: %+v", m)
	}
	if m.HitDice.String() != "1-1" {
		t.Errorf("HitDice = %s, want 1-1", m.HitDice)
	}
}

func TestLoadErrors(t *testing.T) {
	bad := []string{
		"hd: 1\n",                        // field before header
		"monster: X\nbogus: value\n",     // unknown field
		"monster: X\nhd: not-dice\n",     // bad hd
		"monster: X\nac: high\n",         // bad int
		"just some text with no colon\n", // malformed
	}
	for _, src := range bad {
		if err := Load(Bestiary{}, strings.NewReader(src)); err == nil {
			t.Errorf("Load(%q) should error", src)
		}
	}
}

func TestRollNumber(t *testing.T) {
	b := Bestiary{}
	_ = Load(b, strings.NewReader("monster: Goblin\nhd: 1\nnumber: 2d4\n"))
	m, _ := b.Lookup("goblin")
	rng := seeded(2)
	for i := 0; i < 500; i++ {
		n, err := m.RollNumber(rng)
		if err != nil || n < 2 || n > 8 {
			t.Fatalf("RollNumber = %d, %v; want [2,8]", n, err)
		}
	}
}

func TestBuiltin(t *testing.T) {
	b, err := Builtin()
	if err != nil {
		t.Fatalf("Builtin: %v", err)
	}
	if _, ok := b.Lookup("ogre"); !ok {
		t.Error("expected ogre in built-in bestiary")
	}
	rng := seeded(3)
	for _, name := range b.Names() {
		m, _ := b.Lookup(name)
		if hp := m.RollHP(rng); hp < 1 {
			t.Errorf("%s rolled hp %d", name, hp)
		}
	}
}
