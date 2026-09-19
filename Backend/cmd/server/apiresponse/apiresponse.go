// Package apiresponse writes the two JSON envelopes every handler answers with.
//
// The frontend reads the failure message out of a JSON object (API_SCOPE.md §2:
// {"error": "both spouses are male"}), and shows that string to the user as-is.
// http.Error sends text/plain instead, which the browser reports as "The API did
// not answer with JSON (...)" — so a handler that reaches for it loses the
// message it was trying to deliver. Both routers share these helpers so the
// envelopes cannot drift apart.
package apiresponse

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Write answers with body encoded as JSON. Every successful response goes
// through here so the content type is set before the status is written; setting
// a header afterwards has no effect.
func Write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		// The status line is already on the wire, so this can only be logged.
		log.Printf("Failed To Encode The JSON Response: %v", err)
	}
}

// Error answers with {"error": message}. A message that is empty would leave the
// frontend with nothing to show, so callers are expected to pass a user-facing
// sentence — the errors.New texts in the domain packages already read well.
func Error(w http.ResponseWriter, status int, message string) {
	Write(w, status, map[string]string{"error": message})
}

// UserMessage picks the sentence to show for a rejected write.
//
// The domain packages classify failures by wrapping a sentinel with "%w", as in
//
//	fmt.Errorf("%w: name must not be empty", ErrInvalidUpdate)
//
// That wrapping is what makes errors.Is work, but it also puts "invalid update: "
// in front of the text that was written to be read. This strips the sentinel back
// off, so a dialog shows the sentence the package wrote for a person rather than
// the label it wrote for a machine.
func UserMessage(err error, sentinel error) string {
	message := strings.TrimSpace(err.Error())
	trimmed := strings.TrimSpace(strings.TrimPrefix(message, sentinel.Error()+":"))

	if trimmed == "" || trimmed == message {
		return message
	}

	return trimmed
}
