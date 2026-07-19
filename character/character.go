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

// MarshalText and UnmarshalText render a Class as its name.
func (c Class) MarshalText() ([]byte, error) { return []byte(c.String()), nil }

func (c *Class) UnmarshalText(b []byte) error {
	v, err := ParseClass(string(b))
	if err != nil {
		return err
	}
	*c = v
	return nil
}

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

// MarshalText and UnmarshalText render a Race as its name in JSON and other
// text formats, so saved data stays readable and stable across enum changes.
func (r Race) MarshalText() ([]byte, error) { return []byte(r.String()), nil }

func (r *Race) UnmarshalText(b []byte) error {
	v, err := ParseRace(string(b))
	if err != nil {
		return err
	}
	*r = v
	return nil
}

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

// The ability-effect functions below follow Men & Magic (3LBB). Greyhawk and
// later supplements widen these bands; verify against your table's rules.

// ConHPAdjustment returns the per-hit-die hit-point adjustment for a
// Constitution score.
func ConHPAdjustment(con int) int {
	switch {
	case con >= 15:
		return 1
	case con <= 6:
		return -1
	default:
		return 0
	}
}

// DexMissileAdjustment returns the to-hit adjustment for missile fire from a
// Dexterity score.
func DexMissileAdjustment(dex int) int {
	switch {
	case dex >= 13:
		return 1
	case dex <= 8:
		return -1
	default:
		return 0
	}
}

// MaxHirelings returns the maximum number of hirelings a character with the
// given Charisma may retain.
func MaxHirelings(cha int) int {
	switch {
	case cha <= 3:
		return 1
	case cha <= 6:
		return 2
	case cha <= 9:
		return 3
	case cha <= 12:
		return 4
	case cha <= 15:
		return 5
	case cha <= 17:
		return 6
	default:
		return 12
	}
}

// LoyaltyBase returns the loyalty adjustment for a character's hirelings from
// its Charisma score.
func LoyaltyBase(cha int) int {
	switch {
	case cha <= 3:
		return -2
	case cha <= 6:
		return -1
	case cha <= 12:
		return 0
	case cha <= 15:
		return 1
	case cha <= 17:
		return 2
	default:
		return 4
	}
}

// AdditionalLanguages returns the number of bonus languages known for an
// Intelligence score, one per point above 10.
func AdditionalLanguages(intel int) int {
	if intel > 10 {
		return intel - 10
	}
	return 0
}

// Alignment is a character's allegiance in the 3LBB three-way scheme.
type Alignment int

const (
	Law Alignment = iota
	Neutrality
	Chaos
	numAlignments
)

var alignmentNames = [...]string{"Law", "Neutrality", "Chaos"}

func (a Alignment) String() string { return alignmentNames[a] }

// MarshalText and UnmarshalText render an Alignment as its name.
func (a Alignment) MarshalText() ([]byte, error) { return []byte(a.String()), nil }

func (a *Alignment) UnmarshalText(b []byte) error {
	v, err := ParseAlignment(string(b))
	if err != nil {
		return err
	}
	*a = v
	return nil
}

// ParseAlignment converts a name like "lawful" or "chaos" into an Alignment.
func ParseAlignment(s string) (Alignment, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "law", "lawful", "l":
		return Law, nil
	case "neutrality", "neutral", "n":
		return Neutrality, nil
	case "chaos", "chaotic", "c":
		return Chaos, nil
	}
	return 0, fmt.Errorf("unknown alignment %q (want law, neutrality, or chaos)", s)
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

// Character is a rolled 3LBB character. Race, Class, and Scores are set by
// New; Alignment, HP, and Gold are filled in by the Generator's rolls.
type Character struct {
	Race      Race
	Class     Class
	Scores    Scores
	Alignment Alignment
	HP        int
	Gold      int
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

// ConHPAdjustment returns the character's per-hit-die hit-point adjustment.
func (c Character) ConHPAdjustment() int { return ConHPAdjustment(c.Scores[Constitution]) }

// MissileAdjustment returns the character's to-hit adjustment for missile fire.
func (c Character) MissileAdjustment() int { return DexMissileAdjustment(c.Scores[Dexterity]) }

// MaxHirelings returns the most hirelings the character may retain.
func (c Character) MaxHirelings() int { return MaxHirelings(c.Scores[Charisma]) }

// LoyaltyBase returns the character's hireling loyalty adjustment.
func (c Character) LoyaltyBase() int { return LoyaltyBase(c.Scores[Charisma]) }

// AdditionalLanguages returns the character's number of bonus languages.
func (c Character) AdditionalLanguages() int { return AdditionalLanguages(c.Scores[Intelligence]) }

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

// RollHP rolls a level-1 hit die (d6 in the 3LBB) with the Constitution
// adjustment applied, never less than 1.
func (g *Generator) RollHP(con int) int {
	hp := g.rng.Intn(6) + 1 + ConHPAdjustment(con)
	if hp < 1 {
		return 1
	}
	return hp
}

// RollGold rolls a character's starting gold: 3d6 × 10 gp.
func (g *Generator) RollGold() int {
	return (g.rng.Intn(6) + g.rng.Intn(6) + g.rng.Intn(6) + 3) * 10
}

// RollAlignment rolls one of the three alignments with equal chance.
func (g *Generator) RollAlignment() Alignment {
	return Alignment(g.rng.Intn(int(numAlignments)))
}

// Roll rolls a complete character of the given race, using the class
// recommended by prime requisite.
func (g *Generator) Roll(r Race) Character {
	s := g.RollScores()
	c, _ := New(r, RecommendClass(s, r), s)
	c.Alignment = g.RollAlignment()
	c.HP = g.RollHP(s[Constitution])
	c.Gold = g.RollGold()
	return c
}
