package schemas

import _ "embed"

// PackSchemaJSON is the embedded pack.schema.json for validation.
//go:embed pack.schema.json
var PackSchemaJSON []byte
