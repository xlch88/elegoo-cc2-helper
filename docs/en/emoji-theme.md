# Switching the printer to the emoji edition theme

While analyzing the firmware, I found something interesting.

The firmware includes an emoji theme for the `ELEGOO × emoji® Centauri Carbon 2 Combo` edition.

Judging by the images on the official website, this printer has a white body, a more transparent front glass panel and upper plastic cover, and emoji stickers on the body. It displays an emoji theme at startup.

![Screenshot of the ELEGOO × emoji edition product page](../imgs/PixPin_2026-09-09_22-42-57.png)

Screenshot from the [ELEGOO website](https://us.elegoo.com/products/elegoo-emoji-centauri-carbon-2-combo) ([archived page](https://web.archive.org/web/20260907214758/https://us.elegoo.com/products/elegoo-emoji-centauri-carbon-2-combo)).

The emoji theme included with this edition is also compiled into the regular printer's firmware and can be enabled with a few simple steps.

## Preview

### Startup screen

![Startup screen with the emoji theme](../imgs/emoji-theme-1.png)

### Home screen

![Home screen with the emoji theme](../imgs/emoji-theme-2.png)

### Model list preview

![The emoji model list on the printer screen](../imgs/emoji-theme-3.png)

## Switching methods

The following sections cover two methods: custom G-code and SSH.

**Both methods require a full printer reboot after changing the setting. Before you begin, make sure the printer is idle and is not printing, updating, or performing another operation.**

This article only covers switching between themes built into the firmware. It does not change the printer's appearance or install the complete firmware for the emoji edition. Whether the startup logo and bundled models change as well depends on the firmware implementation; a different interface theme alone does not mean those have also been replaced.

### Method 1: Run custom G-code on v01 firmware

This method does not require SSH, but it does require the v01 printer software that still implements the theme-switching command.

In the analyzed `01.03.02.51` version, `SET_UI_THEME` reads the `THEME` parameter and sets the theme. In `02.00.02.00` and `02.01.00.00`, the command handler is a no-op, so it can no longer be used to switch themes.

**If you are [keeping the v01 system and SSH while replacing the printer software binaries and supporting files with their v02 versions](keep-ssh-update.md), use [Method 2](#method-2-run-commands-over-ssh) as well.** What matters is the printer software actually running, not the version number displayed on the screen, which [can be changed manually](disable-ota-update.md).

#### 1. Send the theme-switching command

Use your slicer to add the command to the first layer. In ELEGOO Slicer, for example:

1. Open the slicer and create a new project.
2. Add a simple standard model, such as a cube.
3. Slice it and open Preview.
4. Move the layer slider on the far right to the first layer.
5. Right-click the slider and choose Add custom G-code.
6. Enter the following command and confirm:

```gcode
SET_UI_THEME THEME=emoji
```

7. Send the print file containing this custom G-code to the printer and start printing.

**You do not need to finish printing the model.** Starting the first layer does not mean the theme command and background writes have both finished. Continue in this order:

8. Wait for the inserted G-code to finish executing and confirm that the background writes for the theme setting have finished.
9. Cancel the print and wait for the printer to stop moving.
10. Reboot the entire printer as described in the next section.

This is G-code sent to the printer software, not a Linux command to run in an SSH shell. Use lowercase `emoji`. Do not leave it in the slicer's start G-code and repeat the theme change on every print.

The v01 handler also runs the firmware's model symlink and startup logo update scripts. The model script rebuilds symbolic links in the local model directory. If you have customized those links, use [Method 2](#method-2-run-commands-over-ssh), which only changes the theme variable.

#### 2. Reboot the printer

Once the command has finished processing, reboot the entire printer. If there is no system reboot option, shut it down normally and turn it back on after confirming that it is idle and all writes have finished.

Refreshing a web page, reconnecting a console, or running Klipper's `RESTART` / `FIRMWARE_RESTART` is not a full printer reboot. After rebooting, check whether the printer screen has switched to the emoji theme.

### Method 2: Run commands over SSH

This method applies to printers that already allow SSH login and still have the emoji theme resources, including mixed-version firmware setups that retain SSH.

If you have not connected over SSH before, read the [SSH connection guide](ssh.md). If SSH stopped working after a full upgrade to v02, see [enabling SSH on v02.01 firmware](enable-ssh-on-v02.md). The theme-switching command itself does not enable SSH.

#### 1. Connect to the printer and check the resources

Connect from your computer, replacing `PRINTER_IP` with the printer's actual IP address:

```sh
ssh root@PRINTER_IP
```

After logging in, run these checks in the printer's SSH shell:

```sh
ls -ld /opt/usr/images/theme_emoji
fw_printenv -n theme_type
```

The first command should find the emoji theme resource directory. If it does not exist, stop and check [whether the current firmware includes the theme resources](firmware-unpack.md) and whether they have been deployed correctly. Changing the theme variable does not create missing images.

The second command records the current theme so that you can restore it later. If it reports that `theme_type` is undefined, simply note that. If it reports another error, such as a read or checksum error for the environment storage, do not continue writing to it.

#### 2. Set the emoji theme

Run this in the same SSH shell:

```sh
fw_setenv theme_type emoji && sync
```

After confirming that there were no errors, read the setting back:

```sh
fw_printenv -n theme_type
```

The output should be:

```text
emoji
```

This only changes the selected theme. There is no need to overwrite the image directory, run the model symlink script manually, or write to the boot partition.

#### 3. Reboot the printer

Once the setting has been written successfully and the printer is idle, run this over SSH:

```sh
reboot
```

The SSH connection will close as the printer reboots; this is normal. Wait for the printer to start again, then check the screen's theme.

## Checking after reboot

Check what the printer screen actually displays. If SSH is available, you can also reconnect and run:

```sh
fw_printenv -n theme_type
readlink /opt/usr/images/current_theme
```

In the analyzed firmware, the corresponding results for the emoji theme are:

```text
emoji
/opt/usr/images/theme_emoji
```

If the variable is already `emoji` but the screen has not changed, first check that the entire printer has rebooted and that the running GUI supports and can access these resources. Do not try to fix it by repeatedly sending the theme command or flashing partitions.

## Restoring the default theme

If you are still running the v01 printer software, use this custom G-code:

```gcode
SET_UI_THEME THEME=default
```

Once processing has finished, reboot the entire printer again.

When restoring over SSH, preferably restore the `theme_type` value you recorded earlier. If the variable was previously undefined, remove the setting you added:

```sh
fw_setenv theme_type && sync
```

To explicitly switch back to the default theme, run:

```sh
fw_setenv theme_type default && sync
```

After confirming that the command succeeded, run `reboot`. This restores the theme selection, not the firmware version. It also does not restore custom symbolic links previously rebuilt by the model script.

## Additional notes on `SET_UI_THEME`

In the analyzed v2 firmware (`02.00.02.00` and `02.01.00.00`), the actual processing logic for `SET_UI_THEME` has been removed.  
Although the G-code command remains registered, invoking it does nothing, so it can no longer be used to switch themes.

In v01, the underlying implementation of this command uses a shell, and its parameter handling is relatively permissive, so its capabilities are not strictly limited to switching themes.  
With certain parameter inputs, it can also execute arbitrary shell commands. This article only covers normal theme selection and does not explore that implementation detail.
