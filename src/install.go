package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func install(root, executable string) (err error) {
	for _, name := range []string{"etc/rc.common", "lib/functions/procd.sh"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			return fmt.Errorf("OpenWrt/procd is required: %w", err)
		}
	}
	mounts, err := os.ReadFile(filepath.Join(root, "proc/mounts"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(mounts), " /opt/usr ") {
		return fmt.Errorf("persistent storage /opt/usr is not mounted")
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		return err
	}
	const script = `#!/bin/sh /etc/rc.common
START=99
STOP=10
USE_PROCD=1

start_service() {
    procd_open_instance
    procd_set_param command /bin/sh -c 'while [ ! -x /opt/usr/elegoo-cc2-helper/elegoo-cc2-helper ] || [ ! -S /tmp/elegoo_uds ] || [ ! -w /sys/class/pwm/pwmchip0/pwm0/enable ] || [ ! -w /sys/class/pwm/pwmchip0/pwm0/period ] || [ ! -w /sys/class/pwm/pwmchip0/pwm0/duty_cycle ]; do sleep 1; done; exec /opt/usr/elegoo-cc2-helper/elegoo-cc2-helper --daemon'
    procd_set_param respawn 3600 5 5
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_close_instance
}
`
	var created []string
	defer func() {
		if err != nil {
			for i := len(created) - 1; i >= 0; i-- {
				if cleanupErr := os.Remove(created[i]); cleanupErr != nil {
					fmt.Fprintf(os.Stderr, "Cleanup failed for %s: %v\n", created[i], cleanupErr)
				}
			}
		}
	}()
	directory := filepath.Join(root, "opt/usr/elegoo-cc2-helper")
	if err = os.Mkdir(directory, 0755); err == nil {
		created = append(created, directory)
	} else if !os.IsExist(err) {
		return err
	} else if info, statErr := os.Lstat(directory); statErr != nil || !info.IsDir() {
		return fmt.Errorf("installation directory must be a real directory: %s", directory)
	}
	for _, item := range []struct {
		path   string
		data   []byte
		target string
	}{
		{"opt/usr/elegoo-cc2-helper/elegoo-cc2-helper", binary, ""},
		{"etc/init.d/elegoo-cc2-helper", []byte(script), ""},
		{"etc/rc.d/S99elegoo-cc2-helper", nil, "../init.d/elegoo-cc2-helper"},
		{"etc/rc.d/K10elegoo-cc2-helper", nil, "../init.d/elegoo-cc2-helper"},
	} {
		path := filepath.Join(root, item.path)
		info, statErr := os.Lstat(path)
		if statErr == nil {
			if item.target != "" {
				if target, readErr := os.Readlink(path); readErr == nil && target == item.target {
					continue
				}
			} else if info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
				if data, readErr := os.ReadFile(path); readErr == nil && bytes.Equal(data, item.data) {
					continue
				}
			}
			return fmt.Errorf("refusing to overwrite existing path: %s", path)
		}
		if !os.IsNotExist(statErr) {
			return statErr
		}
		if item.target != "" {
			if err = os.Symlink(item.target, path); err != nil {
				return err
			}
			created = append(created, path)
			continue
		}
		file, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
		if openErr != nil {
			return openErr
		}
		created = append(created, path)
		_, err = file.Write(item.data)
		closeErr := file.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
