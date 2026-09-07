package person

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type UpdateUser struct {
	Name        *string
	DateOfBirth *DateProper
	Gender      *Gender
	Alive       *bool
	DateOfDeath *DateProper
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
		if !update.DateOfBirth.IsValid() {
			return "", false, errors.New("Invalid Date Of Birth Provided")
		}
		props["date_of_birth"] = *update.DateOfBirth
	}

	if update.Gender != nil {
		props["gender"] = *update.Gender // Fixed *& pointer dereference bug
	}

	if update.DateOfDeath != nil {
		if !update.DateOfBirth.IsValid() {
			return "", false, errors.New("Invalid Date Of Birth Provided")
		}
		props["date_of_death"] = *update.DateOfDeath
	}

	if update.DateOfBirth != nil || update.DateOfDeath != nil {
		var dob string
		var dod string

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

		raw_date_of_birth, found := result.Records[0].Get("dob")
		if found {
			data_of_birth, ok := raw_date_of_birth.(string)
			if ok {
				dob = data_of_birth
			}
		}

		raw_date_of_death, found := result.Records[0].Get("dod")
		if found {
			date_of_death, ok := raw_date_of_death.(string)
			if ok {
				dod = date_of_death
			}
		}

		if dob != "" && dod != "" {
			const layout = "02-01-2006"
			parsed_dod, err := time.Parse(layout, dod)
			if err != nil {
				return "", false, err
			}
			parsed_dob, err := time.Parse(layout, dob)
			if err != nil {
				return "", false, err
			}

			if parsed_dob.UnixMilli() > parsed_dod.UnixMilli() {
				return "", false, fmt.Errorf("Date Of Death Is Before Date Of Birth")
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
