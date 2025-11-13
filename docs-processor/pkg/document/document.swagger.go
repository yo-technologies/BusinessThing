package document

import (
	_ "embed"
)

//go:embed document.swagger.json
var swaggerJSON []byte

func GetSwaggerJSON() []byte {
	return swaggerJSON
}
