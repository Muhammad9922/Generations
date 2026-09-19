package person

import (
	"context"
	"errors"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// UpdateUser is a partial profile edit. Every field is a pointer so that an
// omitted field can keep its stored value:
//
//	nil        → the field was not sent, leave it alone
//	&"" (dates)→ the caller asked to clear the date
//	&value     → store the new value
//
// A date has no "absent" spelling of its own once it is decoded into a bare
// DateProper, which is exactly why the clear case needs the empty pointer.
type UpdateUser struct {
	Name        *string
	DateOfBirth *DateProper
	Gender      *Gender
	Alive       *bool
	DateOfDeath *DateProper
}

func Ptr[T any](v T) *T {
	return &v
}

// parseUpdateDate validates one incoming date. The empty string is the clear
// sentinel and passes straight through; every other value must be a real
// calendar day in DD-MM-YYYY.
func parseUpdateDate(field string, value DateProper) (DateProper, error) {
	if value == "" {
		return "", nil
	}

	parsed, err := ParseDateProper(string(value))
	if err != nil {
		return "", fmt.Errorf("%w: invalid %s: %v", ErrInvalidUpdate, field, err)
	}

	return parsed, nil
}

func UpdatePerson(ctx context.Context, driver neo4j.Driver, id string, update UpdateUser) (string, bool, error) {
	props := make(map[string]any)

	if update.Alive != nil {
		props["alive"] = *update.Alive
	}

	if update.Name != nil {
		if *update.Name == "" {
			return "", false, fmt.Errorf("%w: name must not be empty", ErrInvalidUpdate)
		}
		props["name"] = *update.Name
	}

	if update.Gender != nil {
		if !update.Gender.IsValid() {
			return "", false, fmt.Errorf("%w: %q is not a gender this app stores", ErrInvalidUpdate, *update.Gender)
		}
		props["gender"] = *update.Gender
	}

	// Dates are parsed into the canonical DD-MM-YYYY representation up front.
	//
	// The previous implementation validated the YYYY-MM-DD value returned by
	// NormalizeDate with DateProper.IsValid, which only accepts DD-MM-YYYY, so
	// the validation could never pass and every date update failed.
	//
	// An empty value is a clear, not a missing field: the PATCH body spells that
	// as `null` (API_SCOPE.md §4.4) and the HTTP layer turns it into a pointer to
	// the empty DateProper. Without that, a death date could never be taken off
	// someone who turned out to be alive.
	var newBirth, newDeath *DateProper

	if update.DateOfBirth != nil {
		parsed, err := parseUpdateDate("date of birth", *update.DateOfBirth)
		if err != nil {
			return "", false, err
		}
		newBirth = &parsed
	}

	if update.DateOfDeath != nil {
		parsed, err := parseUpdateDate("date of death", *update.DateOfDeath)
		if err != nil {
			return "", false, err
		}
		newDeath = &parsed
	}

	if newBirth != nil || newDeath != nil {
		existing, err := GetPerson(ctx, driver, id)
		if err != nil {
			return "", false, err
		}

		if existing == nil {
			return "", false, fmt.Errorf("%w: user %s does not exist", ErrNotFound, id)
		}

		// Compare the dates as they will be after the update is applied, using
		// the canonical values rather than the raw request payload.
		birth := existing.DateOfBirth
		death := existing.DateOfDeath

		if newBirth != nil {
			birth = *newBirth
		}

		if newDeath != nil {
			death = *newDeath
		}

		// The ordering constraint only applies when both dates are known. A
		// living person has no death date and must still be able to have their
		// birth date corrected.
		if birth != "" && death != "" {
			timeOfBirth, birthOK := birth.ParseTime()
			timeOfDeath, deathOK := death.ParseTime()

			if !birthOK || !deathOK {
				return "", false, fmt.Errorf("%w: unparsable dates |TOD %v| = |TOB %v|", ErrInvalidUpdate, death, birth)
			}

			if timeOfBirth.After(timeOfDeath) {
				return "", false, fmt.Errorf("%w: time of death is before time of birth | tb: %v | td: %v", ErrInvalidUpdate, timeOfBirth, timeOfDeath)
			}
		}

		// Persist native Neo4j dates, exactly like CreateNewPerson does. Writing
		// the string form here was what silently turned the property into a
		// string after the first update.
		//
		// A cleared date is written as an explicit null, which `SET p += $props`
		// turns into "remove the property" rather than "store a null".
		if newBirth != nil {
			if *newBirth == "" {
				props["date_of_birth"] = nil
			} else {
				neoDate, err := newBirth.GetNeoDate()
				if err != nil {
					return "", false, fmt.Errorf("%w: invalid date of birth: %v", ErrInvalidUpdate, err)
				}
				props["date_of_birth"] = *neoDate
			}
		}

		if newDeath != nil {
			if *newDeath == "" {
				props["date_of_death"] = nil
			} else {
				neoDate, err := newDeath.GetNeoDate()
				if err != nil {
					return "", false, fmt.Errorf("%w: invalid date of death: %v", ErrInvalidUpdate, err)
				}
				props["date_of_death"] = *neoDate
			}
		}
	}

	// Single query handles both existence check and property updates
	result, err := neo4j.ExecuteQuery(ctx, driver,
		`
		MATCH (p:Person {id: $id})
		SET p += $props
		RETURN p.name AS name
		`,
		map[string]any{
			"id":    id,
			"props": props,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return "", false, fmt.Errorf("failed to update person: %w", err)
	}

	// If no records returned, the node with given ID does not exist
	if len(result.Records) == 0 {
		return "", false, fmt.Errorf("%w: user %s does not exist", ErrNotFound, id)
	}

	// Safely retrieve the 'name' field
	rawName, found := result.Records[0].Get("name")
	if !found || rawName == nil {
		return "", false, errors.New("profile name property not found on node")
	}

	profileName, ok := rawName.(string)
	if !ok {
		return "", false, errors.New("profile name is not a valid string")
	}

	wasUpdated := result.Summary.Counters().ContainsUpdates()
	return profileName, wasUpdated, nil
}
