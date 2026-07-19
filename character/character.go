// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

// Package character rolls player characters for 3LBB OD&D (Men & Magic, 1974):
// six 3d6 ability scores in order, a class chosen by prime requisite, and the
// experience adjustment and level limits that follow from ability and race.
package character

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Ability is one of the six OD&D ability scores.
type Ability int

const (
	Strength Ability = iota
	Intelligence
	Wisdom
	Constitution
	Dexterity
	Charisma
	numAbilities
)

var abilityNames = [...]string{"Strength", "Intelligence", "Wisdom", "Constitution", "Dexterity", "Charisma"}

func (a Ability) String() string { return abilityNames[a] }

// Scores holds the six ability scores in canonical 3LBB roll order.
type Scores [numAbilities]int

// Class is one of the three original OD&D classes.
type Class int

const (
	FightingMan Class = iota
	MagicUser
	Cleric
	numClasses
)

var classNames = [...]string{"Fighting-Man", "Magic-User", "Cleric"}

func (c Class) String() string { return classNames[c] }

// primeReq maps a class to the ability that serves as its prime requisite.
var primeReq = [numClasses]Ability{
	FightingMan: Strength,
	MagicUser:   Intelligence,
	Cleric:      Wisdom,
}

// PrimeReq returns the ability that is the class's prime requisite.
func (c Class) PrimeReq() Ability { return primeReq[c] }

// Race is a playable character type. OD&D calls these the human and
// non-human types; each restricts class choice and caps advancement.
type Race int

const (
	Human Race = iota
	Dwarf
	Elf
	Hobbit
	numRaces
)

var raceNames = [...]string{"Human", "Dwarf", "Elf", "Hobbit"}

func (r Race) String() string { return raceNames[r] }

// maxLevel[race][class] is the 3LBB level cap for a race in a class:
// -1 marks a class the race may not take, and 0 marks no limit.
var maxLevel = [numRaces][numClasses]int{
	Human:  {FightingMan: 0, MagicUser: 0, Cleric: 0},
	Dwarf:  {FightingMan: 6, MagicUser: -1, Cleric: -1},
	Elf:    {FightingMan: 4, MagicUser: 8, Cleric: -1},
	Hobbit: {FightingMan: 4, MagicUser: -1, Cleric: -1},
}

// Available reports whether the race may take the class in the 3LBB.
func Available(r Race, c Class) bool { return maxLevel[r][c] >= 0 }

// XPModifier returns the percent experience adjustment for a prime-requisite
// score, per the 3LBB table "Bonuses and Penalties to Advancement due to
// Abilities".
func XPModifier(primeReqScore int) int {
	switch {
	case primeReqScore >= 15:
		return 10
	case primeReqScore >= 13:
		return 5
	case primeReqScore >= 9:
		return 0
	case primeReqScore >= 7:
		return -10
	default:
		return -20
	}
}

// RecommendClass returns the class whose prime requisite is highest among the
// classes available to the race. Ties resolve in canonical class order
// (Fighting-Man, then Magic-User, then Cleric).
func RecommendClass(s Scores, r Race) Class {
	best := FightingMan
	bestScore := -1
	for c := Class(0); c < numClasses; c++ {
		if !Available(r, c) {
			continue
		}
		if score := s[c.PrimeReq()]; score > bestScore {
			best, bestScore = c, score
		}
	}
	return best
}

// Character is a rolled 3LBB character.
type Character struct {
	Race   Race
	Class  Class
	Scores Scores
}

// New builds a Character, returning an error if the class is not available
// to the race in the 3LBB.
func New(r Race, c Class, s Scores) (Character, error) {
	if !Available(r, c) {
		return Character{}, fmt.Errorf("a %s cannot be a %s", r, c)
	}
	return Character{Race: r, Class: c, Scores: s}, nil
}

// PrimeReqScore returns the character's score in its class's prime requisite.
func (c Character) PrimeReqScore() int { return c.Scores[c.Class.PrimeReq()] }

// XPModifier returns the character's experience adjustment percent.
func (c Character) XPModifier() int { return XPModifier(c.PrimeReqScore()) }

// MaxLevel returns the character's level cap, or 0 for no limit.
func (c Character) MaxLevel() int { return maxLevel[c.Race][c.Class] }

func (c Character) String() string { return c.Race.String() + " " + c.Class.String() }

// ParseRace converts a name like "elf" or "Halfling" into a Race.
func ParseRace(s string) (Race, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "human", "humans", "man", "men":
		return Human, nil
	case "dwarf", "dwarves", "dwarven":
		return Dwarf, nil
	case "elf", "elves", "elven":
		return Elf, nil
	case "hobbit", "hobbits", "halfling", "halflings":
		return Hobbit, nil
	}
	return 0, fmt.Errorf("unknown race %q (want one of %s)", s, strings.ToLower(strings.Join(raceNames[:], ", ")))
}

// ParseClass converts a name like "fighter" or "magic-user" into a Class.
func ParseClass(s string) (Class, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "fighting-man", "fighting man", "fightingman", "fighter", "fighters", "f":
		return FightingMan, nil
	case "magic-user", "magic user", "magicuser", "mu", "m", "wizard", "wizards":
		return MagicUser, nil
	case "cleric", "clerics", "c":
		return Cleric, nil
	}
	return 0, fmt.Errorf("unknown class %q (want fighting-man, magic-user, or cleric)", s)
}

// Generator rolls characters. It is not safe for concurrent use.
type Generator struct {
	rng *rand.Rand
}

// New returns a Generator seeded from the current time.
func NewGenerator() *Generator {
	return &Generator{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// NewSeeded returns a deterministic Generator, useful for tests and for
// reproducible characters (record the seed in your campaign notes).
func NewSeeded(seed int64) *Generator {
	return &Generator{rng: rand.New(rand.NewSource(seed))}
}

// RollScores rolls 3d6 once for each ability, in canonical 3LBB order.
func (g *Generator) RollScores() Scores {
	var s Scores
	for i := range s {
		s[i] = g.rng.Intn(6) + g.rng.Intn(6) + g.rng.Intn(6) + 3
	}
	return s
}

// Roll rolls ability scores and builds a character of the given race, using
// the class recommended by prime requisite.
func (g *Generator) Roll(r Race) Character {
	s := g.RollScores()
	c, _ := New(r, RecommendClass(s, r), s)
	return c
}
