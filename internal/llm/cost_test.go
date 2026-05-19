package llm

import "testing"

func TestCalculateCostUsesConfiguredModel(t *testing.T) {
	got := CalculateCost("gpt-4o-mini", 1_000_000, 1_000_000)
	want := 0.75
	if got != want {
		t.Fatalf("CalculateCost() = %v, want %v", got, want)
	}
}

func TestCalculateCostUsesLongestPrefix(t *testing.T) {
	got := CalculateCost("gpt-4o-mini-2024-07-18", 1_000_000, 1_000_000)
	want := 0.75
	if got != want {
		t.Fatalf("CalculateCost() = %v, want %v", got, want)
	}
}

func TestCalculateCostUnknownModel(t *testing.T) {
	if got := CalculateCost("unknown-model", 1_000_000, 1_000_000); got != 0 {
		t.Fatalf("CalculateCost() = %v, want 0", got)
	}
}
