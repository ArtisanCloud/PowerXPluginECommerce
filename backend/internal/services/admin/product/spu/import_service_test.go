package spu

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImportService_PartialSuccessSummary(t *testing.T) {
	records := []ImportRecord{
		{Row: 2, Code: "SPU-100", Name: "Alpha", SKUCodes: []string{"SKU-A"}},
		{Row: 3, Code: "SPU-100", Name: "Beta", SKUCodes: []string{"SKU-B"}},
		{Row: 4, Code: "SPU-200", Name: "Gamma", SKUCodes: []string{"SKU-A"}},
	}
	summary := validateImportRecords(records)
	require.Equal(t, 1, summary.SuccessCount)
	require.Len(t, summary.Failures, 2)
	require.Contains(t, summary.Failures[0].Reason, "SPU 编码")
	require.Contains(t, summary.Failures[1].Reason, "SKU 编码")
	require.Equal(t, "partial", statusFromSummary(summary))
}

func TestImportService_AllSuccessSummary(t *testing.T) {
	records := []ImportRecord{
		{Row: 2, Code: "SPU-1", Name: "One"},
		{Row: 3, Code: "SPU-2", Name: "Two", SKUCodes: []string{"SKU-X"}},
	}
	summary := validateImportRecords(records)
	require.Equal(t, 2, summary.SuccessCount)
	require.Empty(t, summary.Failures)
	require.Equal(t, "success", statusFromSummary(summary))
}

func TestPersistFailureReport(t *testing.T) {
	failures := []ImportFailure{
		{Row: 5, Code: "SPU-ERR", Reason: "缺少名称"},
	}
	path, err := persistFailureReport("test-task", failures)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(path) })
	require.FileExists(t, path)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), "SPU-ERR")
	require.Equal(t, filepath.Join("tmp", "spu-import-report-test-task.json"), path)
}
