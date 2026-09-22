package web

import (
	"errors"

	"github.com/adeelahmad/snapback/internal/config"
)

// fieldErrors splits err into per-field form messages and page-level banner
// lines. byPath keeps the first message for each field path; every other
// message, including pathless and non-validation errors, becomes a banner line
// in the order it was reported, so nothing is dropped.
func fieldErrors(err error) (byPath map[string]string, banner []string) {
	byPath = make(map[string]string)
	if err == nil {
		return byPath, nil
	}

	var verr *config.ValidationError
	if !errors.As(err, &verr) {
		return byPath, []string{err.Error()}
	}

	for _, f := range verr.Fields {
		_, seen := byPath[f.Path]
		switch {
		case f.Path == "":
			banner = append(banner, f.Msg)
		case seen:
			banner = append(banner, f.Path+": "+f.Msg)
		default:
			byPath[f.Path] = f.Msg
		}
	}
	return byPath, banner
}
