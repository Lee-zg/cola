# Cola Bookmarks

Cola Bookmarks 是一个本地优先的桌面书签管理器，使用 Go、Wails v2、Vue 3 和 TypeScript 构建。

![CI](https://github.com/Lee-zg/cola/actions/workflows/ci.yml/badge.svg)
![CodeQL](https://github.com/Lee-zg/cola/actions/workflows/codeql.yml/badge.svg)
![Security](https://github.com/Lee-zg/cola/actions/workflows/security.yml/badge.svg)

## 功能特性

- 支持导入 Netscape 书签 HTML、Chrome、Edge、Firefox 书签。
- 支持书签新增、编辑、删除、搜索、分类和标签管理。
- 支持离线规则分析，自动生成标签、关键词和别名。
- 所有书签数据保存在本地 SQLite 数据库。
- 支持本地数据库备份与恢复。
- 可启动 `127.0.0.1` 只读 Web 目录，支持搜索、分类和标签。
- 支持导出离线静态 HTML 目录。
- 导出/Web 主题包架构已预留，支持 `theme.json + CSS + assets`。

## 项目状态

当前版本：`0.1.1`

应用目前处于早期 MVP 阶段。存储、导入导出、本地 Web 目录和治理流程已具备；更丰富的 AI 模型下载与第三方主题生态属于后续工作。

## 开发环境

依赖要求：

- Go 1.25+
- Wails v2.12+
- Node.js 22.13+（`PATH` 中可用 `npm`）
- Windows 平台需安装 WebView2

常用命令：

```powershell
go mod download
cd frontend
npm ci
npm run build
cd ..
go test . ./internal/...
cd frontend
npm run lint
npm run format:check
npm run typecheck
npm run test
npm run build
cd ..
wails dev
```

## 质量与安全

- CI 会执行 Go 测试、前端 lint、格式检查、类型检查、测试、构建和 Windows Wails 构建冒烟测试。
- CodeQL 会扫描 Go 与 TypeScript。
- 已配置 `govulncheck`、`npm audit`、依赖审查、Dependabot、OSSF Scorecard。
- Pull Request 标题遵循 Conventional Commits。
- 通过 `vX.Y.Z` 标签触发发布。

## 参与贡献

请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解分支命名、提交格式、测试要求、评审策略、发布流程和依赖管理规则。

配套文档：

- [开发工作流](docs/WORKFLOW.md)
- [发布指南](docs/RELEASE.md)
- [架构说明](docs/ARCHITECTURE.md)
- [安全策略](SECURITY.md)
- [行为准则](CODE_OF_CONDUCT.md)
- [变更日志](CHANGELOG.md)

当前实现不会上传遥测或书签内容。更大的离线 AI 模型能力将来会以本地下载集成方式提供。

## License

MIT，详见 [LICENSE](LICENSE)。
