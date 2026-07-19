# 链享浏览器 v1.4.5 发布说明

发布日期：2026-07-19
平台：Windows amd64

## 本次更新

- 新增独立“窗口同步器”页面，可从运行中的实例设置一个主窗口和多个被控窗口。
- 支持同步当前活动网页中的点击、Enter/Tab/Esc、输入内容、滚动位置和页面跳转。
- 同步器实时显示实例可用性、主控/被控角色、已转发事件数量、最近事件和连接错误。
- 支持从同步器恢复显示指定浏览器窗口；同步停止后可重新选择窗口和同步内容。
- 同步连接完全运行在本机 CDP 调试通道中，应用退出时会自动释放连接。
- 页面明确提示同步边界：Chrome 地址栏、书签栏、扩展弹窗、下载框和系统对话框不属于当前同步范围。

## Windows 产物

```text
LianXiang-Browser-Setup-1.4.5.exe
LianXiang-Browser-1.4.5-windows-amd64-portable.zip
SHA256SUMS-1.4.5.txt
```

SHA-256：

```text
B4E607EAF7C239560F301A3B43EE56A56C40CF07D233043D92C32AF6D6EB6164  LianXiang-Browser-Setup-1.4.5.exe
5583249FFF782DF0C12E086268B255C99953E396172CB7FB8C11B3A2D2B308DC  LianXiang-Browser-1.4.5-windows-amd64-portable.zip
```

同样的校验值已写入同目录的 `SHA256SUMS-1.4.5.txt`。

- 本地构建程序位于 `build\bin\lianxiang-browser.exe`。
- 安装版包含应用本体、默认配置、Xray 和 sing-box；若发布目录存在有效 Windows Chromium 内核，也会一并打包。
- 便携版解压后直接运行目录中的 `lianxiang-browser.exe`。
- 两种版本首次启动都会初始化空的数据目录，不会打包开发环境的 `data/app.db`、账号、Cookie、Token、代理订阅或 `.env`。

## 升级说明

- 升级前请停止窗口同步并退出链享浏览器；已打开的浏览器实例可按需保留或关闭。
- 安装版可直接运行 v1.4.5 安装包覆盖升级；安装脚本会保留安装目录中的用户数据。
- 便携版建议先备份旧目录，再用新版本程序文件覆盖；不要删除原有 `data` 目录。
- v1.4.5 不新增数据库迁移，现有 v1.4.4 实例、书签、代理、插件和用户数据可直接使用。

## 验证范围

- Go 后端全量测试与静态检查。
- 双 Chrome 实例的输入、点击和滚动真实同步测试。
- TypeScript 检查与 Vite 生产构建。
- Windows 本地 EXE 构建、版本资源与嵌入图标检查。
- NSIS 安装包与便携 ZIP 构建。
- 产物文件名、文件清单、运行时哈希、敏感文件排除和 SHA-256 校验。

## 已知限制

- 当前同步范围是浏览器活动网页，不包括 Chrome 地址栏、书签栏、扩展弹窗、下载框和系统对话框。
- 页面结构差异较大或窗口尺寸差异明显时，基于页面坐标的点击同步可能出现偏差。
- Windows 可执行文件和安装包尚未进行代码签名，系统可能显示未知发布者或 SmartScreen 提示。
- Linux、macOS 发布脚本仍保留，但 v1.4.5 本次只验证 Windows amd64 产物。
- Xray、sing-box 和用户自行配置的 Chromium 内核属于第三方二进制，应独立核对来源和许可。
