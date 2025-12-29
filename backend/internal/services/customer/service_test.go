package customer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceCreateCustomer(t *testing.T) {
	svc := newTestService(t)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-create")
	input := CreateCustomerInput{
		Name:           "测试客户",
		Type:           "individual",
		Phone:          "+8613800000001",
		Email:          "test@example.com",
		MembershipTier: "gold",
		Source:         "website",
		Tags:           []string{"VIP"},
		Notes:          "via unit test",
	}
	result, err := svc.CreateCustomer(ctx, input)
	require.NoError(t, err)
	require.NotEmpty(t, result.ID)
	require.Equal(t, "gold", result.MembershipTier)
	require.Equal(t, "via unit test", result.Notes)

	list, err := svc.ListCustomers(ctx, ListFilters{})
	require.NoError(t, err)
	found := false
	for _, customer := range list.Data {
		if customer.ID == result.ID {
			found = true
			break
		}
	}
	require.True(t, found, "newly created customer should be returned in list")
}

func TestServiceUpdateCustomerConflict(t *testing.T) {
	svc := newTestService(t)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-update")
	first, err := svc.CreateCustomer(ctx, CreateCustomerInput{
		Name:           "A",
		Type:           "individual",
		Phone:          "+8613800000002",
		Email:          "first@example.com",
		MembershipTier: "gold",
		Source:         "website",
	})
	require.NoError(t, err)
	second, err := svc.CreateCustomer(ctx, CreateCustomerInput{
		Name:           "B",
		Type:           "individual",
		Phone:          "+8613800000003",
		Email:          "second@example.com",
		MembershipTier: "gold",
		Source:         "website",
	})
	require.NoError(t, err)
	conflictEmail := first.Email
	err = nil
	_, err = svc.UpdateCustomer(ctx, second.ID, UpdateCustomerInput{Email: &conflictEmail})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCustomerConflict)
}

func TestServiceDeleteRequiresReason(t *testing.T) {
	svc := newTestService(t)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-delete")
	created, err := svc.CreateCustomer(ctx, CreateCustomerInput{
		Name:           "待删除",
		Type:           "individual",
		Phone:          "+8613900000000",
		MembershipTier: "gold",
		Source:         "website",
	})
	require.NoError(t, err)
	_, err = svc.DeleteCustomer(ctx, created.ID, "")
	require.Error(t, err)
	var validationErrs ValidationErrors
	require.True(t, errorsAs(err, &validationErrs))
	require.Len(t, validationErrs, 1)
}

func TestServiceListCustomersDBFilters(t *testing.T) {
	svc := newTestService(t)
	tenant := "tenant-db-filters"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenant)
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("DB Filter %d", i)
		phone := fmt.Sprintf("+8613800001%03d", i)
		email := fmt.Sprintf("db-filter-%d@example.com", i)
		_, err := svc.CreateCustomer(ctx, CreateCustomerInput{
			Name:           name,
			Type:           "enterprise",
			Phone:          phone,
			Email:          email,
			Region:         "华东/上海",
			Source:         "autotest",
			Tags:           []string{"speckit", fmt.Sprintf("cohort-%d", i)},
			MembershipTier: "gold",
		})
		require.NoError(t, err)
	}
	_, err := svc.CreateCustomer(ctx, CreateCustomerInput{
		Name:           "DB Filter Control",
		Type:           "individual",
		Phone:          "+8613800002999",
		Email:          "db-filter-control@example.com",
		Source:         "control-source",
		MembershipTier: "silver",
	})
	require.NoError(t, err)

	result, err := svc.ListCustomers(ctx, ListFilters{Source: "autotest"})
	require.NoError(t, err)
	require.Equal(t, 3, result.Meta.Total)

	keywordResult, err := svc.ListCustomers(ctx, ListFilters{
		Source:  "autotest",
		Keyword: "db filter 1",
	})
	require.NoError(t, err)
	require.Len(t, keywordResult.Data, 1)

	tagResult, err := svc.ListCustomers(ctx, ListFilters{
		Source: "autotest",
		Tags:   []string{"speckit", "cohort-2"},
	})
	require.NoError(t, err)
	require.Len(t, tagResult.Data, 1)
}

func TestServiceListCustomersDBPaginationAndSorting(t *testing.T) {
	svc := newTestService(t)
	tenant := "tenant-db-pagination"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenant)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("Paging %02d", i)
		phone := fmt.Sprintf("+861380010%03d", i)
		email := fmt.Sprintf("paging-%02d@example.com", i)
		_, err := svc.CreateCustomer(ctx, CreateCustomerInput{
			Name:           name,
			Type:           "individual",
			Phone:          phone,
			Email:          email,
			Source:         "paging",
			MembershipTier: "silver",
		})
		require.NoError(t, err)
	}

	result, err := svc.ListCustomers(ctx, ListFilters{
		Source:   "paging",
		Sort:     "name",
		Page:     2,
		PageSize: 2,
	})
	require.NoError(t, err)
	require.Len(t, result.Data, 2)
	require.Equal(t, "Paging 02", result.Data[0].Name)
	require.Equal(t, "Paging 03", result.Data[1].Name)
	require.Equal(t, 5, result.Meta.Total)
}

func TestServiceDeleteCustomerRemovesRecord(t *testing.T) {
	svc := newTestService(t)
	tenant := "tenant-delete-db"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenant)
	created, err := svc.CreateCustomer(ctx, CreateCustomerInput{
		Name:           "Delete Me",
		Type:           "individual",
		Phone:          "+8613800990001",
		Email:          "delete-me@example.com",
		Source:         "cleanup",
		MembershipTier: "gold",
	})
	require.NoError(t, err)

	_, err = svc.DeleteCustomer(ctx, created.ID, "cleanup")
	require.NoError(t, err)

	result, err := svc.ListCustomers(ctx, ListFilters{
		Keyword: "delete me",
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(result.Data))
}

func errorsAs(err error, target *ValidationErrors) bool {
	return err != nil && target != nil && errors.As(err, target)
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	db := newTestDB(t)
	return NewService(&app.Deps{DB: db})
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	models.ForceSchemaForTests("")
	dsn := fmt.Sprintf("file:customer_service_%s?mode=memory&cache=shared", utils.NewUUID())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&customermodel.Customer{}))
	return db
}
