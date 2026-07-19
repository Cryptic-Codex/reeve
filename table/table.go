// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

// Package table rolls on referee tables: dice-range entries that yield text or
// reference nested subtables. Tables load from a simple line format so a
// referee can add their own, and the entries of a table must tile its dice
// range exactly (Validate enforces this).
package table

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Cryptic-Codex/reeve/dice"
)

// Entry is one row of a table. It covers the inclusive roll range [Lo, Hi] and
// yields either literal Text or, when Ref is set, the result of rolling the
// named subtable.
type Entry struct {
	Lo, Hi int
	Text   string
	Ref    string
}

// Table is a roll table: Count dice of Sides sides selecting a range-keyed
// entry.
type Table struct {
	Name    string
	Count   int
	Sides   int
	Entries []Entry
}

// Min returns the lowest possible roll on the table.
func (t *Table) Min() int { return t.Count }

// Max returns the highest possible roll on the table.
func (t *Table) Max() int { return t.Count * t.Sides }

// Dice returns the table's dice in standard notation, e.g. "2d6".
func (t *Table) Dice() string { return fmt.Sprintf("%dd%d", t.Count, t.Sides) }

func (t *Table) rollDice(rng *rand.Rand) int {
	total := 0
	for i := 0; i < t.Count; i++ {
		total += rng.Intn(t.Sides) + 1
	}
	return total
}

func (t *Table) find(roll int) (Entry, bool) {
	for _, e := range t.Entries {
		if roll >= e.Lo && roll <= e.Hi {
			return e, true
		}
	}
	return Entry{}, false
}

// Validate reports whether the table's entries tile its full dice range
// exactly, in ascending order with no gaps or overlaps.
func (t *Table) Validate() error {
	next := t.Min()
	for _, e := range t.Entries {
		if e.Hi < e.Lo {
			return fmt.Errorf("table %q: entry %d-%d is inverted", t.Name, e.Lo, e.Hi)
		}
		if e.Lo != next {
			return fmt.Errorf("table %q: gap or overlap at %d (next entry starts %d)", t.Name, next, e.Lo)
		}
		next = e.Hi + 1
	}
	if next != t.Max()+1 {
		return fmt.Errorf("table %q: entries cover up to %d, want %d", t.Name, next-1, t.Max())
	}
	return nil
}

// Result is the outcome of rolling a table.
type Result struct {
	Roll int
	Text string
}

// Registry holds named tables so entries can reference subtables.
type Registry map[string]*Table

// Names returns the registry's table names in sorted order.
func (r Registry) Names() []string {
	names := make([]string, 0, len(r))
	for n := range r {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Validate validates every table in the registry and confirms that each
// subtable reference resolves.
func (r Registry) Validate() error {
	for _, name := range r.Names() {
		t := r[name]
		if err := t.Validate(); err != nil {
			return err
		}
		for _, e := range t.Entries {
			if e.Ref != "" && r[e.Ref] == nil {
				return fmt.Errorf("table %q references unknown table %q", t.Name, e.Ref)
			}
		}
	}
	return nil
}

// Roll rolls the named table with rng, resolving any subtable references.
func (r Registry) Roll(name string, rng *rand.Rand) (Result, error) {
	t, ok := r[name]
	if !ok {
		return Result{}, fmt.Errorf("unknown table %q", name)
	}
	return r.rollTable(t, rng)
}

func (r Registry) rollTable(t *Table, rng *rand.Rand) (Result, error) {
	roll := t.rollDice(rng)
	e, ok := t.find(roll)
	if !ok {
		return Result{}, fmt.Errorf("table %q has no entry for roll %d", t.Name, roll)
	}

	text := e.Text
	if e.Ref != "" {
		sub, ok := r[e.Ref]
		if !ok {
			return Result{}, fmt.Errorf("table %q references unknown table %q", t.Name, e.Ref)
		}
		subRes, err := r.rollTable(sub, rng)
		if err != nil {
			return Result{}, err
		}
		if text == "" {
			text = subRes.Text
		} else {
			text += " — " + subRes.Text
		}
	}
	return Result{Roll: roll, Text: text}, nil
}

// Load parses tables from src in the line format and adds them to reg.
//
// The format is:
//
//	# comment
//	table: <name> <dice>       start a table, e.g. "table: reaction 2d6"
//	<lo>[-<hi>]: <text>        a leaf entry
//	<lo>[-<hi>]: @<othertable> roll on another table instead
func Load(reg Registry, src io.Reader) error {
	scanner := bufio.NewScanner(src)
	var cur *Table
	line := 0

	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		if rest, ok := strings.CutPrefix(text, "table:"); ok {
			t, err := parseHeader(rest)
			if err != nil {
				return fmt.Errorf("line %d: %w", line, err)
			}
			reg[t.Name] = t
			cur = t
			continue
		}

		if cur == nil {
			return fmt.Errorf("line %d: entry before any table header", line)
		}
		e, err := parseEntry(text)
		if err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
		cur.Entries = append(cur.Entries, e)
	}
	return scanner.Err()
}

func parseHeader(s string) (*Table, error) {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return nil, fmt.Errorf("table header needs a name and dice, e.g. \"table: reaction 2d6\"")
	}
	r, err := dice.Parse(fields[1])
	if err != nil {
		return nil, err
	}
	if r.Modifier != 0 {
		return nil, fmt.Errorf("table dice may not have a modifier: %q", fields[1])
	}
	return &Table{Name: fields[0], Count: r.Count, Sides: r.Sides}, nil
}

func parseEntry(s string) (Entry, error) {
	key, text, ok := strings.Cut(s, ":")
	if !ok {
		return Entry{}, fmt.Errorf("entry needs a \"range: text\" form: %q", s)
	}
	lo, hi, err := parseRange(strings.TrimSpace(key))
	if err != nil {
		return Entry{}, err
	}

	e := Entry{Lo: lo, Hi: hi}
	text = strings.TrimSpace(text)
	if ref, ok := strings.CutPrefix(text, "@"); ok {
		e.Ref = strings.TrimSpace(ref)
	} else {
		e.Text = text
	}
	return e, nil
}

func parseRange(s string) (lo, hi int, err error) {
	if before, after, ok := strings.Cut(s, "-"); ok {
		if lo, err = strconv.Atoi(strings.TrimSpace(before)); err != nil {
			return 0, 0, err
		}
		hi, err = strconv.Atoi(strings.TrimSpace(after))
		return lo, hi, err
	}
	lo, err = strconv.Atoi(s)
	return lo, lo, err
}

// NewRand returns a time-seeded source for rolling tables.
func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
