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
			return "", false, fmt.Errorf("Invalid Name Provided: %s", *update.Name)
		}
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
			return "", false, errors.New("Invalid Date Of Death Provided")
		}
		props["date_of_death"] = *update.DateOfDeath
	}

	if update.DateOfBirth != nil || update.DateOfDeath != nil {
		var dob DateProper
		var dod DateProper

		userObject, err := GetPerson(ctx, driver, id)

		if err != nil {
			return "", false, err
		}

		if userObject == nil {
			return "", false, fmt.Errorf("Error While Retieving Person: %s", id)
		}
		dob = userObject.DateOfBirth
		dod = userObject.DateOfDeath

		if update.DateOfBirth != nil {
			dob = *update.DateOfBirth
		}

		if update.DateOfDeath != nil {
			dod = *update.DateOfDeath
		}

		timeOfDeath := dod.GetMS()
		timeOfBirth := dob.GetMS()

		if timeOfBirth > 0 && timeOfDeath > 0 {
			if timeOfBirth > timeOfDeath {
				return "", false, fmt.Errorf("Time Of Death Is Before Time Of Birth | tb: %v | td: %v", timeOfBirth, timeOfDeath)
			}
		} else {
			return "", false, fmt.Errorf("Invalid Time Recieved For |TOD-MS %v TOD-ORG %v| = |TOB-MS %v TOB-ORG %v|", timeOfDeath, dod, timeOfBirth, dob)
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
