// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package table

import (
	"math/rand"
	"strings"
	"testing"
)

func seeded(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

func TestLoadAndRoll(t *testing.T) {
	const src = `
# a small table
table: color d6
1-2: red
3-4: green
5-6: blue
`
	reg := Registry{}
	if err := Load(reg, strings.NewReader(src)); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := reg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	rng := seeded(1)
	want := map[string]bool{"red": true, "green": true, "blue": true}
	for i := 0; i < 500; i++ {
		res, err := reg.Roll("color", rng)
		if err != nil {
			t.Fatalf("Roll: %v", err)
		}
		if res.Roll < 1 || res.Roll > 6 {
			t.Fatalf("roll out of range: %d", res.Roll)
		}
		if !want[res.Text] {
			t.Fatalf("unexpected result %q", res.Text)
		}
	}
}

func TestValidateTiling(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{"tiles", "table: t d6\n1-3: a\n4-6: b\n", false},
		{"gap", "table: t d6\n1-2: a\n4-6: b\n", true},
		{"overlap", "table: t d6\n1-3: a\n3-6: b\n", true},
		{"short", "table: t d6\n1-3: a\n", true},
		{"over", "table: t d6\n1-3: a\n4-8: b\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := Registry{}
			if err := Load(reg, strings.NewReader(tt.src)); err != nil {
				t.Fatalf("Load: %v", err)
			}
			if err := reg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNestedTables(t *testing.T) {
	const src = `
table: hoard d6
1-3: a few coins
4-6: @gems
table: gems d6
1-3: a dull agate
4-6: a bright ruby
`
	reg := Registry{}
	if err := Load(reg, strings.NewReader(src)); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := reg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	rng := seeded(3)
	sawGem := false
	for i := 0; i < 500; i++ {
		res, err := reg.Roll("hoard", rng)
		if err != nil {
			t.Fatalf("Roll: %v", err)
		}
		if strings.Contains(res.Text, "agate") || strings.Contains(res.Text, "ruby") {
			sawGem = true
		}
	}
	if !sawGem {
		t.Error("nested gems table was never resolved")
	}
}

func TestRollUnknownTable(t *testing.T) {
	reg := Registry{}
	if _, err := reg.Roll("nope", seeded(1)); err == nil {
		t.Error("Roll on unknown table should error")
	}
}

func TestBadReference(t *testing.T) {
	reg := Registry{}
	if err := Load(reg, strings.NewReader("table: t d6\n1-6: @missing\n")); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := reg.Validate(); err == nil {
		t.Error("Validate should reject a dangling reference")
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"1-2: orphan entry with no header\n",
		"table: t\n",            // missing dice
		"table: t d6+1\n",       // modifier not allowed
		"table: t d6\nx-y: a\n", // non-numeric range
	}
	for _, src := range bad {
		if err := Load(Registry{}, strings.NewReader(src)); err == nil {
			t.Errorf("Load(%q) should error", src)
		}
	}
}

func TestBuiltin(t *testing.T) {
	reg, err := Builtin()
	if err != nil {
		t.Fatalf("Builtin: %v", err)
	}
	for _, name := range []string{"reaction", "dressing"} {
		if reg[name] == nil {
			t.Errorf("built-in table %q missing", name)
		}
	}
	rng := seeded(5)
	for _, name := range reg.Names() {
		if _, err := reg.Roll(name, rng); err != nil {
			t.Errorf("Roll(%q): %v", name, err)
		}
	}
}
