# Keep SSH on v01 and update to the v02.01 printer software binaries

SSH is available by default in v01 firmware, allowing you to connect directly to the printer for debugging and maintenance.

v02.01 removes SSH server files from the system rather than simply disabling the service. Files such as `/usr/sbin/sshd`, `/etc/init.d/sshd`, `/etc/ssh/sshd_config`, and `/usr/lib/sftp-server` are missing, so after a full update you can no longer connect through the original SSH service.

If you have already fully updated to `02.01.00.00` and need to restore SSH, see [Enabling SSH on v02.01 firmware](./enable-ssh-on-v02.md) instead. This guide is for printers that still retain the v01 system.

The approach here is to keep the v01 system and SSH environment, while copying only the v02.01 printer software binaries and their accompanying libraries, configuration, and interface resources onto the printer. This updates the printer software, not the complete firmware to v02.01.

**This guide is based on a rootfs comparison between `01.03.02.51` and `02.01.00.00`. The complete update procedure below has not yet been tested end to end on a printer; other versions require a fresh comparison.**

## Summary of differences

### SSH-related differences

v01 includes the complete OpenSSH server components:

| Path                                 | Description                                           |
| ------------------------------------ | ----------------------------------------------------- |
| `/usr/sbin/sshd`                     | SSH server executable                                 |
| `/etc/init.d/sshd`                   | procd startup script that runs `/usr/sbin/sshd -D`    |
| `/etc/ssh/sshd_config`               | SSH server configuration                              |
| `/usr/lib/sftp-server`               | SFTP subsystem                                        |
| `/lib/upgrade/keep.d/openssh-server` | List of SSH host key files to preserve during updates |
| `sshd:x:22:sshd` in `/etc/group`     | sshd group                                            |

v02.01 still includes the SSH client, `ssh-keygen`, `/etc/ssh/ssh_config`, and the `sshd` user in `/etc/passwd`, but lacks the server files and the `sshd` group.

In other words, v02.01 has not removed the entire SSH runtime environment; it has removed the server components needed to provide remote login.

### Printer startup chain

Both firmware versions use the same printer startup path:

```text
/etc/init.d/printer
  -> /opt/bin/run_printer.sh start
  -> elegoo_printer /opt/inst/printer_dsp.cfg -s /opt/usr/cfg/autosave.cfg -a /tmp/elegoo_uds
```

The comparison shows that `/etc/init.d/printer` and `/opt/bin/run_printer.sh` are identical. Updating the printer software therefore focuses on the binaries, libraries, configuration, and resources under `/opt`, rather than the system init scripts.

### Main changes in v02.01

The differences directly related to the printer software between `01.03.02.51` and `02.01.00.00` are mainly in these locations:

| Path                                         | Change                           |
| -------------------------------------------- | -------------------------------- |
| `/opt/bin/elegoo_printer`                    | Updated main printer executable  |
| `/opt/bin/ai_camera`                         | Updated camera/vision software   |
| `/opt/bin/ec-eeb001-gui`                     | Updated screen GUI               |
| `/opt/inst/daemon-000/daemon-000`            | Updated background program       |
| `/opt/inst/daemon-000/update_whitelist.conf` | Added in v02.01                  |
| `/opt/inst/printer_dsp.cfg`                  | Updated printer configuration    |
| `/opt/lib/libapi_module.a`                   | Updated dependency library       |
| `/opt/lib/libcommunication.a`                | Updated communication library    |
| `/opt/lib/libelegoo_extras.so`               | Updated extension library        |
| `/opt/lib/libturbo_core.so`                  | Updated dependency library       |
| `/opt/inst/images/`                          | Updated some interface resources |
| `/opt/inst/factory/`                         | Changed factory example G-code   |

To reduce the chance of missing files, synchronize the complete `/opt/bin`, `/opt/lib`, and `/opt/inst` directories rather than copying only a few selected binaries.

## Recommended approach

The overall process is:

1. Keep the printer booting the v01 firmware.
2. Log in through the SSH service provided by v01.
3. Package `/opt/bin`, `/opt/lib`, and `/opt/inst` from the v02.01 rootfs.
4. Upload the package to the printer.
5. Stop the printer-related processes.
6. Back up the three existing directories under `/opt`, along with the images and fonts currently in use.
7. Extract the new package into a temporary directory, then synchronize the programs and resources, removing files left over from the old version.
8. Reboot the printer and confirm that SSH still works.

This approach does not replace v01 SSH files such as `/usr/sbin/sshd`, `/etc/init.d/sshd`, or `/etc/ssh/sshd_config`, nor does it replace the entire rootfs.

## Preparing the v02.01 printer software update package on your computer

