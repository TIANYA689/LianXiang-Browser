# 链享浏览器 v1.4.0 发布说明

发布日期：2026-07-18  
平台：Windows amd64

## 本次更新

- 默认书签支持分组和多级目录，`工作/AI` 会写入为真实的嵌套书签文件夹。
- 支持导入 Chrome、Edge、Firefox 导出的 HTML 书签文件，保留目录结构并跳过重复 URL。
- 书签分组完整接入 SQLite、配置文件、备份合并和已有未运行实例同步。
- 实例列表和代理池根据当前数据自动分配列宽，减少空白并避免中文表头逐字竖排。
- 保留 v1.3.1 的 Chrome 用户数据目录导入、占用检测和指纹变更风险提醒。

## Windows 产物

```text
LianXiang-Browser-Setup-1.4.0.exe
LianXiang-Browser-1.4.0-windows-amd64-portable.zip
```

SHA-256：

```text
D98CE690A55000B0B1C83E2EF503F738DBFD107559613B651A73F27DDC00E657  LianXiang-Browser-Setup-1.4.0.exe
C047B64B53B6E7263EBE1557975F936AA3948A36D26909E7017808F8E4EC8AE2  LianXiang-Browser-1.4.0-windows-amd64-portable.zip
```

- 安装版包含应用本体、默认配置、Xray 和 sing-box；本次未检测到可打包的 Windows Chromium 内核，因此 `chrome` 目录只包含配置说明。
- 便携版解压后直接运行目录中的 `lianxiang-browser.exe`。
- 两种版本首次启动都会初始化空的数据目录，不会打包开发环境的 `data/app.db`、账号、Cookie 或代理订阅。

## 升级说明

- 从旧版本升级时建议先退出链享浏览器和所有由其启动的浏览器实例。
- 安装版可直接运行新安装包覆盖升级；安装脚本会保留安装目录中的用户数据。
- 便携版建议先备份旧目录，再用新版本程序文件覆盖；不要误删原有 `data` 目录。
- 首次启动 v1.4.0 时会自动执行数据库迁移，为默认书签增加分组字段。

## 验证范围

- Go 后端全量测试。
- TypeScript 检查与 Vite 生产构建。
- Windows 本地 EXE 构建。
- NSIS 安装包与便携 ZIP 构建。
- 产物版本、文件清单、运行时哈希和 SHA-256 校验。

## 已知限制

- Windows 可执行文件和安装包尚未进行代码签名，系统可能显示未知发布者。
- Linux、macOS 发布脚本仍保留，但 v1.4.0 本次只验证 Windows amd64 产物。
- Xray、sing-box 和用户自行配置的 Chromium 内核属于第三方二进制，应独立核对来源和许可。
