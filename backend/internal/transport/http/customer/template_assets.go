package customer

import _ "embed"

const importTemplateFilename = "customer_import_template.csv"

//go:embed assets/customer_import_template.csv
var customerImportTemplate []byte
