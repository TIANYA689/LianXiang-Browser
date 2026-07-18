# 链享浏览器 v1.4.2 发布说明

发布日期：2026-07-18
平台：Windows amd64

## 本次更新

- 启用新的“联结 L”品牌图标，以蓝青双向链路表现多浏览器环境、代理连接与共享能力。
- Windows 主程序、安装程序、系统托盘、Linux 应用图标、前端 Logo 和 favicon 统一使用新图标。
- Windows 主图标提供 16、24、32、48、64、128、256 像素层级，系统托盘图标提供 16 至 64 像素层级。
- 图标保留透明圆角，改善标题栏、任务栏、桌面快捷方式和深浅色背景下的显示效果。

## Windows 产物

```text
LianXiang-Browser-Setup-1.4.2.exe
LianXiang-Browser-1.4.2-windows-amd64-portable.zip
```

SHA-256：

```text
9E39A9321F754A1DB2C21167AC44B7CBA407F6EC305652ABBF4F6D95B735DAE9  LianXiang-Browser-Setup-1.4.2.exe
0C215E44C07F21E614CAFACC15DE78F1126E4E649550CB80B965C1F6415297EC  LianXiang-Browser-1.4.2-windows-amd64-portable.zip
```

- 安装版包含应用本体、默认配置、Xray 和 sing-box；若发布目录存在有效 Windows Chromium 内核，也会一并打包。
- 便携版解压后直接运行目录中的 `lianxiang-browser.exe`。
- 两种版本首次启动都会初始化空的数据目录，不会打包开发环境的 `data/app.db`、账号、Cookie 或代理订阅。

## 升级说明

- 升级前退出链享浏览器，并关闭所有由其启动或待导入的 Chrome 窗口。
- 安装版可直接运行 v1.4.2 安装包覆盖升级；安装脚本会保留安装目录中的用户数据。
- 便携版建议先备份旧目录，再用新版本程序文件覆盖；不要删除原有 `data` 目录。
- v1.4.2 不新增数据库迁移，现有 v1.4.1 数据可直接使用。
- Windows 可能缓存旧快捷方式图标；重新固定快捷方式或刷新图标缓存后即可显示新图标。

## 验证范围

- Go 后端全量测试。
- TypeScript 检查与 Vite 生产构建。
- Windows 本地 EXE 构建、版本资源与嵌入图标检查。
- NSIS 安装包与便携 ZIP 构建。
- 产物文件名、文件清单、运行时哈希、敏感文件排除和 SHA-256 校验。

## 已知限制

- Windows 可执行文件和安装包尚未进行代码签名，系统可能显示未知发布者。
- Linux、macOS 发布脚本仍保留，但 v1.4.2 本次只验证 Windows amd64 产物。
- Xray、sing-box 和用户自行配置的 Chromium 内核属于第三方二进制，应独立核对来源和许可。
