# CC2 固件默认 SSH 密码

已升级到 v02.01 且 SSH 不可用，可先看[在 v02.01 上开启 SSH](./enable-ssh-on-v02.md)；仍在 v01、希望保留 SSH 更新打印程序二进制文件及配套文件，则看[保留 v01 SSH 环境](./keep-ssh-update.md)。

打印机已开启 SSH 后，在电脑终端或 PowerShell 中执行，**把 `PRINTER_IP` 换成打印机的实际 IP 地址**：

```sh
ssh root@PRINTER_IP
```

下面是社区记录的默认密码，适用固件版本尚未核对；如果已经修改过密码，请使用自己的密码。

```text
MTY4ODE2
```

参考：[Discord 社区消息](https://discord.com/channels/1367538416539013122/1434248003459354684/1488450815801819236)。

登录后建议在打印机 SSH 终端执行下面的命令修改默认密码，保留当前连接，另开一个终端确认新密码能够登录后再退出：

```sh
passwd
```

确认 SSH 可用后，可以按各篇说明[安装蜂鸣器提示音](./beep.md#一键安装)、[通过 SSH 切换 emoji 主题](./emoji-theme.md#方法二通过-ssh-执行指令)，或[关闭开机更新提示](./disable-ota-update.md#修改步骤)。
