package main

import "testing"

func TestFurElise(t *testing.T) {
	if len(furElise) != 40 {
		t.Fatalf("unexpected score length: %d", len(furElise))
	}
	for i, frequency := range []int{1319, 1245, 1319, 1245, 1319, 988, 1175, 1047, 880} {
		if furElise[i][0] != frequency {
			t.Fatalf("incorrect opening note %d", i)
		}
	}
	duration, highest := 0, 0
	for i, note := range furElise {
		if note[0] < 0 || note[1] <= 0 || note[0] > 0 && note[1] <= 10 {
			t.Fatalf("invalid note or articulation duration at %d: %v", i, note)
		}
		duration += note[1]
		if note[0] > highest {
			highest = note[0]
		}
	}
	if duration != 10000 {
		t.Fatalf("unexpected theme duration: %d ms", duration)
	}
	if highest != 1319 || furElise[len(furElise)-1][0] != 880 {
		t.Fatal("incorrect opening-theme range or final A")
	}
}
