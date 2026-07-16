# 链享浏览器二开安全审计记录

审计时间：2026-07-17

参考仓库：`https://github.com/black-ant/Ant-Browser`

## 结论

未在源码层发现明显后门、远控上报、未知私有域名数据回传或开机持久化恶意逻辑。

原项目存在可修复风险，链享浏览器当前已先做安全基线处理后再进入二开：

- 前端依赖审计已修复到 `npm audit` 0 vulnerabilities。
- Go 代码可达依赖漏洞已修复到 `govulncheck ./...` 0 vulnerabilities。
- Launch API 默认改为启用鉴权。
- API 鉴权配置为空时改为 fail-closed，不再静默放行 `/api/*` 请求。
- API Key 不写入仓库，运行时通过 `LIANXIANG_BROWSER_API_KEY` 环境变量注入。
- API Header 已改为 `X-LianXiang-Api-Key`。
- 构建脚本代理改为显式 `LIANXIANG_BUILD_PROXY` 配置，不再写入用户级 npm 代理配置。
- 应用名、Go module、前端包名、输出文件名、状态目录等已换成链享浏览器命名。

## 原仓库发现的风险

1. 前端生产依赖存在漏洞

原始 `npm audit --omit=dev` 发现 6 个生产依赖漏洞：4 个 high、2 个 moderate，主要涉及 `react-router-dom`、`@remix-run/router`、`js-yaml`、`lodash` 相关链路。

2. 开发依赖存在漏洞

安装依赖后完整 `npm audit` 还会暴露 Vite/esbuild、Rollup、glob/minimatch/picomatch、PostCSS 等开发工具链漏洞。链享浏览器已升级 Vite 到安全版本，并修复 Vite 8 对 `manualChunks` 的兼容变更。

3. Launch API 原默认无鉴权

原配置 `launch_server.auth.enabled: false`。服务绑定在 `127.0.0.1`，没有直接暴露局域网，但本机其他程序仍可调用 API。链享浏览器已改为默认请求鉴权。

4. 仓库内置代理运行时二进制

仓库包含 `xray.exe`、`sing-box.exe` 及 Linux/macOS 运行时二进制。源码审计无法证明这些二进制内容，只能核对哈希和来源清单。Windows 两个二进制当前哈希：

- `bin/xray.exe`: `103DA2750F4348A266AE61632C322F95CF3E18DCE99EB588E685379F041E97C5`
- `bin/sing-box.exe`: `94166C7C4142E4EB8DF6CED07208EF286DD700155B9C70024C14F7C57B09149F`

发布前建议从官方 Xray/sing-box Release 重新拉取并复核哈希。

5. Go 依赖代码可达漏洞

`govulncheck` 初次扫描发现 `golang.org/x/net@v0.35.0` 的 GO-2026-5026，调用路径进入代理运行代码。链享浏览器已升级到 `golang.org/x/net@v0.55.0`，并同步升级相关 `golang.org/x/*` 依赖；当前最低 Go 版本为 1.25。

## 已验证

- `npm audit`
- `npm run build`
- `go test ./...`
- `govulncheck ./...`
- Wails `v2.12.0` Windows/amd64 完整构建
- Wails 桌面程序启动冒烟测试
- YAML 配置解析：`config.yaml`、`publish/config.init.yaml`、`publish/config.init.mac.yaml`、`publish/config.init.linux.yaml`
- 品牌残留扫描：未发现 `Ant Browser`、`Ant Chrome`、`ant-chrome`、`ant-browser`、`X-Ant-Api-Key` 等旧标识残留

## 发布前仍需人工复核

- `bin/xray.exe`、`bin/sing-box.exe` 等内置二进制不属于源码审计范围，应从官方 Release 重新下载并核对来源及哈希。
- 不要把真实 API Key 写入 `config.yaml`、`.env.example` 或日志；运行 API 时设置 `LIANXIANG_BROWSER_API_KEY`。
- 当前安全扫描结论覆盖源码和依赖调用路径，不等同于对代理二进制内部行为的逆向审计。

Go 验证命令：

```powershell
Set-Location "D:\aaaaa\LianXiang-Browser"
gofmt -w backend\bootstrap.go backend\internal\config\config_io.go backend\internal\config\config_defaults.go backend\internal\launchcode\auth.go backend\internal\launchcode\auth_test.go backend\internal\fsutil\path.go
go test ./...
govulncheck ./...
```
