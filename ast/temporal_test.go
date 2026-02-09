// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ast

import (
	"testing"
	"time"
)

func TestTemporalBound(t *testing.T) {
	tests := []struct {
		name     string
		bound    TemporalBound
		wantStr  string
		wantType TemporalBoundType
	}{
		{
			name:     "timestamp bound",
			bound:    NewTimestampBound(time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)),
			wantStr:  "2024-03-15T10:30:00Z",
			wantType: TimestampBound,
		},
		{
			name:     "variable bound",
			bound:    NewVariableBound(Variable{"T"}),
			wantStr:  "T",
			wantType: VariableBound,
		},
		{
			name:     "negative infinity",
			bound:    NegativeInfinity(),
			wantStr:  "_",
			wantType: NegativeInfinityBound,
		},
		{
			name:     "positive infinity",
			bound:    PositiveInfinity(),
			wantStr:  "_",
			wantType: PositiveInfinityBound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bound.String(); got != tt.wantStr {
				t.Errorf("TemporalBound.String() = %v, want %v", got, tt.wantStr)
			}
			if tt.bound.Type != tt.wantType {
				t.Errorf("TemporalBound.Type = %v, want %v", tt.bound.Type, tt.wantType)
			}
		})
	}
}

func TestTemporalBoundEquals(t *testing.T) {
	t1 := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	t2 := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	t3 := time.Date(2024, 3, 16, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		a    TemporalBound
		b    TemporalBound
		want bool
	}{
		{
			name: "same timestamp",
			a:    NewTimestampBound(t1),
			b:    NewTimestampBound(t2),
			want: true,
		},
		{
			name: "different timestamps",
			a:    NewTimestampBound(t1),
			b:    NewTimestampBound(t3),
			want: false,
		},
		{
			name: "same variable",
			a:    NewVariableBound(Variable{"T"}),
			b:    NewVariableBound(Variable{"T"}),
			want: true,
		},
		{
			name: "different variables",
			a:    NewVariableBound(Variable{"T1"}),
			b:    NewVariableBound(Variable{"T2"}),
			want: false,
		},
		{
			name: "both negative infinity",
			a:    NegativeInfinity(),
			b:    NegativeInfinity(),
			want: true,
		},
		{
			name: "both positive infinity",
			a:    PositiveInfinity(),
			b:    PositiveInfinity(),
			want: true,
		},
		{
			name: "negative vs positive infinity",
			a:    NegativeInfinity(),
			b:    PositiveInfinity(),
			want: false,
		},
		{
			name: "timestamp vs variable",
			a:    NewTimestampBound(t1),
			b:    NewVariableBound(Variable{"T"}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equals(tt.b); got != tt.want {
				t.Errorf("TemporalBound.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterval(t *testing.T) {
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		interval  Interval
		wantStr   string
		isEternal bool
		isPoint   bool
	}{
		{
			name:      "concrete interval",
			interval:  NewInterval(NewTimestampBound(t1), NewTimestampBound(t2)),
			wantStr:   "@[2020-01-01T00:00:00Z, 2023-06-15T00:00:00Z]",
			isEternal: false,
			isPoint:   false,
		},
		{
			name:      "eternal interval",
			interval:  EternalInterval(),
			wantStr:   "",
			isEternal: true,
			isPoint:   false,
		},
		{
			name:      "point interval",
			interval:  NewPointInterval(t1),
			wantStr:   "@[2020-01-01T00:00:00Z]",
			isEternal: false,
			isPoint:   true,
		},
		{
			name:      "half-open interval (unbounded end)",
			interval:  NewInterval(NewTimestampBound(t1), PositiveInfinity()),
			wantStr:   "@[2020-01-01T00:00:00Z, _]",
			isEternal: false,
			isPoint:   false,
		},
		{
			name:      "interval with variable",
			interval:  NewInterval(NewTimestampBound(t1), NewVariableBound(Variable{"T"})),
			wantStr:   "@[2020-01-01T00:00:00Z, T]",
			isEternal: false,
			isPoint:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.String(); got != tt.wantStr {
				t.Errorf("Interval.String() = %v, want %v", got, tt.wantStr)
			}
			if got := tt.interval.IsEternal(); got != tt.isEternal {
				t.Errorf("Interval.IsEternal() = %v, want %v", got, tt.isEternal)
			}
			if got := tt.interval.IsPoint(); got != tt.isPoint {
				t.Errorf("Interval.IsPoint() = %v, want %v", got, tt.isPoint)
			}
		})
	}
}

func TestIntervalContains(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	interval := NewInterval(NewTimestampBound(start), NewTimestampBound(end))

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{
			name: "time within interval",
			t:    time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "time at start",
			t:    start,
			want: true,
		},
		{
			name: "time at end",
			t:    end,
			want: true,
		},
		{
			name: "time before interval",
			t:    time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "time after interval",
			t:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interval.Contains(tt.t); got != tt.want {
				t.Errorf("Interval.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntervalContainsWithUnbounded(t *testing.T) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	// Interval from 2020-01-01 to +infinity
	openEnd := NewInterval(NewTimestampBound(start), PositiveInfinity())

	// Interval from -infinity to 2020-01-01
	openStart := NewInterval(NegativeInfinity(), NewTimestampBound(start))

	// Eternal interval
	eternal := EternalInterval()

	testTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	pastTime := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)

	if !openEnd.Contains(testTime) {
		t.Error("Open-end interval should contain future time")
	}
	if openEnd.Contains(pastTime) {
		t.Error("Open-end interval should not contain time before start")
	}

	if openStart.Contains(testTime) {
		t.Error("Open-start interval should not contain time after end")
	}
	if !openStart.Contains(pastTime) {
		t.Error("Open-start interval should contain past time")
	}

	if !eternal.Contains(testTime) || !eternal.Contains(pastTime) {
		t.Error("Eternal interval should contain all times")
	}
}

func TestIntervalOverlaps(t *testing.T) {
	// Interval A: 2020-01-01 to 2022-12-31
	a := NewInterval(
		NewTimestampBound(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
		NewTimestampBound(time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)),
	)

	tests := []struct {
		name string
		b    Interval
		want bool
	}{
		{
			name: "overlapping interval",
			b: NewInterval(
				NewTimestampBound(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
				NewTimestampBound(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)),
			),
			want: true,
		},
		{
			name: "contained interval",
			b: NewInterval(
				NewTimestampBound(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
				NewTimestampBound(time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC)),
			),
			want: true,
		},
		{
			name: "before interval",
			b: NewInterval(
				NewTimestampBound(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)),
				NewTimestampBound(time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)),
			),
			want: false,
		},
		{
			name: "after interval",
			b: NewInterval(
				NewTimestampBound(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				NewTimestampBound(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
			),
			want: false,
		},
		{
			name: "touching at end",
			b: NewInterval(
				NewTimestampBound(time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)),
				NewTimestampBound(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)),
			),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := a.Overlaps(tt.b); got != tt.want {
				t.Errorf("Interval.Overlaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemporalAtom(t *testing.T) {
	alice, _ := Name("/alice")
	engineering, _ := Name("/engineering")
	atom := NewAtom("team_member", alice, engineering)

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	interval := NewInterval(NewTimestampBound(start), NewTimestampBound(end))

	tests := []struct {
		name    string
		ta      TemporalAtom
		wantStr string
	}{
		{
			name:    "atom without temporal annotation",
			ta:      TemporalAtom{Atom: atom, Interval: nil},
			wantStr: "team_member(/alice,/engineering)",
		},
		{
			name:    "atom with temporal annotation",
			ta:      TemporalAtom{Atom: atom, Interval: &interval},
			wantStr: "team_member(/alice,/engineering)@[2020-01-01T00:00:00Z, 2023-06-15T00:00:00Z]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ta.String(); got != tt.wantStr {
				t.Errorf("TemporalAtom.String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func TestTemporalOperator(t *testing.T) {
	interval := NewInterval(
		NewTimestampBound(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		NewTimestampBound(time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)),
	)

	tests := []struct {
		name    string
		op      TemporalOperator
		wantStr string
	}{
		{
			name:    "diamond minus",
			op:      TemporalOperator{Type: DiamondMinus, Interval: interval},
			wantStr: "<-[2024-01-01T00:00:00Z, 2024-01-08T00:00:00Z]",
		},
		{
			name:    "box minus",
			op:      TemporalOperator{Type: BoxMinus, Interval: interval},
			wantStr: "[-[2024-01-01T00:00:00Z, 2024-01-08T00:00:00Z]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.String(); got != tt.wantStr {
				t.Errorf("TemporalOperator.String() = %v, want %v", got, tt.wantStr)
			}
		})
	}
}

func TestDateHelpers(t *testing.T) {
	// Ensure we start with UTC
	SetDefaultTimezone(time.UTC)
	defer SetDefaultTimezone(time.UTC) // Reset after test

	t.Run("Date uses default timezone", func(t *testing.T) {
		d := Date(2024, 1, 15)
		if d.Location() != time.UTC {
			t.Errorf("Date() location = %v, want UTC", d.Location())
		}
		if d.Year() != 2024 || d.Month() != 1 || d.Day() != 15 {
			t.Errorf("Date() = %v, want 2024-01-15", d)
		}
		if d.Hour() != 0 || d.Minute() != 0 || d.Second() != 0 {
			t.Errorf("Date() should be midnight, got %v", d)
		}
	})

	t.Run("DateTime uses default timezone", func(t *testing.T) {
		d := DateTime(2024, 1, 15, 10, 30)
		if d.Hour() != 10 || d.Minute() != 30 {
			t.Errorf("DateTime() = %v, want 10:30", d)
		}
	})

	t.Run("DateTimeSec uses default timezone", func(t *testing.T) {
		d := DateTimeSec(2024, 1, 15, 10, 30, 45)
		if d.Second() != 45 {
			t.Errorf("DateTimeSec() second = %d, want 45", d.Second())
		}
	})

	t.Run("SetTimezone with string changes behavior", func(t *testing.T) {
		err := SetTimezone("America/New_York")
		if err != nil {
			t.Skip("Could not load America/New_York timezone")
		}

		d := Date(2024, 1, 15)
		if d.Location().String() != "America/New_York" {
			t.Errorf("After SetTimezone(America/New_York), Date() location = %v, want America/New_York", d.Location())
		}

		// Reset to UTC
		SetTimezone("UTC")
		d = Date(2024, 1, 15)
		if d.Location() != time.UTC {
			t.Errorf("After reset, Date() location = %v, want UTC", d.Location())
		}
	})

	t.Run("SetTimezone with abbreviations", func(t *testing.T) {
		err := SetTimezone("PST")
		if err != nil {
			t.Skip("Could not load PST timezone")
		}
		if GetDefaultTimezone().String() != "America/Los_Angeles" {
			t.Errorf("SetTimezone(PST) should map to America/Los_Angeles, got %v", GetDefaultTimezone())
		}
		SetTimezone("UTC")
	})

	t.Run("SetTimezone with Local", func(t *testing.T) {
		err := SetTimezone("Local")
		if err != nil {
			t.Fatalf("SetTimezone(Local) failed: %v", err)
		}
		if GetDefaultTimezone() != time.Local {
			t.Errorf("SetTimezone(Local) should use time.Local")
		}
		SetTimezone("UTC")
	})

	t.Run("DateIn with string timezone", func(t *testing.T) {
		SetDefaultTimezone(time.UTC)

		d := DateIn(2024, 1, 15, "America/New_York")
		if d.Location().String() != "America/New_York" {
			t.Errorf("DateIn(..., America/New_York) location = %v, want America/New_York", d.Location())
		}

		// Default should still be UTC
		d2 := Date(2024, 1, 15)
		if d2.Location() != time.UTC {
			t.Errorf("Date() after DateIn() should still use UTC, got %v", d2.Location())
		}
	})

	t.Run("DateIn with abbreviation", func(t *testing.T) {
		d := DateIn(2024, 1, 15, "PST")
		if d.Location().String() != "America/Los_Angeles" {
			t.Errorf("DateIn(..., PST) location = %v, want America/Los_Angeles", d.Location())
		}
	})

	t.Run("GetDefaultTimezone returns current setting", func(t *testing.T) {
		SetDefaultTimezone(time.UTC)
		if got := GetDefaultTimezone(); got != time.UTC {
			t.Errorf("GetDefaultTimezone() = %v, want UTC", got)
		}
	})

	t.Run("SetDefaultTimezone with nil defaults to UTC", func(t *testing.T) {
		SetDefaultTimezone(nil)
		if got := GetDefaultTimezone(); got != time.UTC {
			t.Errorf("SetDefaultTimezone(nil) should default to UTC, got %v", got)
		}
	})

	t.Run("MustSetTimezone panics on invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustSetTimezone with invalid tz should panic")
			}
			SetTimezone("UTC") // Reset
		}()
		MustSetTimezone("Invalid/Timezone")
	})
}

func TestDateInterval(t *testing.T) {
	// Explicitly set UTC to avoid test pollution
	SetDefaultTimezone(time.UTC)
	defer SetDefaultTimezone(time.UTC)

	interval := DateInterval(2023, 1, 1, 2024, 12, 31)

	startTime := interval.Start.Time().UTC()
	endTime := interval.End.Time().UTC()

	if startTime.Year() != 2023 || startTime.Month() != 1 || startTime.Day() != 1 {
		t.Errorf("DateInterval start = %v, want 2023-01-01", startTime)
	}
	if endTime.Year() != 2024 || endTime.Month() != 12 || endTime.Day() != 31 {
		t.Errorf("DateInterval end = %v, want 2024-12-31", endTime)
	}
}

func TestTimeInterval(t *testing.T) {
	start := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)

	interval := TimeInterval(start, end)

	if !interval.Start.Time().Equal(start) {
		t.Errorf("TimeInterval start = %v, want %v", interval.Start.Time(), start)
	}
	if !interval.End.Time().Equal(end) {
		t.Errorf("TimeInterval end = %v, want %v", interval.End.Time(), end)
	}
}
