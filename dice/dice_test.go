// Copyright (C) 2026 Cryptic Codex LLC
// SPDX-License-Identifier: GPL-3.0-or-later
package dice

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Roll
		wantErr bool
	}{
		// the cases from our discussion
		{"2d6 + 6", Roll{2, 6, 6}, false},
		{"d6", Roll{1, 6, 0}, false},

		// basics
		{"3d6", Roll{3, 6, 0}, false},
		{"1d20", Roll{1, 20, 0}, false},
		{"d100", Roll{1, 100, 0}, false},

		// modifiers and whitespace
		{"3d6+1", Roll{3, 6, 1}, false},
		{"3d6-2", Roll{3, 6, -2}, false},
		{"3d6 - 2", Roll{3, 6, -2}, false},
		{"  2d8+3  ", Roll{2, 8, 3}, false},
		{"D20", Roll{1, 20, 0}, false}, // case-insensitive

		// garbage should error loudly
		{"", Roll{}, true},
		{"banana", Roll{}, true},
		{"2d", Roll{}, true},
		{"2dd6", Roll{}, true},
		{"2d6+6+3", Roll{}, true},
		{"2d6+", Roll{}, true},
		{"-2d6", Roll{}, true},
		{"0d6", Roll{}, true},    // count out of range
		{"2000d6", Roll{}, true}, // count out of range
		{"2d0", Roll{}, true},    // sides out of range
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

// Rolling is random, so we test properties, not exact values.
func TestRoll(t *testing.T) {
	r := Roll{Count: 3, Sides: 6, Modifier: 2}

	for i := 0; i < 1000; i++ {
		res := r.Roll()

		if len(res.Rolls) != 3 {
			t.Fatalf("expected 3 rolls, got %d", len(res.Rolls))
		}

		sum := 0
		for _, d := range res.Rolls {
			if d < 1 || d > 6 {
				t.Fatalf("die out of range [1,6]: %d", d)
			}
			sum += d
		}

		if res.Total != sum+2 {
			t.Fatalf("total %d != sum of rolls %d + modifier 2", res.Total, sum)
		}
		if res.Total < 5 || res.Total > 20 {
			t.Fatalf("3d6+2 total out of range [5,20]: %d", res.Total)
		}
	}
}

func TestRollD1(t *testing.T) {
	// degenerate but legal: d1 always rolls 1
	res := Roll{Count: 5, Sides: 1}.Roll()
	if res.Total != 5 {
		t.Errorf("5d1 should total 5, got %d", res.Total)
	}
}
