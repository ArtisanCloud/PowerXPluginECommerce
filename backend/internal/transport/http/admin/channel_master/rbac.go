package channel_master

import (
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
)

// RBACEntries declares route level permissions for channel master API.
func RBACEntries(prefix string) map[string]authx.Permission {
	base := strings.TrimRight(prefix, "/") + "/admin/channels"
	res := func(suffix string) string {
		return "com.powerx.plugin.ecommerce:" + suffix
	}
	return map[string]authx.Permission{
		"GET:" + base:                                      {Resource: res("channel.master"), Action: "read"},
		"GET:" + base + "/platforms":                       {Resource: res("channel.master"), Action: "read"},
		"GET:" + base + "/owners":                          {Resource: res("channel.master"), Action: "read"},
		"POST:" + base:                                     {Resource: res("channel.master"), Action: "create"},
		"PATCH:" + base + "/:channelId":                    {Resource: res("channel.master"), Action: "update"},
		"POST:" + base + "/:channelId/submit":              {Resource: res("channel.master"), Action: "update"},
		"POST:" + base + "/:channelId/approval":            {Resource: res("channel.master"), Action: "approve"},
		"GET:" + base + "/:channelId":                      {Resource: res("channel.master"), Action: "read"},
		"GET:" + base + "/:channelId/credentials":          {Resource: res("channel.credential"), Action: "read"},
		"POST:" + base + "/:channelId/credentials":         {Resource: res("channel.credential"), Action: "manage"},
		"POST:" + base + "/:channelId/credentials/test":    {Resource: res("channel.credential"), Action: "manage"},
		"GET:" + base + "/:channelId/alerts":               {Resource: res("channel.alert"), Action: "read"},
		"PATCH:" + base + "/:channelId/alerts/:alertId":    {Resource: res("channel.alert"), Action: "manage"},
		"POST:" + base + "/:channelId/sync":                {Resource: res("channel.sync"), Action: "trigger"},
		"GET:" + base + "/:channelId/sync-history":         {Resource: res("channel.sync"), Action: "read"},
		"GET:" + base + "/:channelId/tasks":                {Resource: res("channel.task"), Action: "read"},
		"POST:" + base + "/:channelId/tasks":               {Resource: res("channel.task"), Action: "manage"},
		"PATCH:" + base + "/:channelId/tasks/:taskLinkId":  {Resource: res("channel.task"), Action: "manage"},
		"DELETE:" + base + "/:channelId/tasks/:taskLinkId": {Resource: res("channel.task"), Action: "manage"},
		"GET:" + base + "/:channelId/notes":                {Resource: res("channel.note"), Action: "read"},
		"POST:" + base + "/:channelId/notes":               {Resource: res("channel.note"), Action: "create"},
		"GET:" + base + "/:channelId/strategy":             {Resource: res("channel.strategy"), Action: "read"},
		"PATCH:" + base + "/:channelId/strategy":           {Resource: res("channel.strategy"), Action: "manage"},
	}
}
