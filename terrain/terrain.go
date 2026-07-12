// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

// Package terrain generates random wilderness terrain hex by hex,
// transitioning from the current hex's terrain (after AD&D DMG Appendix B).
package terrain

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Type is a base terrain type a hex can hold.
type Type int

const (
	Plain Type = iota
	Scrub
	Forest
	Rough
	Desert
	Hills
	Mountains
	Marsh
	numTypes
)

var typeNames = [...]string{"plain", "scrub", "forest", "rough", "desert", "hills", "mountains", "marsh"}

func (t Type) String() string { return typeNames[t] }

// Parse converts a name like "plain" or "Mountains" into a Type.
func Parse(s string) (Type, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	for i, n := range typeNames {
		if s == n {
			return Type(i), nil
		}
	}
	return 0, fmt.Errorf("unknown terrain %q (want one of %s)", s, strings.Join(typeNames[:], ", "))
}

// Feature is a special result occupying a hex alongside its base terrain.
type Feature int

const (
	None       Feature = iota
	Pond               // pool, tarn, lake — terrain matches the previous hex
	Depression         // gorge, rift, valley — terrain matches the previous hex
	Pass               // a pass leading through a mountain range
)

func (f Feature) String() string {
	switch f {
	case Pond:
		return "pond"
	case Depression:
		return "depression"
	case Pass:
		return "pass"
	}
	return ""
}

// Hex is one generated wilderness space.
type Hex struct {
	Type         Type
	Feature      Feature
	Secondary    Type // e.g. hills within forest
	HasSecondary bool
	Roll         int // the d20 that produced this hex
}

func (h Hex) String() string {
	s := h.Type.String()
	if h.HasSecondary {
		s += " (with " + h.Secondary.String() + ")"
	}
	switch h.Feature {
	case Pond:
		s += " with a pond"
	case Depression:
		s += " with a depression"
	case Pass:
		s += " (a pass leads through)"
	}
	return s
}

// result is what a table row can yield: a terrain, or a pond/depression.
type result int

const (
	rPond result = result(numTypes) + iota
	rDepression
)

type span struct {
	lo, hi int
	out    result
}

// transitions holds, per current terrain, the d20 spans for the next hex.
// Spans must tile 1..20 exactly; terrain_test.go enforces this.
var transitions = [numTypes][]span{
	Plain: {
		{1, 11, result(Plain)}, {12, 12, result(Scrub)}, {13, 13, result(Forest)},
		{14, 14, result(Rough)}, {15, 15, result(Desert)}, {16, 16, result(Hills)},
		{17, 17, result(Mountains)}, {18, 18, result(Marsh)}, {19, 19, rPond}, {20, 20, rDepression},
	},
	Scrub: {
		{1, 3, result(Plain)}, {4, 11, result(Scrub)}, {12, 13, result(Forest)},
		{14, 14, result(Rough)}, {15, 15, result(Desert)}, {16, 16, result(Hills)},
		{17, 17, result(Mountains)}, {18, 18, result(Marsh)}, {19, 19, rPond}, {20, 20, rDepression},
	},
	Forest: {
		{1, 1, result(Plain)}, {2, 4, result(Scrub)}, {5, 14, result(Forest)},
		{15, 15, result(Rough)}, {16, 16, result(Hills)},
		{17, 17, result(Mountains)}, {18, 18, result(Marsh)}, {19, 19, rPond}, {20, 20, rDepression},
	},
	Rough: {
		{1, 2, result(Plain)}, {3, 4, result(Scrub)}, {5, 5, result(Forest)},
		{6, 8, result(Rough)}, {9, 10, result(Desert)}, {11, 15, result(Hills)},
		{16, 17, result(Mountains)}, {18, 18, result(Marsh)}, {19, 19, rPond}, {20, 20, rDepression},
	},
	Desert: {
		{1, 3, result(Plain)}, {4, 5, result(Scrub)},
		{6, 8, result(Rough)}, {9, 14, result(Desert)}, {15, 15, result(Hills)},
		{16, 17, result(Mountains)}, {18, 18, result(Marsh)}, {19, 19, rPond}, {20, 20, rDepression},
	},
	Hills: {
		{1, 1, result(Plain)}, {2, 3, result(Scrub)}, {4, 5, result(Forest)},
		{6, 7, result(Rough)}, {8, 8, result(Desert)}, {9, 14, result(Hills)},
		{15, 16, result(Mountains)}, {17, 17, result(Marsh)}, {18, 19, rPond}, {20, 20, rDepression},
	},
	Mountains: {
		{1, 1, result(Plain)}, {2, 2, result(Scrub)}, {3, 3, result(Forest)},
		{4, 5, result(Rough)}, {6, 6, result(Desert)}, {7, 10, result(Hills)},
		{11, 18, result(Mountains)}, {19, 19, rPond}, {20, 20, rDepression},
	},
	Marsh: {
		{1, 2, result(Plain)}, {3, 4, result(Scrub)}, {5, 6, result(Forest)},
		{7, 7, result(Rough)}, {8, 8, result(Hills)},
		{9, 15, result(Marsh)}, {16, 19, rPond}, {20, 20, rDepression},
	},
}

// Generator produces successive hexes. It is not safe for concurrent use.
type Generator struct {
	rng *rand.Rand
}

// New returns a Generator seeded from the current time.
func New() *Generator {
	return &Generator{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// NewSeeded returns a deterministic Generator, useful for tests and
// for reproducible wilderness (record the seed in your campaign notes).
func NewSeeded(seed int64) *Generator {
	return &Generator{rng: rand.New(rand.NewSource(seed))}
}

// Next rolls the next hex entered from a hex of the given terrain.
// For hexes with a pond or depression, pass the hex's base Type.
func (g *Generator) Next(current Type) Hex {
	roll := g.rng.Intn(20) + 1
	h := Hex{Roll: roll}

	var out result = -1
	for _, sp := range transitions[current] {
		if roll >= sp.lo && roll <= sp.hi {
			out = sp.out
			break
		}
	}

	switch out {
	case rPond:
		h.Type, h.Feature = current, Pond
	case rDepression:
		h.Type, h.Feature = current, Depression
	default:
		h.Type = Type(out)
		switch h.Type {
		case Forest:
			if g.rng.Intn(10) == 0 { // 1 in 10 also includes hills
				h.Secondary, h.HasSecondary = Hills, true
			}
		case Hills:
			if g.rng.Intn(10) == 0 { // 1 in 10 also includes forest
				h.Secondary, h.HasSecondary = Forest, true
			}
		case Mountains:
			if g.rng.Intn(20) == 0 { // 1 in 20 have a pass
				h.Feature = Pass
			}
		}
	}
	return h
}

// Walk generates n successive hexes starting from the given terrain,
// feeding each hex's base terrain into the next roll.
func (g *Generator) Walk(start Type, n int) []Hex {
	hexes := make([]Hex, n)
	cur := start
	for i := range hexes {
		hexes[i] = g.Next(cur)
		cur = hexes[i].Type // pond/depression keep base terrain, so this is right
	}
	return hexes
}
