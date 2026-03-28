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
	&LogisticsModel.RoutingRule{},
	&LogisticsModel.ETAPolicy{},
	&LogisticsModel.ETARecord{},
	&LogisticsModel.RiskRule{},
	&LogisticsModel.BlacklistEntry{},
	&LogisticsModel.RiskHit{},
	&LogisticsModel.ExceptionOrchestrationRule{},
	&LogisticsModel.ExceptionOrchestrationRun{},
	&LogisticsModel.AddressValidation{},
	&LogisticsModel.RedeliveryTask{},
	&LogisticsModel.Waybill{},
	&LogisticsModel.TrackingEvent{},
	&LogisticsModel.TrackingSyncJob{},
	&LogisticsModel.TrackingSyncSchedule{},
	&LogisticsModel.GatewayFailureEvent{},
	&LogisticsModel.GatewayUsage{},
	&LogisticsModel.LabelPrintTask{},
	&LogisticsModel.BillingCase{},
	&LogisticsModel.NotificationTemplate{},
	&LogisticsModel.NotificationRecord{},
	&FulfillmentModel.Task{},
	&FulfillmentModel.TaskLog{},
	&FulfillmentModel.Wave{},
	&FulfillmentModel.WaveTaskLink{},
	&FulfillmentModel.Outbound{},
	&FulfillmentModel.PickItem{},
	&FulfillmentModel.PackOrder{},
	&FulfillmentModel.WaveStrategy{},
	&FulfillmentModel.Exception{},
	&ReverseModel.Waybill{},
	&ReverseModel.TrackingEvent{},
	&ReverseModel.WarehouseResult{},
	&ReverseModel.InspectionRule{},
	&ReverseModel.WaybillInspection{},
}
