package config

import "github.com/adeelahmad/snapback/internal/errcode"

// FieldError is one invalid field. Path is the YAML path, such as
// repositories[0].password_file. Code is empty for ordinary errors and names
// a specific errcode for the v0.1 rejections.
type FieldError struct {
	Path string
	Msg  string
	Code errcode.Code
}
