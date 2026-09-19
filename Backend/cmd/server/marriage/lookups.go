package marriageRouter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// unknownPersonError reports an id that belongs to nobody.
//
// It carries its own user-facing sentence, because the dialogs show whatever the
// API answers with, and its type lets a handler tell "the client named a person
// that does not exist" apart from "the lookup itself broke".
type unknownPersonError struct {
	id string
}

func (e unknownPersonError) Error() string {
	return fmt.Sprintf("No Person Exists With ID %v", e.id)
}

// requirePeople rejects any id that does not belong to a person.
//
// Every id the marriage endpoints receive is supposed to belong to somebody
// already: the dialogs create a person through POST /people first and then pass
// the id it handed back. integration.CreateFamily would instead try to *create*
// anyone it cannot find, so a stale or mistyped id would surface as "The
// Person's Name Must Be Given" from its create path rather than as the real
// problem.
func requirePeople(ctx context.Context, driver neo4j.Driver, ids ...string) error {
	for _, id := range ids {
		exists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{ID: id})
		if err != nil {
			return fmt.Errorf("failed to look up person %s: %w", id, err)
		}
		if !exists {
			return unknownPersonError{id: id}
		}
	}

	return nil
}

// requireMarriage wraps marriage.CheckMarriageExistence so a missing id comes
// back as marriage.ErrNotFound rather than as a query-level error.
func requireMarriage(ctx context.Context, driver neo4j.Driver, id string) error {
	exists, err := marriage.CheckMarriageExistence(ctx, driver, id)
	if err != nil {
		return fmt.Errorf("failed to look up marriage %s: %w", id, err)
	}
	if !exists {
		return fmt.Errorf("%w: %s", marriage.ErrNotFound, id)
	}

	return nil
}

// respondLookupError answers a failed requirePeople / requireMarriage call.
//
// A person who does not exist is the client's mistake, so it is a 404 when the
// id came from the path and a 400 when it came from the body and the client
// could have checked first (API_SCOPE.md §4.5). A lookup that itself broke is
// ours, so it stays a 500 rather than being dressed up as "not found".
func respondLookupError(w http.ResponseWriter, err error, idCameFromBody bool) {
	var unknown unknownPersonError

	switch {
	case errors.As(err, &unknown):
		if idCameFromBody {
			writeError(w, http.StatusBadRequest, unknown.Error())
		} else {
			writeError(w, http.StatusNotFound, "Person not found")
		}
	case errors.Is(err, marriage.ErrNotFound):
		writeError(w, http.StatusNotFound, "Marriage not found")
	default:
		log.Printf("Unable To Look Up A Record: %v", err)
		writeError(w, http.StatusInternalServerError, "Unable To Look Up The Record")
	}
}
