# 获取并解包 CC2 固件

这篇介绍如何在电脑上打开 CC2 固件，查看里面的程序和配置。**以 Windows 为例，浏览器解包器配合 7-Zip 就够了**，Linux 命令行步骤也保留在后面。封装格式参考 [OpenCentauri 软件文档](https://docs.opencentauri.cc/software/)中的 CC2 专页。

## 获取固件

优先使用 ELEGOO 提供的对应型号原包；历史版本也可以在 [OpenCentauri CC2 固件归档](https://docs.opencentauri.cc/software/updates-cc2/#firmware-update-archive)中选择版本，点击 **Download** 下载。

- 确认是 **Centauri Carbon 2**，并核对版本及地区；普通的 [Updates 页面](https://docs.opencentauri.cc/software/updates/)介绍的是 CC1，不能把它的固件或 `unpack.py` 命令直接套到 CC2 上。
- 归档链接指向社区保存的文件，不是 ELEGOO 官方下载站；做原版对比时不要选标为 **repacked** 的重新打包版本。
- 截至 2026-09-10，该归档页未列出 `02.01.00.00`；要跟做 v02.01 打印程序更新，需要另行取得该版本原包，不能拿 `02.00.02.00` 代替。

## 去掉两层 `.sig` 封装

CC2 的常见更新包结构是：

```text
firmware.zip.sig
  -> firmware.zip
      -> firmware.swu.sig
          -> firmware.swu
      -> firmware.json.sig
```

在**电脑浏览器**中打开 [CC2 专用解包器](https://docs.opencentauri.cc/extras/cc2_update_decrypt.html)：

1. 选择下载的 `.zip.sig`，点击 **Unpack**，保存输出的 `.zip`。
2. 用 7-Zip 打开这个 `.zip`，取出其中的 `.swu.sig`。
3. 再把 `.swu.sig` 交给同一个解包器，保存输出的 `.swu`。

这个工具每次只处理一层 `.sig`，不会自动展开 ZIP；页面显示的 **Version** 是封装格式版本，不是打印机固件版本，具体行为见[解包器源码](https://github.com/OpenCentauri/OpenCentauri/blob/main/docs/extras/cc2_update_decrypt.html)。

如果手头已经是 `.swu.sig`，从第 3 步开始；已经是解密后的 `.swu`，直接进入下一节。旁边的 `.json.sig` 是元数据，不是 rootfs 镜像。

## 用 7-Zip 打开 rootfs（Windows）

电脑安装 [7-Zip](https://www.7-zip.org/) 后，继续操作，不需要安装 WSL 或 Linux：

1. 右键刚才得到的 `.swu` 文件，选择 **7-Zip → 打开压缩包**。
2. 在里面找到名为 **`rootfs`** 的文件，选中它，点击 **提取**，保存到电脑上的文件夹，例如 `D:\CC2`。
3. 右键提取出来的 **`rootfs` 文件**，再次选择 **7-Zip → 打开压缩包**。它没有扩展名，直接打开即可，不用改后缀。
4. 现在就能浏览固件里的目录和文件，需要哪个文件就选中并提取出来。

Windows 11 的右键菜单如果没有 7-Zip，先点“显示更多选项”；也可以打开 **7-Zip File Manager**，在里面找到文件并打开。

整个过程就是：

```text
.swu 文件
  → 用 7-Zip 打开，提取 rootfs
    → 再用 7-Zip 打开 rootfs，查看固件内容
```

常用文件都在 `opt` 目录下：

```text
opt/bin/                                  打印程序
opt/lib/                                  程序依赖库
opt/inst/                                 默认配置、图片和其他资源
opt/inst/firmware_version/versions.json    固件版本记录
```

例如，把 `versions.json` 提取出来，用记事本打开，就能查看其中的 `model` 和 `ota_version`。这里查看的是原包记录，不需要修改版本号；设备上手写版本号的情况，见[更新提示与版本号的说明](./disable-ota-update.md#修改后的效果)。如果要导出整个 rootfs，选一个新目录，例如 `D:\CC2\rootfs-files`，不要与原来的 `rootfs` 文件同名。

**只是查看或取出文件，到这里就完成了。** 如果还要制作打印程序更新包，请使用[下面的 Linux 提取方法](#在-linux-中提取-rootfs可选)，从原包提取打印程序二进制文件及配套的库、配置和界面资源，保留文件权限、所有者和符号链接。

## 在 Linux 中提取 rootfs（可选）

下面在**电脑的 Linux Bash 终端**执行，Windows 可用 **WSL2 Ubuntu**，不要在打印机 SSH 中运行。解包目录放在 Linux 文件系统中，例如 WSL 的家目录，不要把最终 rootfs 解到 `/mnt/c`，以免丢失 Linux 权限或符号链接。

需要 GNU cpio、`file` 和支持 XZ 的 `unsquashfs`。Ubuntu / Debian 可通过系统软件源安装：

```bash
sudo apt-get update && sudo apt-get install cpio squashfs-tools file
```

可用 `unsquashfs -version` 查看工具版本；下文参数参照 [Squashfs-tools 4.7.5 手册](https://github.com/plougher/squashfs-tools/blob/master/Documentation/4.7.5/USAGE-UNSQUASHFS.md)，并不要求恰好使用这个版本。

将代码中的 `/path/to/firmware.swu` 替换为刚才保存的 **`.swu` 文件绝对路径**。这段先从 SWU 中取出 `rootfs` 镜像，再将其展开到新建目录：

```bash
(
  set -eu
  SWU="/path/to/firmware.swu"
  test -f "$SWU"
  WORK=$(mktemp -d "$HOME/cc2-unpack.XXXXXX")
  mkdir "$WORK/swu"
  cd "$WORK/swu"
  cpio -id --no-absolute-filenames rootfs sw-description < "$SWU"
  test -f rootfs
  file rootfs
  unsquashfs -s rootfs
  sudo unsquashfs -d "$WORK/rootfs" rootfs
  test -x "$WORK/rootfs/opt/bin/elegoo_printer"
  test -d "$WORK/rootfs/opt/lib"
  test -f "$WORK/rootfs/opt/inst/printer_dsp.cfg"
  grep -E '"(model|ota_version)"' "$WORK/rootfs/opt/inst/firmware_version/versions.json"
  printf 'Rootfs directory: %s\n' "$WORK/rootfs"
)
```

已检查的 `02.01.00.00` 包中，`rootfs` 是 **SquashFS 4.0 / XZ** 镜像；它没有扩展名，但不是目录。`unsquashfs` 无需挂载镜像即可展开，使用 `sudo` 是为了保留原始所有者、权限及特殊文件。

命令报错时不要把半成品拿去打包；尤其是找不到 `rootfs` 或格式识别失败时，先检查是否还停留在 `.zip`、`.swu.sig` 那一层。

## 接着准备打印程序更新包

使用上面的 Linux 命令后，会输出一个类似这样的目录：

```text
/home/yourname/cc2-unpack.ABC123/rootfs
```

其中应有 `opt/bin`、`opt/lib`、`opt/inst` 等目录。核对输出中的 `model` 和 `ota_version`；如果要继续[保留 v01 SSH 环境，更新 v02.01 打印程序二进制文件](./keep-ssh-update.md)，该版本应为 `02.01.00.00`。

将那篇文章中的 `/path/to/cc2-v02.01-rootfs` 换成这里输出的 **rootfs 目录**，不是 `.swu` 文件，也不是 `swu/rootfs` 镜像文件，就可以继续打包。

本文命令已做语法与资料核对，本次未重新下载固件或执行完整解包。
