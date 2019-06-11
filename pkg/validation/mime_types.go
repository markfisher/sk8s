package validation

import (
	"github.com/knative/pkg/apis"
	"strings"
)

func MimeType(mimeType, field string) *apis.FieldError {
	errs := &apis.FieldError{}

	index := strings.Index(mimeType, "/")
	if index == -1 || index == len(mimeType)-1 {
		errs = errs.Also(apis.ErrInvalidValue(mimeType, field))
	}

	return errs
}
