// api/docs/docs.go
package docs

import (
	"github.com/swaggo/swag"
)

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "Capstone Core API",
	Description:      "This is the API documentation for Capstone Core System.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

// docTemplate holds the Swagger template
const docTemplate = `{
    "swagger": "2.0",
    "info": {{ marshal .Info }},
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {}
}`
