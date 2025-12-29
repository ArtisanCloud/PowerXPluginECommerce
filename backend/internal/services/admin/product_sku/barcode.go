package product_sku

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

const (
	defaultBarcodeLength = 12
	maxAutoBarcodeCount  = 100
)

// GenerateBarcodes validates or auto-generates barcodes and prepares printable labels.
func (s *Service) GenerateBarcodes(ctx context.Context, skuID string, req BarcodeGenerateRequest) (*BarcodeBatchResult, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = BarcodeModeAuto
	}

	switch mode {
	case BarcodeModeAuto:
		return s.generateAutoBarcodes(ctx, tenantID, skuID, req)
	case BarcodeModeManual:
		return s.validateManualBarcodes(ctx, tenantID, skuID, req)
	default:
		return nil, fmt.Errorf("unsupported barcode mode %s", mode)
	}
}

func (s *Service) generateAutoBarcodes(ctx context.Context, tenantID, skuID string, req BarcodeGenerateRequest) (*BarcodeBatchResult, error) {
	count := req.Count
	if count <= 0 {
		count = 1
	}
	if count > maxAutoBarcodeCount {
		return nil, fmt.Errorf("count exceeds limit (%d)", maxAutoBarcodeCount)
	}
	length := req.Length
	if length <= 0 {
		length = defaultBarcodeLength
	}
	prefix := strings.TrimSpace(req.Prefix)
	if len(prefix) >= length {
		return nil, errors.New("prefix length must be less than total length")
	}

	seen := map[string]struct{}{}
	results := make([]BarcodeCheckResult, 0, count)
	labels := make([]BarcodeLabel, 0, count)
	attempts := 0
	for len(results) < count {
		attempts++
		if attempts > count*10 {
			return nil, errors.New("unable to find unique barcodes, try again")
		}
		candidate, err := randomBarcode(length - len(prefix))
		if err != nil {
			return nil, err
		}
		barcode := normalizeBarcode(prefix + candidate)
		if _, ok := seen[barcode]; ok {
			continue
		}
		unique, _, err := s.ensureUniqueBarcode(ctx, tenantID, skuID, barcode)
		if err != nil {
			return nil, err
		}
		if !unique {
			continue
		}
		seen[barcode] = struct{}{}
		results = append(results, BarcodeCheckResult{Barcode: barcode, Unique: true})
		labels = append(labels, BarcodeLabel{Barcode: barcode, SVG: buildBarcodeSVG(barcode)})
	}
	s.emitEvent(ctx, "barcode_batch_generated", map[string]any{
		"tenant_id": tenantID,
		"sku_id":    skuID,
		"count":     len(results),
		"mode":      "auto",
	})
	return &BarcodeBatchResult{Items: results, Labels: labels}, nil
}

func (s *Service) validateManualBarcodes(ctx context.Context, tenantID, skuID string, req BarcodeGenerateRequest) (*BarcodeBatchResult, error) {
	if len(req.Codes) == 0 {
		return nil, errors.New("codes are required for manual mode")
	}
	results := make([]BarcodeCheckResult, 0, len(req.Codes))
	labels := make([]BarcodeLabel, 0, len(req.Codes))
	for _, raw := range req.Codes {
		barcode := normalizeBarcode(raw)
		if barcode == "" {
			results = append(results, BarcodeCheckResult{Barcode: raw, Unique: false})
			continue
		}
		unique, conflict, err := s.ensureUniqueBarcode(ctx, tenantID, skuID, barcode)
		if err != nil {
			return nil, err
		}
		results = append(results, BarcodeCheckResult{Barcode: barcode, Unique: unique, ConflictSKUID: conflict})
		if unique {
			labels = append(labels, BarcodeLabel{Barcode: barcode, SVG: buildBarcodeSVG(barcode)})
		}
	}
	s.emitEvent(ctx, "barcode_validation_completed", map[string]any{
		"tenant_id": tenantID,
		"sku_id":    skuID,
		"count":     len(results),
		"mode":      "manual",
	})
	return &BarcodeBatchResult{Items: results, Labels: labels}, nil
}

func (s *Service) ensureUniqueBarcode(ctx context.Context, tenantID, skuID, barcode string) (bool, string, error) {
	if s.SKURepo == nil || s.SKURepo.DB == nil {
		return false, "", errors.New("sku repository unavailable")
	}
	barcode = normalizeBarcode(barcode)
	if barcode == "" {
		return false, "", errors.New("barcode is required")
	}
	existing, err := s.SKURepo.FindByBarcode(ctx, tenantID, barcode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, "", nil
		}
		return false, "", err
	}
	if existing == nil {
		return true, "", nil
	}
	if skuID != "" && existing.ID == skuID {
		return true, "", nil
	}
	return false, existing.ID, nil
}

func randomBarcode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}
	var sb strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		sb.WriteByte(byte('0' + n.Int64()))
	}
	return sb.String(), nil
}

func normalizeBarcode(raw string) string {
	trimmed := strings.TrimSpace(raw)
	return strings.ToUpper(trimmed)
}

func buildBarcodeSVG(code string) string {
	width := 200
	if l := len(code) * 10; l > width {
		width = l
	}
	bgID := hex.EncodeToString([]byte(utils.NewUUID()))
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="80" role="img" aria-labelledby="barcode-%s"><title id="barcode-%s">%s</title><rect width="100%%" height="100%%" fill="white"/><text x="50%%" y="50%%" dominant-baseline="middle" text-anchor="middle" font-family="monospace" font-size="20">%s</text></svg>`, width, bgID, bgID, code, code)
}
