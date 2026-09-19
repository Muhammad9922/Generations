package person

import "errors"

// Sentinel errors let callers (and the HTTP layer) tell a bad request apart
// from a genuine server-side failure, instead of having to match on message
// text.
var (
	// ErrInvalidUpdate reports data supplied by the caller that can never be
	// applied, such as an unparsable date or an empty name.
	ErrInvalidUpdate = errors.New("invalid update")

	// ErrNotFound reports that no person with the requested id exists.
	ErrNotFound = errors.New("person not found")
)
