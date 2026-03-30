package membership

import "time"

type CreateTierRequest struct {
	Name    string
	Code    string
	Status  string
	Rules   []byte
	ActorID string
}

type CreateBenefitRequest struct {
	Name    string
	Type    string
	Status  string
	Items   []byte
	ActorID string
}

type UpdateTierStatusRequest struct {
	Status  string
	ActorID string
}

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

type ListTokenTransactionsRequest struct {
	CustomerID  string
	TokenCode   string
	SourceType  string
	SourceID    string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Page        int
	PageSize    int
}

type RedeemPointsBenefitRequest struct {
	CustomerID string
	BenefitID  string
	PointsCost int64
	Reason     string
	ActorID    string
	SourceID   string
}

type RedeemPointsBenefitResult struct {
	BenefitID       string
	CustomerID      string
	TransactionID   string
	Balance         int64
	GrantedServices []string
}
