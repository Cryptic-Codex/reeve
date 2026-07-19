// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package campaign

import (
	"testing"

	"github.com/Cryptic-Codex/reeve/character"
	"github.com/Cryptic-Codex/reeve/monster"
)

func TestCreateListLoad(t *testing.T) {
	s := At(t.TempDir())

	if names, err := s.List(); err != nil || len(names) != 0 {
		t.Fatalf("empty store List = %v, %v", names, err)
	}

	if _, err := s.Create("Keep on the Borderlands"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	names, err := s.List()
	if err != nil || len(names) != 1 || names[0] != "keep-on-the-borderlands" {
		t.Fatalf("List = %v, %v", names, err)
	}

	cur, err := s.Current()
	if err != nil || cur != "keep-on-the-borderlands" {
		t.Fatalf("Current = %q, %v", cur, err)
	}

	if _, err := s.Create("keep-on-the-borderlands"); err == nil {
		t.Error("creating a duplicate campaign should error")
	}
}

func TestRoundTrip(t *testing.T) {
	s := At(t.TempDir())
	c, err := s.Create("dungeon")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ch, _ := character.New(character.Elf, character.MagicUser,
		character.Scores{11, 15, 12, 9, 13, 8})
	ch.Alignment = character.Law
	ch.HP = 3
	ch.Gold = 120
	c.AddCharacter(ch)

	m, ok := mustBestiary(t).Lookup("ogre")
	if !ok {
		t.Fatal("ogre missing from bestiary")
	}
	c.AddMonster(*m)

	c.Notes = append(c.Notes, "the party fled the ogre")
	if err := s.Save(c); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Load("dungeon")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Characters) != 1 || got.Characters[0] != ch {
		t.Errorf("character round-trip mismatch: %+v", got.Characters)
	}
	if len(got.Monsters) != 1 || got.Monsters[0].Name != "Ogre" ||
		got.Monsters[0].HitDice.String() != "4+1" {
		t.Errorf("monster round-trip mismatch: %+v", got.Monsters)
	}
	if len(got.Notes) != 1 || got.Notes[0] != "the party fled the ogre" {
		t.Errorf("notes round-trip mismatch: %+v", got.Notes)
	}
}

func TestNoCurrent(t *testing.T) {
	s := At(t.TempDir())
	if _, err := s.LoadCurrent(); err == nil {
		t.Error("LoadCurrent with no current campaign should error")
	}
}

func TestInvalidName(t *testing.T) {
	s := At(t.TempDir())
	if _, err := s.Create("bad/name"); err == nil {
		t.Error("Create with a slash should error")
	}
}

func mustBestiary(t *testing.T) monster.Bestiary {
	t.Helper()
	b, err := monster.Builtin()
	if err != nil {
		t.Fatalf("bestiary: %v", err)
	}
	return b
}
