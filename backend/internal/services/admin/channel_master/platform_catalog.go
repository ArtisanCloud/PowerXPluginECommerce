package channel_master

import "context"

// PlatformDescriptor exposes supported sales platforms for channel onboarding.
type PlatformDescriptor struct {
	Code         string   `json:"code"`
	Label        string   `json:"label"`
	Description  string   `json:"description,omitempty"`
	ChannelTypes []string `json:"channelTypes"`
	Regions      []string `json:"regions,omitempty"`
	Deprecated   bool     `json:"deprecated,omitempty"`
}

// ChannelTypeDescriptor enumerates available channel authorization flows.
type ChannelTypeDescriptor struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type RegionDescriptor struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type CityDescriptor struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type CountryDescriptor struct {
	Code        string           `json:"code"`
	Label       string           `json:"label"`
	Description string           `json:"description,omitempty"`
	Cities      []CityDescriptor `json:"cities"`
}

// PlatformCatalog aggregates platform + channel type metadata for UI consumption.
type PlatformCatalog struct {
	Platforms    []PlatformDescriptor    `json:"platforms"`
	ChannelTypes []ChannelTypeDescriptor `json:"channelTypes"`
	Regions      []RegionDescriptor      `json:"regions"`
	Countries    []CountryDescriptor     `json:"countries"`
}

var defaultPlatformCatalog = PlatformCatalog{
	Platforms: []PlatformDescriptor{
		{
			Code:         "tmall",
			Label:        "天猫 Tmall",
			Description:  "阿里平台旗舰或国际店铺",
			ChannelTypes: []string{ChannelTypePlatformOAuth, ChannelTypePlatformManual},
			Regions:      []string{"cn-mainland", "sea"},
		},
		{
			Code:         "jd",
			Label:        "京东 JD",
			Description:  "京东主站/POP 店铺",
			ChannelTypes: []string{ChannelTypePlatformManual},
			Regions:      []string{"cn-mainland"},
		},
		{
			Code:         "douyin",
			Label:        "抖音 Douyin",
			Description:  "抖音小店 / 抖音跨境",
			ChannelTypes: []string{ChannelTypePlatformOAuth},
			Regions:      []string{"cn-mainland", "sea"},
		},
		{
			Code:         "offline",
			Label:        "线下 Offline",
			Description:  "线下直营/加盟门店",
			ChannelTypes: []string{ChannelTypeOffline},
			Regions:      []string{"cn-mainland"},
		},
	},
	ChannelTypes: []ChannelTypeDescriptor{
		{Code: ChannelTypePlatformOAuth, Label: "平台授权 / Platform OAuth", Description: "跳转第三方完成 OAuth"},
		{Code: ChannelTypePlatformManual, Label: "手动凭证 / Manual Credential", Description: "运营手动录入凭证"},
		{Code: ChannelTypeOffline, Label: "线下渠道 / Offline", Description: "无系统授权，需线下附件"},
	},
	Regions: []RegionDescriptor{
		{Code: "cn-mainland", Label: "中国大陆 / Mainland China"},
		{Code: "hmt", Label: "港澳台 / HK-MO-TW"},
		{Code: "sea", Label: "东南亚 / Southeast Asia"},
		{Code: "na", Label: "北美 / North America"},
	},
	Countries: []CountryDescriptor{
		{
			Code:  "CN",
			Label: "中国 China",
			Cities: []CityDescriptor{
				{Code: "cn-beijing", Label: "北京 Beijing"},
				{Code: "cn-shanghai", Label: "上海 Shanghai"},
				{Code: "cn-shenzhen", Label: "深圳 Shenzhen"},
				{Code: "cn-guangzhou", Label: "广州 Guangzhou"},
			},
		},
		{
			Code:  "SG",
			Label: "新加坡 Singapore",
			Cities: []CityDescriptor{
				{Code: "sg-singapore", Label: "新加坡 Singapore"},
			},
		},
		{
			Code:  "US",
			Label: "美国 United States",
			Cities: []CityDescriptor{
				{Code: "us-nyc", Label: "纽约 New York"},
				{Code: "us-la", Label: "洛杉矶 Los Angeles"},
				{Code: "us-sf", Label: "旧金山 San Francisco"},
			},
		},
		{
			Code:  "GB",
			Label: "英国 United Kingdom",
			Cities: []CityDescriptor{
				{Code: "gb-london", Label: "伦敦 London"},
				{Code: "gb-manchester", Label: "曼彻斯特 Manchester"},
			},
		},
		{
			Code:  "AE",
			Label: "阿联酋 United Arab Emirates",
			Cities: []CityDescriptor{
				{Code: "ae-dubai", Label: "迪拜 Dubai"},
				{Code: "ae-abu-dhabi", Label: "阿布扎比 Abu Dhabi"},
			},
		},
	},
}

// PlatformCatalog returns the configured platform + channel-type catalog. Currently sourced from constants
// but can be replaced with configuration or database dictionary in the future.
func (s *Service) PlatformCatalog(ctx context.Context) PlatformCatalog {
	return clonePlatformCatalog(defaultPlatformCatalog)
}

func clonePlatformCatalog(src PlatformCatalog) PlatformCatalog {
	cat := PlatformCatalog{}
	if len(src.Platforms) > 0 {
		cat.Platforms = make([]PlatformDescriptor, len(src.Platforms))
		for i, p := range src.Platforms {
			cat.Platforms[i] = PlatformDescriptor{
				Code:         p.Code,
				Label:        p.Label,
				Description:  p.Description,
				ChannelTypes: append([]string(nil), p.ChannelTypes...),
				Regions:      append([]string(nil), p.Regions...),
				Deprecated:   p.Deprecated,
			}
		}
	}
	if len(src.ChannelTypes) > 0 {
		cat.ChannelTypes = make([]ChannelTypeDescriptor, len(src.ChannelTypes))
		copy(cat.ChannelTypes, src.ChannelTypes)
	}
	if len(src.Regions) > 0 {
		cat.Regions = make([]RegionDescriptor, len(src.Regions))
		copy(cat.Regions, src.Regions)
	}
	if len(src.Countries) > 0 {
		cat.Countries = make([]CountryDescriptor, len(src.Countries))
		for i, country := range src.Countries {
			cat.Countries[i] = CountryDescriptor{
				Code:        country.Code,
				Label:       country.Label,
				Description: country.Description,
				Cities:      append([]CityDescriptor(nil), country.Cities...),
			}
		}
	}
	return cat
}
