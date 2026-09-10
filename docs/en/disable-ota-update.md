# Disable the firmware update prompt at startup

If you do not want to be prompted to update every time the printer starts, you can change `ota_version` in `/opt/inst/firmware_version/versions.json` to a higher version number, such as `04.95.04.95`.

This changes the version number used by the system display and update checks. **It does not actually upgrade the firmware or disable background update requests.**

## Version number format

**Use the `xx.xx.xx.xx` format: four groups of two digits, separated by periods.** You can refer to the example version as v04.95.04.95, but the value written to JSON must be `04.95.04.95`, without the `v`.

The field to change is shown below. Do not overwrite the entire JSON file with this snippet:

```text
"ota_version": "04.95.04.95"
```

Do not use `v04.95.04.95`, `4.95.4.95`, or another format. Doing so can cause repeated update check failures and crash the daemon process. Change only `ota_version`, not `os_version`, `elegoo_version`, `canvas_version`, or any other fields.

## Steps

SSH must already be enabled on the printer; see [SSH connection and default password](ssh.md) for connection instructions. If SSH is not yet enabled on v02.01, start with the [SSH recovery guide](enable-ssh-on-v02.md). Run the following in your computer's terminal or PowerShell. **Replace `PRINTER_IP` with your printer's actual IP address:**

```sh
ssh root@PRINTER_IP
```

Make sure the printer is idle and is not updating its firmware, then paste the following into the **printer's SSH terminal**:

```sh
(
    set -eu
    cd /opt/inst/firmware_version
    if [ ! -e versions.json.before-ota ]; then
        cp -p versions.json versions.json.before-ota
    fi
    sed -i 's/"ota_version"[[:space:]]*:[[:space:]]*"[^"]*"/"ota_version": "04.95.04.95"/' versions.json
    grep -E '"ota_version"[[:space:]]*:[[:space:]]*"04\.95\.04\.95"' versions.json
    sync
)
```

These commands change only the value of `ota_version` and leave the other fields untouched. Before the first change, they save a backup as `versions.json.before-ota`; running them again will not overwrite that backup. Note the original version number in the backup, as you will need it if you want to restore it later.

Check that there are no errors and the output shows `"ota_version": "04.95.04.95"`, then reboot the printer from the same SSH terminal:

```sh
reboot
```

## After the change

After rebooting, open **Settings → System Version** on the printer. The current version is shown as `04.95.04.95`, with a message below indicating that no newer version was found. The screenshot below shows the actual interface after the change:

![After changing ota_version, the system shows 04.95.04.95 and reports that no newer version was found](../imgs/ota-version-04.95.04.95.png)

This version number is now a manually entered value, so do not use it to identify the actual installed firmware. Flashing firmware or [copying `/opt/inst` again](keep-ssh-update.md#replacing-program-files-and-resources-on-the-printer) may also overwrite this change.

## Restore the update prompt

If you have not updated the firmware or replaced the version file since making this change, restore the original file in the printer's SSH terminal:

```sh
cd /opt/inst/firmware_version &&
    cp -p versions.json.before-ota versions.json &&
    sync
```

After the command succeeds, run `reboot`. If you have since updated the printer software binaries and accompanying files, do not restore the entire old backup. Instead, change only `ota_version` back to the version recorded in the firmware package that is actually installed. To inspect `versions.json` in the original package, follow [Opening rootfs with 7-Zip](firmware-unpack.md#opening-rootfs-with-7-zip-windows).
