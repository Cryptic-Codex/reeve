// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package dice

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
)

type Roll struct {
	Count, Sides, Modifier int
}

type Result struct {
	Rolls []int
	Total int
}

var diceRe = regexp.MustCompile(`^(\d*)d(\d+)\s*(?:([+-])\s*(\d+))?$`)

// Parse parses a string in the format of "<C>d<S> +/- <M>" to create
// a roll object. This function trims whitespace, only allows for one use
// of the die designator 'd', and disallows (for now) multiplication or division
// modifiers
func Parse(s string) (Roll, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	m := diceRe.FindStringSubmatch(s)
	if m == nil {
		return Roll{}, fmt.Errorf("invalid dice notation: %q", s)
	}

	count := 1
	if m[1] != "" {
		count, _ = strconv.Atoi(m[1])
	}
	sides, _ := strconv.Atoi(m[2])

	if count < 1 || count > 1000 {
		return Roll{}, fmt.Errorf("dice count out of range: %d", count)
	}
	if sides < 1 {
		return Roll{}, fmt.Errorf("sides must be at least 1: %d", sides)
	}

	mod := 0
	if m[3] != "" {
		mod, _ = strconv.Atoi(m[4])
		if m[3] == "-" {
			mod = -mod
		}
	}

	return Roll{Count: count, Sides: sides, Modifier: mod}, nil
}

func (r Roll) Roll() Result {
	total := 0
	rolls := make([]int, r.Count)
	for i := 0; i < r.Count; i++ {
		rolls[i] = rand.IntN(r.Sides) + 1
		total += rolls[i]
	}
	if r.Modifier != 0 {
		total += r.Modifier
	}
	return Result{Rolls: rolls, Total: total}
}
