// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package character

import "testing"

func TestXPModifier(t *testing.T) {
	tests := []struct {
		score int
		want  int
	}{
		{3, -20}, {6, -20},
		{7, -10}, {8, -10},
		{9, 0}, {12, 0},
		{13, 5}, {14, 5},
		{15, 10}, {18, 10},
	}
	for _, tt := range tests {
		if got := XPModifier(tt.score); got != tt.want {
			t.Errorf("XPModifier(%d) = %d, want %d", tt.score, got, tt.want)
		}
	}
}

func TestAvailable(t *testing.T) {
	tests := []struct {
		race  Race
		class Class
		want  bool
	}{
		{Human, FightingMan, true}, {Human, MagicUser, true}, {Human, Cleric, true},
		{Dwarf, FightingMan, true}, {Dwarf, MagicUser, false}, {Dwarf, Cleric, false},
		{Elf, FightingMan, true}, {Elf, MagicUser, true}, {Elf, Cleric, false},
		{Hobbit, FightingMan, true}, {Hobbit, MagicUser, false}, {Hobbit, Cleric, false},
	}
	for _, tt := range tests {
		if got := Available(tt.race, tt.class); got != tt.want {
			t.Errorf("Available(%s, %s) = %v, want %v", tt.race, tt.class, got, tt.want)
		}
	}
}

func TestRecommendClass(t *testing.T) {
	// scores indexed as Str, Int, Wis, Con, Dex, Cha.
	tests := []struct {
		name   string
		scores Scores
		race   Race
		want   Class
	}{
		{"high wisdom human is a cleric", Scores{10, 11, 16, 9, 12, 8}, Human, Cleric},
		{"high strength human is a fighter", Scores{16, 11, 10, 9, 12, 8}, Human, FightingMan},
		{"elf ignores wisdom", Scores{13, 10, 18, 11, 12, 8}, Elf, FightingMan},
		{"elf picks magic-user on intelligence", Scores{9, 15, 12, 11, 12, 8}, Elf, MagicUser},
		{"dwarf is always a fighter", Scores{6, 18, 18, 9, 12, 8}, Dwarf, FightingMan},
		{"ties favor fighting-man", Scores{13, 13, 13, 9, 12, 8}, Human, FightingMan},
	}
	for _, tt := range tests {
		if got := RecommendClass(tt.scores, tt.race); got != tt.want {
			t.Errorf("%s: RecommendClass = %s, want %s", tt.name, got, tt.want)
		}
	}
}

func TestNewRejectsUnavailable(t *testing.T) {
	if _, err := New(Dwarf, MagicUser, Scores{}); err == nil {
		t.Error("New(Dwarf, MagicUser) should error")
	}
	if _, err := New(Human, Cleric, Scores{}); err != nil {
		t.Errorf("New(Human, Cleric) should succeed, got %v", err)
	}
}

func TestParseRace(t *testing.T) {
	tests := map[string]Race{
		"human": Human, "Men": Human,
		"dwarf": Dwarf, "Dwarves": Dwarf,
		"elf": Elf, "elven": Elf,
		"hobbit": Hobbit, "halfling": Hobbit, "  Halflings ": Hobbit,
	}
	for in, want := range tests {
		if got, err := ParseRace(in); err != nil || got != want {
			t.Errorf("ParseRace(%q) = %s, %v; want %s", in, got, err, want)
		}
	}
	if _, err := ParseRace("orc"); err == nil {
		t.Error("ParseRace(orc) should error")
	}
}

func TestParseClass(t *testing.T) {
	tests := map[string]Class{
		"fighter": FightingMan, "fighting-man": FightingMan, "F": FightingMan,
		"magic-user": MagicUser, "mu": MagicUser, "wizard": MagicUser,
		"cleric": Cleric, "c": Cleric,
	}
	for in, want := range tests {
		if got, err := ParseClass(in); err != nil || got != want {
			t.Errorf("ParseClass(%q) = %s, %v; want %s", in, got, err, want)
		}
	}
	if _, err := ParseClass("paladin"); err == nil {
		t.Error("ParseClass(paladin) should error")
	}
}

func TestAbilityEffects(t *testing.T) {
	if got := ConHPAdjustment(15); got != 1 {
		t.Errorf("ConHPAdjustment(15) = %d, want 1", got)
	}
	if got := ConHPAdjustment(6); got != -1 {
		t.Errorf("ConHPAdjustment(6) = %d, want -1", got)
	}
	if got := ConHPAdjustment(10); got != 0 {
		t.Errorf("ConHPAdjustment(10) = %d, want 0", got)
	}
	if got := DexMissileAdjustment(13); got != 1 {
		t.Errorf("DexMissileAdjustment(13) = %d, want 1", got)
	}
	if got := DexMissileAdjustment(8); got != -1 {
		t.Errorf("DexMissileAdjustment(8) = %d, want -1", got)
	}
	if got := MaxHirelings(3); got != 1 {
		t.Errorf("MaxHirelings(3) = %d, want 1", got)
	}
	if got := MaxHirelings(18); got != 12 {
		t.Errorf("MaxHirelings(18) = %d, want 12", got)
	}
	if got := LoyaltyBase(18); got != 4 {
		t.Errorf("LoyaltyBase(18) = %d, want 4", got)
	}
	if got := LoyaltyBase(10); got != 0 {
		t.Errorf("LoyaltyBase(10) = %d, want 0", got)
	}
	if got := AdditionalLanguages(14); got != 4 {
		t.Errorf("AdditionalLanguages(14) = %d, want 4", got)
	}
	if got := AdditionalLanguages(9); got != 0 {
		t.Errorf("AdditionalLanguages(9) = %d, want 0", got)
	}
}

func TestParseAlignment(t *testing.T) {
	tests := map[string]Alignment{
		"law": Law, "Lawful": Law,
		"neutrality": Neutrality, "neutral": Neutrality,
		"chaos": Chaos, "chaotic": Chaos,
	}
	for in, want := range tests {
		if got, err := ParseAlignment(in); err != nil || got != want {
			t.Errorf("ParseAlignment(%q) = %s, %v; want %s", in, got, err, want)
		}
	}
	if _, err := ParseAlignment("lawful-good"); err == nil {
		t.Error("ParseAlignment(lawful-good) should error")
	}
}

func TestRollExtras(t *testing.T) {
	g := NewSeeded(7)
	for i := 0; i < 1000; i++ {
		if hp := g.RollHP(3); hp < 1 {
			t.Fatalf("RollHP never below 1, got %d", hp)
		}
		gold := g.RollGold()
		if gold < 30 || gold > 180 || gold%10 != 0 {
			t.Fatalf("RollGold out of range or not a multiple of 10: %d", gold)
		}
		if a := g.RollAlignment(); a < Law || a >= numAlignments {
			t.Fatalf("RollAlignment out of range: %d", a)
		}
	}
}

// Rolling is random, so we test properties, not exact values.
func TestRollScores(t *testing.T) {
	g := NewSeeded(1)
	for i := 0; i < 1000; i++ {
		s := g.RollScores()
		for a, score := range s {
			if score < 3 || score > 18 {
				t.Fatalf("%s score out of range [3,18]: %d", Ability(a), score)
			}
		}
	}
}

func TestRollRecommendsAvailableClass(t *testing.T) {
	g := NewSeeded(42)
	for _, r := range []Race{Human, Dwarf, Elf, Hobbit} {
		for i := 0; i < 100; i++ {
			c := g.Roll(r)
			if !Available(c.Race, c.Class) {
				t.Fatalf("rolled %s, an unavailable combination", c)
			}
		}
	}
}
