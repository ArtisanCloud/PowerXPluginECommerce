package channel_master

import channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"

// ChannelListQuery exposes optional filters for GET /channels.
type ChannelListQuery struct {
	Keyword  string   `form:"keyword"`
	Platform string   `form:"platform"`
	Status   []string `form:"status[]"`
	Owner    string   `form:"owner"`
	Region   string   `form:"region"`
	Page     int      `form:"page"`
	PageSize int      `form:"pageSize"`
}

// CreateChannelRequest mirrors the OpenAPI create payload.
type CreateChannelRequest struct {
	Name        string   `json:"name"`
	Platform    string   `json:"platform"`
	StoreID     string   `json:"storeId"`
	Region      string   `json:"region"`
	OwnerUUID   string   `json:"ownerUuid"`
	ChannelType string   `json:"channelType"`
	Tags        []string `json:"tags"`
}

// ToServiceInput converts HTTP requests to service DTOs.
func (req CreateChannelRequest) ToServiceInput() channelservice.CreateChannelInput {
	return channelservice.CreateChannelInput{
		Name:        req.Name,
		Platform:    req.Platform,
		StoreID:     req.StoreID,
		Region:      req.Region,
		OwnerUUID:   req.OwnerUUID,
		ChannelType: req.ChannelType,
		Tags:        req.Tags,
	}
}
