# Restoring SSH through HTTP uploads

This script uploads the prepared SSH components to the CC2 so that SSH can start after a reboot. The [upload script](../../poc/http-upload-ssh-02/upload-ssh-over-http.mjs) and [bundled files](../../poc/http-upload-ssh-02/files/) are in the repository. For how the endpoint works, see [enabling SSH on v02.01](enable-ssh-on-v02.md).

Uploading and writing files has been verified on `02.01.00.00` firmware, but the complete SSH recovery procedure has not yet been tested on a physical printer.

## Preparation

- Use a computer on the same local network as the printer. Windows PowerShell, Linux terminals, and macOS terminals are all suitable.
- [Enable LAN MODE](enable-ssh-on-v02.md#enable-lan-mode) on the printer, enable the access code, and note the IP address and access code shown on the screen. SSH does not need to be enabled for uploading.
- Install Node.js on the computer. The script only uses built-in modules, so there is no need to run `npm install`. Based on its `node:` module imports and top-level `await`, the minimum version is **14.13.1**. Choose the current LTS release when installing. See the [Node.js module documentation](https://nodejs.org/download/release/v14.13.1/docs/api/esm.html#esm_node_imports).

Download the [repository](https://github.com/xlch88/elegoo-cc2-helper) using **Code → Download ZIP** and extract it. Open a terminal in the project root:

```sh
cd poc/http-upload-ssh-02
node --version
```

Keep the entire `files` directory next to the script. Do not download only the `.mjs` file.

## Uploading

**Replace `PRINTER_IP` with the printer's IP address and `LAN_ACCESS_CODE` with the LAN MODE access code shown on its screen:**

```sh
node upload-ssh-over-http.mjs --host="http://PRINTER_IP" --token="LAN_ACCESS_CODE"
```

An example IP address is `192.168.1.123`; keep `http://` in `--host`. The access code authenticates uploads and is not the SSH login password.

The script prints lines in the form `upload /target/path (...) ... ok`, followed by `done. reboot the printer...` when everything is complete. It exits if a request fails; files already written are not restored automatically.

### Arguments

| Argument                   | Environment variable | Description                             |
| -------------------------- | -------------------- | --------------------------------------- |
| `--host=http://PRINTER_IP` | `CC2_HOST`           | Printer address, including the protocol |
| `--token=LAN_ACCESS_CODE`  | `CC2_TOKEN`          | LAN MODE access code                    |

Use the `--name=value` format, not `--host address`. **Nonempty environment variables take precedence over command-line arguments.** The script only recognizes the two arguments above and ignores others. It has no `--help`, `--dry-run`, backup, or rollback functionality.

## Files that will be overwritten

The script recursively reads the `files` directory beside it and writes each file to the printer's root directory using its relative path. The six bundled files map to:

| Local file                                | Path on the printer                  |
| ----------------------------------------- | ------------------------------------ |
| `files/usr/sbin/sshd`                     | `/usr/sbin/sshd`                     |
| `files/etc/init.d/sshd`                   | `/etc/init.d/sshd`                   |
| `files/etc/ssh/sshd_config`               | `/etc/ssh/sshd_config`               |
| `files/usr/lib/sftp-server`               | `/usr/lib/sftp-server`               |
| `files/lib/upgrade/keep.d/openssh-server` | `/lib/upgrade/keep.d/openssh-server` |
| `files/etc/rc.local`                      | `/etc/rc.local`                      |

These are all **whole-file replacements**, not configuration merges. `/etc/rc.local` is uploaded last. If these files have been modified on your printer, save the originals and compare them first. Do not put unrelated files in `files`; the script will upload those too.

The bundled third-party files were extracted from ELEGOO Centauri Carbon 2 v01 firmware. They are not covered by this project's WTFPL license, and their rights belong to their respective rights holders.

## Rebooting and connecting

After all uploads finish, reboot the printer manually. The script does not reboot it or start SSH immediately.

At boot, `/etc/rc.local` adds the `sshd` group, restores executable permissions, and starts the service. Missing SSH host keys are generated on the printer by the startup script. The upload package does not contain fixed private keys or change the root password.

Once the printer has started, connect from a terminal on your computer, again replacing the IP address:

```sh
ssh root@PRINTER_IP
```

Use the printer's existing root password. See [SSH connection and password](ssh.md) for details.
