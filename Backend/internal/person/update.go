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
		props["name"] = *update.Name
	}

	if update.DateOfBirth != nil {
		if update.DateOfBirth != nil && !update.DateOfBirth.IsValid() {
			return "", false, errors.New("Invalid Date Of Birth Provided")
		}
		props["date_of_birth"] = *update.DateOfBirth
	}

	if update.Gender != nil {
		props["gender"] = *update.Gender // Fixed *& pointer dereference bug
	}

	if update.DateOfDeath != nil {
		if update.DateOfDeath != nil && !update.DateOfDeath.IsValid() {
			return "", false, errors.New("Invalid Date Of Birth Provided")
		}
		props["date_of_death"] = *update.DateOfDeath
	}

	if update.DateOfBirth != nil || update.DateOfDeath != nil {
		var dob DateProper
		var dod DateProper

		const query = `
			MATCH (p:Person {id: $id})
			RETURN p.date_of_birth as dob, p.date_of_death as dod
		`

		params := map[string]any{
			"id": id,
		}

		result, err := neo4j.ExecuteQuery(
			ctx,
			driver,
			query,
			params,
			neo4j.EagerResultTransformer,
		)

		if err != nil {
			return "", false, err
		}

		// 1. Guard against empty records to prevent panic
		if len(result.Records) == 0 {
			return "", false, fmt.Errorf("Person with id %s not found", id)
		}

		record := result.Records[0]

		// Populate existing values from DB
		if raw, found := record.Get("dob"); found && raw != nil {
			if str, ok := raw.(string); ok {
				dob = DateProper(str)
			}
		}

		if raw, found := record.Get("dod"); found && raw != nil {
			if str, ok := raw.(string); ok {
				dod = DateProper(str)
			}
		}

		// 2. Override with new update values if provided
		if update.DateOfBirth != nil {
			dob = DateProper(*update.DateOfBirth)
		}
		if update.DateOfDeath != nil {
			dod = DateProper(*update.DateOfDeath)
		}

		// 3. Validate individual date formats independently
		var parsed_dob, parsed_dod int64

		if dob != "" {
			parsed_dob = int64(dob.GetMS())
			if parsed_dob == 0 {
				return "", false, fmt.Errorf("Invalid Date Of Birth Provided: %s", dob)
			}
		}

		if dod != "" {
			parsed_dod = int64(dod.GetMS())
			if parsed_dod == 0 {
				return "", false, fmt.Errorf("Invalid Date Of Death Provided: %s", dod)
			}
		}

		// 4. Validate relative order when both dates are present
		if dob != "" && dod != "" {
			if parsed_dob > parsed_dod {
				return "", false, errors.New("Date Of Death Is Before Date Of Birth")
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
		return "", false, fmt.Errorf("user %s does not exist", id)
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
