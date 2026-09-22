package doctor

import "github.com/adeelahmad/snapback/internal/errcode"

// Check is one doctor check result.
type Check struct {
	Name   string       `json:"name"`
	Status string       `json:"status"`
	Code   errcode.Code `json:"code"`
	Detail string       `json:"detail"`
	Fix    string       `json:"fix"`
}
