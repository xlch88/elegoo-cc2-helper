# 通过 HTTP 上传恢复 SSH

这份脚本把准备好的 SSH 组件上传到 CC2，供重启时启动 SSH。代码和文件在 [`poc/http-upload-ssh-02`](../poc/http-upload-ssh-02)，接口原理见 [v02.01 开启 SSH](./enable-ssh-on-v02.md)。

目前已在 `02.01.00.00` 固件上验证上传写文件，整套 SSH 恢复流程尚未完成实机验证。

## 准备

- 在与打印机同一局域网的电脑上操作，Windows PowerShell、Linux 和 macOS 终端均可。
- 打印机[开启 LAN MODE](./enable-ssh-on-v02.md#开启-lan-mode)，启用访问码，记下屏幕上的 IP 和访问码；上传时不需要打印机已经开启 SSH。
- 电脑安装 Node.js，脚本只用内置模块，不需要运行 `npm install`。按所用的 `node:` 模块导入和顶层 `await`，最低版本为 **14.13.1**，安装时选当前 LTS 即可，见 [Node.js 模块文档](https://nodejs.org/download/release/v14.13.1/docs/api/esm.html#esm_node_imports)。

从 [仓库](https://github.com/xlch88/elegoo-cc2-helper)的 **Code → Download ZIP** 下载并解压，在项目根目录打开终端：

```sh
cd poc/http-upload-ssh-02
node --version
```

保留脚本旁边的整个 `files` 目录，不要只下载 `.mjs` 文件。

## 上传

**把 `PRINTER_IP` 换成打印机 IP，把 `LAN_ACCESS_CODE` 换成屏幕上的 LAN MODE 访问码：**

```sh
node upload-ssh-over-http.mjs --host="http://PRINTER_IP" --token="LAN_ACCESS_CODE"
```

IP 例如 `192.168.1.123`，`--host` 中保留 `http://`。访问码是上传接口的凭据，不是 SSH 登录密码。

脚本会依次打印 `upload /目标路径 (...) ... ok`，全部完成后打印 `done. reboot the printer...`。请求失败会退出；已经写入的文件不会自动恢复。

### 参数

| 参数                       | 环境变量    | 说明                 |
| -------------------------- | ----------- | -------------------- |
| `--host=http://PRINTER_IP` | `CC2_HOST`  | 打印机地址，包含协议 |
| `--token=LAN_ACCESS_CODE`  | `CC2_TOKEN` | LAN MODE 访问码      |

参数要使用 `--名称=值`，不能写成 `--host 地址`；**非空环境变量优先于命令行参数**。脚本只识别上表两项，其他参数会被忽略，没有 `--help`、`--dry-run`、备份或回滚功能。

## 会覆盖哪些文件

脚本递归读取自身旁边的 `files` 目录，按相对路径写到打印机根目录。随包的六个文件对应：

| 本地文件                                  | 打印机路径                           |
| ----------------------------------------- | ------------------------------------ |
| `files/usr/sbin/sshd`                     | `/usr/sbin/sshd`                     |
| `files/etc/init.d/sshd`                   | `/etc/init.d/sshd`                   |
| `files/etc/ssh/sshd_config`               | `/etc/ssh/sshd_config`               |
| `files/usr/lib/sftp-server`               | `/usr/lib/sftp-server`               |
| `files/lib/upgrade/keep.d/openssh-server` | `/lib/upgrade/keep.d/openssh-server` |
| `files/etc/rc.local`                      | `/etc/rc.local`                      |

这些都是**整文件覆盖**，不是合并配置；其中 `/etc/rc.local` 最后上传。改过这些文件的设备，应先保存原文件并比对。不要把无关文件放进 `files`，脚本也会上传它们。

随包的 `sshd`、`sftp-server` 等是从 v01 固件提取的第三方组件，不属于本仓库原创内容的 WTFPL 授权。

## 重启并连接

全部上传完成后，手动重启打印机。脚本本身不会重启打印机或立即启动 SSH。

开机时，`/etc/rc.local` 会补上 `sshd` 用户组、恢复程序执行权限并启动服务；缺失的 SSH 主机密钥由启动脚本在打印机上生成，上传包不含固定私钥，也不修改 root 密码。

待打印机启动后，在电脑终端连接，仍需替换 IP：

```sh
ssh root@PRINTER_IP
```

使用打印机现有的 root 密码，详见 [SSH 连接与密码](./ssh.md)。
