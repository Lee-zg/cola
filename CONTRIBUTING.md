# 参与 Cola Bookmarks 贡献指南

感谢你参与改进 Cola Bookmarks。该项目是本地优先桌面软件，因此隐私、可预测行为和可维护性优先于快速堆叠功能。

## 开发环境准备

依赖要求：

- Go 1.25 或更高版本
- Node.js 22.13 或更高版本
- npm 10 或更高版本
- Wails v2.12.0 或更高版本
- Windows 构建需要 WebView2

安装与校验：

```powershell
go mod download
cd frontend
npm ci
npm run build
cd ..
go test . ./internal/...
cd frontend
npm run lint
npm run typecheck
npm run test
npm run build
cd ..
wails dev
```

## 分支策略

- `main` 分支始终应保持可发布状态。
- 功能开发使用 `feature/<short-description>`。
- 缺陷修复使用 `fix/<short-description>`。
- 纯文档修改使用 `docs/<short-description>`。
- CI/构建维护使用 `chore/<short-description>`。
- 发布准备使用 `release/vX.Y.Z`。

请保持分支聚焦，不要在同一个 PR 中混入功能开发、重构、格式化和依赖升级。

## 提交信息格式

采用 Conventional Commits：

```text
<type>(optional-scope): <summary>
```

允许的类型：

- `feat`：面向用户的新功能
- `fix`：缺陷修复
- `docs`：文档
- `test`：测试
- `refactor`：不改变行为的代码重构
- `perf`：性能优化
- `build`：构建系统或依赖调整
- `ci`：持续集成配置
- `chore`：日常维护
- `security`：安全修复或加固

示例：

```text
feat(importer): support Firefox bookmark import
fix(exporter): escape bookmark descriptions in static HTML
ci: add CodeQL workflow
```

若存在破坏性变更，提交正文必须包含 `BREAKING CHANGE:`。

## Pull Request 要求

每个 PR 应包含：

- 清晰的变更说明。
- 相关 issue（如适用）。
- 测试证据（包含执行命令）。
- UI 变更的截图或短录屏。
- 涉及隐私、本地存储或安全时的影响说明。

发起评审前请执行：

```powershell
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
```

## 代码评审

合并到 `main` 前至少需要一位维护者批准。

评审重点：

- 数据隐私和本地化保证是否被破坏。
- 导入、导出、本地 Web 和主题包的安全边界是否清晰。
- 新行为与回归风险是否有测试覆盖。
- 模块职责与 API 是否可维护。
- 常见桌面窗口尺寸下的 UI 行为是否正常。

普通 PR 建议使用 squash merge，除非保留独立提交历史有明确价值。

## 测试策略

- Go 单测放在对应包目录，命名为 `*_test.go`。
- 集成测试优先覆盖公开 `App` 门面或服务边界。
- 前端测试使用 Vitest。
- 导入器、导出器、备份或本地 Web 服务改动必须补测试。
- 涉及 HTML 安全处理的改动必须覆盖转义或拒绝策略测试。

## 发布流程

1. 按 SemVer 更新 `VERSION`。
2. 更新 `CHANGELOG.md`。
3. 本地执行完整校验命令。
4. 从 `release/vX.Y.Z` 向 `main` 发起发布 PR。
5. PR 合并后创建并推送标签 `vX.Y.Z`。
6. GitHub Actions 自动构建发布制品并创建 GitHub Release。

## 依赖管理

- Dependabot 每周会为 Go、npm、GitHub Actions 依赖创建 PR。
- 依赖升级 PR 应保持小而独立，不要混入功能代码。
- 安全更新在 CI 通过后应尽快评估并合并。
- 新增运行时依赖需要在 PR 里给出简要理由。

## 安全与隐私要求

- 不得引入书签数据遥测或网络上传。
- 本地 Web 服务默认必须绑定 `127.0.0.1`。
- 导入/导出 HTML 必须进行清洗或转义。
- 主题包不得执行第三方 JavaScript。
