// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package dungeon

import (
	"math/rand"
	"testing"

	"github.com/Cryptic-Codex/reeve/monster"
)

func newGen(t *testing.T, seed int64) *Generator {
	t.Helper()
	b, err := monster.Builtin()
	if err != nil {
		t.Fatalf("bestiary: %v", err)
	}
	return New(b, rand.New(rand.NewSource(seed)))
}

func TestEncounterTablesResolve(t *testing.T) {
	b, err := monster.Builtin()
	if err != nil {
		t.Fatalf("bestiary: %v", err)
	}
	for level, names := range encounterTables {
		for _, name := range names {
			if _, ok := b.Lookup(name); !ok {
				t.Errorf("level %d references %q, missing from bestiary", level, name)
			}
		}
	}
}

func TestRollEncounter(t *testing.T) {
	g := newGen(t, 1)
	for _, level := range []int{-1, 1, 2, 3, 99} {
		for i := 0; i < 200; i++ {
			enc, err := g.RollEncounter(level)
			if err != nil {
				t.Fatalf("RollEncounter(%d): %v", level, err)
			}
			if enc.Number < 1 || len(enc.HitPoints) != enc.Number {
				t.Fatalf("bad number/hp: %d / %v", enc.Number, enc.HitPoints)
			}
			for _, hp := range enc.HitPoints {
				if hp < 1 {
					t.Fatalf("hp below 1: %d", hp)
				}
			}
			if enc.Reaction < Hostile || enc.Reaction > Friendly {
				t.Fatalf("reaction out of range: %d", enc.Reaction)
			}
		}
	}
}

func TestStock(t *testing.T) {
	g := newGen(t, 5)
	rooms, err := g.Stock(1, 100)
	if err != nil {
		t.Fatalf("Stock: %v", err)
	}
	if len(rooms) != 100 {
		t.Fatalf("expected 100 rooms, got %d", len(rooms))
	}
	monsters, treasure := 0, 0
	for i, r := range rooms {
		if r.Number != i+1 {
			t.Errorf("room %d numbered %d", i+1, r.Number)
		}
		if r.Monster != nil {
			monsters++
		}
		if r.Treasure {
			treasure++
		}
		if r.Trapped && r.Monster != nil {
			t.Error("guarded treasure should not be trapped")
		}
	}
	// Roughly a third of rooms hold monsters; just confirm both occur.
	if monsters == 0 || treasure == 0 {
		t.Errorf("expected some monsters (%d) and treasure (%d)", monsters, treasure)
	}
}
