// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package monster

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

//go:embed data/bestiary.txt
var builtinData string

// Bestiary is a collection of monsters keyed by lower-cased name.
type Bestiary map[string]*Monster

// Builtin returns the bestiary bundled with reeve.
func Builtin() (Bestiary, error) {
	b := Bestiary{}
	if err := Load(b, strings.NewReader(builtinData)); err != nil {
		return nil, err
	}
	return b, nil
}

// Lookup finds a monster by name, case-insensitively.
func (b Bestiary) Lookup(name string) (*Monster, bool) {
	m, ok := b[strings.ToLower(strings.TrimSpace(name))]
	return m, ok
}

// Names returns the bestiary's monster names in sorted order.
func (b Bestiary) Names() []string {
	names := make([]string, 0, len(b))
	for _, m := range b {
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return names
}

// Load parses monsters from src in the block format and adds them to b.
//
// Each monster is a block of "key: value" lines started by a "monster: Name"
// line; blank lines and lines beginning with '#' are ignored.
func Load(b Bestiary, src io.Reader) error {
	scanner := bufio.NewScanner(src)
	var cur *Monster
	line := 0

	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		key, val, ok := strings.Cut(text, ":")
		if !ok {
			return fmt.Errorf("line %d: expected \"key: value\": %q", line, text)
		}
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.TrimSpace(val)

		if key == "monster" {
			cur = &Monster{Name: val, Attacks: 1, Damage: "1d6"}
			b[strings.ToLower(val)] = cur
			continue
		}
		if cur == nil {
			return fmt.Errorf("line %d: field %q before any monster header", line, key)
		}
		if err := cur.set(key, val); err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
	}
	return scanner.Err()
}

func (m *Monster) set(key, val string) error {
	atoi := func() (int, error) { return strconv.Atoi(val) }
	var err error
	switch key {
	case "hd":
		m.HitDice, err = ParseHD(val)
	case "ac":
		m.ArmorClass, err = atoi()
	case "move":
		m.Move = val
	case "attacks":
		m.Attacks, err = atoi()
	case "damage":
		m.Damage = val
	case "number":
		m.NumberAppear = val
	case "lair":
		m.InLair, err = atoi()
	case "treasure":
		m.TreasureType = val
	case "alignment":
		m.Alignment = val
	case "notes":
		m.Notes = val
	default:
		return fmt.Errorf("unknown field %q", key)
	}
	return err
}
