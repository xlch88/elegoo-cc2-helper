# 保留 SSH 并更新到 v02.01 固件

在 v01 固件中，SSH 默认可用，可以直接连接设备进行调试和维护。

从 v02 固件开始，系统层移除了 SSH server 相关组件，不是单纯关闭服务，而是缺少了 `/usr/sbin/sshd`、`/etc/init.d/sshd`、`/etc/ssh/sshd_config`、`/usr/lib/sftp-server` 等文件，因此完整升级到 v02 后就无法继续通过 SSH 连接打印机。

一种相对可控的做法是：保留 v01 的系统层和 SSH 环境，只把 v02.01 的打印业务组件同步到设备上。这样可以获得 v02.01 的主要打印程序更新，同时避免完整替换 rootfs 后丢失 SSH。

**本文基于 `01.03.02.51` 与 `02.01.00.00` 的 rootfs 对比结果编写，其他版本需要重新比对。**

# 差异结论

## SSH 相关差异

v01 中存在完整的 OpenSSH server 组件：

| 路径                                 | 说明                                     |
| ------------------------------------ | ---------------------------------------- |
| `/usr/sbin/sshd`                     | SSH server 主程序                        |
| `/etc/init.d/sshd`                   | procd 启动脚本，启动 `/usr/sbin/sshd -D` |
| `/etc/ssh/sshd_config`               | SSH server 配置                          |
| `/usr/lib/sftp-server`               | SFTP 子系统                              |
| `/lib/upgrade/keep.d/openssh-server` | 升级时保留 SSH host key 的列表           |
| `/etc/group` 中的 `sshd:x:22:sshd`   | sshd 组                                  |

v02.01 中仍保留了 SSH client、`ssh-keygen`、`/etc/ssh/ssh_config` 和 `/etc/passwd` 里的 `sshd` 用户，但缺少 server 侧文件和 `sshd` 组。

这说明 v02.01 不是缺少所有 SSH 运行环境，而是移除了对外提供登录服务所需的 server 组件。

## 打印启动链差异

两版固件的打印入口保持一致：

```text
/etc/init.d/printer
  -> /opt/bin/run_printer.sh start
  -> elegoo_printer /opt/inst/printer_dsp.cfg -s /opt/usr/cfg/autosave.cfg -a /tmp/elegoo_uds
```

对比结果中，`/etc/init.d/printer` 和 `/opt/bin/run_printer.sh` 内容一致。因此，要更新打印业务能力，重点不在系统 init 脚本，而在 `/opt` 下面的业务程序、库、配置和资源。

## v02.01 主要变化位置

对比 `01.03.02.51` 和 `02.01.00.00` 后，和打印业务直接相关的变化主要集中在这些位置：

| 路径                                         | 变化                    |
| -------------------------------------------- | ----------------------- |
| `/opt/bin/elegoo_printer`                    | 主打印程序更新          |
| `/opt/bin/ai_camera`                         | 摄像头/视觉相关程序更新 |
| `/opt/bin/ec-eeb001-gui`                     | 屏幕 GUI 程序更新       |
| `/opt/inst/daemon-000/daemon-000`            | 后台组件更新            |
| `/opt/inst/daemon-000/update_whitelist.conf` | v02.01 新增             |
| `/opt/inst/printer_dsp.cfg`                  | 打印配置更新            |
| `/opt/lib/libapi_module.a`                   | 业务库更新              |
| `/opt/lib/libcommunication.a`                | 通信库更新              |
| `/opt/lib/libelegoo_extras.so`               | 扩展库更新              |
| `/opt/lib/libturbo_core.so`                  | 业务库更新              |
| `/opt/inst/images/`                          | 部分界面资源更新        |
| `/opt/inst/factory/`                         | 出厂示例 G-code 有变化  |

为了减少遗漏，建议按目录整体同步 `/opt/bin`、`/opt/lib`、`/opt/inst`，不要只挑几个二进制文件复制。

# 推荐方案

总体思路如下：

1. 设备保持 v01 固件启动。
2. 通过 v01 的 SSH 登录设备。
3. 从 v02.01 rootfs 中打包 `/opt/bin`、`/opt/lib`、`/opt/inst`。
4. 上传到设备。
5. 停止打印业务进程。
6. 备份设备当前 `/opt` 业务目录。
7. 解包 v02.01 的 `/opt` 业务文件。
8. 重启设备并确认 SSH 仍可连接。

