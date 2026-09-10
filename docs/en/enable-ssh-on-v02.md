# Enabling SSH on v02.01.00.00 firmware

ELEGOO removed some SSH server components from v02.01.00.00 firmware, preventing users from continuing to connect to the printer remotely over SSH for debugging and other operations. This article explains how to restore SSH on that version.

Analysis of the upload process in v02.01.00.00 shows that the local network file upload endpoint can be used to restore the missing SSH components and enable SSH again.

Please note that this method carries some risk and may cause the printer to stop working correctly. The author provides the technical analysis and procedure; you use them at your own risk.

**The firmware version used when writing this article was `02.01.00.00`. Other versions may differ, so proceed carefully.**

Uploading and writing files has been tested on this version, but the complete SSH recovery procedure has not yet been verified on a physical printer.

## Enable LAN MODE

This requires LAN MODE (local-network-only mode). After enabling SSH, you can turn off local-network-only mode to resume using cloud features.

Configure this on the printer screen:

Step 1: Tap the Settings icon → Network, then connect to Wi-Fi.

Step 2: Open Settings, scroll down to LAN-only mode, open it, and enable it.

Step 3: Enable the access code and tap the refresh button to generate a new one. Note the IP address and access code shown on the screen.

## Upload endpoint behavior

The relevant endpoint is `/upload/udisk`.  
It appears to be intended mainly for receiving print files over the local network. In the tested version, its filename handling places relatively loose restrictions on the destination path, allowing it to be used to restore missing system components.

### Minimal test

The following example only writes a test file named `/1.txt`, containing the single byte `1`.

Run it in a **Bash terminal on a computer** on the same local network as the printer (Windows users can use WSL). Do not run it directly in the printer's SSH terminal or in PowerShell. Replace `PRINTER_IP` with the printer's IP address and `LAN_ACCESS_CODE` with the access code shown on its screen.

```bash
#!/usr/bin/env bash
set -euo pipefail

PRINTER="http://PRINTER_IP"
TOKEN="LAN_ACCESS_CODE"

printf '1' | curl -i -X POST "$PRINTER/upload/udisk" \
  -H "X-Token: $TOKEN" \
  -H "X-File-Name: ..%2F..%2F1.txt" \
  -H "X-File-MD5: c4ca4238a0b923820dcc509a6f75849b" \
  -H "Content-Range: bytes 0-0/1" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @-
```

### Endpoint behavior

When tested on `02.01.00.00`, `/upload/udisk` read the destination filename from the `X-File-Name` request header, URL-decoded it on the server, and joined it to form the path written to disk. Normally, it should only write inside the upload directory. If the filename contains `../../`, the resulting path may fall outside the default upload directory.

For an actual request, therefore, do not supply `/tmp/cc2-upload-test.txt` directly. Instead, URL-encode `../../tmp/cc2-upload-test.txt` and put it in `X-File-Name`, for example:

```text
..%2F..%2Ftmp%2Fcc2-upload-test.txt
```

After the server decodes this value, it is equivalent to:

```text
../../tmp/cc2-upload-test.txt
```

For a destination of `/etc/rc.local`, the corresponding `X-File-Name` is:

```text
..%2F..%2Fetc%2Frc.local
```

### Request format

The upload request is not `multipart/form-data`. The file contents are sent directly as the HTTP POST body to:

```text
POST http://PRINTER_IP/upload/udisk
```

The headers required in the current tests are:

| Header           | Purpose                                                   |
| ---------------- | --------------------------------------------------------- |
| `X-Token`        | Access code shown under LAN MODE on the printer screen    |
| `X-File-Name`    | URL-encoded destination path                              |
| `X-File-MD5`     | MD5 of the request body                                   |
| `Content-Range`  | File range, in the form `bytes 0-(file_size-1)/file_size` |
| `Content-Type`   | `application/octet-stream`                                |
| `Content-Length` | File size                                                 |

The minimal test above includes the complete request; curl sets `Content-Length` automatically.

On success, the endpoint returns HTTP 200 and an `error_code` of `0` in the JSON response. In earlier testing, the printer returned a result like this:

```json
{ "error_code": 0, "offset": 1 }
```

