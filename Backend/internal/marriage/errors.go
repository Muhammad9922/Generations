package marriage

import "errors"

// Sentinel errors let callers (and the HTTP layer) tell a bad request apart from
// a genuine server-side failure, instead of having to match on message text.
var (
	// ErrNotFound reports that no marriage with the requested id exists.
	ErrNotFound = errors.New("marriage not found")

	// ErrInvalidUpdate reports dates the caller supplied that can never be
	// stored, such as a malformed date or a marriage that ends before it starts.
	ErrInvalidUpdate = errors.New("invalid marriage update")
)
