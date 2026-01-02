package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	"gopkg.in/yaml.v3"
)

// loadRBACInfoFromFile attempts to parse backend/etc/rbac.yaml (or CONFIG_PATH/rbac.yaml)
// so that admin /rbac endpoint reflects the deployed manifest.
func loadRBACInfoFromFile() (*contracts.RBACInfo, error) {
	path := locateRBACYAML()
	if path == "" {
		return nil, fmt.Errorf("rbac.yaml not found")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc rbacDocument
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	info := doc.toContracts()
	if info == nil {
		return nil, fmt.Errorf("rbac.yaml missing resources or permissions")
	}
	return info, nil
}

func locateRBACYAML() string {
	for _, candidate := range rbacCandidates() {
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		return candidate
	}
	return ""
}

func rbacCandidates() []string {
	var candidates []string
	if raw := strings.TrimSpace(os.Getenv("CONFIG_PATH")); raw != "" {
		if info, err := os.Stat(raw); err == nil {
			if info.IsDir() {
				candidates = append(candidates, filepath.Join(raw, "rbac.yaml"))
			} else {
				dir := filepath.Dir(raw)
				candidates = append(candidates, filepath.Join(dir, "rbac.yaml"))
			}
		}
	}
	candidates = append(candidates,
		filepath.Join("config", "rbac.yaml"),
		filepath.Join("backend", "etc", "rbac.yaml"),
		"rbac.yaml",
	)
	return candidates
}

type rbacDocument struct {
	Resources   []rbacResourceDoc   `yaml:"resources"`
	Roles       []rbacRoleDoc       `yaml:"roles"`
	Permissions []rbacPermissionDoc `yaml:"permissions"`
}

type rbacResourceDoc struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Actions     []rbacActionDoc `yaml:"actions"`
}

type rbacActionDoc struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type rbacRoleDoc struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Permissions []string `yaml:"permissions"`
}

type rbacPermissionDoc struct {
	Resource string `yaml:"resource"`
	Action   string `yaml:"action"`
}

func (doc rbacDocument) toContracts() *contracts.RBACInfo {
	if len(doc.Resources) == 0 && len(doc.Roles) == 0 && len(doc.Permissions) == 0 {
		return nil
	}
	info := &contracts.RBACInfo{
		Resources:   make([]contracts.Resource, 0, len(doc.Resources)),
		Roles:       make([]contracts.Role, 0, len(doc.Roles)),
		Permissions: make([]contracts.Permission, 0, len(doc.Permissions)),
	}
	for _, res := range doc.Resources {
		if strings.TrimSpace(res.Name) == "" {
			continue
		}
		resource := contracts.Resource{
			Name:        strings.TrimSpace(res.Name),
			Description: strings.TrimSpace(res.Description),
			Actions:     make([]contracts.Action, 0, len(res.Actions)),
		}
		for _, act := range res.Actions {
			if strings.TrimSpace(act.Name) == "" {
				continue
			}
			resource.Actions = append(resource.Actions, contracts.Action{
				Name:        strings.TrimSpace(act.Name),
				Description: strings.TrimSpace(act.Description),
			})
		}
		info.Resources = append(info.Resources, resource)
	}
	for _, role := range doc.Roles {
		if strings.TrimSpace(role.Name) == "" {
			continue
		}
		info.Roles = append(info.Roles, contracts.Role{
			Name:        strings.TrimSpace(role.Name),
			Description: strings.TrimSpace(role.Description),
			Permissions: normalizeRolePermissions(role.Permissions),
		})
	}
	if len(doc.Permissions) > 0 {
		for _, perm := range doc.Permissions {
			if strings.TrimSpace(perm.Resource) == "" || strings.TrimSpace(perm.Action) == "" {
				continue
			}
			info.Permissions = append(info.Permissions, contracts.Permission{
				Resource: strings.TrimSpace(perm.Resource),
				Action:   strings.TrimSpace(perm.Action),
			})
		}
	} else {
		info.Permissions = compilePermissions(info.Resources, nil)
	}
	return info
}

