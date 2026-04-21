package provider

import "testing"

func TestParseFlexibleTimeString(t *testing.T) {
	const (
		expectedNoTZParsed = "2024-01-02T03:04:05Z"
		expectedTZParsed   = "2024-01-02T01:04:05Z"
	)

	t.Run("parses timestamp without timezone", func(t *testing.T) {
		got, err := parseFlexibleTimeString("2024-01-02T03:04:05")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got.Format("2006-01-02T15:04:05Z07:00") != expectedNoTZParsed {
			t.Fatalf("unexpected parsed time: %s", got.Format("2006-01-02T15:04:05Z07:00"))
		}
	})

	t.Run("parses RFC3339 timestamp with timezone", func(t *testing.T) {
		got, err := parseFlexibleTimeString("2024-01-02T03:04:05+02:00")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got.UTC().Format("2006-01-02T15:04:05Z07:00") != expectedTZParsed {
			t.Fatalf("unexpected parsed time: %s", got.UTC().Format("2006-01-02T15:04:05Z07:00"))
		}
	})

	t.Run("returns error for invalid timestamp", func(t *testing.T) {
		_, err := parseFlexibleTimeString("not-a-time")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParseTimeString(t *testing.T) {
	const expectedNormalizedTime = "2024-01-02T01:04:05Z"

	t.Run("normalizes timestamp to UTC RFC3339", func(t *testing.T) {
		got, err := parseTimeString("2024-01-02T03:04:05+02:00")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != expectedNormalizedTime {
			t.Fatalf("unexpected normalized time: %s", got)
		}
	})

	t.Run("preserves configured value when time matches exactly", func(t *testing.T) {
		preserved := "2024-01-02T03:04:05+02:00"
		got, err := parseTimeString("2024-01-02T01:04:05Z", preserved)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != preserved {
			t.Fatalf("expected preserved value %q, got %q", preserved, got)
		}
	})

	t.Run("ignores invalid preserve value and returns normalized time", func(t *testing.T) {
		got, err := parseTimeString("2024-01-02T03:04:05", "invalid")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != "2024-01-02T03:04:05Z" {
			t.Fatalf("unexpected normalized time: %s", got)
		}
	})

	t.Run("returns error for invalid timestamp", func(t *testing.T) {
		_, err := parseTimeString("not-a-time")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParseTimeStringWithTZ(t *testing.T) {
	t.Run("normalizes timestamp to explicit UTC offset", func(t *testing.T) {
		got, err := parseTimeStringWithTZ("2024-01-02T03:04:05+02:00")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != "2024-01-02T01:04:05+00:00" {
			t.Fatalf("unexpected normalized time: %s", got)
		}
	})

	t.Run("returns error for invalid timestamp", func(t *testing.T) {
		_, err := parseTimeStringWithTZ("not-a-time")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
