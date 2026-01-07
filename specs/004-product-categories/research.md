# Research: 商品类目与类目模板管理（Phase 0）

## Decisions

### D1：停用类目在“可展示类目树”中的表现

- **Decision**：停用类目及其整棵子树都不返回，前台完全不可见。
- **Rationale**：规则最直观、避免前台出现“不可点击入口”、减少后续筛选/导航的歧义。
- **Alternatives considered**：返回但标记不可用；隐藏父节点但提升子节点（会破坏信息架构一致性）。

### D2：类目模板继承/覆盖规则

- **Decision**：子类目未绑定模板时，继承最近祖先“已发布模板”；子类目绑定模板后完全覆盖，不做字段合并。
- **Rationale**：覆盖规则清晰，避免字段合并带来的冲突优先级与兼容性问题。
- **Alternatives considered**：字段合并/叠加（复杂度高，测试面广）；强制每类目显式绑定（运营成本高）。

### D3：类目唯一性规则

- **Decision**：租户内 `code` 全局唯一；同一父节点下 `displayName` 唯一；`alias/slug` 全局唯一。
- **Rationale**：便于运营管理与导入校验，并减少前台路由/SEO 冲突风险。
- **Alternatives considered**：仅 code 唯一；不做约束（会导致大量数据清洗成本）。

### D4：模板发布/回滚对存量商品的影响

- **Decision**：只影响“新建/编辑时校验”；提供“影响范围预览 + 批量重检任务”；不自动改变存量商品状态。
- **Rationale**：避免线上大规模自动状态变更导致运营风险，同时保留可控的治理手段。
- **Alternatives considered**：发布后自动全量重检并改变状态（风险高、易误伤）。

### D5：批量导入/导出格式

- **Decision**：仅支持 CSV。
- **Rationale**：最通用、易生成/解析、Excel 可直接打开；可在后续迭代补充 xlsx。
- **Alternatives considered**：仅 xlsx；CSV+xlsx（初期成本更高）；仅 JSON（运营不友好）。

## Design Choice Notes (for Phase 1)

- 类目树存储建议优先选择“父子关系 + 排序 + 预计算路径/层级”的组合，以满足树查询与前台导航性能目标，并减少复杂树算法依赖。
- 审计事件以“关键动作可追踪”为目标，后端 service 层统一写入 admin_console_audit_events，避免在 handler 层散落。

