package person

import (
	"context"
	"errors"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type UpdateUser struct {
	Name        *string
	DateOfBirth *DateOfBirth
	Gender      *Gender
	Alive       *bool
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
		props["date_of_birth"] = *update.DateOfBirth
	}

	if update.Gender != nil {
		props["gender"] = *update.Gender // Fixed *& pointer dereference bug
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
