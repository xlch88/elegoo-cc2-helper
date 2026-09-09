package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall(t *testing.T) {
	for _, scenario := range []string{"install", "conflict", "unmounted", "symlink_directory"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			for _, name := range []string{"etc/init.d", "etc/rc.d", "lib/functions", "proc", "opt/usr"} {
				if err := os.MkdirAll(filepath.Join(root, name), 0755); err != nil {
					t.Fatal(err)
				}
			}
			for name, data := range map[string]string{
				"etc/rc.common":          "original framework",
				"etc/rc.local":           "original local startup",
				"lib/functions/procd.sh": "original procd",
				"proc/mounts":            "/dev/by-name/UDISK /opt/usr ext4 rw 0 0\n",
				"source":                 "test executable",
			} {
				if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0644); err != nil {
					t.Fatal(err)
				}
			}
			script := filepath.Join(root, "etc/init.d/elegoo-cc2-helper")
			directory := filepath.Join(root, "opt/usr/elegoo-cc2-helper")
			switch scenario {
			case "conflict":
				if err := os.WriteFile(script, []byte("existing service"), 0755); err != nil {
					t.Fatal(err)
				}
			case "unmounted":
				if err := os.WriteFile(filepath.Join(root, "proc/mounts"), nil, 0644); err != nil {
					t.Fatal(err)
				}
			case "symlink_directory":
				if err := os.Symlink(filepath.Join(root, "etc"), directory); err != nil {
					t.Fatal(err)
				}
			}
			err := install(root, filepath.Join(root, "source"))
			if scenario != "install" {
				if err == nil {
					t.Fatal("expected installation to fail")
				}
				if _, err := os.Stat(filepath.Join(directory, "elegoo-cc2-helper")); !os.IsNotExist(err) {
					t.Fatal("failed installation left a binary behind", err)
				}
				if scenario == "conflict" {
					data, err := os.ReadFile(script)
					if err != nil || string(data) != "existing service" {
						t.Fatal("existing service was changed", err)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := install(root, filepath.Join(root, "source")); err != nil {
				t.Fatal("repeated installation failed", err)
			}
			binary := filepath.Join(directory, "elegoo-cc2-helper")
			data, err := os.ReadFile(binary)
			if err != nil || string(data) != "test executable" {
				t.Fatal("executable was not copied", err)
			}
			info, err := os.Stat(binary)
			if err != nil || info.Mode().Perm()&0100 == 0 {
				t.Fatal("copied binary is not executable", err)
			}
			for _, link := range []string{"S99elegoo-cc2-helper", "K10elegoo-cc2-helper"} {
				target, err := os.Readlink(filepath.Join(root, "etc/rc.d", link))
				if err != nil || target != "../init.d/elegoo-cc2-helper" {
					t.Fatal("invalid startup link", err)
				}
			}
			data, err = os.ReadFile(script)
			if err != nil || !strings.Contains(string(data), "exec /opt/usr/elegoo-cc2-helper/elegoo-cc2-helper --daemon") {
				t.Fatal("service does not use the installed executable", err)
			}
			if output, err := exec.Command("sh", "-n", script).CombinedOutput(); err != nil {
				t.Fatalf("invalid shell script: %v: %s", err, output)
			}
			for name, expected := range map[string]string{"etc/rc.common": "original framework", "etc/rc.local": "original local startup", "lib/functions/procd.sh": "original procd"} {
				data, err := os.ReadFile(filepath.Join(root, name))
				if err != nil || string(data) != expected {
					t.Fatalf("original file changed: %s: %v", name, err)
				}
			}
		})
	}
}
