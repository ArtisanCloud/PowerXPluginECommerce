package membership

import "time"

type GrantEntitlementRequest struct {
	CustomerID  string
	ServiceCode string
	Quantity    int64
	ValidFrom   *time.Time
	ValidTo     *time.Time
	StackPolicy string
	Reason      string
	ActorID     string
	SourceID    string
}

type RevokeEntitlementRequest struct {
	EntitlementID string
	Reason        string
	ActorID       string
}

type AdjustTokenRequest struct {
	CustomerID string
	TokenCode  string
	Delta      int64
	Reason     string
	ActorID    string
	SourceID   string
}