func normalizeRolePermissions(perms []string) []string {
	seen := make(map[string]struct{}, len(perms))
	var normalized []string
	for _, perm := range perms {
		perm = strings.TrimSpace(perm)
		if perm == "" {
			continue
		}
		if _, ok := seen[perm]; ok {
			continue
		}
		seen[perm] = struct{}{}
		normalized = append(normalized, perm)
	}
	return normalized
}

func compilePermissions(resources []contracts.Resource, explicit []contracts.Permission) []contracts.Permission {
	type key struct {
		resource string
		action   string
	}
	seen := make(map[key]struct{})
	var perms []contracts.Permission
	for _, perm := range explicit {
		k := key{resource: perm.Resource, action: perm.Action}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		perms = append(perms, perm)
	}
	for _, res := range resources {
		for _, act := range res.Actions {
			k := key{resource: res.Name, action: act.Name}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			perms = append(perms, contracts.Permission{
				Resource: res.Name,
				Action:   act.Name,
			})
		}
	}
	return perms
}

func defaultRBACInfo() *contracts.RBACInfo {
	resources := defaultRBACResources()
	perms := compilePermissions(resources, nil)
	return &contracts.RBACInfo{
		Resources:   resources,
		Roles:       defaultRoles(),
		Permissions: perms,
	}
}

