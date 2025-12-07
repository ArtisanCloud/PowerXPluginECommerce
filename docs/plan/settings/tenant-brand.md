# 租户与品牌设置 PRD

> 规划尚未存在的租户/品牌配置页面，负责租户基本信息、域名、品牌化（logo、颜色、文案）、门户配置、多环境（Dev/Stage/Prod）。

## 1. 目标
- 允许租户管理员配置基础信息（公司、联系、地址）、域名/URL、品牌样式、门户文案，并管理不同环境（Dev/Stage/Prod）的配置。

## 2. 信息架构
- 租户信息：名称、ID、联系人、地址、行业、许可证、合同、有效期。
- 域名配置：主域名、回调域、白名单、SSL、重定向。
- 品牌化：Logo、配色、字体、登录/门户文案、favicon、邮件模板样式。
- 环境管理：环境列表（Dev/Stage/Prod）、配置、同步、数据隔离、API Key。
- 门户设置：客户门户、供应商门户的开关、URL、认证方式、品牌样式。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 租户信息管理 | 查看/编辑租户信息、合同、有效期、模块授权 |
| 域名/SSL | 配置域名、SSL 证书、DNS、CNAME |
| 品牌设置 | 上传 Logo、颜色、文案，自定义登录页、邮件模板 |
| 环境 | 创建/管理环境，控制配置同步、数据复制、权限 |
| 门户配置 | 开启/关闭客户/供应商门户，配置登录方式（SSO/OAuth）、品牌 |
| 审计 | 记录所有租户级配置更改 |

## 4. 数据 & API
- 表：`tenants`、`tenant_configs`、`tenant_domains`、`tenant_brands`、`tenant_environments`。
- API：`GET/POST /api/settings/tenant`、`PATCH /api/settings/tenant`、`POST /api/settings/tenant/environments`、`POST /api/settings/tenant/brand`。

## 5. 权限 & 审计
- 权限：`settings.tenant.manage`、`settings.tenant.brand`、`settings.tenant.env`。
- 审计：租户信息变更、域名、品牌、门户配置。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 配置变化生效时间 | < 10 分钟 |
| 环境同步成功率 | ≥ 99% |

## 7. Backlog
- 自助门户定制器（拖拽式）。
- 多租户模板管理。
- 自动化域名证书续期。
