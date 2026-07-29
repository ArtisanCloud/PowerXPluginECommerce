package product_spec

import "testing"

func TestValidateReplaceRequestRejectsEmptyOptions(t *testing.T) {
	err := validateReplaceRequest(ReplaceRequest{Groups: []GroupInput{{
		Code: "size",
		Name: "Size",
	}}})
	if err == nil {
		t.Fatal("expected empty option list to fail")
	}
}

func TestValidateReplaceRequestRejectsDuplicateOptionCodes(t *testing.T) {
	err := validateReplaceRequest(ReplaceRequest{Groups: []GroupInput{{
		Code: "size",
		Name: "Size",
		Options: []OptionInput{
			{Code: "10cm", Name: "10cm"},
			{Code: "10CM", Name: "10cm duplicate"},
		},
	}}})
	if err == nil {
		t.Fatal("expected duplicate option code to fail")
	}
}

func TestNormalizeSpecStatusRejectsLegacyDeprecatedStatus(t *testing.T) {
	if _, err := normalizeSpecStatus("deprecated"); err == nil {
		t.Fatal("expected deprecated status to fail")
	}
}
