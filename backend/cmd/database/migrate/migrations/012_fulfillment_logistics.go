package migrations

import (
	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	ReverseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/reverse"
)

// FulfillmentLogisticsTables enumerates 011-fulfillment-logistics models.
var FulfillmentLogisticsTables = []interface{}{
	&LogisticsModel.Carrier{},
	&LogisticsModel.CarrierService{},
	&LogisticsModel.RateTemplate{},
	&LogisticsModel.RateZone{},
	&LogisticsModel.Waybill{},
	&LogisticsModel.TrackingEvent{},
	&FulfillmentModel.Task{},
	&FulfillmentModel.TaskLog{},
	&FulfillmentModel.Wave{},
	&FulfillmentModel.WaveTaskLink{},
	&FulfillmentModel.Exception{},
	&ReverseModel.Waybill{},
	&ReverseModel.TrackingEvent{},
	&ReverseModel.WarehouseResult{},
}
