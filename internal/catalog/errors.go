package catalog

import "errors"

// ErrNotFound is returned when a product id does not exist.
var ErrNotFound = errors.New("product not found")
