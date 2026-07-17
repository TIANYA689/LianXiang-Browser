# 链享浏览器 v1.4.1 发布说明

发布日期：2026-07-18  
平台：Windows amd64

## 本次更新

- 修复 Chrome 用户数据导入时，Windows PowerShell 未正确接收目录参数并返回 `exit status 1` 的问题。
- 优化 Chrome 目录占用检测，只查询包含 `--user-data-dir` 的相关进程，并降低系统负载较高时的超时误报。
- PowerShell 查询失败时保留实际错误信息，便于继续定位 CIM、权限或系统服务问题。
- 延续 v1.4.0 的书签分组、HTML 书签导入、书签同步和列表布局优化。

## Windows 产物

```text
LianXiang-Browser-Setup-1.4.1.exe
LianXiang-Browser-1.4.1-windows-amd64-portable.zip
```

SHA-256：

```text
C3DBFC0DEAB2445DA6C8E23E63B0B090626CB48F20370A2197811BA8583C8E10  LianXiang-Browser-Setup-1.4.1.exe
B41AD68FA63681BBE66BFD8562E0F145EE5EC29CE0E06C9C5E4F7CAAA487B25F  LianXiang-Browser-1.4.1-windows-amd64-portable.zip
```

- 安装版包含应用本体、默认配置、Xray 和 sing-box；若发布目录存在有效 Windows Chromium 内核，也会一并打包。
- 便携版解压后直接运行目录中的 `lianxiang-browser.exe`。
- 两种版本首次启动都会初始化空的数据目录，不会打包开发环境的 `data/app.db`、账号、Cookie 或代理订阅。

## 升级说明

- 升级前退出链享浏览器，并关闭所有由其启动或待导入的 Chrome 窗口。
- 安装版可直接运行 v1.4.1 安装包覆盖升级；安装脚本会保留安装目录中的用户数据。
- 便携版建议先备份旧目录，再用新版本程序文件覆盖；不要删除原有 `data` 目录。
- v1.4.1 不新增数据库迁移，现有 v1.4.0 数据可直接使用。

## 验证范围

- Go 后端全量测试。
- TypeScript 检查与 Vite 生产构建。
- Windows 本地 EXE 构建及版本资源检查。
- NSIS 安装包与便携 ZIP 构建。
- 产物文件名、文件清单、运行时哈希和 SHA-256 校验。

## 已知限制

- Windows 可执行文件和安装包尚未进行代码签名，系统可能显示未知发布者。
- Linux、macOS 发布脚本仍保留，但 v1.4.1 本次只验证 Windows amd64 产物。
- Xray、sing-box 和用户自行配置的 Chromium 内核属于第三方二进制，应独立核对来源和许可。
