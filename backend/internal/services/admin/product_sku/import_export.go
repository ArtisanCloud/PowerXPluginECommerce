package product_sku

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
)

const (
	maxImportRows   = 1000
	maxExportRows   = 5000
	maxImportBytes  = 2 << 20 // 2MB
	defaultExport   = "csv"
	dataURLTemplate = "data:text/%s;base64,%s"
)

// ImportSkus parses an uploaded CSV payload and creates an import bulk task snapshot.
func (s *Service) ImportSkus(ctx context.Context, req SkuImportRequest) (*BulkTaskResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if len(req.Content) == 0 {
		return nil, errors.New("import payload is empty")
	}
	if len(req.Content) > maxImportBytes {
		return nil, fmt.Errorf("import payload exceeds %d bytes", maxImportBytes)
	}
	rows, err := parseSkuImportRows(req.Content)
	if err != nil {
		return nil, err
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "upsert"
	}
	scope := map[string]any{
		"mode":      mode,
		"row_count": len(rows),
	}
	op := map[string]any{
		"mode":      mode,
		"file_name": req.FileName,
		"rows":      rows,
	}
	stats := &BulkTaskStats{Succeeded: len(rows)}
	result := map[string]any{
		"mode": mode,
		"rows": len(rows),
	}
	return s.createInstantTask(ctx, tenantID, "import", scope, op, stats, result, nil)
}

// ExportSkus builds a CSV snapshot of filtered SKUs and stores the payload inline for download.
func (s *Service) ExportSkus(ctx context.Context, req SkuExportRequest) (*BulkTaskResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = defaultExport
	}
	if req.Limit <= 0 || req.Limit > maxExportRows {
		req.Limit = maxExportRows
	}
	filters := repo.SkuExportFilters{
		SPUID:   strings.TrimSpace(req.Filters["spu_id"]),
		Status:  strings.TrimSpace(req.Filters["status"]),
		Keyword: strings.TrimSpace(req.Filters["keyword"]),
		Limit:   req.Limit,
	}
	skus, err := s.SKURepo.ListForExport(ctx, tenantID, filters)
	if err != nil {
		return nil, err
	}
	payload := buildSkuExportCSV(skus)
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	downloadURL := fmt.Sprintf(dataURLTemplate, format, encoded)
	scope := map[string]any{
		"filters": filters,
	}
	op := map[string]any{
		"format":  format,
		"filters": filters,
	}
	stats := &BulkTaskStats{Succeeded: len(skus)}
	result := map[string]any{
		"format":       format,
		"rows":         len(skus),
		"download_url": downloadURL,
		"file_name":    fmt.Sprintf("sku-export-%s.%s", utils.NewUUID(), format),
	}
	return s.createInstantTask(ctx, tenantID, "export", scope, op, stats, result, nil)
}

func (s *Service) createInstantTask(
	ctx context.Context,
	tenantID string,
	taskType string,
	scope any,
	operation any,
	stats *BulkTaskStats,
	result map[string]any,
	errorReport *string,
) (*BulkTaskResponse, error) {
	scopeJSON, _ := json.Marshal(scope)
	opJSON, _ := json.Marshal(operation)
	task := &productskumodel.ProductSKUBulkTask{
		TaskID:           utils.NewUUID(),
		TenantUUID:       tenantID,
		TaskType:         taskType,
		Scope:            datatypes.JSON(scopeJSON),
		Operation:        datatypes.JSON(opJSON),
		ApprovalRequired: false,
		ApprovalState:    BulkTaskStatusApproved,
		Status:           BulkTaskStatusSucceeded,
	}
	if stats != nil {
		task.AffectedCount = stats.Succeeded
	}
	task.Result = encodeTaskResult(stats, result)
	if errorReport != nil {
		task.ErrorReportURL = *errorReport
	}
	if err := s.BulkTaskRepo.DB.WithContext(ctx).Create(task).Error; err != nil {
		return nil, err
	}
	resp := &BulkTaskResponse{
		TaskID:           task.TaskID,
		ApprovalRequired: false,
		Status:           task.Status,
		ErrorReportURL:   task.ErrorReportURL,
	}
	if stats != nil {
		resp.Stats = stats
	}
	if len(result) > 0 {
		resp.Result = result
	}
	s.emitEvent(ctx, "bulk_task_instant_completed", map[string]any{
		"tenant_id":  tenantID,
		"task_id":    task.TaskID,
		"task_type":  taskType,
		"status":     task.Status,
		"succeeded":  statsSuccess(stats),
		"failed":     statsFailed(stats),
		"mode":       taskType,
		"has_result": len(result) > 0,
	})
	return resp, nil
}

type csvColumnIndex map[string]int

func parseSkuImportRows(content []byte) ([]SkuImportRow, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.TrimLeadingSpace = true
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	index := make(csvColumnIndex)
	for idx, col := range headers {
		index[strings.ToLower(strings.TrimSpace(col))] = idx
	}
	if _, ok := index["sku_code"]; !ok {
		return nil, errors.New("csv header must include sku_code")
	}
	rows := make([]SkuImportRow, 0, len(content)/64)
	for rowIdx := 1; rowIdx <= maxImportRows; rowIdx++ {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row %d: %w", rowIdx, err)
		}
		row := SkuImportRow{
			SKUCode: readCSVValue(record, index, "sku_code"),
		}
		if row.SKUCode == "" {
			continue
		}
		if val := readCSVValue(record, index, "sku_id"); val != "" {
			row.SKUID = val
		}
		if val := readCSVValue(record, index, "price"); val != "" {
			if price, err := parseDecimal(val); err == nil {
				row.Price = &price
			}
		}
		if val := readCSVValue(record, index, "inventory"); val != "" {
			if qty, err := parseInteger(val); err == nil {
				row.Inventory = &qty
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, errors.New("no valid rows found in import payload")
	}
	return rows, nil
}

func readCSVValue(record []string, index csvColumnIndex, key string) string {
	pos, ok := index[key]
	if !ok || pos >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[pos])
}

func parseDecimal(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}

func parseInteger(input string) (int64, error) {
	return strconv.ParseInt(input, 10, 64)
}

func buildSkuExportCSV(skus []productskumodel.ProductSKU) string {
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)
	_ = writer.Write([]string{"sku_code", "status", "barcode", "min_order_qty"})
	for _, sku := range skus {
		writer.Write([]string{
			sku.SKUCode,
			sku.Status,
			sku.Barcode,
			fmt.Sprintf("%d", sku.MinOrderQty),
		})
	}
	writer.Flush()
	return buffer.String()
}
