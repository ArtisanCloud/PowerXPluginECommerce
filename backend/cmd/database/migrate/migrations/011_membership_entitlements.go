package migrations

import (
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
)

// MembershipEntitlementTables defines membership/entitlement related tables.
var MembershipEntitlementTables = []interface{}{
	&membershipModel.MembershipTier{},
	&membershipModel.MembershipBenefit{},
	&membershipModel.MembershipAssignment{},
	&membershipModel.Entitlement{},
	&membershipModel.TokenAccount{},
	&membershipModel.TokenTransaction{},
}
