# CC2 02 HTTP upload SSH patch POC

把 01 固件里的 OpenSSH server 文件通过 02 固件的 HTTP 上传接口写到设备根目录。

完整的环境准备、参数和重启连接步骤见[使用说明](../../docs/http-upload-ssh.md)。

## 用法

```bash
CC2_HOST=http://192.168.x.x CC2_TOKEN='填设备 token' node upload-ssh-over-http.mjs
```

或者：

```bash
node upload-ssh-over-http.mjs --host=http://192.168.x.x --token='填设备 token'
```

脚本会走 `/upload/udisk`，用 `X-File-Name=../../目标路径` 写入以下文件：

```text
/usr/sbin/sshd
/etc/init.d/sshd
/etc/ssh/sshd_config
/usr/lib/sftp-server
/lib/upgrade/keep.d/openssh-server
/etc/rc.local
```

`/etc/rc.local` 保留原来的 `/etc/init_dir.sh` 和 `/etc/time_sync.sh`，并在它们之间补：

```sh
grep -q '^sshd:' /etc/group || echo 'sshd:x:22:sshd' >> /etc/group
chmod 0755 /usr/sbin/sshd /etc/init.d/sshd /usr/lib/sftp-server 2>/dev/null || true
mkdir -m 0700 -p /var/empty
pidof sshd >/dev/null 2>&1 || /etc/init.d/sshd start
```

上传完成后重启设备即可。
