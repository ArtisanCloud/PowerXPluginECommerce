package migrations

import channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"

// ChannelMasterTables enumerates 003-channel-master models.
var ChannelMasterTables = []interface{}{
	&channelmodel.ChannelMaster{},
	&channelmodel.ChannelTaskLink{},
	&channelmodel.ChannelNote{},
	&channelmodel.ChannelSyncHistory{},
	&channelmodel.ChannelCredential{},
	&channelmodel.ChannelAlert{},
	&channelmodel.ChannelMetric{},
	&channelmodel.ChannelConfig{},
}
