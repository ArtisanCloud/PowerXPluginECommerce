package product_spec

import (
	"encoding/json"

	"gorm.io/datatypes"
)

func decodeMeta(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func encodeMeta(meta map[string]any) (datatypes.JSON, error) {
	if len(meta) == 0 {
		return datatypes.JSON([]byte("null")), nil
	}
	body, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(body), nil
}

