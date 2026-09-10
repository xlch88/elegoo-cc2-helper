# 蜂鸣器提示音程序说明

`elegoo-cc2-helper` 是一个运行在 ELEGOO Centauri Carbon 2 上的小工具，用来通过机身蜂鸣器播放打印状态提示音。

它不会控制打印流程，也不会修改 G-code。程序只做两件事：

1. 通过 PWM 驱动蜂鸣器播放简短旋律。
2. 在服务模式下监听 `/tmp/elegoo_uds`，根据打印状态播放不同提示音。

## 一键安装

因为打印机系统里的 `wget` 只能下载普通 HTTP，不能直接下载 GitHub HTTPS 文件，所以一键安装命令需要在用户自己的电脑上执行，不是在打印机 SSH 里执行。

下面命令会在电脑上下载 GitHub 最新 release，然后通过 `scp` 上传到打印机，再通过 `ssh` 执行 `--install` 注册开机自启动。

安装会覆盖本程序已有的文件，可用于更新；不会自动启动或重启服务，手动启动方法见下方[自启动安装](#自启动安装)。

> **重要：安装前必须先开启打印机 SSH。**  
> 按当前固件选择对应说明：
>
> 1. 已升级到 v02.01：[如何在 v02.01.00.00 固件上开启 SSH](./enable-ssh-on-v02.md)。
> 2. 仍在 v01，准备更新打印程序：[保留 v01 SSH 环境，更新 v02.01 打印程序二进制文件](./keep-ssh-update.md)。
>
> 连接命令和密码见 [SSH 连接与默认密码](./ssh.md)。

> **重要：运行后出现 `Printer IP` 提示时，输入您的打印机 IP 地址。**
> 例如打印机屏幕上显示的 IP 是 `192.168.1.123`，就输入 `192.168.1.123`。

### Windows PowerShell

打开 PowerShell 5.1 或更高版本，确认 `ssh` 和 `scp` 可用，整段复制粘贴执行：

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

执行过程中如果提示输入密码，就输入打印机 SSH 密码。

### Linux / macOS

打开 Bash 或 zsh 终端，确认 `ssh`、`scp` 和支持 HTTPS 的 `curl` 或 `wget` 可用，整段复制粘贴执行：

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

执行过程中如果提示输入密码，就输入打印机 SSH 密码。

## 功能

当前内置提示音包括：

| 提示音      | 触发场景               |
| ----------- | ---------------------- |
| `startup`   | 程序启动服务模式时播放 |
| `printing`  | 打印开始时播放         |
| `paused`    | 打印暂停时播放         |
| `resumed`   | 从暂停恢复打印时播放   |
| `cancelled` | 打印取消时播放         |
| `complete`  | 打印完成时播放         |

另外还提供一个测试音乐模式，用于播放一小段《致爱丽丝》开头旋律，方便确认蜂鸣器和 PWM 输出是否正常。

## 使用方式

以下命令在打印机 SSH 终端执行。安装后先进入程序目录：

```sh
cd /opt/usr/elegoo-cc2-helper
```

直接运行程序或使用 `-h`、`--help` 会显示帮助信息：

```sh
./elegoo-cc2-helper
./elegoo-cc2-helper -h
./elegoo-cc2-helper --help
```

测试所有提示音：

```sh
./elegoo-cc2-helper -t
./elegoo-cc2-helper --test
```

测试音乐：

```sh
./elegoo-cc2-helper --test-music
```

以服务模式运行：

```sh
./elegoo-cc2-helper -d
./elegoo-cc2-helper --daemon
```

注册开机自启动：

```sh
./elegoo-cc2-helper --install
```

## 服务模式

服务模式会连接 Klipper 兼容的 Unix socket：

```text
/tmp/elegoo_uds
```

程序会订阅 `print_stats.state` 状态变化，并根据状态播放提示音。

状态发生变化时，处理逻辑如下：

| 状态变化                           | 播放提示音  |
| ---------------------------------- | ----------- |
| 除 `paused` 外的状态 -> `printing` | `printing`  |
| `paused` -> `printing`             | `resumed`   |
| 任意状态 -> `paused`               | `paused`    |
| 任意状态 -> `cancelled`            | `cancelled` |
| 任意状态 -> `complete`             | `complete`  |

服务刚连接时收到的第一个状态只作为基线，不播放打印状态提示音，避免程序启动后因为当前状态重复响一遍。

服务模式连接到 socket 后、发送订阅请求前，会主动播放一次 `startup`。

## PWM 输出

程序直接使用系统 PWM 节点控制蜂鸣器：

```text
/sys/class/pwm/pwmchip0/pwm0/enable
/sys/class/pwm/pwmchip0/pwm0/period
/sys/class/pwm/pwmchip0/pwm0/duty_cycle
```

播放每个音符时，程序会：

1. 关闭 PWM 输出。
2. 根据音符频率计算 `period`。
3. 将 `duty_cycle` 设置为 `period / 2`。
4. 打开 PWM 输出。
5. 等待音符时长结束。
6. 再次关闭 PWM 输出。

因为占空比已经使用 50%，音量主要受蜂鸣器硬件、驱动电路和机器结构限制，程序侧能调节的空间有限。

## 自启动安装

执行 `--install` 时，程序会把当前二进制复制到：

```text
/opt/usr/elegoo-cc2-helper/elegoo-cc2-helper
```

并创建 procd 启动脚本：

```text
/etc/init.d/elegoo-cc2-helper
```

同时创建启动和停止链接：

```text
/etc/rc.d/S99elegoo-cc2-helper
/etc/rc.d/K10elegoo-cc2-helper
```

安装会覆盖以上位置的旧程序、启动脚本和启动链接，不修改其他系统启动文件。程序和脚本先写入临时文件，再替换目标文件，避免直接截断正在运行的程序。

安装完成后不会自动启动或重启服务。首次安装后可以重启设备或手动运行服务脚本；更新时，已经运行的进程会继续使用旧版本，需要手动重启服务才会运行新版本。

在打印机 SSH 终端启动服务：

```sh
/etc/init.d/elegoo-cc2-helper start
```

更新已运行的服务后，在打印机 SSH 终端执行：

```sh
/etc/init.d/elegoo-cc2-helper restart
```

## 构建

源码位于 `src` 目录，在电脑的项目根目录运行构建脚本，生成设备可执行文件：

```sh
cd src
./build.sh
```

构建产物会输出到：

```text
src/dist/elegoo-cc2-helper
```

当前构建目标是 Linux ARM，适合在 CC2 设备上运行。

## 注意事项

1. 程序需要能写入 PWM 节点，因此通常需要 root 权限运行。
2. 服务模式需要 `/tmp/elegoo_uds` 存在并可连接。
3. 程序只监听打印状态，不发送暂停、恢复、取消等控制命令。
4. 如果手动运行测试音，请确认周围环境允许蜂鸣器发声。
5. 如果设备固件路径或 PWM 节点发生变化，需要重新确认程序是否仍然适配。
