// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package terrain

import "testing"

// Every column of the transition table must cover 1..20 exactly once —
// this catches transcription errors against the source table.
func TestTransitionsTile(t *testing.T) {
	for ty := Type(0); ty < numTypes; ty++ {
		var covered [21]bool
		for _, sp := range transitions[ty] {
			if sp.lo < 1 || sp.hi > 20 || sp.lo > sp.hi {
				t.Fatalf("%s: bad span %+v", ty, sp)
			}
			for i := sp.lo; i <= sp.hi; i++ {
				if covered[i] {
					t.Errorf("%s: roll %d covered twice", ty, i)
				}
				covered[i] = true
			}
		}
		for i := 1; i <= 20; i++ {
			if !covered[i] {
				t.Errorf("%s: roll %d not covered", ty, i)
			}
		}
	}
}

// Spot-check the dashes: transitions the table says can't happen.
func TestImpossibleTransitions(t *testing.T) {
	forbidden := map[Type][]Type{
		Forest:    {Desert},            // no desert from forest
		Desert:    {Forest},            // no forest from desert
		Mountains: {Marsh},             // no marsh from mountains
		Marsh:     {Desert, Mountains}, // neither from marsh
	}
	g := NewSeeded(1)
	for from, banned := range forbidden {
		for i := 0; i < 5000; i++ {
			h := g.Next(from)
			for _, b := range banned {
				if h.Type == b && h.Feature != Pond && h.Feature != Depression {
					t.Fatalf("impossible transition %s -> %s (roll %d)", from, b, h.Roll)
				}
			}
		}
	}
}

// Ponds and depressions must inherit the current hex's terrain.
func TestPondDepressionInherit(t *testing.T) {
	g := NewSeeded(2)
	for i := 0; i < 5000; i++ {
		h := g.Next(Marsh)
		if (h.Feature == Pond || h.Feature == Depression) && h.Type != Marsh {
			t.Fatalf("feature hex should keep terrain marsh, got %s", h.Type)
		}
	}
}

// Rough sanity check of the distribution: from plain, plain should
// dominate (11/20 = 55%); allow generous tolerance.
func TestPlainSticky(t *testing.T) {
	g := NewSeeded(3)
	n, plains := 20000, 0
	for i := 0; i < n; i++ {
		if h := g.Next(Plain); h.Type == Plain && h.Feature == None {
			plains++
		}
	}
	frac := float64(plains) / float64(n)
	if frac < 0.50 || frac > 0.60 {
		t.Errorf("plain->plain fraction %.3f, want ~0.55", frac)
	}
}

// Same seed, same wilderness — reproducibility for campaign notes.
func TestSeededDeterminism(t *testing.T) {
	a := NewSeeded(505).Walk(Forest, 50)
	b := NewSeeded(505).Walk(Forest, 50)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("walks diverge at hex %d: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestParse(t *testing.T) {
	if ty, err := Parse(" Mountains "); err != nil || ty != Mountains {
		t.Errorf("Parse(Mountains) = %v, %v", ty, err)
	}
	if _, err := Parse("tundra"); err == nil {
		t.Error("Parse(tundra) should error (synonyms not accepted yet)")
	}
}