这个方案不替换 `/usr/sbin/sshd`、`/etc/init.d/sshd`、`/etc/ssh/sshd_config` 等 v01 的 SSH 文件，也不替换整个 rootfs。

# 在电脑上准备 v02.01 业务包

假设已经解包 v02.01 rootfs，并且目录为：

```text
runtime/elegoo/compare-02.01.00.00
```

在电脑上执行：

```sh
cd runtime/elegoo/compare-02.01.00.00
sudo tar --numeric-owner -czf /tmp/cc2-v02.01-opt.tar.gz opt/bin opt/lib opt/inst
sha256sum /tmp/cc2-v02.01-opt.tar.gz
```

如果不想带出厂示例 G-code，可以排除 `/opt/inst/factory`，这样包会小很多：

```sh
cd runtime/elegoo/compare-02.01.00.00
sudo tar --numeric-owner -czf /tmp/cc2-v02.01-opt.tar.gz \
  --exclude='opt/inst/factory' \
  opt/bin opt/lib opt/inst
sha256sum /tmp/cc2-v02.01-opt.tar.gz
```

# 上传到 v01 设备

把业务包上传到打印机：

```sh
scp /tmp/cc2-v02.01-opt.tar.gz root@打印机IP:/tmp/
```

连接打印机：

```sh
ssh root@打印机IP
```

# 在设备上替换业务文件

先停止打印业务进程：

```sh
/etc/init.d/printer stop || true
/opt/bin/run_printer.sh stop || true
killall elegoo_printer 2>/dev/null || true
killall ai_camera 2>/dev/null || true
killall ec-eeb001-gui 2>/dev/null || true
```

备份当前业务目录：

```sh
mkdir -p /opt/usr/backup
BACKUP=/opt/usr/backup/v01-opt-before-v02.01-$(date +%Y%m%d-%H%M%S).tar.gz
tar -C / -czf "$BACKUP" opt/bin opt/lib opt/inst
ls -lh "$BACKUP"
```

解包 v02.01 业务文件：

```sh
tar -C / -xzf /tmp/cc2-v02.01-opt.tar.gz
chmod 0755 /opt/bin/elegoo_printer /opt/bin/ai_camera /opt/bin/ec-eeb001-gui /opt/bin/run_printer.sh 2>/dev/null || true
chmod 0755 /opt/inst/daemon-000/daemon-000 2>/dev/null || true
sync
```

重启设备：

```sh
reboot
```

# 重启后确认

设备重启后，先确认 SSH 仍然可用：

```sh
ssh root@打印机IP
```

再确认打印业务进程是否启动：

```sh
ps | grep -E 'elegoo_printer|ai_camera|ec-eeb001-gui|daemon-000' | grep -v grep
```

确认打印入口仍然使用同一条 UDS 路径：

```sh
ps | grep elegoo_printer | grep -v grep
ls -l /tmp/elegoo_uds
```

正常情况下，`elegoo_printer` 仍应带有：

```text
-a /tmp/elegoo_uds
```

# 回滚方法

如果更新后业务程序异常，可以通过 SSH 登录后恢复备份：

```sh
/etc/init.d/printer stop || true
/opt/bin/run_printer.sh stop || true
killall elegoo_printer 2>/dev/null || true
killall ai_camera 2>/dev/null || true
killall ec-eeb001-gui 2>/dev/null || true

tar -C / -xzf /opt/usr/backup/你的备份文件.tar.gz
sync
reboot
```

# 注意事项

1. 这个方案的核心是保留 v01 系统层，只替换 v02.01 打印业务层。
2. 不建议直接把 v02.01 的整个 rootfs 覆盖到设备上，否则 SSH server 相关文件仍会被移除。
3. 不建议覆盖 `/etc/passwd`、`/etc/shadow`、`/etc/group`、`/etc/ssh`、`/usr/sbin/sshd`、`/etc/init.d/sshd`。
4. `/opt/usr/cfg/autosave.cfg` 是用户运行时配置，不在 v02.01 rootfs 的默认 `/opt/inst` 包内，正常不需要覆盖。
5. 如果设备已经有自己的校准、网络、账号或打印历史数据，不要批量覆盖 `/opt/usr`。
6. 出厂示例 G-code 体积较大，且不是运行 v02.01 打印业务的必要条件，可以按需排除。
7. 本文只描述基于 rootfs 差异的维护思路，实际执行前请确认设备电源稳定，并避免在打印过程中操作。
