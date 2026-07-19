// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

// Package campaign persists referee data — characters, monsters, and notes —
// as one JSON file per campaign, with a notion of the current campaign. Domain
// structs are stored directly, so the same types drive the CLI, storage, and
// any future web layer.
package campaign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Cryptic-Codex/reeve/character"
	"github.com/Cryptic-Codex/reeve/monster"
)

// Campaign is a saved collection of referee data.
type Campaign struct {
	Name       string
	Created    time.Time
	Notes      []string              `json:",omitempty"`
	Characters []character.Character `json:",omitempty"`
	Monsters   []monster.Monster     `json:",omitempty"`
}

// AddCharacter appends a rolled character to the campaign.
func (c *Campaign) AddCharacter(ch character.Character) { c.Characters = append(c.Characters, ch) }

// AddMonster appends a monster stat block to the campaign.
func (c *Campaign) AddMonster(m monster.Monster) { c.Monsters = append(c.Monsters, m) }

// Store is a directory holding campaign files and a pointer to the current one.
type Store struct {
	root string
}

// Open returns the default store, rooted at $REEVE_HOME or, failing that, the
// user config directory's "reeve" folder.
func Open() (*Store, error) {
	if env := os.Getenv("REEVE_HOME"); env != "" {
		return &Store{root: env}, nil
	}
	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Store{root: filepath.Join(cfg, "reeve")}, nil
}

// At returns a store rooted at dir, useful for tests.
func At(dir string) *Store { return &Store{root: dir} }

var nameRe = regexp.MustCompile(`^[a-z0-9_-]+$`)

func normalize(name string) (string, error) {
	n := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-")
	if !nameRe.MatchString(n) {
		return "", fmt.Errorf("invalid campaign name %q (use letters, digits, - and _)", name)
	}
	return n, nil
}

func (s *Store) campaignsDir() string    { return filepath.Join(s.root, "campaigns") }
func (s *Store) path(name string) string { return filepath.Join(s.campaignsDir(), name+".json") }
func (s *Store) currentPath() string     { return filepath.Join(s.root, "current") }

// List returns the names of all saved campaigns, sorted.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.campaignsDir())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if n, ok := strings.CutSuffix(e.Name(), ".json"); ok {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names, nil
}

// Create makes a new empty campaign, saves it, and makes it current.
func (s *Store) Create(name string) (*Campaign, error) {
	n, err := normalize(name)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(s.path(n)); err == nil {
		return nil, fmt.Errorf("campaign %q already exists", n)
	}
	c := &Campaign{Name: n, Created: time.Now()}
	if err := s.Save(c); err != nil {
		return nil, err
	}
	return c, s.SetCurrent(n)
}

// Load reads a campaign by name.
func (s *Store) Load(name string) (*Campaign, error) {
	n, err := normalize(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(s.path(n))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("no campaign named %q", n)
	}
	if err != nil {
		return nil, err
	}
	var c Campaign
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("campaign %q is corrupt: %w", n, err)
	}
	return &c, nil
}

// Save writes a campaign to disk as indented JSON.
func (s *Store) Save(c *Campaign) error {
	if err := os.MkdirAll(s.campaignsDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(c.Name), data, 0o644)
}

// Current returns the current campaign's name, or "" if none is set.
func (s *Store) Current() (string, error) {
	data, err := os.ReadFile(s.currentPath())
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SetCurrent records name as the current campaign.
func (s *Store) SetCurrent(name string) error {
	n, err := normalize(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.currentPath(), []byte(n+"\n"), 0o644)
}

// LoadCurrent loads the current campaign, erroring if none is set.
func (s *Store) LoadCurrent() (*Campaign, error) {
	n, err := s.Current()
	if err != nil {
		return nil, err
	}
	if n == "" {
		return nil, fmt.Errorf("no current campaign (use `reeve campaign new <name>`)")
	}
	return s.Load(n)
}
