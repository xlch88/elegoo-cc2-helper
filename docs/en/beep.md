# Buzzer notification sounds

`elegoo-cc2-helper` is a small tool that runs on the ELEGOO Centauri Carbon 2 and uses its built-in buzzer to play sounds for different printing states.

It does not control the print process or modify G-code. The program does just two things:

1. Drive the buzzer with PWM to play short melodies.
2. In service mode, listen to `/tmp/elegoo_uds` and play sounds in response to print state changes.

## One-click installation

The printer's built-in `wget` only supports plain HTTP and cannot download files from GitHub over HTTPS. Run the installation commands on your own computer, not in an SSH session on the printer.

The commands below download the latest GitHub release to your computer, upload it to the printer with `scp`, and run `--install` over `ssh` to register the service for startup at boot.

Installation overwrites this program's existing files, so it can also be used for updates. It does not automatically start or restart the service; see [Boot autostart installation](#boot-autostart-installation) below for the commands.

> **Important: SSH must be enabled on the printer before installation.**
> Choose the instructions for your current firmware:
>
> 1. Already upgraded to v02.01: [Enabling SSH on v02.01.00.00 firmware](./enable-ssh-on-v02.md).
> 2. Still on v01 and planning to update the printer software: [Keeping the v01 SSH environment while updating to the v02.01 printer software binaries](./keep-ssh-update.md).
>
> For connection commands and the password, see [SSH connection and default password](./ssh.md).

> **Important: When the `Printer IP` prompt appears, enter your printer's IP address.**
> For example, if the printer screen shows `192.168.1.123`, enter `192.168.1.123`.

### Windows PowerShell

Open PowerShell 5.1 or later, make sure `ssh` and `scp` are available, and paste the entire block:

```powershell
& {
    $ErrorActionPreference = "Stop"
    $PrinterIp = Read-Host "Printer IP"
    $Url = "https://github.com/xlch88/elegoo-cc2-helper/releases/latest/download/elegoo-cc2-helper"
    $LocalFile = [IO.Path]::GetTempFileName()
    $RemoteFile = "/tmp/elegoo-cc2-helper-$([IO.Path]::GetFileName($LocalFile))"
    try {
        Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $LocalFile
        if ((Get-Item $LocalFile).Length -eq 0) { throw "Downloaded file is empty" }
        scp $LocalFile "root@${PrinterIp}:$RemoteFile"
        if ($LASTEXITCODE -ne 0) { throw "Upload failed" }
        ssh "root@$PrinterIp" "trap 'rm -f $RemoteFile' 0; chmod +x $RemoteFile && $RemoteFile --install"
        if ($LASTEXITCODE -ne 0) { throw "Remote installation failed" }
    } finally {
        Remove-Item -LiteralPath $LocalFile -Force
    }
}
```

If prompted for a password, enter the printer's SSH password.

### Linux / macOS

Open a Bash or zsh terminal. Make sure `ssh`, `scp`, and an HTTPS-capable `curl` or `wget` are available, then paste the entire block:

```sh
(
    set -eu
    printf 'Printer IP: '
    IFS= read -r PRINTER_IP </dev/tty
    URL="https://github.com/xlch88/elegoo-cc2-helper/releases/latest/download/elegoo-cc2-helper"
    TMP=$(mktemp "${TMPDIR:-/tmp}/elegoo-cc2-helper.XXXXXX")
    trap 'rm -f "$TMP"' 0
    trap 'exit 1' HUP INT TERM
    REMOTE="/tmp/${TMP##*/}"
    if command -v curl >/dev/null 2>&1; then
        curl -fL -o "$TMP" "$URL"
    else
        wget -O "$TMP" "$URL"
    fi
    [ -s "$TMP" ] || { echo "Downloaded file is empty" >&2; exit 1; }
    scp "$TMP" "root@$PRINTER_IP:$REMOTE"
    ssh "root@$PRINTER_IP" "trap 'rm -f $REMOTE' 0; chmod +x $REMOTE && $REMOTE --install"
)
```

If prompted for a password, enter the printer's SSH password.

## Features

The following notification sounds are built in:

