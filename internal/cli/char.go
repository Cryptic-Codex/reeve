// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"io"

	"github.com/Cryptic-Codex/reeve/character"
)

var charGen = character.NewGenerator()

// doChar handles `char [race] [class]` for both the CLI and the menu.
// Race defaults to human; class defaults to the recommendation by prime
// requisite. A lone argument may be either a race or a class, so `char cleric`
// rolls a human cleric.
func doChar(out io.Writer, args []string) {
	args, save := popFlag(args, "--save")

	race := character.Human
	class := character.FightingMan
	haveClass := false

	switch len(args) {
	case 0:
		// all defaults

	case 1:
		if r, rerr := character.ParseRace(args[0]); rerr == nil {
			race = r
		} else if c, cerr := character.ParseClass(args[0]); cerr == nil {
			class, haveClass = c, true
		} else {
			fmt.Fprintln(out, rerr)
			return
		}

	default:
		r, err := character.ParseRace(args[0])
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}
		c, err := character.ParseClass(args[1])
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}
		race, class, haveClass = r, c, true
	}

	scores := charGen.RollScores()
	if !haveClass {
		class = character.RecommendClass(scores, race)
	}

	c, err := character.New(race, class, scores)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	c.Alignment = charGen.RollAlignment()
	c.HP = charGen.RollHP(scores[character.Constitution])
	c.Gold = charGen.RollGold()

	printCharacter(out, c)
	if save {
		saveCharacter(out, c)
	}
}

func printCharacter(out io.Writer, c character.Character) {
	fmt.Fprint(out, c)
	if max := c.MaxLevel(); max > 0 {
		fmt.Fprintf(out, "  (max level %d)", max)
	}
	fmt.Fprintln(out)

	prime := c.Class.PrimeReq()
	for i, score := range c.Scores {
		a := character.Ability(i)
		fmt.Fprintf(out, "  %-13s %2d", a, score)
		if a == prime {
			fmt.Fprintf(out, "   prime requisite, %s", xpNote(c.XPModifier()))
		}
		if note := abilityNote(c, a); note != "" {
			fmt.Fprintf(out, "   %s", note)
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "  %-13s %2d\n", "Hit Points", c.HP)
	fmt.Fprintf(out, "  %-13s %d gp\n", "Gold", c.Gold)
	fmt.Fprintf(out, "  %-13s %s\n", "Alignment", c.Alignment)
}

// abilityNote returns the parenthetical play effect for an ability, or "" when
// the score carries no adjustment worth showing.
func abilityNote(c character.Character, a character.Ability) string {
	switch a {
	case character.Constitution:
		if adj := c.ConHPAdjustment(); adj != 0 {
			return fmt.Sprintf("(%+d hp/die)", adj)
		}
	case character.Dexterity:
		if adj := c.MissileAdjustment(); adj != 0 {
			return fmt.Sprintf("(%+d to missiles)", adj)
		}
	case character.Charisma:
		note := fmt.Sprintf("(%s", plural(c.MaxHirelings(), "hireling"))
		if loy := c.LoyaltyBase(); loy != 0 {
			note += fmt.Sprintf(", %+d loyalty", loy)
		}
		return note + ")"
	case character.Intelligence:
		if n := c.AdditionalLanguages(); n > 0 {
			return fmt.Sprintf("(+%s)", plural(n, "language"))
		}
	}
	return ""
}

// plural formats a count with its noun, adding "s" for anything but one.
func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// xpNote renders an experience adjustment percent as a short phrase.
func xpNote(mod int) string {
	switch {
	case mod > 0:
		return fmt.Sprintf("+%d%% experience", mod)
	case mod < 0:
		return fmt.Sprintf("%d%% experience", mod)
	default:
		return "no experience adjustment"
	}
}
