# elegoo-cc2-helper

[English](README_EN.md)

这是我折腾 ELEGOO Centauri Carbon 2（CC2）留下的代码和笔记：SSH、固件差异、主题切换，还有自己写的蜂鸣器提示音程序。

`elegoo-cc2-helper` 用 Go 编写，让打印机在开始、暂停、恢复、取消和完成打印时播放不同的提示音。

作者：[Dark495](https://github.com/xlch88)

## 文章目录

- [蜂鸣器提示音：安装与使用](docs/beep.md)
- [SSH 连接与默认密码](docs/ssh.md)
- [在 v02.01 固件上恢复 SSH](docs/enable-ssh-on-v02.md)
- [通过局域网上传 SSH 组件](docs/http-upload-ssh.md)
- [下载与解包 CC2 固件](docs/firmware-unpack.md)
- [保留 SSH，更新 v02.01 打印程序二进制文件](docs/keep-ssh-update.md)
- [关闭开机固件更新提示](docs/disable-ota-update.md)
- [切换 emoji 主题](docs/emoji-theme.md)

## 免责声明

个人项目，与 ELEGOO 官方无关。只在自己的或已获授权的设备上操作；改系统前做好备份，操作风险自行承担。各篇文章会注明适用版本和实测情况。

## 许可

Copyright © 2026 Dark495。原创代码、脚本和文档文字采用 [WTFPL v2](LICENSE)，随意使用和修改。第三方固件、界面素材、模型和商标不在这份授权范围内。
