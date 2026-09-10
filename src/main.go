package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Override at build time with -ldflags "-X main.version=1.2.3".
var version = "0.1.0"

func main() {
	border := strings.Repeat("-", 64)
	fmt.Printf("+%s+\n|  %-60s  |\n|  %-60s  |\n|  %-60s  |\n|  %-60s  |\n|  %-60s  |\n+%s+\n\n",
		border,
		"ELEGOO CC2 HELPER",
		"Printer notifications & startup service",
		"",
		"Version: "+version+"   /   Author: Dark495",
		"GitHub: https://github.com/xlch88/elegoo-cc2-helper",
		border)
	const usage = `Usage: elegoo-cc2-helper [option]

Options:
  -h, --help    Show this help (also shown with no arguments).
  -d, --daemon  Run in service mode: play the startup sound and monitor /tmp/elegoo_uds.
  -t, --test    Play all six sounds once, print each sound name, then exit.
  --test-music  Play the opening theme of Fur Elise once (about 10 seconds).
  --install    Install or overwrite autostart files (root required); do not start the service.

Press Ctrl+C to stop playback or service mode.
`
	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if len(os.Args) > 2 {
		fmt.Fprint(os.Stderr, "Expected at most one option.\n\n"+usage)
		os.Exit(2)
	}
	switch mode {
	case "", "-h", "--help":
		fmt.Print(usage)
		return
	case "-d", "--daemon", "-t", "--test", "--test-music":
	case "--install":
		if os.Geteuid() != 0 {
			fmt.Fprintln(os.Stderr, "Installation requires root.")
			os.Exit(1)
		}
		executable, err := os.Executable()
		if err == nil {
			err = install("/", executable)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "Installation failed:", err)
			os.Exit(1)
		}
		fmt.Println("Autostart registered. The service has not been started.")
		fmt.Println("Service: /etc/init.d/elegoo-cc2-helper")
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown option: %s\n\n%s", mode, usage)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if mode == "--test-music" {
		fmt.Println("Playing: Fur Elise - Ludwig van Beethoven (opening theme, monophonic)")
		playNotes(ctx, furElise[:], 10)
		return
	}
	if mode == "-t" || mode == "--test" {
		for i, sound := range []string{"startup", "printing", "paused", "resumed", "cancelled", "complete"} {
			if i > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
			}
			if ctx.Err() != nil {
				return
			}
			play(ctx, sound)
		}
		return
	}
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "unix", "/tmp/elegoo_uds")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	defer pwm("enable", 0)
	cancelClose := context.AfterFunc(ctx, func() { conn.Close() })
	defer cancelClose()
	fmt.Println("Service mode: monitoring /tmp/elegoo_uds")
	play(ctx, "startup")
	if err := listen(conn, func(state string) { play(ctx, state) }); err != nil && ctx.Err() == nil {
		panic(err)
	}
}
