// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package treasure

import (
	"math/rand"
	"testing"
)

func TestRollTypes(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, letter := range Types() {
		for i := 0; i < 500; i++ {
			h, err := Roll(letter, rng)
			if err != nil {
				t.Fatalf("Roll(%s): %v", letter, err)
			}
			if h.Type != letter {
				t.Errorf("Type = %s, want %s", h.Type, letter)
			}
			if h.Copper < 0 || h.Silver < 0 || h.Gold < 0 || h.Magic < 0 {
				t.Fatalf("negative amount in %+v", h)
			}
			if h.TotalGP != h.valueGP() {
				t.Errorf("TotalGP %d != valueGP %d", h.TotalGP, h.valueGP())
			}
			for _, v := range h.Gems {
				if v < 10 || v > 1000 {
					t.Errorf("gem value out of range: %d", v)
				}
			}
			for _, v := range h.Jewelry {
				if v < 300 || v > 1800 {
					t.Errorf("jewelry value out of range: %d", v)
				}
			}
		}
	}
}

func TestCaseInsensitive(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	if _, err := Roll("a", rng); err != nil {
		t.Errorf("Roll(a) should match type A: %v", err)
	}
}

func TestUnknownType(t *testing.T) {
	if _, err := Roll("Z", rand.New(rand.NewSource(1))); err == nil {
		t.Error("Roll(Z) should error")
	}
}
