# Default SSH password for CC2 firmware

If you have upgraded to v02.01 and SSH is unavailable, see [Enabling SSH on v02.01](./enable-ssh-on-v02.md). If you are still on v01 and want to keep SSH while updating the printer software binaries and accompanying files, see [Keeping the v01 SSH environment](./keep-ssh-update.md).

Once SSH is enabled on the printer, run the following in your computer's terminal or PowerShell. **Replace `PRINTER_IP` with your printer's actual IP address:**

```sh
ssh root@PRINTER_IP
```

The following default password was recorded by the community; the firmware versions it applies to have not been verified. If you have changed the password, use your own instead.

```text
MTY4ODE2
```

Source: [Discord community message](https://discord.com/channels/1367538416539013122/1434248003459354684/1488450815801819236).

After logging in, it is recommended that you change the default password by running the following in the printer's SSH terminal. Keep the current connection open and verify that the new password works in another terminal before disconnecting:

```sh
passwd
```

Once you have confirmed SSH works, follow the respective guides to [install buzzer notification sounds](./beep.md#one-click-installation), [switch to the emoji theme over SSH](./emoji-theme.md#method-2-run-commands-over-ssh), or [disable the firmware update prompt at startup](./disable-ota-update.md#steps).
