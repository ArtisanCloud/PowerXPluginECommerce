package product_category

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	mappingrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MappingImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type MappingImportResult struct {
	Total     int                  `json:"total"`
	Succeeded int                  `json:"succeeded"`
	Failed    int                  `json:"failed"`
	Errors    []MappingImportError `json:"errors,omitempty"`
}

type MappingImportOptions struct {
	CategoryID string
}

func (s *Service) ImportMappingsCSV(ctx context.Context, reader io.Reader, opts MappingImportOptions) (*MappingImportResult, error) {
	if s == nil || !s.Ready() || s.deps == nil || s.deps.DB == nil || s.MappingRepo == nil {
		return nil, errors.New("mapping import service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, errors.New("csv reader is required")
	}
	defaultCategoryID := strings.TrimSpace(opts.CategoryID)

	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	for i := range header {
		header[i] = strings.TrimSpace(strings.ToLower(header[i]))
	}
	idx := func(name string) int {
		for i, h := range header {
			if h == strings.ToLower(name) {
				return i
			}
		}
		return -1
	}
	iCategoryID := idx("categoryid")
	iChannel := idx("channel")
	iPlatform := idx("platformcategoryid")
	iStrategy := idx("strategy")
	iSync := idx("syncstatus")
	iMeta := idx("metadata")

	if iChannel < 0 || iPlatform < 0 {
		return nil, errors.New("csv header must include channel, platformCategoryId")
	}
	if iCategoryID < 0 && defaultCategoryID == "" {
		return nil, errors.New("csv header missing categoryId and no default categoryId provided")
	}

	type rowData struct {
		row     int
		entity  *productcategory.CategoryMapping
		channel string
		key     string
	}
	rows := make([]rowData, 0)
	errs := make([]MappingImportError, 0)
	totalRows := 0
	seen := map[string]struct{}{}
	now := time.Now().UTC()

	line := 1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		line++
		totalRows++
		if err != nil {
			errs = append(errs, MappingImportError{Row: line, Field: "", Message: err.Error()})
			continue
		}
		get := func(i int) string {
			if i < 0 || i >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[i])
		}
		categoryID := defaultCategoryID
		if iCategoryID >= 0 {
			if v := get(iCategoryID); v != "" {
				categoryID = v
			}
		}
		channel := get(iChannel)
		platformID := get(iPlatform)
		strategy := strings.ToLower(get(iStrategy))
		syncStatus := strings.ToLower(get(iSync))
		metaRaw := get(iMeta)

		if categoryID == "" {
			errs = append(errs, MappingImportError{Row: line, Field: "categoryId", Message: "categoryId is required"})
			continue
		}
		if channel == "" {
			errs = append(errs, MappingImportError{Row: line, Field: "channel", Message: "channel is required"})
			continue
		}
		if platformID == "" {
			errs = append(errs, MappingImportError{Row: line, Field: "platformCategoryId", Message: "platformCategoryId is required"})
			continue
		}
		if strategy == "" {
			strategy = productcategory.MappingStrategyManual
		}
		if syncStatus == "" {
			syncStatus = productcategory.MappingSyncPending
		}
		meta := map[string]any{}
		if metaRaw != "" {
			if err := json.Unmarshal([]byte(metaRaw), &meta); err != nil {
				errs = append(errs, MappingImportError{Row: line, Field: "metadata", Message: "metadata must be valid json"})
				continue
			}
		}
		key := fmt.Sprintf("%s|%s|%s", strings.ToLower(categoryID), strings.ToLower(channel), strings.ToLower(platformID))
		if _, ok := seen[key]; ok {
			errs = append(errs, MappingImportError{Row: line, Field: "channel", Message: "duplicate mapping in file"})
			continue
		}
		seen[key] = struct{}{}
		metaBody, _ := json.Marshal(meta)
		rows = append(rows, rowData{
			row:     line,
			channel: channel,
			key:     key,
			entity: &productcategory.CategoryMapping{
				ID:                 utils.NewUUID(),
				TenantUUID:         tenantID,
				CategoryID:         categoryID,
				Channel:            channel,
				PlatformCategoryID: platformID,
				Strategy:           strategy,
				SyncStatus:         syncStatus,
				Metadata:           datatypes.JSON(metaBody),
				CreatedAt:          now,
				UpdatedAt:          now,
			},
		})
	}

	result := &MappingImportResult{
		Total:  totalRows,
		Errors: errs,
	}
	if len(rows) == 0 {
		result.Failed = totalRows
		return result, nil
	}

	saveErr := s.MappingRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := mappingrepo.NewMappingRepository(tx)
		for _, row := range rows {
			_, err := repoTx.Upsert(ctx, tx, row.entity)
			if err != nil {
				result.Errors = append(result.Errors, MappingImportError{
					Row:     row.row,
					Field:   "channel",
					Message: err.Error(),
				})
				continue
			}
			result.Succeeded++
		}
		return nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	result.Failed = result.Total - result.Succeeded
	sort.SliceStable(result.Errors, func(i, j int) bool { return result.Errors[i].Row < result.Errors[j].Row })
	_ = s.writeCategoryAudit(ctx, tenantID, "mapping_import", defaultCategoryID, map[string]any{
		"total":     result.Total,
		"succeeded": result.Succeeded,
		"failed":    result.Failed,
	})
	return result, nil
}

type MappingExportOptions struct {
	CategoryID string
}

func (s *Service) ExportMappingsCSV(ctx context.Context, opts MappingExportOptions) ([]byte, string, error) {
	if s == nil || !s.Ready() || s.MappingRepo == nil {
		return nil, "", errors.New("mapping export service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, "", err
	}
	categoryID := strings.TrimSpace(opts.CategoryID)
	if categoryID == "" {
		return nil, "", errors.New("categoryId is required")
	}
	records, err := s.MappingRepo.ListByCategory(ctx, tenantID, categoryID)
	if err != nil {
		return nil, "", err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"categoryId", "channel", "platformCategoryId", "strategy", "syncStatus", "metadata"})
	for _, rec := range records {
		rec.Normalize()
		meta := strings.TrimSpace(string(rec.Metadata))
		_ = w.Write([]string{
			rec.CategoryID,
			rec.Channel,
			rec.PlatformCategoryID,
			rec.Strategy,
			rec.SyncStatus,
			meta,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("category_mappings_%s.csv", categoryID)
	_ = s.writeCategoryAudit(ctx, tenantID, "mapping_export", categoryID, map[string]any{
		"count": len(records),
	})
	return buf.Bytes(), filename, nil
}
