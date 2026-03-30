package manifestx

import "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/manifest"

// Plugin returns the manifest definition consumed by the framework/router layer.
func Plugin() manifest.Plugin {
	return manifest.Plugin{
		ID:      "com.powerx.plugins.ecommerce",
		Name:    "PowerX Ecommerce Plugin",
		Version: "0.1.0",
		Permissions: []string{
			"iam.user.read",
			"iam.role.read",
			"iam.department.read",
		},
		Menus: []manifest.Menu{
			{
				Path:  "/_p/com.powerx.plugins.ecommerce/admin/templates/intro",
				Title: "模板介绍",
			},
			{
				Path:  "/_p/com.powerx.plugins.ecommerce/admin/templates/crud",
				Title: "模板管理",
			},
		},
	}
}
