package spu

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// decodeVersionPayload safely unmarshals version payload JSON into a map for mutation.
func decodeVersionPayload(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

// encodeVersionPayload marshals the payload map back into JSON format.
func encodeVersionPayload(payload map[string]any) datatypes.JSON {
	if payload == nil {
		payload = map[string]any{}
	}
	body, _ := json.Marshal(payload)
	return datatypes.JSON(body)
}
