package core

import (
	_ "embed"
)

//go:embed core.swagger.json
var SwaggerJSON []byte

func GetSwaggerJSON() []byte {
	return SwaggerJSON
}
