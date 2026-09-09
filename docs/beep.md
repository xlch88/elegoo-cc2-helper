# 蜂鸣器提示音程序说明

`elegoo-cc2-helper` 是一个运行在 Elegoo Centauri Carbon 2 上的小工具，用来通过机身蜂鸣器播放打印状态提示音。

它不会控制打印流程，也不会修改 G-code。程序只做两件事：

1. 通过 PWM 驱动蜂鸣器播放简短旋律。
2. 在服务模式下监听 `/tmp/elegoo_uds`，根据打印状态播放不同提示音。

# 一键安装

因为打印机系统里的 `wget` 只能下载普通 HTTP，不能直接下载 GitHub HTTPS 文件，所以一键安装命令需要在用户自己的电脑上执行，不是在打印机 SSH 里执行。

下面命令会在电脑上下载 GitHub 最新 release，然后通过 `scp` 上传到打印机，再通过 `ssh` 执行 `--install` 注册开机自启动。

> **重要：安装前必须先开启打印机 SSH。**  
> 如果您还没有开启 SSH，请先阅读：
>
> 1. [如何在 v02.01.00.00 固件上开启 SSH](./enable-ssh-on-v02.md)
> 2. [保留 SSH 并更新到 v02.01 打印固件](./keep-ssh-update.md)

> **重要：命令里的 `PRINTER_IP` / PowerShell 里输入的 `Printer IP` 必须换成您的打印机 IP 地址。**  
> 例如打印机屏幕上显示的 IP 是 `192.168.1.123`，就输入 `192.168.1.123`。

## Windows PowerShell

打开 PowerShell，整段复制粘贴执行：

```powershell
$PrinterIp = Read-Host "Printer IP"
$Url = "https://github.com/xlch88/elegoo-cc2-helper/releases/latest/download/elegoo-cc2-helper"
$LocalFile = Join-Path $env:TEMP "elegoo-cc2-helper"
Invoke-WebRequest -Uri $Url -OutFile $LocalFile
scp $LocalFile "root@${PrinterIp}:/tmp/elegoo-cc2-helper"
ssh "root@$PrinterIp" "chmod +x /tmp/elegoo-cc2-helper && /tmp/elegoo-cc2-helper --install"
```

执行过程中如果提示输入密码，就输入打印机 SSH 密码。

## Linux / macOS

打开终端，整段复制粘贴执行：

```sh
read -r -p "Printer IP: " PRINTER_IP
URL="https://github.com/xlch88/elegoo-cc2-helper/releases/latest/download/elegoo-cc2-helper"
TMP="/tmp/elegoo-cc2-helper"
if command -v curl >/dev/null 2>&1; then
  curl -L -o "$TMP" "$URL"
else
  wget -O "$TMP" "$URL"
fi
scp "$TMP" "root@$PRINTER_IP:/tmp/elegoo-cc2-helper"
ssh "root@$PRINTER_IP" "chmod +x /tmp/elegoo-cc2-helper && /tmp/elegoo-cc2-helper --install"
```

执行过程中如果提示输入密码，就输入打印机 SSH 密码。

# 功能

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

# 使用方式

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

# 服务模式

服务模式会连接 Klipper 兼容的 Unix socket：

```text
/tmp/elegoo_uds
```

程序会订阅 `print_stats.state` 状态变化，并根据状态播放提示音。

当前处理的状态逻辑如下：

| 状态变化                | 播放提示音  |
| ----------------------- | ----------- |
| 任意状态 -> `printing`  | `printing`  |
| `paused` -> `printing`  | `resumed`   |
| 任意状态 -> `paused`    | `paused`    |
| 任意状态 -> `cancelled` | `cancelled` |
| 任意状态 -> `complete`  | `complete`  |

服务刚连接时收到的第一个状态只作为基线，不播放打印状态提示音，避免程序启动后因为当前状态重复响一遍。

不过服务模式启动成功后会主动播放一次 `startup`，用于确认程序已经开始运行。

# PWM 输出

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

# 自启动安装

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

安装过程只新增这个程序自己的文件。如果目标路径已经存在不同内容，程序会拒绝覆盖，避免误改其他文件。

安装完成后不会立即启动服务，需要重启设备或手动运行服务脚本。

# 构建

源码位于 `src` 目录，可以使用项目内的构建脚本生成设备可执行文件：

```sh
cd src
./build.sh
```

构建产物会输出到：

```text
src/dist/elegoo-cc2-helper
```

当前构建目标是 Linux ARM，适合在 CC2 设备上运行。

# 注意事项

1. 程序需要能写入 PWM 节点，因此通常需要 root 权限运行。
2. 服务模式需要 `/tmp/elegoo_uds` 存在并可连接。
3. 程序只监听打印状态，不发送暂停、恢复、取消等控制命令。
4. 如果手动运行测试音，请确认周围环境允许蜂鸣器发声。
5. 如果设备固件路径或 PWM 节点发生变化，需要重新确认程序是否仍然适配。
