# 如何在 v02.01.00.00 固件上开启 SSH

ELEGOO 在 v02.01.00.00 固件中移除了部分 SSH 服务端组件，导致用户无法继续通过 SSH 远程连接打印机进行调试和操作。本文介绍在该版本上恢复 SSH 的方法。

通过对 v02.01.00.00 固件上传流程的分析，可以借助局域网文件上传接口补全 SSH 相关组件，从而重新开启 SSH 功能。

请注意，这种方法存在一定的风险，可能会导致打印机无法正常工作，作者仅提供技术分析和操作方法，使用者需自行承担风险。

**编写此文档时，笔者使用的固件版本为 `02.01.00.00`，其他版本可能存在差异，请谨慎操作。**

上传写文件已在该版本实测，整套 SSH 恢复流程尚未完成实机验证。

## 开启 LAN MODE

此功能需要启用 LAN MODE（仅局域网模式）。开启 SSH 后，可以关闭仅局域网模式，恢复使用云功能。

具体需要在机器屏幕上进行设置。

第一步：点击设置图标 -> 网络，连接 Wi-Fi。

第二步：点击设置 -> 下滑找到“仅局域网”，点击进入，然后开启“仅局域网”模式。

第三步：启用访问码，并点击刷新按钮生成新的访问码。记录下屏幕上显示的 IP 地址和访问码。

## 上传接口行为说明

相关行为位于 `/upload/udisk` 接口。  
该接口看起来主要用于接收局域网内上传的打印文件。在当前测试版本中，文件名处理逻辑对目标路径的限制较宽，因此可以用于补全缺失的系统组件。

### 最小验证

下面的示例只写入一个名为 `/1.txt` 的测试文件，文件内容为单字节 `1`。

在与打印机同一局域网的**电脑 Bash 终端**中运行（Windows 可使用 WSL），不要在打印机 SSH 终端或 PowerShell 中直接执行。将 `PRINTER_IP` 和 `LAN_ACCESS_CODE` 分别替换为打印机 IP 和屏幕上的访问码。

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

### 接口行为

在 `02.01.00.00` 上测试时，`/upload/udisk` 会从请求头 `X-File-Name` 读取目标文件名，并在服务端进行 URL 解码后拼接成落盘路径。正常情况下它应当只写入上传目录；如果文件名中包含 `../../`，最终路径可能会落到默认上传目录之外。

因此，实际请求里不要直接写 `/tmp/cc2-upload-test.txt`，而是把 `../../tmp/cc2-upload-test.txt` 做 URL 编码后放到 `X-File-Name` 中，例如：

```text
..%2F..%2Ftmp%2Fcc2-upload-test.txt
```

这个值在服务端解码后等价于：

```text
../../tmp/cc2-upload-test.txt
```

如果目标路径是 `/etc/rc.local`，对应的 `X-File-Name` 会是：

```text
..%2F..%2Fetc%2Frc.local
```

### 请求格式

上传请求不是 `multipart/form-data`，而是把文件内容直接作为 HTTP POST 请求体发送到：

```text
POST http://PRINTER_IP/upload/udisk
```

当前测试中需要带上的请求头如下：

| 请求头           | 作用                                             |
| ---------------- | ------------------------------------------------ |
| `X-Token`        | 屏幕 LAN MODE 里显示的访问码                     |
| `X-File-Name`    | URL 编码后的目标路径                             |
| `X-File-MD5`     | 请求体的 MD5                                     |
| `Content-Range`  | 文件范围，格式为 `bytes 0-(文件大小-1)/文件大小` |
| `Content-Type`   | `application/octet-stream`                       |
| `Content-Length` | 文件大小                                         |

上面的最小验证已给出完整请求，`Content-Length` 由 curl 自动设置。

成功时接口返回 HTTP 200，并且 JSON 里的 `error_code` 为 `0`。此前测试时，设备返回过类似下面的结果：

```json
{ "error_code": 0, "offset": 1 }
```

