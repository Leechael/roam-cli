# roam-cli（Golang）重构规划（更新版）

## 你的要求（已吸收）
1. `save_markdown` / `get_journaling_by_date` 名字太长，不适合 CLI 主命令。
2. 但这两类能力必须保留。
3. 必须提供 **low-level blocks 操作** 与 **batch 操作** API（CLI 入口）。
4. token / graph 继续走环境变量注入。

---

## 命令命名策略

### A. 用户友好的短命令（主入口）
- `save` 代替 `save_markdown`
- `journal` 代替 `get_journaling_by_date`

### B. 兼容别名（可选）
为了和旧 MCP 语义对齐，可保留别名：
- `save-markdown` -> `save`
- `journaling` / `get-journaling-by-date` -> `journal`

> 这样既保持可用性，也不牺牲 CLI 体验。

---

## CLI 命令设计（MVP + low-level）

```bash
roam-cli get <identifier> [--raw]
roam-cli search <terms...> [--page <title>] [--ignore-case] [--limit 20]
roam-cli q '<datalog>' [--args ...]
roam-cli save --title "New Page" [--file note.md]
roam-cli journal [--date 2026-02-12] [--topic "xxx"]

# low-level block API
roam-cli block create --parent <uid> --text "..." [--order last] [--uid <uid>] [--open]
roam-cli block update --uid <uid> --text "..."
roam-cli block delete --uid <uid>
roam-cli block get --uid <uid> [--raw]

# low-level batch API
roam-cli batch run --file actions.json
roam-cli batch run --stdin
```

---

## low-level 输入/输出约定

### `batch run`
- 输入为 Roam `actions` JSON 数组（与 `/write` 的 `batch-actions` 对齐）
- 示例：
```json
[
  {"action":"create-page","page":{"title":"T","uid":"...","children-view-type":"bullet"}},
  {"action":"create-block","location":{"parent-uid":"...","order":"last"},"block":{"uid":"...","string":"hello","open":true}}
]
```

### `block create/update/delete`
- 默认输出 JSON（便于脚本链式处理）
- 支持 `--quiet` 仅返回 uid / 成功状态

---

## 环境变量

必填：
- `ROAM_API_TOKEN`
- `ROAM_API_GRAPH`

可选：
- `ROAM_API_BASE_URL`（默认 `https://api.roamresearch.com/api/graph`）
- `ROAM_TIMEOUT_SECONDS`（默认 10）

优先级：CLI 参数 > 环境变量 > 默认值。

---

## 目录结构建议

```text
roamresearch-skills/
  go.mod
  cmd/
    roam-cli/
      main.go
  internal/
    cli/
      root.go
      get.go
      search.go
      q.go
      save.go
      journal.go
      block.go        # create/update/delete/get
      batch.go        # batch run
    roam/
      client.go
      types.go
      query.go
      write.go
    format/
      markdown.go
    parser/
      identifier.go
    config/
      env.go
  README.md
```

---

## 与现有 MCP 的映射

- `save_markdown` -> `save`（alias: `save-markdown`）
- `get_journaling_by_date` -> `journal`（alias 可选）
- `query` -> `q`
- `get` -> `get`
- `search` -> `search`

另外新增：
- `block *`（low-level）
- `batch run`（low-level）

---

## 里程碑（调整后）

### M1
- 项目骨架 + root command + env 注入
- HTTP client + `/q` + `/write` 基础封装

### M2
- `get / search / q / journal`

### M3
- `save`（markdown -> actions）
- `block create/update/delete/get`
- `batch run --file/--stdin`

### M4（可选）
- 异步任务与本地 task db（对齐当前 Python MCP 的后台写入模型）

---

## 验收标准（新增 low-level）

1. 高层命令可用：`get/search/q/save/journal`
2. 低层命令可用：`block *`、`batch run`
3. 所有写操作支持脚本化（stdin/file/json 输出）
4. 环境变量注入可无缝替换现有 Python 使用方式
