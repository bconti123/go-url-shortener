package main

import "testing"

// TestSaveThenGet verifies the basic round-trip: a URL we save under a code
// comes back when we get that code.
func TestSaveThenGet(t *testing.T) {
	s := newStore()

	code := s.save("https://go.dev")

	url, ok := s.get(code)
	if !ok {
		t.Fatalf("get(%q) returned ok=false, want true", code)
	}
	if url != "https://go.dev" {
		t.Errorf("get(%q) = %q, want %q", code, url, "https://go.dev")
	}
}

// TestGetIncrement verifies that each call to get increments the click count for that code.
func TestGetIncrement(t *testing.T) {
	s := newStore()

	code := s.save("https://go.dev")

	// Call get three times to simulate three clicks.
	for range 3 {
		_, ok := s.get(code)
		if !ok {
			t.Fatalf("get(%q) returned ok=false, want true", code)
		}
	}

	count, ok := s.stats(code)
	if !ok {
		t.Fatalf("stats(%q) returned ok=false, want true", code)
	}
	if count != 3 {
		t.Errorf("stats(%q) = %d, want %d", code, count, 3)
	}
}

// TestStatsDoesNotCount verifies that calling stats does not increment the click count.
func TestStatsDoesNotCount(t *testing.T) {
	s := newStore()

	code := s.save("https://go.dev")

	// Call stats three times to check the count without incrementing it.
	for range 3 {
		count, ok := s.stats(code)
		if !ok {
			t.Fatalf("stats(%q) returned ok=false, want true", code)
		}
		if count != 0 {
			t.Errorf("stats(%q) = %d, want %d", code, count, 0)
		}
	}

	// Now call get once and verify that the count increments to 1.
	_, ok := s.get(code)
	if !ok {
		t.Fatalf("get(%q) returned ok=false, want true", code)
	}

	count, ok := s.stats(code)
	if !ok {
		t.Fatalf("stats(%q) returned ok=false, want true", code)
	}
	if count != 1 {
		t.Errorf("stats(%q) = %d, want %d", code, count, 1)
	}
}

// TestGetUnknownCode verifies that getting an unknown code returns ok=false.
func TestGetUnknownCode(t *testing.T) {
	s := newStore()

	_, ok := s.get("unknown")
	if ok {
		t.Errorf("get(%q) returned ok=true, want false", "unknown")
	}
}
