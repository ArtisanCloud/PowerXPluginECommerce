package product_sku

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
)

const maxGeneratorCandidates = 512

// GenerateSkusFromSPU builds cartesian combinations for the provided specs and defaults.
func (s *Service) GenerateSkusFromSPU(ctx context.Context, spuID string, req SkuGeneratorRequest) (*SkuGeneratorResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(spuID) == "" {
		return nil, errors.New("spu id is required")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	baseCode, err := s.resolveSPUCode(ctx, tenantID, spuID)
	if err != nil {
		return nil, err
	}
	specMatrix := buildSpecMatrix(req.SpecSelections)
	var combinations [][]specOption
	if len(specMatrix) == 0 {
		combinations = [][]specOption{{}}
	} else {
		combinations, err = expandCartesian(specMatrix)
		if err != nil {
			return nil, err
		}
	}
	existing, _, _, err := s.lookupExistingSpecHashes(ctx, tenantID, spuID)
	if err != nil {
		return nil, err
	}
	candidates := make([]SkuGeneratorCandidate, 0, len(combinations))
	for _, combo := range combinations {
		specs := convertToSkuSpecs(combo)
		hash := hashSpecValues(specs)
		_, exists := existing[hash]
		candidate := SkuGeneratorCandidate{
			Specs:          specs,
			PreviewSKUCode: buildPreviewCode(baseCode, combo),
			DefaultValues:  req.Defaults,
			Exists:         exists,
			Selected:       !exists,
		}
		if exists {
			candidate.ConflictReasons = []string{"combination already exists"}
		}
		candidates = append(candidates, candidate)
	}
	return &SkuGeneratorResponse{Candidates: candidates}, nil
}

type specOption struct {
	SpecID    string
	SpecName  string
	ValueID   string
	ValueName string
	ValueCode string
}

func buildSpecMatrix(selections []SpecSelection) [][]specOption {
	matrix := make([][]specOption, 0, len(selections))
	for _, sel := range selections {
		cleanID := strings.TrimSpace(sel.SpecID)
		if cleanID == "" {
			continue
		}
		name := sel.SpecName
		if strings.TrimSpace(name) == "" {
			name = cleanID
		}
		meta := map[string]SpecValueSelection{}
		for _, v := range sel.Values {
			if strings.TrimSpace(v.ValueID) == "" {
				continue
			}
			meta[v.ValueID] = v
		}
		added := map[string]struct{}{}
		options := make([]specOption, 0, len(sel.ValueIDs))
		for _, raw := range sel.ValueIDs {
			id := strings.TrimSpace(raw)
			if id == "" {
				continue
			}
			if _, dup := added[id]; dup {
				continue
			}
			added[id] = struct{}{}
			value := meta[id]
			valueName := value.ValueName
			if strings.TrimSpace(valueName) == "" {
				valueName = id
			}
			options = append(options, specOption{
				SpecID:    cleanID,
				SpecName:  name,
				ValueID:   id,
				ValueName: valueName,
				ValueCode: value.ValueCode,
			})
		}
		if len(options) > 0 {
			matrix = append(matrix, options)
		}
	}
	return matrix
}

func expandCartesian(matrix [][]specOption) ([][]specOption, error) {
	if len(matrix) == 0 {
		return [][]specOption{{}}, nil
	}
	total := 1
	for _, cols := range matrix {
		total *= len(cols)
		if total > maxGeneratorCandidates {
			return nil, fmt.Errorf("too many combinations (%d), please narrow your selection", total)
		}
	}
	result := make([][]specOption, 0, total)
	var dfs func(idx int, current []specOption)
	dfs = func(idx int, current []specOption) {
		if idx == len(matrix) {
			result = append(result, append([]specOption(nil), current...))
			return
		}
		for _, opt := range matrix[idx] {
			current = append(current, opt)
			dfs(idx+1, current)
			current = current[:len(current)-1]
		}
	}
	dfs(0, []specOption{})
	return result, nil
}

func convertToSkuSpecs(combo []specOption) []SkuSpec {
	specs := make([]SkuSpec, 0, len(combo))
	for _, opt := range combo {
		specs = append(specs, SkuSpec{
			SpecID:    opt.SpecID,
			SpecName:  opt.SpecName,
			ValueID:   opt.ValueID,
			ValueName: opt.ValueName,
		})
	}
	return specs
}

func hashSpecValues(specs []SkuSpec) string {
	if len(specs) == 0 {
		return "__default__"
	}
	copySpecs := append([]SkuSpec(nil), specs...)
	sort.Slice(copySpecs, func(i, j int) bool {
		if copySpecs[i].SpecID == copySpecs[j].SpecID {
			return copySpecs[i].ValueID < copySpecs[j].ValueID
		}
		return copySpecs[i].SpecID < copySpecs[j].SpecID
	})
	parts := make([]string, 0, len(copySpecs))
	for _, spec := range copySpecs {
		parts = append(parts, spec.SpecID+":"+spec.ValueID)
	}
	return strings.Join(parts, "|")
}

func buildPreviewCode(base string, combo []specOption) string {
	tokens := []string{sanitizeCode(base)}
	for _, opt := range combo {
		valueToken := opt.ValueCode
		if strings.TrimSpace(valueToken) == "" {
			valueToken = opt.ValueName
		}
		if strings.TrimSpace(valueToken) == "" {
			valueToken = opt.ValueID
		}
		tokens = append(tokens, sanitizeCode(valueToken))
	}
	return strings.ToUpper(strings.Join(tokens, "-"))
}

func sanitizeCode(input string) string {
	if input == "" {
		return ""
	}
	builder := strings.Builder{}
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('-')
	}
	out := strings.Trim(builder.String(), "-")
	if out == "" {
		return "SKU"
	}
	return out
}

func (s *Service) resolveSPUCode(ctx context.Context, tenantID, spuID string) (string, error) {
	var code string
	if err := s.deps.DB.WithContext(ctx).
		Model(&productmodel.SPU{}).
		Where("tenant_uuid = ? AND id = ?", tenantID, spuID).
		Pluck("code", &code).Error; err != nil {
		return "", err
	}
	if strings.TrimSpace(code) == "" {
		return "", fmt.Errorf("spu %s not found", spuID)
	}
	return code, nil
}

func (s *Service) lookupExistingSpecHashes(ctx context.Context, tenantID, spuID string) (map[string]productskumodel.ProductSKU, map[string]struct{}, map[string]productskumodel.ProductSKU, error) {
	if s == nil || s.SKURepo == nil {
		return nil, nil, nil, errors.New("sku repository unavailable")
	}
	rows, err := s.SKURepo.ListBySPU(ctx, tenantID, spuID)
	if err != nil {
		return nil, nil, nil, err
	}
	result := make(map[string]productskumodel.ProductSKU, len(rows))
	codes := make(map[string]struct{}, len(rows))
	codeRecords := make(map[string]productskumodel.ProductSKU, len(rows))
	for _, row := range rows {
		code := strings.ToLower(strings.TrimSpace(row.SKUCode))
		if code != "" {
			codes[code] = struct{}{}
			codeRecords[code] = row
		}
		var specs []SkuSpec
		if len(row.SpecValues) > 0 {
			_ = json.Unmarshal(row.SpecValues, &specs)
		}
		hash := hashSpecValues(specs)
		if hash == "" {
			continue
		}
		result[hash] = row
	}
	return result, codes, codeRecords, nil
}
