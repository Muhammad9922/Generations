package person

import (
	"context"
	"errors"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

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
		props["gender"] = *update.Gender
	}

	// Dates are parsed into the canonical DD-MM-YYYY representation up front.
	//
	// The previous implementation validated the YYYY-MM-DD value returned by
	// NormalizeDate with DateProper.IsValid, which only accepts DD-MM-YYYY, so
	// the validation could never pass and every date update failed.
	var newBirth, newDeath *DateProper

	if update.DateOfBirth != nil {
		parsed, err := ParseDateProper(string(*update.DateOfBirth))
		if err != nil {
			return "", false, fmt.Errorf("%w: invalid date of birth: %v", ErrInvalidUpdate, err)
		}
		newBirth = &parsed
	}

	if update.DateOfDeath != nil {
		parsed, err := ParseDateProper(string(*update.DateOfDeath))
		if err != nil {
			return "", false, fmt.Errorf("%w: invalid date of death: %v", ErrInvalidUpdate, err)
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
		if newBirth != nil {
			neoDate, err := newBirth.GetNeoDate()
			if err != nil {
				return "", false, fmt.Errorf("%w: invalid date of birth: %v", ErrInvalidUpdate, err)
			}
			props["date_of_birth"] = *neoDate
		}

		if newDeath != nil {
			neoDate, err := newDeath.GetNeoDate()
			if err != nil {
				return "", false, fmt.Errorf("%w: invalid date of death: %v", ErrInvalidUpdate, err)
			}
			props["date_of_death"] = *neoDate
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
