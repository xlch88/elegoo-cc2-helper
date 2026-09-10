# Getting and unpacking CC2 firmware

This guide shows how to open CC2 firmware on your computer and inspect its programs and configuration. **On Windows, a browser-based unpacker and 7-Zip are all you need.** Optional Linux command-line steps are included below. The package format is described on the CC2-specific page in the [OpenCentauri software documentation](https://docs.opencentauri.cc/software/).

## Getting the firmware

Use an original package provided by ELEGOO for your exact model whenever possible. For older releases, you can also select a version in the [OpenCentauri CC2 firmware archive](https://docs.opencentauri.cc/software/updates-cc2/#firmware-update-archive) and click **Download**.

- Make sure the firmware is for **Centauri Carbon 2**, and check the version and region. The general [Updates page](https://docs.opencentauri.cc/software/updates/) covers CC1; its firmware and `unpack.py` commands cannot be used directly for CC2.
- The archive links to community-hosted files, not an official ELEGOO download site. When comparing original firmware, do not choose a version marked **repacked**.
- As of September 10, 2026, the archive does not list `02.01.00.00`. To follow the v02.01 printer software update guide, obtain the original package for that version separately; `02.00.02.00` is not a substitute.

## Removing the two layers of `.sig` packaging

A typical CC2 update package has this structure:

```text
firmware.zip.sig
  -> firmware.zip
      -> firmware.swu.sig
          -> firmware.swu
      -> firmware.json.sig
```

Open the [CC2-specific unpacker](https://docs.opencentauri.cc/extras/cc2_update_decrypt.html) in a **browser on your computer**:

1. Select the downloaded `.zip.sig`, click **Unpack**, and save the resulting `.zip`.
2. Open that `.zip` with 7-Zip and extract the `.swu.sig` inside.
3. Open the `.swu.sig` with the same unpacker and save the resulting `.swu`.

The tool processes only one `.sig` layer at a time and does not automatically extract the ZIP. The **Version** shown on the page is the package format version, not the printer firmware version. See the [unpacker source code](https://github.com/OpenCentauri/OpenCentauri/blob/main/docs/extras/cc2_update_decrypt.html) for details.

If you already have a `.swu.sig`, start at step 3. If you already have a decrypted `.swu`, go straight to the next section. The accompanying `.json.sig` contains metadata, not the rootfs image.

## Opening rootfs with 7-Zip (Windows)

Install [7-Zip](https://www.7-zip.org/) on your computer, then follow these steps. WSL or Linux is not required:

1. Right-click the `.swu` file you just saved and select **7-Zip → Open archive**.
2. Find the file named **`rootfs`**, select it, and click **Extract**. Save it to a folder on your computer, such as `D:\CC2`.
3. Right-click the extracted **`rootfs` file** and select **7-Zip → Open archive** again. It has no extension; open it as-is without renaming it.
4. You can now browse the directories and files in the firmware. Select and extract any files you need.

If 7-Zip is missing from the Windows 11 context menu, click **Show more options** first. Alternatively, open **7-Zip File Manager**, find the file, and open it there.

The whole process is:

```text
.swu file
  → Open with 7-Zip and extract rootfs
    → Open rootfs with 7-Zip to browse the firmware
```

The commonly used files are under `opt`:

```text
opt/bin/                                  Printer software
opt/lib/                                  Required libraries
opt/inst/                                 Default configuration, images, and other resources
opt/inst/firmware_version/versions.json    Firmware version record
```

For example, extract `versions.json` and open it in Notepad to view `model` and `ota_version`. These are the values recorded in the original package; you do not need to change them. For manually changed version numbers on a printer, see [update prompts and version numbers](./disable-ota-update.md#after-the-change). To extract the entire rootfs, choose a new directory such as `D:\CC2\rootfs-files`, rather than using the same name as the existing `rootfs` file.

**If you only need to inspect or extract files, you are done.** To build a printer software update package, use the [Linux extraction method below](#extracting-rootfs-on-linux-optional) to extract the printer software binaries and their accompanying libraries, configuration, and interface resources from the original package while preserving permissions, ownership, and symbolic links.

## Extracting rootfs on Linux (optional)

Run the following in a **Linux Bash terminal on your computer**. Windows users can use **WSL2 Ubuntu**. Do not run these commands in an SSH session on the printer. Use a directory on the Linux filesystem, such as your WSL home directory; do not extract the final rootfs to `/mnt/c`, where Linux permissions or symbolic links may be lost.

You need GNU cpio, `file`, and an `unsquashfs` build with XZ support. On Ubuntu / Debian, install them from the system repositories:

```bash
sudo apt-get update && sudo apt-get install cpio squashfs-tools file
```

Use `unsquashfs -version` to check the tool version. The options below follow the [Squashfs-tools 4.7.5 manual](https://github.com/plougher/squashfs-tools/blob/master/Documentation/4.7.5/USAGE-UNSQUASHFS.md); you do not need that exact version.

Replace `/path/to/firmware.swu` in the code with the **absolute path to the `.swu` file** you saved earlier. This first extracts the `rootfs` image from the SWU, then unpacks it into a new directory:

```bash
(
  set -eu
  SWU="/path/to/firmware.swu"
  test -f "$SWU"
  WORK=$(mktemp -d "$HOME/cc2-unpack.XXXXXX")
  mkdir "$WORK/swu"
  cd "$WORK/swu"
  cpio -id --no-absolute-filenames rootfs sw-description < "$SWU"
  test -f rootfs
  file rootfs
  unsquashfs -s rootfs
  sudo unsquashfs -d "$WORK/rootfs" rootfs
  test -x "$WORK/rootfs/opt/bin/elegoo_printer"
  test -d "$WORK/rootfs/opt/lib"
  test -f "$WORK/rootfs/opt/inst/printer_dsp.cfg"
  grep -E '"(model|ota_version)"' "$WORK/rootfs/opt/inst/firmware_version/versions.json"
  printf 'Rootfs directory: %s\n' "$WORK/rootfs"
)
```

In the inspected `02.01.00.00` package, `rootfs` is a **SquashFS 4.0 / XZ** image. It has no extension, but it is a file, not a directory. `unsquashfs` extracts it without mounting the image; `sudo` preserves the original ownership, permissions, and special files.

If a command fails, do not package the incomplete output. In particular, if `rootfs` is missing or its format cannot be recognized, first check that you have not stopped at the `.zip` or `.swu.sig` layer.

## Preparing the printer software update package

The Linux commands above print a directory path similar to this:

```text
/home/yourname/cc2-unpack.ABC123/rootfs
```

It should contain directories such as `opt/bin`, `opt/lib`, and `opt/inst`. Check the `model` and `ota_version` values in the output. To continue with [Keep SSH on v01 and update to the v02.01 printer software binaries](./keep-ssh-update.md), the version must be `02.01.00.00`.

Replace `/path/to/cc2-v02.01-rootfs` in that guide with the **rootfs directory** printed here, not the `.swu` file or the `swu/rootfs` image file. You can then continue with packaging.

The commands in this article have been checked for syntax and against the referenced documentation. No firmware was downloaded again, and the complete unpacking process was not run during those checks.