这里的 `offset` 表示接口返回的进度位置，不建议单独用它判断最终结果。判断请求是否被接受时，主要看 HTTP 状态码和 `error_code`。

## SSH 组件补全思路

v02 固件移除 SSH 相关文件后，可以参考 v01 固件中的 OpenSSH 服务端组件，将缺失文件补回设备根文件系统。示例的 `files` 目录和上传脚本已放在 [`poc/http-upload-ssh-02`](../poc/http-upload-ssh-02)，下载、执行和重启步骤见[通过 HTTP 上传恢复 SSH](./http-upload-ssh.md)。脚本会计算各文件的 MD5 和长度。

需要自行提取或核对组件时，可按[获取并解包 CC2 固件](./firmware-unpack.md)查看 rootfs。

示例的 `files` 目录映射到设备根目录，例如 `files/usr/sbin/sshd` 对应设备上的 `/usr/sbin/sshd`。

当前示例包含的文件如下：

| 本地文件                                  | 设备目标路径                         | 作用                       |
| ----------------------------------------- | ------------------------------------ | -------------------------- |
| `files/usr/sbin/sshd`                     | `/usr/sbin/sshd`                     | SSH 服务端主程序           |
| `files/etc/init.d/sshd`                   | `/etc/init.d/sshd`                   | OpenWrt/procd 风格启动脚本 |
| `files/etc/ssh/sshd_config`               | `/etc/ssh/sshd_config`               | SSH 服务端配置             |
| `files/usr/lib/sftp-server`               | `/usr/lib/sftp-server`               | SFTP 子系统                |
| `files/lib/upgrade/keep.d/openssh-server` | `/lib/upgrade/keep.d/openssh-server` | 固件升级保留文件列表       |
| `files/etc/rc.local`                      | `/etc/rc.local`                      | 开机时修复权限并启动 SSH   |

上传脚本会先处理其他文件，最后处理 `/etc/rc.local`。这样做是为了避免启动逻辑已经写入，但二进制或配置文件尚未就绪。

### `/etc/rc.local` 做了什么

示例里的 `/etc/rc.local` 是一个准备好的完整文件，不是自动合并补丁。它保留了原来的 `/etc/init_dir.sh` 和 `/etc/time_sync.sh`，并在中间加入 SSH 启动逻辑：

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

因为 HTTP 上传只写文件内容，不保证可执行权限，所以这里在开机时重新给 `sshd`、`/etc/init.d/sshd`、`sftp-server` 设置 `0755` 权限，然后创建 `/var/empty`，最后启动 SSH 服务。

`/etc/init.d/sshd` 启动时会检查并生成缺失的主机密钥：

```text
/etc/ssh/ssh_host_rsa_key
/etc/ssh/ssh_host_ecdsa_key
/etc/ssh/ssh_host_ed25519_key
```

也就是说，示例不会上传固定的私钥，主机密钥会在设备上生成。

重启后的登录命令和密码见 [SSH 连接与默认密码](./ssh.md)。

## 注意事项

1. `/etc/rc.local` 是覆盖写入，不是智能合并。如果您的固件或设备里已经改过这个文件，上传前应先自行备份并比对。
2. 示例不会修改 root 密码，也不会写入 SSH 登录密码。LAN MODE 的访问码只用于 HTTP 上传接口，不等于 SSH 密码。
3. `sshd_config` 中启用了 `PermitRootLogin yes`，并配置了 `AuthorizedKeysFile .ssh/authorized_keys` 和 SFTP 子系统，但具体能否登录仍取决于设备上的账号、密码或 `authorized_keys` 状态。
4. 上传脚本只负责写文件，完成后不会自动重启设备，也不会立即启动 SSH。按当前设计，需要上传完成后手动重启设备，让 `/etc/rc.local` 在开机过程中启动 SSH。
5. 本文档依据当前示例和 `02.01.00.00` 上的上传行为编写，其他 v02 版本需要重新确认接口行为和系统文件差异。
