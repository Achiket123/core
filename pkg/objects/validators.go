package objects

import (
	"errors"
	"fmt"
	"strings"

	"github.com/theopenlane/core/pkg/objects/storage"
)

var ErrUnsupportedMimeType = errors.New("unsupported mime type uploaded")

// ValidationFunc is a type that can be used to dynamically validate a file
type ValidationFunc func(f storage.File) error

// MimeTypeValidator makes sure we only accept a valid mimetype.
// It takes in an array of supported mimes
func MimeTypeValidator(validMimeTypes ...string) ValidationFunc {
	return func(f storage.File) error {
		for _, mimeType := range validMimeTypes {
			if strings.EqualFold(strings.ToLower(mimeType), f.ContentType) {
				return nil
			}
		}

		return fmt.Errorf("%w: %s", ErrUnsupportedMimeType, f.ContentType)
	}
}

// ChainValidators returns a validator that accepts multiple validating criteria
func ChainValidators(validators ...ValidationFunc) ValidationFunc {
	return func(f storage.File) error {
		for _, validator := range validators {
			if err := validator(f); err != nil {
				return err
			}
		}

		return nil
	}
}
