# 链享浏览器 v1.4.4 发布说明

发布日期：2026-07-19
平台：Windows amd64

## 本次更新

- 默认书签新增全局启用/关闭开关，关闭后不会写入新启动的浏览器实例。
- 支持只对指定实例关闭某条书签，其他实例不受影响。
- 实例启动时同时过滤书签栏内容和“启动打开”网址，避免关闭后仍自动打开对应网站。
- 保存书签后会同步已停止实例；运行中的实例会跳过，停止后再次启动即可按新设置生效。
- 关闭同步只移除由链享浏览器写入的书签节点，不会删除用户手工创建的同 URL 书签。
- SQLite 新增书签关闭状态和单实例关闭列表，并兼容配置文件、备份导入和旧数据库自动迁移。
- 延续新的“联结 L”品牌图标，Windows 主程序、安装程序、系统托盘和前端图标保持统一。

## Windows 产物

```text
LianXiang-Browser-Setup-1.4.4.exe
LianXiang-Browser-1.4.4-windows-amd64-portable.zip
SHA256SUMS-1.4.4.txt
```

SHA-256：

```text
2945D15C58BC961A9737B386A91F9E3D8D9575DDC5D0377EA2EB3214E4F16E9E  LianXiang-Browser-Setup-1.4.4.exe
16FB5EC404594C519580EF476C1D5FA353FE23355FA2E393FE3CDC330724681D  LianXiang-Browser-1.4.4-windows-amd64-portable.zip
```

同样的校验值已写入同目录的 `SHA256SUMS-1.4.4.txt`。

- 本地构建程序位于 `build\bin\lianxiang-browser.exe`。
- 安装版包含应用本体、默认配置、Xray 和 sing-box；若发布目录存在有效 Windows Chromium 内核，也会一并打包。
- 便携版解压后直接运行目录中的 `lianxiang-browser.exe`。
- 两种版本首次启动都会初始化空的数据目录，不会打包开发环境的 `data/app.db`、账号、Cookie、Token 或代理订阅。

## 升级说明

- 升级前退出链享浏览器，并停止需要更新书签状态的浏览器实例。
- 安装版可直接运行 v1.4.4 安装包覆盖升级；安装脚本会保留安装目录中的用户数据。
- 便携版建议先备份旧目录，再用新版本程序文件覆盖；不要删除原有 `data` 目录。
- 首次启动 v1.4.4 时会自动执行数据库迁移，增加书签关闭配置字段，已有实例、书签和用户数据会保留。
- 关闭书签后，运行中的实例不会被直接修改；停止并再次启动实例后即可生效。

## 验证范围

- Go 后端全量测试。
- TypeScript 检查与 Vite 生产构建。
- Windows 本地 EXE 构建、版本资源与嵌入图标检查。
- NSIS 安装包与便携 ZIP 构建。
- 产物文件名、文件清单、运行时哈希、敏感文件排除和 SHA-256 校验。

## 已知限制

- Windows 可执行文件和安装包尚未进行代码签名，系统可能显示未知发布者或 SmartScreen 提示。
- Linux、macOS 发布脚本仍保留，但 v1.4.4 本次只验证 Windows amd64 产物。
- Xray、sing-box 和用户自行配置的 Chromium 内核属于第三方二进制，应独立核对来源和许可。
