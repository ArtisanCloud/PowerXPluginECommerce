package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		spuID    = flag.String("spu-id", "", "target SPU ID")
		tenantID = flag.String("tenant-id", "", "tenant UUID")
		reason   = flag.String("reason", "", "deletion reason")
		dryRun   = flag.Bool("dry-run", true, "only print SQL statements")
	)
	flag.Parse()
	if *spuID == "" || *tenantID == "" {
		log.Fatal("spu-id and tenant-id are required")
	}
	if *reason == "" {
		log.Fatal("reason is required to keep audit trace")
	}
	fmt.Fprintf(os.Stdout, "Preparing to hard delete SPU %s (tenant %s)\n", *spuID, *tenantID)
	fmt.Fprintln(os.Stdout, "Steps:")
	fmt.Fprintln(os.Stdout, "1. Verify soft-delete date >= 30 days and ensure no external references remain.")
	fmt.Fprintln(os.Stdout, "2. Execute the following SQL statements inside a transaction:")
	sql := fmt.Sprintf(`
DELETE FROM product_spu_audit_logs WHERE tenant_uuid = '%s' AND spu_id = '%s';
DELETE FROM product_spu_subscription_plans WHERE tenant_uuid = '%s' AND spu_id = '%s';
DELETE FROM product_spu_channels WHERE tenant_uuid = '%s' AND spu_id = '%s';
DELETE FROM product_spu_versions WHERE tenant_uuid = '%s' AND spu_id = '%s';
DELETE FROM product_spu_import_tasks WHERE tenant_uuid = '%s' AND detail->>'spuId' = '%s';
DELETE FROM product_spu_export_tasks WHERE tenant_uuid = '%s' AND detail->>'spuId' = '%s';
DELETE FROM product_spus WHERE tenant_uuid = '%s' AND id = '%s';
`, *tenantID, *spuID,
		*tenantID, *spuID,
		*tenantID, *spuID,
		*tenantID, *spuID,
		*tenantID, *spuID,
		*tenantID, *spuID,
		*tenantID, *spuID,
	)
	fmt.Fprintln(os.Stdout, sql)
	if *dryRun {
		fmt.Fprintln(os.Stdout, "Dry-run mode enabled. No changes executed.")
	} else {
		fmt.Fprintln(os.Stdout, "Execute these statements manually or extend this tool to connect to DB.")
	}
	fmt.Fprintf(os.Stdout, "Reason recorded: %s\n", *reason)
}
