package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		directory int
		fail      bool
	}{
		{"install", -1, false},
		{"replace_files", -1, false},
		{"replace_symlinks", -1, false},
		{"replace_directory_symlinks", -1, false},
		{"replace_dangling_symlinks", -1, false},
		{"repair_permissions", -1, false},
		{"unmounted", -1, true},
		{"symlink_directory", -1, true},
		{"binary_directory", 0, true},
		{"script_directory", 1, true},
		{"start_link_directory", 2, true},
		{"stop_link_directory", 3, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := t.TempDir()
			for _, name := range []string{"etc/init.d", "etc/rc.d", "lib/functions", "proc", "opt/usr"} {
				if err := os.MkdirAll(filepath.Join(root, name), 0755); err != nil {
					t.Fatal(err)
				}
			}
			originals := map[string]string{
				"etc/rc.common":          "original framework",
				"etc/rc.local":           "original local startup",
				"lib/functions/procd.sh": "original procd",
				"unrelated":              "unrelated file",
			}
			for name, data := range originals {
				if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0644); err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() {
				for name, expected := range originals {
					data, err := os.ReadFile(filepath.Join(root, name))
					if err != nil || string(data) != expected {
						t.Errorf("original file changed: %s: %v", name, err)
					}
				}
			})
			if err := os.WriteFile(filepath.Join(root, "proc/mounts"), []byte("/dev/by-name/UDISK /opt/usr ext4 rw 0 0\n"), 0644); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(root, "source")
			if err := os.WriteFile(source, []byte("test executable"), 0644); err != nil {
				t.Fatal(err)
			}
			directory := filepath.Join(root, "opt/usr/elegoo-cc2-helper")
			targets := []string{
				filepath.Join(directory, "elegoo-cc2-helper"),
				filepath.Join(root, "etc/init.d/elegoo-cc2-helper"),
				filepath.Join(root, "etc/rc.d/S99elegoo-cc2-helper"),
				filepath.Join(root, "etc/rc.d/K10elegoo-cc2-helper"),
			}
			if scenario.name == "symlink_directory" {
				if err := os.Symlink(filepath.Join(root, "etc"), directory); err != nil {
					t.Fatal(err)
				}
			} else if scenario.name != "unmounted" && scenario.name != "install" && scenario.name != "repair_permissions" {
				if err := os.Mkdir(directory, 0755); err != nil {
					t.Fatal(err)
				}
			}
			switch scenario.name {
			case "replace_files":
				for i, target := range targets {
					if err := os.WriteFile(target, []byte("old file content"), 0644); err != nil {
						t.Fatal(err)
					}
					if err := os.Link(target, filepath.Join(root, fmt.Sprintf("retained-%d", i))); err != nil {
						t.Fatal(err)
					}
				}
			case "replace_symlinks", "replace_directory_symlinks", "replace_dangling_symlinks":
				for _, target := range targets {
					link := filepath.Join(root, "unrelated")
					if scenario.name == "replace_directory_symlinks" {
						link = filepath.Join(root, "etc")
					} else if scenario.name == "replace_dangling_symlinks" {
						link = filepath.Join(root, "missing")
					}
					if err := os.Symlink(link, target); err != nil {
						t.Fatal(err)
					}
				}
			case "repair_permissions":
				if err := install(root, source); err != nil {
					t.Fatal(err)
				}
				for target, mode := range map[string]os.FileMode{targets[0]: 0644, targets[1]: 0700} {
					if err := os.Chmod(target, mode); err != nil {
						t.Fatal(err)
					}
				}
			case "unmounted":
				if err := os.WriteFile(filepath.Join(root, "proc/mounts"), nil, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if scenario.directory >= 0 {
				if err := os.Mkdir(targets[scenario.directory], 0755); err != nil {
					t.Fatal(err)
				}
			}
			err := install(root, source)
			if scenario.fail {
				if err == nil {
					t.Fatal("expected installation to fail")
				}
				if scenario.directory >= 0 {
					info, err := os.Lstat(targets[scenario.directory])
					if err != nil || !info.IsDir() {
						t.Fatal("existing directory was changed", err)
					}
				} else {
					if _, err := os.Lstat(targets[0]); !os.IsNotExist(err) {
						t.Fatal("failed installation left a binary behind", err)
					}
					if scenario.name == "symlink_directory" {
						target, err := os.Readlink(directory)
						if err != nil || target != filepath.Join(root, "etc") {
							t.Fatal("installation directory symlink was changed", err)
						}
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := install(root, source); err != nil {
				t.Fatal("repeated installation failed", err)
			}
			data, err := os.ReadFile(targets[0])
			if err != nil || string(data) != "test executable" {
				t.Fatal("executable was not copied", err)
			}
			for _, target := range targets[:2] {
				info, err := os.Lstat(target)
				if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0755 {
					t.Fatal("installed file must be regular with mode 0755", target, err)
				}
			}
			for _, link := range targets[2:] {
				target, err := os.Readlink(link)
				if err != nil || target != "../init.d/elegoo-cc2-helper" {
					t.Fatal("invalid startup link", err)
				}
			}
			data, err = os.ReadFile(targets[1])
			if err != nil || !strings.Contains(string(data), "exec /opt/usr/elegoo-cc2-helper/elegoo-cc2-helper --daemon") {
				t.Fatal("service does not use the installed executable", err)
			}
			if output, err := exec.Command("sh", "-n", targets[1]).CombinedOutput(); err != nil {
				t.Fatalf("invalid shell script: %v: %s", err, output)
			}
			if scenario.name == "replace_files" {
				for i := range targets {
					data, err := os.ReadFile(filepath.Join(root, fmt.Sprintf("retained-%d", i)))
					if err != nil || string(data) != "old file content" {
						t.Fatal("replacement modified the old inode in place", i, err)
					}
				}
			}
			if scenario.name == "replace_dangling_symlinks" {
				if _, err := os.Lstat(filepath.Join(root, "missing")); !os.IsNotExist(err) {
					t.Fatal("installation followed a dangling symlink", err)
				}
			}
		})
	}
}
