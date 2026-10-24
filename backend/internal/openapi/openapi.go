// Package openapi встраивает в бинарник спецификацию API и страницу Swagger UI.
package openapi

import _ "embed"

//go:embed openapi.yaml
var Spec []byte

//go:embed swagger.html
var SwaggerUI []byte
