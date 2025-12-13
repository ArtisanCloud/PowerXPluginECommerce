package spu

import (
	productmetrics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/product"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

func resolveSPUMetrics(deps *app.Deps, component string) *productmetrics.SPUMetrics {
	if deps == nil {
		return nil
	}
	if deps.ProductMetrics != nil && deps.ProductMetrics.SPU != nil {
		return deps.ProductMetrics.SPU
	}
	if logger := deps.RuntimeLogger(nil, component, nil); logger != nil {
		return productmetrics.NewSPUMetrics(logger)
	}
	return nil
}