| Sound       | When it plays                           |
| ----------- | --------------------------------------- |
| `startup`   | When the program starts in service mode |
| `printing`  | When printing starts                    |
| `paused`    | When printing is paused                 |
| `resumed`   | When a paused print resumes             |
| `cancelled` | When a print is cancelled               |
| `complete`  | When a print finishes                   |

There is also a test music mode that plays a short excerpt from the opening of _Für Elise_, useful for checking the buzzer and PWM output.

## Usage

Run the following commands in an SSH terminal on the printer. After installation, first enter the program directory:

```sh
cd /opt/usr/elegoo-cc2-helper
```

Running the program without arguments, or with `-h` or `--help`, displays help:

```sh
./elegoo-cc2-helper
./elegoo-cc2-helper -h
./elegoo-cc2-helper --help
```

Test all notification sounds:

```sh
./elegoo-cc2-helper -t
./elegoo-cc2-helper --test
```

Test the music:

```sh
./elegoo-cc2-helper --test-music
```

Run in service mode:

```sh
./elegoo-cc2-helper -d
./elegoo-cc2-helper --daemon
```

Register automatic startup at boot:

```sh
./elegoo-cc2-helper --install
```

## Service mode

Service mode connects to the Klipper-compatible Unix socket:

```text
/tmp/elegoo_uds
```

The program subscribes to `print_stats.state` changes and plays the corresponding sounds.

When the state changes, it follows these rules:

| State change                                | Sound       |
| ------------------------------------------- | ----------- |
| Any state other than `paused` -> `printing` | `printing`  |
| `paused` -> `printing`                      | `resumed`   |
| Any state -> `paused`                       | `paused`    |
| Any state -> `cancelled`                    | `cancelled` |
| Any state -> `complete`                     | `complete`  |

The first state received after connecting is used only as a baseline. It does not trigger a print state sound, so starting the program does not replay a notification for the current state.

In service mode, the program plays `startup` once after connecting to the socket and before sending the subscription request.

## PWM output

The program controls the buzzer directly through the system PWM nodes:

```text
/sys/class/pwm/pwmchip0/pwm0/enable
/sys/class/pwm/pwmchip0/pwm0/period
/sys/class/pwm/pwmchip0/pwm0/duty_cycle
```

For each note, the program:

1. Disables PWM output.
2. Calculates `period` from the note's frequency.
3. Sets `duty_cycle` to `period / 2`.
4. Enables PWM output.
5. Waits for the duration of the note.
6. Disables PWM output again.

The duty cycle is already set to 50%, so volume is mainly limited by the buzzer hardware, its driver circuit, and the printer's physical construction. There is little room for further adjustment in software.

## Boot autostart installation

Running `--install` copies the current executable to:

```text
/opt/usr/elegoo-cc2-helper/elegoo-cc2-helper
```

It also creates a procd service script:

```text
/etc/init.d/elegoo-cc2-helper
```

And the startup and shutdown links:

```text
/etc/rc.d/S99elegoo-cc2-helper
/etc/rc.d/K10elegoo-cc2-helper
```

Installation replaces the old executable, service script, and startup links at these paths without changing other system startup files. The executable and script are written to temporary files before replacing their destinations, rather than truncating a running executable in place.

Installation does not automatically start or restart the service. After the first installation, reboot the printer or start the service manually. When updating, an already-running process continues to use the old version until you manually restart the service.

To start the service, run this in an SSH terminal on the printer:

```sh
/etc/init.d/elegoo-cc2-helper start
```

After updating a running service, run this in an SSH terminal on the printer:

```sh
/etc/init.d/elegoo-cc2-helper restart
```

## Building

The source code is in `src`. Run the build script from the project root on your computer to produce an executable for the printer:

```sh
cd src
./build.sh
```

The build output is written to:

```text
src/dist/elegoo-cc2-helper
```

The current build target is Linux ARM, for use on the CC2.

## Notes

1. The program needs write access to the PWM nodes, so it normally requires root privileges.
2. Service mode requires `/tmp/elegoo_uds` to exist and accept connections.
3. The program only listens for print state changes. It does not send pause, resume, cancel, or other control commands.
4. Before manually testing sounds, make sure it is appropriate to sound the buzzer in your surroundings.
5. If firmware paths or PWM nodes change, check whether the program is still compatible.
