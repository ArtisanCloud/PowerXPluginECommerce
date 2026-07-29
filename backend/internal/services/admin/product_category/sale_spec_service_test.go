package product_category

import "testing"

func TestValidateSaleSpecInputsRejectsEmptyOptions(t *testing.T) {
	err := validateSaleSpecInputs([]SaleSpecGroupInput{{
		Code: "size",
		Name: "Size",
	}})
	if err == nil {
		t.Fatal("expected empty option list to fail")
	}
}

func TestValidateSaleSpecInputsRejectsDuplicateCodes(t *testing.T) {
	err := validateSaleSpecInputs([]SaleSpecGroupInput{
		{
			Code: "size",
			Name: "Size",
			Options: []SaleSpecOptionInput{{
				Code: "10cm",
				Name: "10cm",
			}},
		},
		{
			Code: "SIZE",
			Name: "Size Again",
			Options: []SaleSpecOptionInput{{
				Code: "20cm",
				Name: "20cm",
			}},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate group code to fail")
	}
}

func TestNormalizeSaleSpecStatusRejectsUnknownStatus(t *testing.T) {
	if _, err := normalizeSaleSpecStatus("deprecated"); err == nil {
		t.Fatal("expected unknown status to fail")
	}
}
