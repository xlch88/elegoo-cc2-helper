# 保留 v01 SSH 环境，更新 v02.01 打印程序二进制文件

在 v01 固件中，SSH 默认可用，可以直接连接设备进行调试和维护。

v02.01 的系统层移除了 SSH 服务端相关组件，不是单纯关闭服务，而是缺少了 `/usr/sbin/sshd`、`/etc/init.d/sshd`、`/etc/ssh/sshd_config`、`/usr/lib/sftp-server` 等文件，因此完整升级后就无法继续通过原来的 SSH 服务连接打印机。

如果已经完整升级到 `02.01.00.00` 且需要恢复 SSH，改看[在 v02.01 固件上开启 SSH](./enable-ssh-on-v02.md)；本文针对仍保留 v01 系统的设备。

这里采用的做法是：保留 v01 的系统层和 SSH 环境，只把 v02.01 的打印程序二进制文件及配套的库、配置和界面资源同步到设备上。这是打印程序更新，不是整包升级到 v02.01 固件。

**本文基于 `01.03.02.51` 与 `02.01.00.00` 的 rootfs 对比结果编写；下面的整套更新流程尚未实机验证，其他版本需要重新比对。**

## 差异结论

### SSH 相关差异

v01 中存在完整的 OpenSSH 服务端组件：

| 路径                                 | 说明                                     |
| ------------------------------------ | ---------------------------------------- |
| `/usr/sbin/sshd`                     | SSH 服务端主程序                         |
| `/etc/init.d/sshd`                   | procd 启动脚本，启动 `/usr/sbin/sshd -D` |
| `/etc/ssh/sshd_config`               | SSH 服务端配置                           |
| `/usr/lib/sftp-server`               | SFTP 子系统                              |
| `/lib/upgrade/keep.d/openssh-server` | 升级时保留 SSH 主机密钥的列表            |
| `/etc/group` 中的 `sshd:x:22:sshd`   | sshd 组                                  |

v02.01 中仍保留了 SSH 客户端、`ssh-keygen`、`/etc/ssh/ssh_config` 和 `/etc/passwd` 里的 `sshd` 用户，但缺少服务端文件和 `sshd` 组。

这说明 v02.01 不是缺少所有 SSH 运行环境，而是移除了对外提供登录服务所需的服务端组件。

### 打印启动链差异

两版固件的打印入口保持一致：

```text
/etc/init.d/printer
  -> /opt/bin/run_printer.sh start
  -> elegoo_printer /opt/inst/printer_dsp.cfg -s /opt/usr/cfg/autosave.cfg -a /tmp/elegoo_uds
```

对比结果中，`/etc/init.d/printer` 和 `/opt/bin/run_printer.sh` 内容一致。因此，要更新打印程序，重点不在系统 init 脚本，而在 `/opt` 下面的程序二进制文件、库、配置和资源。

### v02.01 主要变化位置

对比 `01.03.02.51` 和 `02.01.00.00` 后，和打印程序直接相关的变化主要集中在这些位置：

| 路径                                         | 变化                    |
| -------------------------------------------- | ----------------------- |
| `/opt/bin/elegoo_printer`                    | 主打印程序更新          |
| `/opt/bin/ai_camera`                         | 摄像头/视觉相关程序更新 |
| `/opt/bin/ec-eeb001-gui`                     | 屏幕 GUI 程序更新       |
| `/opt/inst/daemon-000/daemon-000`            | 后台程序更新            |
| `/opt/inst/daemon-000/update_whitelist.conf` | v02.01 新增             |
| `/opt/inst/printer_dsp.cfg`                  | 打印配置更新            |
| `/opt/lib/libapi_module.a`                   | 依赖库更新              |
| `/opt/lib/libcommunication.a`                | 通信库更新              |
| `/opt/lib/libelegoo_extras.so`               | 扩展库更新              |
| `/opt/lib/libturbo_core.so`                  | 依赖库更新              |
| `/opt/inst/images/`                          | 部分界面资源更新        |
| `/opt/inst/factory/`                         | 出厂示例 G-code 有变化  |

为了减少遗漏，建议按目录整体同步 `/opt/bin`、`/opt/lib`、`/opt/inst`，不要只挑几个二进制文件复制。

## 推荐方案

总体思路如下：

1. 设备保持 v01 固件启动。
2. 通过 v01 的 SSH 登录设备。
3. 从 v02.01 rootfs 中打包 `/opt/bin`、`/opt/lib`、`/opt/inst`。
4. 上传到设备。
5. 停止打印相关进程。
6. 备份设备当前 `/opt` 下的三个目录，以及实际使用的图片和字体。
7. 将新包解到临时目录，再同步程序和资源，清除旧版残留文件。
8. 重启设备并确认 SSH 仍可连接。

