# 关闭开机固件更新提示

不想每次开机都被提醒升级，可以把 `/opt/inst/firmware_version/versions.json` 中的 `ota_version` 改成一个较高的版本号，例如 `04.95.04.95`。

修改的是系统显示和更新检查使用的版本号，**不会真的升级固件，也不是关闭后台更新请求**。

## 版本号格式

**必须使用 `xx.xx.xx.xx` 格式：四段数字，每段两位，用英文句点分隔。** 示例版本可以称为 v04.95.04.95，但写入 JSON 的值是 `04.95.04.95`，不能带 `v`。

以下是需要修改的字段，不要用这个片段覆盖整个 JSON 文件：

```text
"ota_version": "04.95.04.95"
```

不要写成 `v04.95.04.95`、`4.95.4.95` 或其他格式，否则可能导致更新检查反复失败，进而使 daemon 进程崩溃。只改 `ota_version`，不要改 `os_version`、`elegoo_version`、`canvas_version` 等其他字段。

## 修改步骤

打印机需要已经开启 SSH，连接方法见 [SSH 连接与默认密码](ssh.md)；v02.01 尚未开启 SSH 时，先看[恢复 SSH 教程](enable-ssh-on-v02.md)。在电脑终端或 PowerShell 中执行，**将 `PRINTER_IP` 换成打印机的实际 IP 地址**：

```sh
ssh root@PRINTER_IP
```

确认打印机空闲、没有正在更新固件后，在**打印机 SSH 终端**粘贴执行：

```sh
(
    set -eu
    cd /opt/inst/firmware_version
    if [ ! -e versions.json.before-ota ]; then
        cp -p versions.json versions.json.before-ota
    fi
    sed -i 's/"ota_version"[[:space:]]*:[[:space:]]*"[^"]*"/"ota_version": "04.95.04.95"/' versions.json
    grep -E '"ota_version"[[:space:]]*:[[:space:]]*"04\.95\.04\.95"' versions.json
    sync
)
```

这段命令只替换 `ota_version` 的值，其余字段不动；首次修改前会保存 `versions.json.before-ota`，重复执行不会覆盖这份备份。记下备份中的原始版本号，之后需要恢复时会用到。

确认没有报错，输出为 `"ota_version": "04.95.04.95"` 后，在同一个 SSH 终端重启打印机：

```sh
reboot
```

## 修改后的效果

重启后进入打印机的“设置 → 系统版本”，当前版本显示为 `04.95.04.95`，下方提示“暂无检测到更新版本”。下图是修改后的实际界面：

![修改 ota_version 后，系统显示 04.95.04.95，并提示暂无检测到更新版本](imgs/ota-version-04.95.04.95.png)

这里的版本号已经是手动填写的值，之后判断实际固件版本时不要再以它为依据。刷入固件或[重新同步 `/opt/inst`](keep-ssh-update.md#在设备上替换程序文件及配套资源) 也可能覆盖这次修改。

## 恢复更新提示

如果修改后没有再更新过固件或替换版本文件，在打印机 SSH 终端恢复原文件：

```sh
cd /opt/inst/firmware_version &&
    cp -p versions.json.before-ota versions.json &&
    sync
```

确认命令成功后执行 `reboot`。如果之后更新过打印程序二进制文件及配套文件，不要整份恢复旧备份，应只把 `ota_version` 改回当前实际安装的固件包所记录的版本号；查看原包中的 `versions.json`，可按[用 7-Zip 打开 rootfs](firmware-unpack.md#用-7-zip-打开-rootfswindows)操作。
