package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	configpkg "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NewIsolatedDB returns a database connection that has been fully cleared.
// When TEST_DATABASE_URL is provided we reuse that Postgres instance, otherwise we fall back to sqlite.
func NewIsolatedDB(t *testing.T) (*gorm.DB, bool) {
	t.Helper()
	basemodels.ForceSchemaForTests("")

	if dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL")); dsn != "" {
		return openPostgresDB(t, dsn), true
	}

	if dsn, ok := loadDSNFromEnvConfig(t); ok {
		return openPostgresDB(t, dsn), true
	}
	if dsn, ok := loadDSNFromDefaultConfigs(t); ok {
		return openPostgresDB(t, dsn), true
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	dropAllTables(t, db, false)
	return db, false
}

func openPostgresDB(t *testing.T, dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	dropAllTables(t, db, true)
	return db
}

func loadDSNFromEnvConfig(t *testing.T) (string, bool) {
	configPath := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	if configPath == "" {
		return "", false
	}
	if dsn, ok := loadDSNFromPath(t, configPath); ok {
		return dsn, true
	}
	return "", false
}

func loadDSNFromDefaultConfigs(t *testing.T) (string, bool) {
	candidates := []string{
		filepath.Join("etc", "config.test.yaml"),
		filepath.Join("..", "etc", "config.test.yaml"),
		filepath.Join("backend", "etc", "config.test.yaml"),
		filepath.Join("etc", "config.ci.yaml"),
		filepath.Join("..", "etc", "config.ci.yaml"),
		filepath.Join("backend", "etc", "config.ci.yaml"),
	}
	for _, candidate := range candidates {
		if dsn, ok := loadDSNFromPath(t, candidate); ok {
			return dsn, true
		}
	}
	return "", false
}

func loadDSNFromPath(t *testing.T, path string) (string, bool) {
	resolved := filepath.Clean(path)
	if !filepath.IsAbs(resolved) {
		if _, err := os.Stat(resolved); err != nil {
			wd, wdErr := os.Getwd()
			if wdErr == nil {
				resolved = filepath.Join(wd, resolved)
			}
		}
	}
	if _, err := os.Stat(resolved); err != nil {
		return "", false
	}

	cfg, err := loadConfigWithPath(resolved)
	if err != nil {
		t.Logf("failed to load config from %s: %v", resolved, err)
		return "", false
	}
	if cfg == nil || cfg.Database == nil {
		return "", false
	}
	driver := strings.ToLower(strings.TrimSpace(cfg.Database.Driver))
	dsn := strings.TrimSpace(cfg.Database.DSN)
	if driver != "postgres" || dsn == "" {
		return "", false
	}
	return dsn, true
}

func loadConfigWithPath(path string) (*configpkg.Config, error) {
	original := os.Getenv("CONFIG_PATH")
	defer func() {
		if original == "" {
			_ = os.Unsetenv("CONFIG_PATH")
		} else {
			_ = os.Setenv("CONFIG_PATH", original)
		}
	}()
	if err := os.Setenv("CONFIG_PATH", path); err != nil {
		return nil, err
	}
	return configpkg.Load()
}

func dropAllTables(t *testing.T, db *gorm.DB, isPostgres bool) {
	t.Helper()
	var tables []string
	if isPostgres {
		require.NoError(t, db.Raw(`SELECT tablename FROM pg_tables WHERE schemaname = current_schema()`).Scan(&tables).Error)
		for _, table := range tables {
			require.NoError(t, db.Exec(`DROP TABLE IF EXISTS "`+table+`" CASCADE`).Error)
		}
		return
	}

	require.NoError(t, db.Raw(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`).Scan(&tables).Error)
	for _, table := range tables {
		require.NoError(t, db.Exec(`DROP TABLE IF EXISTS "`+table+`"`).Error)
	}
}
