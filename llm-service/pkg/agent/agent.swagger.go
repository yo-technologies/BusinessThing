package agent

import (
	_ "embed"
)

//go:embed agent.swagger.json
var swaggerJSON []byte

func GetSwaggerJSON() []byte {
	return swaggerJSON
}