这个方案不替换 `/usr/sbin/sshd`、`/etc/init.d/sshd`、`/etc/ssh/sshd_config` 等 v01 的 SSH 文件，也不替换整个 rootfs。

## 在电脑上准备 v02.01 打印程序更新包

从官方渠道获取并解包对应型号的 v02.01 固件，保留文件权限和符号链接；解包步骤见[固件解包说明中的 Linux 提取方法](./firmware-unpack.md#在-linux-中提取-rootfs可选)。假设解包目录为：

```text
/path/to/cc2-v02.01-rootfs
```

在电脑的 Linux Bash 环境执行，需要 GNU tar 和 `sha256sum`；将路径替换为自己的解包目录：

```bash
(
  set -eu
  cd /path/to/cc2-v02.01-rootfs
  sudo tar --numeric-owner -czf /tmp/cc2-v02.01-opt.tar.gz opt/bin opt/lib opt/inst
  cd /tmp
  tar -tzf cc2-v02.01-opt.tar.gz >/dev/null
  sha256sum cc2-v02.01-opt.tar.gz > cc2-v02.01-opt.tar.gz.sha256
)
```

如果不想带出厂示例 G-code，可以改用下面这段，包会小很多；后面的更新命令会保留设备原有的 `factory` 目录：

```bash
(
  set -eu
  cd /path/to/cc2-v02.01-rootfs
  sudo tar --numeric-owner -czf /tmp/cc2-v02.01-opt.tar.gz \
    --exclude='opt/inst/factory' opt/bin opt/lib opt/inst
  cd /tmp
  tar -tzf cc2-v02.01-opt.tar.gz >/dev/null
  sha256sum cc2-v02.01-opt.tar.gz > cc2-v02.01-opt.tar.gz.sha256
)
```

## 上传到 v01 设备

先按 [SSH 连接说明](./ssh.md)连接打印机并检查空间，**将 `PRINTER_IP` 替换为打印机 IP**：

```sh
ssh root@PRINTER_IP
```

在打印机 SSH 中执行：

```sh
df -h / /opt/usr /tmp
du -sh /opt/bin /opt/lib /opt/inst /opt/usr/images /opt/usr/fonts
command -v rsync
command -v sha256sum
```

`/tmp` 要放得下上传包，`/opt/usr` 要同时容纳解包目录和备份，系统分区也要留出更新余量。按解包后的大小估算，不要只看压缩包大小。

回到电脑终端，将包和校验文件一起上传：

```sh
scp /tmp/cc2-v02.01-opt.tar.gz /tmp/cc2-v02.01-opt.tar.gz.sha256 root@PRINTER_IP:/tmp/
```

## 在设备上替换程序文件及配套资源

下面的命令在**打印机 SSH 中执行**，依次校验上传包、停止进程、备份并同步文件。打印机应处于空闲状态，配置和数据库请先另存一份到电脑。

原来的停止脚本漏掉了 `daemon-000`，这里逐个停止相关进程并检查退出结果。备份包含程序目录和 `/opt/usr` 下实际使用的图片、字体；它不包含打印历史、数据库等用户数据。

```sh
(
  set -eu
  command -v rsync >/dev/null
  grep -q ' /opt/usr ' /proc/mounts
  cd /tmp
  sha256sum -c cc2-v02.01-opt.tar.gz.sha256

  WORK=$(mktemp -d /opt/usr/cc2-update.XXXXXX)
  trap 'rm -rf "$WORK"' 0
  tar -C "$WORK" -xzpf /tmp/cc2-v02.01-opt.tar.gz
  for dir in bin lib inst; do
    test -d "$WORK/opt/$dir"
    test ! -L "$WORK/opt/$dir"
  done
  test -d "$WORK/opt/inst/images"
  test ! -L "$WORK/opt/inst/images"
  test -d "$WORK/opt/inst/fonts"
  test ! -L "$WORK/opt/inst/fonts"
  for dir in bin lib inst usr/images usr/fonts; do
    test -d "/opt/$dir"
    test ! -L "/opt/$dir"
  done

  PROCESSES="daemon-000 elegoo_printer ai_camera ec-eeb001-gui mosquitto eeb001-factory"
  for name in $PROCESSES; do
    if pidof "$name" >/dev/null; then killall "$name"; fi
  done
  sleep 2
  if pidof $PROCESSES >/dev/null; then
    echo "Printer processes are still running" >&2
    exit 1
  fi

  mkdir -p /opt/usr/backup
  BACKUP=$(mktemp -d /opt/usr/backup/v01-before-v02.01.XXXXXX)
  tar -C / -czf "$BACKUP/opt.tar.gz" opt/bin opt/lib opt/inst opt/usr/images opt/usr/fonts
  tar -tzf "$BACKUP/opt.tar.gz" >/dev/null
  (cd "$BACKUP" && sha256sum opt.tar.gz > opt.tar.gz.sha256)
  printf 'Backup: %s\n' "$BACKUP"

  if pidof $PROCESSES >/dev/null; then
    echo "Printer processes restarted; update stopped" >&2
    exit 1
  fi
  for dir in bin lib; do
    rsync -ac --delete "$WORK/opt/$dir/" "/opt/$dir/"
  done
  if [ -d "$WORK/opt/inst/factory" ]; then
    rsync -ac --delete "$WORK/opt/inst/" /opt/inst/
  else
    rsync -ac --delete --exclude='/factory/' "$WORK/opt/inst/" /opt/inst/
  fi
  rsync -ac --delete /opt/inst/images/ /opt/usr/images/
  rsync -ac --delete /opt/inst/fonts/ /opt/usr/fonts/
  sync
  echo "Update complete"
)
```

这里使用 `rsync --delete` 清除旧版残留，而不是直接叠加解压。它也会移除目标目录中的自定义文件，所以先做备份；校准、账号和打印记录所在的其他 `/opt/usr` 目录不参与同步。

固件只在 `ota_flag` 为空、为 `true` 或缺少默认主题目录时自动同步图片和字体。上面直接同步实际使用目录，不依赖这个开机条件，也不修改 `ota_flag`。

记录输出的 `Backup` 路径，并在电脑上保存一份，例如将下面路径中的 `XXXXXX` 换成实际目录名：

```sh
scp -r root@PRINTER_IP:/opt/usr/backup/v01-before-v02.01.XXXXXX ./
```

**看到 `Update complete` 且备份已保存后**，在打印机 SSH 中重启；若中途报错，先按[回滚方法](#回滚方法)恢复，不要直接重启：

```sh
reboot
```

## 重启后确认

设备重启后，先确认 SSH 仍然可用：

```sh
ssh root@PRINTER_IP
```

再确认打印相关进程是否启动：

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

更新后若要切换界面主题，参见[通过 SSH 设置 emoji 主题](./emoji-theme.md#方法二通过-ssh-执行指令)。如果之前改过 `ota_version`，同步 `/opt/inst` 后的影响见[更新提示与手写版本号说明](./disable-ota-update.md#修改后的效果)。

## 回滚方法

如果更新后打印程序异常，在打印机 SSH 中恢复备份。把 `BACKUP` 改为前面记录的目录；先校验、解包到空目录，再同步回去，避免留下新版新增文件：

```sh
(
  set -eu
  BACKUP=/opt/usr/backup/v01-before-v02.01.REPLACE_ME
  command -v rsync >/dev/null
  grep -q ' /opt/usr ' /proc/mounts
  (cd "$BACKUP" && sha256sum -c opt.tar.gz.sha256)
  WORK=$(mktemp -d /opt/usr/cc2-rollback.XXXXXX)
  trap 'rm -rf "$WORK"' 0
  tar -C "$WORK" -xzpf "$BACKUP/opt.tar.gz"
  for dir in bin lib inst usr/images usr/fonts; do
    test -d "$WORK/opt/$dir"
    test ! -L "$WORK/opt/$dir"
    test -d "/opt/$dir"
    test ! -L "/opt/$dir"
  done

  PROCESSES="daemon-000 elegoo_printer ai_camera ec-eeb001-gui mosquitto eeb001-factory"
  for name in $PROCESSES; do
    if pidof "$name" >/dev/null; then killall "$name"; fi
  done
  sleep 2
  if pidof $PROCESSES >/dev/null; then
    echo "Printer processes are still running" >&2
    exit 1
  fi
  for dir in bin lib inst usr/images usr/fonts; do
    rsync -ac --delete "$WORK/opt/$dir/" "/opt/$dir/"
  done
  sync
  echo "Rollback complete"
)
```

出现 `Rollback complete` 表示程序、图片和字体已恢复。若新版修改过数据库或配置，先用更新前另存的用户数据备份恢复对应文件，再执行 `reboot`。

## 注意事项

1. 这个方案的核心是保留 v01 系统层，只替换 v02.01 打印程序二进制文件及配套的库、配置和界面资源。
2. 不要用 v02.01 的整个 rootfs 替换设备系统，否则仍会失去原有 SSH 服务端。
3. 不建议覆盖 `/etc/passwd`、`/etc/shadow`、`/etc/group`、`/etc/ssh`、`/usr/sbin/sshd`、`/etc/init.d/sshd`。
4. `/opt/usr/cfg/autosave.cfg` 是用户运行时配置，不在 v02.01 rootfs 的默认 `/opt/inst` 包内，正常不需要覆盖。
5. 如果设备已经有自己的校准、网络、账号或打印历史数据，不要批量覆盖 `/opt/usr`。
6. 出厂示例 G-code 体积较大，且不是运行 v02.01 打印程序的必要条件，可以按需排除。
7. 本文只描述基于 rootfs 差异的维护思路，实际执行前请确认设备电源稳定，并避免在打印过程中操作。
