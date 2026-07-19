// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/Cryptic-Codex/reeve/campaign"
	"github.com/Cryptic-Codex/reeve/character"
	"github.com/Cryptic-Codex/reeve/monster"
)

// doCampaign handles `campaign [list|new|use|show] ...`.
func doCampaign(out io.Writer, args []string) {
	s, err := campaign.Open()
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	if len(args) == 0 || args[0] == "list" {
		listCampaigns(out, s)
		return
	}

	switch args[0] {
	case "new", "create":
		if len(args) < 2 {
			fmt.Fprintln(out, "usage: campaign new <name>")
			return
		}
		c, err := s.Create(strings.Join(args[1:], " "))
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}
		fmt.Fprintf(out, "created campaign %q (now current)\n", c.Name)

	case "use", "switch":
		if len(args) < 2 {
			fmt.Fprintln(out, "usage: campaign use <name>")
			return
		}
		c, err := s.Load(strings.Join(args[1:], " "))
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}
		if err := s.SetCurrent(c.Name); err != nil {
			fmt.Fprintln(out, err)
			return
		}
		fmt.Fprintf(out, "current campaign is now %q\n", c.Name)

	case "show":
		var c *campaign.Campaign
		if len(args) >= 2 {
			c, err = s.Load(strings.Join(args[1:], " "))
		} else {
			c, err = s.LoadCurrent()
		}
		if err != nil {
			fmt.Fprintln(out, err)
			return
		}
		showCampaign(out, c)

	default:
		fmt.Fprintf(out, "unknown campaign command %q (try list, new, use, show)\n", args[0])
	}
}

func listCampaigns(out io.Writer, s *campaign.Store) {
	names, err := s.List()
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	if len(names) == 0 {
		fmt.Fprintln(out, "no campaigns yet — create one with `campaign new <name>`")
		return
	}
	current, _ := s.Current()
	fmt.Fprintln(out, "campaigns:")
	for _, n := range names {
		mark := " "
		if n == current {
			mark = "*"
		}
		fmt.Fprintf(out, " %s %s\n", mark, n)
	}
}

func showCampaign(out io.Writer, c *campaign.Campaign) {
	fmt.Fprintf(out, "%s  (created %s)\n", c.Name, c.Created.Format("2006-01-02"))

	fmt.Fprintf(out, "\ncharacters (%d):\n", len(c.Characters))
	for _, ch := range c.Characters {
		fmt.Fprintf(out, "  %-22s HP %-2d  %d gp  %s\n", ch, ch.HP, ch.Gold, ch.Alignment)
	}

	fmt.Fprintf(out, "\nmonsters (%d):\n", len(c.Monsters))
	for _, m := range c.Monsters {
		fmt.Fprintf(out, "  %-14s HD %s\n", m.Name, m.HitDice)
	}

	if len(c.Notes) > 0 {
		fmt.Fprintf(out, "\nnotes (%d):\n", len(c.Notes))
		for _, n := range c.Notes {
			fmt.Fprintf(out, "  - %s\n", n)
		}
	}
}

// saveCharacter appends a rolled character to the current campaign.
func saveCharacter(out io.Writer, ch character.Character) {
	s, err := campaign.Open()
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	c, err := s.LoadCurrent()
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	c.AddCharacter(ch)
	if err := s.Save(c); err != nil {
		fmt.Fprintln(out, err)
		return
	}
	fmt.Fprintf(out, "saved to campaign %q\n", c.Name)
}

// saveMonster appends a monster stat block to the current campaign.
func saveMonster(out io.Writer, m monster.Monster) {
	s, err := campaign.Open()
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	c, err := s.LoadCurrent()
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}
	c.AddMonster(m)
	if err := s.Save(c); err != nil {
		fmt.Fprintln(out, err)
		return
	}
	fmt.Fprintf(out, "saved to campaign %q\n", c.Name)
}

// popFlag removes the first occurrence of flag from args, reporting whether it
// was present.
func popFlag(args []string, flag string) ([]string, bool) {
	out := make([]string, 0, len(args))
	found := false
	for _, a := range args {
		if a == flag {
			found = true
			continue
		}
		out = append(out, a)
	}
	return out, found
}
