package common

import (
	"errors"

	cmlerrors "github.com/rschmied/gocmlclient/pkg/errors"
)

// IsNotFound reports true when CML returns a 404-class error.
func IsNotFound(err error) bool {
	return errors.Is(err, cmlerrors.ErrElementNotFound) || errors.Is(err, cmlerrors.ErrAPINotFound)
}