Here, `offset` is the progress position reported by the endpoint. It is not recommended as the sole measure of the final result. To determine whether the request was accepted, primarily check the HTTP status code and `error_code`.

## Restoring the SSH components

After the SSH files were removed from v02 firmware, the OpenSSH server components in v01 firmware can serve as a reference for restoring the missing files to the printer's root filesystem. The example [upload script](../../poc/http-upload-ssh-02/upload-ssh-over-http.mjs) and [files directory](../../poc/http-upload-ssh-02/files/) are included in the repository. For downloading, running the script, and rebooting, see [restoring SSH through HTTP uploads](http-upload-ssh.md). The script calculates each file's MD5 and length.

If you need to extract or check the components yourself, follow [getting and unpacking CC2 firmware](firmware-unpack.md) to inspect the rootfs.

The example's `files` directory maps to the printer's root directory. For example, `files/usr/sbin/sshd` corresponds to `/usr/sbin/sshd` on the printer.

The example currently includes these files:

| Local file                                | Destination on the printer           | Purpose                                          |
| ----------------------------------------- | ------------------------------------ | ------------------------------------------------ |
| `files/usr/sbin/sshd`                     | `/usr/sbin/sshd`                     | SSH server executable                            |
| `files/etc/init.d/sshd`                   | `/etc/init.d/sshd`                   | OpenWrt/procd-style startup script               |
| `files/etc/ssh/sshd_config`               | `/etc/ssh/sshd_config`               | SSH server configuration                         |
| `files/usr/lib/sftp-server`               | `/usr/lib/sftp-server`               | SFTP subsystem                                   |
| `files/lib/upgrade/keep.d/openssh-server` | `/lib/upgrade/keep.d/openssh-server` | List of files to retain during firmware upgrades |
| `files/etc/rc.local`                      | `/etc/rc.local`                      | Fix permissions and start SSH at boot            |

The upload script processes the other files first and `/etc/rc.local` last. This avoids writing the startup logic before the binaries and configuration files are in place.

### What `/etc/rc.local` does

The example's `/etc/rc.local` is a prepared, complete file, not a patch that merges automatically. It preserves the original `/etc/init_dir.sh` and `/etc/time_sync.sh` calls and adds SSH startup logic between them:

```sh
if ! grep -q '^sshd:' /etc/group; then
    echo 'sshd:x:22:sshd' >> /etc/group
fi

chmod 0755 /usr/sbin/sshd /etc/init.d/sshd /usr/lib/sftp-server 2>/dev/null || true
mkdir -m 0700 -p /var/empty
if [ -x /etc/init.d/sshd ] && ! pidof sshd >/dev/null 2>&1; then
    /etc/init.d/sshd start
fi
```

HTTP uploads only write file contents and do not guarantee executable permissions. This logic therefore sets `0755` permissions on `sshd`, `/etc/init.d/sshd`, and `sftp-server` at boot, creates `/var/empty`, and then starts the SSH service.

When `/etc/init.d/sshd` starts, it checks for and generates any missing host keys:

```text
/etc/ssh/ssh_host_rsa_key
/etc/ssh/ssh_host_ecdsa_key
/etc/ssh/ssh_host_ed25519_key
```

In other words, the example does not upload fixed private keys; host keys are generated on the printer.

For the login command and password after rebooting, see [SSH connection and the default password](ssh.md).

## Notes

1. `/etc/rc.local` is overwritten, not intelligently merged. If this file has already been modified in your firmware or on your printer, back it up and compare it before uploading.
2. The example does not change the root password or write an SSH login password. The LAN MODE access code is only for the HTTP upload endpoint; it is not the SSH password.
3. `sshd_config` enables `PermitRootLogin yes` and configures `AuthorizedKeysFile .ssh/authorized_keys` and the SFTP subsystem. Whether login succeeds still depends on the account, password, or `authorized_keys` on the printer.
4. The upload script only writes files. It does not reboot the printer automatically or start SSH immediately. As currently designed, it requires a manual reboot after uploading so that `/etc/rc.local` starts SSH during boot.
5. This article is based on the current example and the upload behavior observed on `02.01.00.00`. Other v02 versions require rechecking the endpoint behavior and system file differences.
