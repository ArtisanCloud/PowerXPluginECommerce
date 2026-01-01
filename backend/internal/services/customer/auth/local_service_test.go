package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLocalServiceRegisterLoginAndAuthenticate(t *testing.T) {
	svc, ctx := newLocalServiceForTest(t)

	registerResult, err := svc.Register(ctx, RegisterInput{
		Name:       "Alice",
		Identifier: "alice@example.com",
		Password:   "secret123",
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerResult.Token)

	loginResult, err := svc.Login(ctx, LoginInput{
		Identifier: "alice@example.com",
		Password:   "secret123",
	})
	require.NoError(t, err)
	require.Equal(t, registerResult.CustomerID, loginResult.CustomerID)

	authCtx, err := svc.Authenticate(ctx, loginResult.Token)
	require.NoError(t, err)
	require.Equal(t, loginResult.CustomerID, authCtx.CustomerID)
	require.Equal(t, loginResult.TenantUUID, authCtx.TenantUUID)
}

func TestLocalServiceRejectsDuplicateIdentifier(t *testing.T) {
	svc, ctx := newLocalServiceForTest(t)

	_, err := svc.Register(ctx, RegisterInput{
		Name:       "Bob",
		Identifier: "bob@example.com",
		Password:   "secret123",
	})
	require.NoError(t, err)

	_, err = svc.Register(ctx, RegisterInput{
		Name:       "Bob Clone",
		Identifier: "bob@example.com",
		Password:   "secret123",
	})
	require.ErrorIs(t, err, ErrIdentifierExists)
}

func newLocalServiceForTest(t *testing.T) (*LocalService, context.Context) {
	t.Helper()
	models.ForceSchemaForTests("")
	dsn := fmt.Sprintf("file:customer_local_auth_%s?mode=memory&cache=shared", utils.NewUUID())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&customermodel.Customer{}, &customermodel.CustomerAccount{}))

	svc, err := NewLocalService(db, LocalConfig{
		JWTSecret: []byte("test-secret"),
		Issuer:    "test-issuer",
		Audience:  "test-miniapp",
	})
	require.NoError(t, err)
	ctx := authx.ContextWithTenantUUID(context.Background(), "00000000-0000-0000-0000-000000000001")
	return svc, ctx
}
