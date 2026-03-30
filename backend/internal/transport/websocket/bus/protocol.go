package bus

import (
	"encoding/json"
	"time"
)

const (
	CmdSubscribe   = "subscribe"
	CmdUnsubscribe = "unsubscribe"
	CmdPing        = "ping"

	MsgTypeWelcome = "welcome"
	MsgTypeAck     = "ack"
	MsgTypeError   = "error"
	MsgTypeEvent   = "event"
)

type Envelope struct {
	Topic     string          `json:"topic,omitempty"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp int64           `json:"ts"`
	TraceID   string          `json:"trace_id,omitempty"`
}

func NewEnvelope(typ, topic string, payload any, traceID string) (Envelope, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return Envelope{}, err
		}
		raw = b
	}
	return Envelope{
		Topic:     topic,
		Type:      typ,
		Payload:   raw,
		Timestamp: time.Now().UTC().UnixMilli(),
		TraceID:   traceID,
	}, nil
}

type Command struct {
	Type   string   `json:"type"`
	Topic  string   `json:"topic,omitempty"`
	Topics []string `json:"topics,omitempty"`
	ReqID  string   `json:"req_id,omitempty"`
}

type WelcomePayload struct {
	Protocol     string `json:"protocol"`
	Server       string `json:"server"`
	HeartbeatSec int    `json:"heartbeat_sec,omitempty"`
}

type AckPayload struct {
	ReqID   string   `json:"req_id,omitempty"`
	OK      bool     `json:"ok"`
	Message string   `json:"message,omitempty"`
	Topics  []string `json:"topics,omitempty"`
}

type ErrorPayload struct {
	ReqID   string `json:"req_id,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}