Obtain the v02.01 firmware for the correct model from an official source and unpack it, preserving file permissions and symbolic links. See [the Linux extraction section of the firmware unpacking guide](./firmware-unpack.md#extracting-rootfs-on-linux-optional) for instructions. Suppose the extracted rootfs is at:

```text
/path/to/cc2-v02.01-rootfs
```

Run the following in **Linux Bash on your computer**. GNU tar and `sha256sum` are required; replace the path with your own extraction directory:

```bash
(
  set -eu
  cd /path/to/cc2-v02.01-rootfs
  sudo tar --numeric-owner -czf /tmp/cc2-v02.01-opt.tar.gz opt/bin opt/lib opt/inst
  cd /tmp
  tar -tzf cc2-v02.01-opt.tar.gz >/dev/null
  sha256sum cc2-v02.01-opt.tar.gz > cc2-v02.01-opt.tar.gz.sha256
)
```

If you do not want to include the factory example G-code, use the following instead to make the package much smaller. The update commands later in this guide will preserve the printer's existing `factory` directory:

```bash
(
  set -eu
  cd /path/to/cc2-v02.01-rootfs
  sudo tar --numeric-owner -czf /tmp/cc2-v02.01-opt.tar.gz \
    --exclude='opt/inst/factory' opt/bin opt/lib opt/inst
  cd /tmp
  tar -tzf cc2-v02.01-opt.tar.gz >/dev/null
  sha256sum cc2-v02.01-opt.tar.gz > cc2-v02.01-opt.tar.gz.sha256
)
```

## Uploading to the v01 printer

Connect to the printer using the [SSH connection guide](./ssh.md) and check its available space. **Replace `PRINTER_IP` with your printer's IP address**:

```sh
ssh root@PRINTER_IP
```

Run these commands in the printer's SSH session:

```sh
df -h / /opt/usr /tmp
du -sh /opt/bin /opt/lib /opt/inst /opt/usr/images /opt/usr/fonts
command -v rsync
command -v sha256sum
```

`/tmp` must have enough room for the uploaded package, `/opt/usr` must accommodate both the extracted files and the backup, and the system partition also needs free space for the update. Estimate based on the unpacked size, not just the compressed package size.

Back in your computer's terminal, upload both the package and its checksum file:

```sh
scp /tmp/cc2-v02.01-opt.tar.gz /tmp/cc2-v02.01-opt.tar.gz.sha256 root@PRINTER_IP:/tmp/
```

## Replacing program files and resources on the printer

Run the following commands **in the printer's SSH session**. They verify the uploaded package, stop the processes, create a backup, and synchronize the files. The printer must be idle; save a separate copy of its configuration and databases on your computer first.

The original stop script misses `daemon-000`, so this procedure stops the relevant processes individually and checks that they have exited. The backup includes the program directories and the images and fonts actually used under `/opt/usr`; it does not include print history, databases, or other user data.

```sh
(
  set -eu
  command -v rsync >/dev/null
  grep -q ' /opt/usr ' /proc/mounts
  cd /tmp
  sha256sum -c cc2-v02.01-opt.tar.gz.sha256

  WORK=$(mktemp -d /opt/usr/cc2-update.XXXXXX)
  trap 'rm -rf "$WORK"' 0
  tar -C "$WORK" -xzpf /tmp/cc2-v02.01-opt.tar.gz
  for dir in bin lib inst; do
    test -d "$WORK/opt/$dir"
    test ! -L "$WORK/opt/$dir"
  done
  test -d "$WORK/opt/inst/images"
  test ! -L "$WORK/opt/inst/images"
  test -d "$WORK/opt/inst/fonts"
  test ! -L "$WORK/opt/inst/fonts"
  for dir in bin lib inst usr/images usr/fonts; do
    test -d "/opt/$dir"
    test ! -L "/opt/$dir"
  done

  PROCESSES="daemon-000 elegoo_printer ai_camera ec-eeb001-gui mosquitto eeb001-factory"
  for name in $PROCESSES; do
    if pidof "$name" >/dev/null; then killall "$name"; fi
  done
  sleep 2
  if pidof $PROCESSES >/dev/null; then
    echo "Printer processes are still running" >&2
    exit 1
  fi

  mkdir -p /opt/usr/backup
  BACKUP=$(mktemp -d /opt/usr/backup/v01-before-v02.01.XXXXXX)
  tar -C / -czf "$BACKUP/opt.tar.gz" opt/bin opt/lib opt/inst opt/usr/images opt/usr/fonts
  tar -tzf "$BACKUP/opt.tar.gz" >/dev/null
  (cd "$BACKUP" && sha256sum opt.tar.gz > opt.tar.gz.sha256)
  printf 'Backup: %s\n' "$BACKUP"

  if pidof $PROCESSES >/dev/null; then
    echo "Printer processes restarted; update stopped" >&2
    exit 1
  fi
  for dir in bin lib; do
    rsync -ac --delete "$WORK/opt/$dir/" "/opt/$dir/"
  done
  if [ -d "$WORK/opt/inst/factory" ]; then
    rsync -ac --delete "$WORK/opt/inst/" /opt/inst/
  else
    rsync -ac --delete --exclude='/factory/' "$WORK/opt/inst/" /opt/inst/
  fi
  rsync -ac --delete /opt/inst/images/ /opt/usr/images/
  rsync -ac --delete /opt/inst/fonts/ /opt/usr/fonts/
  sync
  echo "Update complete"
)
```

This uses `rsync --delete` to remove files left over from the old version instead of extracting over the existing directories. It also removes custom files in the destination directories, which is why the backup comes first. Other directories under `/opt/usr`, including those containing calibration, account information, and print records, are not synchronized.

The firmware only synchronizes images and fonts automatically when `ota_flag` is empty, is set to `true`, or the default theme directory is missing. The commands above synchronize the actual resource directories directly, without relying on that startup condition or changing `ota_flag`.

Record the printed `Backup` path and save a copy on your computer. For example, replace `XXXXXX` in the following path with the actual directory suffix:

```sh
scp -r root@PRINTER_IP:/opt/usr/backup/v01-before-v02.01.XXXXXX ./
```

**Only after you see `Update complete` and have saved the backup**, reboot from the printer's SSH session. If an error occurs during the procedure, follow the [rollback instructions](#rollback) first rather than rebooting immediately:

```sh
reboot
```

## Checks after reboot

After the printer reboots, first confirm that SSH still works:

```sh
ssh root@PRINTER_IP
```

Then check that the printer-related processes have started:

```sh
ps | grep -E 'elegoo_printer|ai_camera|ec-eeb001-gui|daemon-000' | grep -v grep
```

Confirm that the printer process still uses the same UDS path:

```sh
ps | grep elegoo_printer | grep -v grep
ls -l /tmp/elegoo_uds
```

Normally, the `elegoo_printer` arguments should still include:

```text
-a /tmp/elegoo_uds
```

To change the interface theme after the update, see [Setting the emoji theme over SSH](./emoji-theme.md#method-2-run-commands-over-ssh). If you previously changed `ota_version`, see [Update prompts and manually edited version numbers](./disable-ota-update.md#after-the-change) for the effect of synchronizing `/opt/inst`.

## Rollback

If the printer software behaves incorrectly after the update, restore the backup from the printer's SSH session. Set `BACKUP` to the directory you recorded earlier. The commands verify the backup, extract it into an empty directory, and then synchronize it back, so that files added by the newer version are not left behind:

```sh
(
  set -eu
  BACKUP=/opt/usr/backup/v01-before-v02.01.REPLACE_ME
  command -v rsync >/dev/null
  grep -q ' /opt/usr ' /proc/mounts
  (cd "$BACKUP" && sha256sum -c opt.tar.gz.sha256)
  WORK=$(mktemp -d /opt/usr/cc2-rollback.XXXXXX)
  trap 'rm -rf "$WORK"' 0
  tar -C "$WORK" -xzpf "$BACKUP/opt.tar.gz"
  for dir in bin lib inst usr/images usr/fonts; do
    test -d "$WORK/opt/$dir"
    test ! -L "$WORK/opt/$dir"
    test -d "/opt/$dir"
    test ! -L "/opt/$dir"
  done

  PROCESSES="daemon-000 elegoo_printer ai_camera ec-eeb001-gui mosquitto eeb001-factory"
  for name in $PROCESSES; do
    if pidof "$name" >/dev/null; then killall "$name"; fi
  done
  sleep 2
  if pidof $PROCESSES >/dev/null; then
    echo "Printer processes are still running" >&2
    exit 1
  fi
  for dir in bin lib inst usr/images usr/fonts; do
    rsync -ac --delete "$WORK/opt/$dir/" "/opt/$dir/"
  done
  sync
  echo "Rollback complete"
)
```

`Rollback complete` means the programs, images, and fonts have been restored. If the newer version modified databases or configuration, restore the relevant files from the separate user-data backup made before the update, then run `reboot`.

## Notes

1. The key to this approach is keeping the v01 system while replacing only the printer software binaries and their accompanying libraries, configuration, and interface resources with the v02.01 versions.
2. Do not replace the printer's system with the entire v02.01 rootfs, or you will still lose the original SSH server.
3. Avoid overwriting `/etc/passwd`, `/etc/shadow`, `/etc/group`, `/etc/ssh`, `/usr/sbin/sshd`, and `/etc/init.d/sshd`.
4. `/opt/usr/cfg/autosave.cfg` is the user's runtime configuration. It is not part of the default `/opt/inst` package from the v02.01 rootfs and normally does not need to be overwritten.
5. If the printer already contains your calibration, network settings, account information, or print history, do not overwrite `/opt/usr` wholesale.
6. The factory example G-code takes up substantial space and is not required to run the v02.01 printer software, so it can be excluded if desired.
7. This guide describes a maintenance approach based on rootfs differences. Before proceeding, ensure that the printer has a stable power supply, and do not perform these operations while printing.
