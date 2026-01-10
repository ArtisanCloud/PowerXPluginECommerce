// cmd/database/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/cmd/database/migrate"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/cmd/database/seed"
	pluginbootstrap "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/bootstrap"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/db"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	iamservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/iam"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s [migrate|seed|setup|refresh|pricing-base-upsert]", os.Args[0])
	}
	cmd := os.Args[1]

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	models.InitSchemaFrom(cfg.Database.Schema) // 必须在所有 DB 操作之前

	iamResolver := pluginbootstrap.NewIAMResolver(cfg)
	includeIAM := iamResolver.Mode() == iamservice.IAMModeLocal

	ctx := context.Background()
	// 连接数据库
	db, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	switch cmd {
	case "migrate":
		cmdFlags := flag.NewFlagSet(cmd, flag.ExitOnError)
		_ = cmdFlags.Parse(os.Args[2:])
		if err := migrate.MigratePluginModels(ctx, db, includeIAM); err != nil {
			log.Fatal("migrate failed:", err)
		}
		fmt.Println("migrate ok")

	case "seed":
		cmdFlags := flag.NewFlagSet(cmd, flag.ExitOnError)
		_ = cmdFlags.Parse(os.Args[2:])
		if includeIAM {
			if err := iamservice.SeedLocalAdmin(ctx, db, cfg); err != nil {
				log.Fatal("iam seed failed:", err)
			}
		}
		if err := seed.SeedPluginData(ctx, db); err != nil {
			log.Fatal("seed failed:", err)
		}
		fmt.Println("seed ok")

	case "setup":
		cmdFlags := flag.NewFlagSet(cmd, flag.ExitOnError)
		_ = cmdFlags.Parse(os.Args[2:])
		if err := migrate.MigratePluginModels(ctx, db, includeIAM); err != nil {
			log.Fatal("migrate failed:", err)
		}
		fmt.Println("migrate ok")

		if includeIAM {
			if err := iamservice.SeedLocalAdmin(ctx, db, cfg); err != nil {
				log.Fatal("iam seed failed:", err)
			}
		}
		if err := seed.SeedPluginData(ctx, db); err != nil {
			log.Fatal("seed failed:", err)
		}
		fmt.Println("seed ok")

	case "refresh":
		cmdFlags := flag.NewFlagSet(cmd, flag.ExitOnError)
		_ = cmdFlags.Parse(os.Args[2:])
		// 先 drop database（或 drop all tables）
		if err := migrate.ResetDatabase(ctx, db, cfg.Database); err != nil {
			log.Fatal("reset failed:", err)
		}
		fmt.Println("reset ok")

		// 再 migrate
		if err := migrate.MigratePluginModels(ctx, db, includeIAM); err != nil {
			log.Fatal("migrate failed:", err)
		}
		fmt.Println("migrate ok")

		// 最后 seed
		if includeIAM {
			if err := iamservice.SeedLocalAdmin(ctx, db, cfg); err != nil {
				log.Fatal("iam seed failed:", err)
			}
		}
		if err := seed.SeedPluginData(ctx, db); err != nil {
			log.Fatal("seed failed:", err)
		}
		fmt.Println("seed ok")

	case "pricing-base-upsert":
		cmdFlags := flag.NewFlagSet(cmd, flag.ExitOnError)
		currency := cmdFlags.String("currency", "USD", "base pricebook currency when created")
		dryRun := cmdFlags.Bool("dry-run", false, "only print actions without writing to DB")
		tenants := cmdFlags.String("tenant-uuids", "", "comma-separated tenant UUID list (optional)")
		fromIAM := cmdFlags.Bool("from-iam", true, "discover tenants from iam_tenants (when available)")
		fromData := cmdFlags.Bool("from-existing-data", true, "discover tenants from existing tenant_uuid columns in business tables")
		if err := cmdFlags.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		if err := upsertBasePricebooks(ctx, db, basePricebookUpsertOptions{
			Currency:         *currency,
			DryRun:           *dryRun,
			TenantUUIDs:      *tenants,
			DiscoverFromIAM:  *fromIAM,
			DiscoverFromData: *fromData,
		}); err != nil {
			log.Fatal("pricing-base-upsert failed:", err)
		}
		fmt.Println("pricing-base-upsert ok")

	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}
