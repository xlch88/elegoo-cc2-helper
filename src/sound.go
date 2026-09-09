package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
)

func pwm(name string, value int) {
	if err := os.WriteFile("/sys/class/pwm/pwmchip0/pwm0/"+name, []byte(strconv.Itoa(value)), 0); err != nil {
		panic(err)
	}
}

func play(ctx context.Context, state string) {
	fmt.Println("Playing:", state)
	var notes [][2]int // Hz, milliseconds; 0 Hz = rest
	gap := 30
	switch state {
	case "startup":
		notes = [][2]int{{2637, 160}, {2960, 160}, {3520, 240}, {0, 80}, {2960, 180}, {3520, 400}}
	case "printing":
		gap = 15
		notes = [][2]int{{2349, 110}, {2960, 110}, {3520, 220}, {0, 40}, {2349, 110}, {2960, 110}, {3520, 350}}
	case "paused":
		notes = [][2]int{{2960, 85}, {2960, 85}, {2960, 85}, {0, 180}, {2960, 85}, {2960, 85}, {2960, 140}}
	case "resumed":
		notes = [][2]int{{2349, 120}, {2960, 120}, {3520, 300}}
	case "cancelled":
		notes = [][2]int{{3520, 160}, {2960, 160}, {2349, 400}}
	case "complete":
		notes = [][2]int{{3520, 140}, {0, 80}, {3520, 140}, {0, 100}, {2637, 180}, {2960, 180}, {3520, 650}}
	}
	playNotes(ctx, notes, gap)
}

func playNotes(ctx context.Context, notes [][2]int, gap int) {
	defer pwm("enable", 0)
	for _, note := range notes {
		if ctx.Err() != nil {
			return
		}
		pwm("enable", 0)
		if note[0] != 0 {
			period := 1_000_000_000 / note[0]
			pwm("duty_cycle", 0)
			pwm("period", period)
			pwm("duty_cycle", period/2)
			pwm("enable", 1)
			note[1] -= gap
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(note[1]) * time.Millisecond):
		}
		pwm("enable", 0)
		if note[0] != 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(gap) * time.Millisecond):
			}
		}
	}
}