func defaultRBACResources() []contracts.Resource {
	return []contracts.Resource{
		{
			Name:        "com.powerx.plugin.ecommerce:product.category",
			Description: "商品类目树管理与前台展示数据",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看类目树、类目列表与审计记录"},
				{Name: "manage", Description: "创建/编辑/迁移/启停类目"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.category.template",
			Description: "类目模板配置与版本管理",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看模板、预览与历史版本"},
				{Name: "manage", Description: "创建/编辑/发布/回滚模板"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.category.mapping",
			Description: "渠道类目映射维护",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看渠道类目映射"},
				{Name: "manage", Description: "新增/编辑/删除渠道类目映射"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.category.import",
			Description: "类目/映射批量导入导出",
			Actions: []contracts.Action{
				{Name: "read", Description: "导出 CSV 与查看导出结果"},
				{Name: "manage", Description: "导入 CSV 并触发批量处理"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.sku",
			Description: "SKU 生成与基础信息管理",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看 SKU 列表、生成器预览与详情"},
				{Name: "manage", Description: "创建/编辑/删除 SKU，运行生成器"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.sku.bulk",
			Description: "批量调价、库存和导入导出任务",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看批量任务列表、任务状态与日志"},
				{Name: "manage", Description: "提交/审批/重试任务与上传模板"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.sku.channel",
			Description: "渠道映射与发布",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看渠道映射、发布日志"},
				{Name: "manage", Description: "新增/编辑/发布渠道映射"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.sku.inventory",
			Description: "库存与 SLA 监控",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看库存快照、预警与同步时间"},
				{Name: "manage", Description: "触发手动同步或调整库存参数"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.sku.serial",
			Description: "序列号与批次记录",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看序列号/批次记录"},
				{Name: "manage", Description: "录入或导出序列号/批次"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:product.sku.barcode",
			Description: "条码校验与打印",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看条码记录"},
				{Name: "manage", Description: "生成/校验条码并导出模板"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:channel.master",
			Description: "渠道主数据与审批",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看渠道配置"},
				{Name: "create", Description: "新增渠道"},
				{Name: "update", Description: "更新渠道配置或提交审核"},
				{Name: "approve", Description: "渠道审批或发布"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:channel.credential",
			Description: "渠道凭证",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看凭证"},
				{Name: "manage", Description: "新增/更新凭证并测试连通性"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:channel.alert",
			Description: "渠道告警",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看渠道告警"},
				{Name: "manage", Description: "处理或关闭告警"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:channel.strategy",
			Description: "渠道策略配置",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看策略"},
				{Name: "manage", Description: "更新策略或适配参数"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:channel.sync",
			Description: "渠道同步与巡检",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看同步历史"},
				{Name: "trigger", Description: "手动发起同步"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:channel.task",
			Description: "渠道任务看板",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看任务列表与详情"},
				{Name: "manage", Description: "创建、更新或重试任务"},
			},
		},
		{
			Name:        "com.powerx.plugin.ecommerce:channel.note",
			Description: "渠道备注",
			Actions: []contracts.Action{
				{Name: "read", Description: "查看备注"},
				{Name: "create", Description: "新增备注"},
			},
		},
	}
}

func defaultRoles() []contracts.Role {
	return []contracts.Role{
		{
			Name:        "product_category.viewer",
			Description: "查看类目树、模板与渠道映射",
			Permissions: []string{
				"com.powerx.plugin.ecommerce:product.category:read",
				"com.powerx.plugin.ecommerce:product.category.template:read",
				"com.powerx.plugin.ecommerce:product.category.mapping:read",
				"com.powerx.plugin.ecommerce:product.category.import:read",
			},
		},
		{
			Name:        "product_category.operator",
			Description: "维护类目树、模板与渠道映射，并执行导入导出",
			Permissions: []string{
				"com.powerx.plugin.ecommerce:product.category:manage",
				"com.powerx.plugin.ecommerce:product.category.template:manage",
				"com.powerx.plugin.ecommerce:product.category.mapping:manage",
				"com.powerx.plugin.ecommerce:product.category.import:manage",
			},
		},
		{
			Name:        "product_sku.viewer",
			Description: "查看 SKU、库存、渠道映射与日志",
			Permissions: []string{
				"com.powerx.plugin.ecommerce:product.sku:read",
				"com.powerx.plugin.ecommerce:product.sku.bulk:read",
				"com.powerx.plugin.ecommerce:product.sku.channel:read",
				"com.powerx.plugin.ecommerce:product.sku.inventory:read",
				"com.powerx.plugin.ecommerce:product.sku.serial:read",
				"com.powerx.plugin.ecommerce:product.sku.barcode:read",
			},
		},
		{
			Name:        "product_sku.operator",
			Description: "运营可生成 SKU、执行批量任务并维护渠道映射/条码/序列号",
			Permissions: []string{
				"com.powerx.plugin.ecommerce:product.sku:manage",
				"com.powerx.plugin.ecommerce:product.sku.bulk:manage",
				"com.powerx.plugin.ecommerce:product.sku.channel:manage",
				"com.powerx.plugin.ecommerce:product.sku.inventory:manage",
				"com.powerx.plugin.ecommerce:product.sku.serial:manage",
				"com.powerx.plugin.ecommerce:product.sku.barcode:manage",
			},
		},
		{
			Name:        "channel_center.viewer",
			Description: "查看渠道主数据、凭证、同步历史与任务",
			Permissions: []string{
				"com.powerx.plugin.ecommerce:channel.master:read",
				"com.powerx.plugin.ecommerce:channel.credential:read",
				"com.powerx.plugin.ecommerce:channel.alert:read",
				"com.powerx.plugin.ecommerce:channel.strategy:read",
				"com.powerx.plugin.ecommerce:channel.sync:read",
				"com.powerx.plugin.ecommerce:channel.task:read",
				"com.powerx.plugin.ecommerce:channel.note:read",
			},
		},
		{
			Name:        "channel_center.operator",
			Description: "渠道运营，可创建/审批渠道并维护凭证、策略与任务",
			Permissions: []string{
				"com.powerx.plugin.ecommerce:channel.master:create",
				"com.powerx.plugin.ecommerce:channel.master:update",
				"com.powerx.plugin.ecommerce:channel.master:approve",
				"com.powerx.plugin.ecommerce:channel.credential:manage",
				"com.powerx.plugin.ecommerce:channel.alert:manage",
				"com.powerx.plugin.ecommerce:channel.strategy:manage",
				"com.powerx.plugin.ecommerce:channel.sync:trigger",
				"com.powerx.plugin.ecommerce:channel.task:manage",
				"com.powerx.plugin.ecommerce:channel.note:create",
			},
		},
	}
}
