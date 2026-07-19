// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

// Package treasure generates hoards by 3LBB treasure type (Monsters &
// Treasure). The type values here are approximate OD&D lair treasures; verify
// against Vol 2 for your table.
package treasure

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

// category is one line of a treasure type: a percent chance of appearing and,
// when it does, an amount of (n d sides) × mult.
type category struct {
	chance, n, sides, mult int
}

func (c category) roll(rng *rand.Rand) int {
	if c.chance == 0 || rng.Intn(100) >= c.chance {
		return 0
	}
	total := 0
	for i := 0; i < c.n; i++ {
		total += rng.Intn(c.sides) + 1
	}
	return total * c.mult
}

// Type is a treasure type: the odds and amounts of each kind of loot. Gems,
// Jewelry, and Magic amounts are counts rather than coins.
type Type struct {
	Letter               string
	Copper, Silver, Gold category
	Gems, Jewelry, Magic category
}

// types holds the built-in lair treasure types (approximate OD&D values).
var types = map[string]Type{
	"A": {
		Letter: "A",
		Copper: category{25, 1, 6, 1000}, Silver: category{30, 1, 6, 1000}, Gold: category{35, 1, 4, 1000},
		Gems: category{50, 6, 6, 1}, Jewelry: category{50, 6, 6, 1}, Magic: category{30, 1, 3, 1},
	},
	"B": {
		Letter: "B",
		Copper: category{50, 1, 8, 1000}, Silver: category{25, 1, 6, 1000}, Gold: category{25, 1, 4, 1000},
		Gems: category{25, 1, 6, 1}, Jewelry: category{25, 1, 3, 1}, Magic: category{10, 1, 1, 1},
	},
	"C": {
		Letter: "C",
		Copper: category{20, 1, 12, 1000}, Silver: category{30, 1, 6, 1000}, Gold: category{10, 1, 4, 1000},
		Gems: category{25, 1, 4, 1}, Jewelry: category{20, 1, 4, 1}, Magic: category{10, 1, 2, 1},
	},
	"D": {
		Letter: "D",
		Copper: category{10, 1, 8, 1000}, Silver: category{15, 1, 12, 1000}, Gold: category{60, 1, 6, 1000},
		Gems: category{30, 1, 8, 1}, Jewelry: category{25, 1, 6, 1}, Magic: category{15, 1, 2, 1},
	},
	"E": {
		Letter: "E",
		Copper: category{5, 1, 10, 1000}, Silver: category{30, 1, 12, 1000}, Gold: category{25, 1, 8, 1000},
		Gems: category{15, 1, 12, 1}, Jewelry: category{10, 1, 8, 1}, Magic: category{25, 1, 3, 1},
	},
}

// Types returns the available treasure-type letters, sorted.
func Types() []string {
	ts := make([]string, 0, len(types))
	for l := range types {
		ts = append(ts, l)
	}
	sort.Strings(ts)
	return ts
}

// Hoard is a rolled treasure hoard. Gems and Jewelry list the gp value of each
// item; Magic is a count of magic items to roll on the magic tables.
type Hoard struct {
	Type                 string
	Copper, Silver, Gold int
	Gems, Jewelry        []int
	Magic                int
	TotalGP              int
}

// Roll generates a hoard of the given treasure type.
func Roll(letter string, rng *rand.Rand) (Hoard, error) {
	t, ok := types[strings.ToUpper(strings.TrimSpace(letter))]
	if !ok {
		return Hoard{}, fmt.Errorf("unknown treasure type %q (have %s)", letter, strings.Join(Types(), ", "))
	}

	h := Hoard{
		Type:   t.Letter,
		Copper: t.Copper.roll(rng),
		Silver: t.Silver.roll(rng),
		Gold:   t.Gold.roll(rng),
		Magic:  t.Magic.roll(rng),
	}
	if n := t.Gems.roll(rng); n > 0 {
		h.Gems = make([]int, n)
		for i := range h.Gems {
			h.Gems[i] = gemValue(rng)
		}
	}
	if n := t.Jewelry.roll(rng); n > 0 {
		h.Jewelry = make([]int, n)
		for i := range h.Jewelry {
			h.Jewelry[i] = jewelryValue(rng)
		}
	}
	h.TotalGP = h.valueGP()
	return h, nil
}

// valueGP totals the hoard's worth in gold, taking 10 sp and 100 cp to the gp.
func (h Hoard) valueGP() int {
	gp := h.Gold + h.Silver/10 + h.Copper/100
	for _, v := range h.Gems {
		gp += v
	}
	for _, v := range h.Jewelry {
		gp += v
	}
	return gp
}

// gemValue rolls the gp value of a single gem.
func gemValue(rng *rand.Rand) int {
	switch roll := rng.Intn(100) + 1; {
	case roll <= 20:
		return 10
	case roll <= 45:
		return 50
	case roll <= 75:
		return 100
	case roll <= 95:
		return 500
	default:
		return 1000
	}
}

// jewelryValue rolls the gp value of a single piece of jewelry (3d6 × 100).
func jewelryValue(rng *rand.Rand) int {
	return (rng.Intn(6) + rng.Intn(6) + rng.Intn(6) + 3) * 100
}
