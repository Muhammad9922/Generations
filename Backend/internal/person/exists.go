package person

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// PersonQuery defines lookup filters for person existence.
type PersonQuery struct {
	ID   string
	Name string
}

// CheckPersonExistence checks if a person matching the ID or Name exists in the database.
func CheckPersonExistence(ctx context.Context, driver neo4j.Driver, params PersonQuery) (bool, error) {
	query := `
		MATCH (p:Person)
		WHERE ($id <> "" AND p.id = $id) OR ($name <> "" AND p.name = $name)
		RETURN count(p) > 0 AS exists
	`
	queryParams := map[string]any{
		"id":   params.ID,
		"name": params.Name,
	}

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		queryParams,
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check person existence: %w", err)
	}

	if len(result.Records) == 0 {
		return false, nil
	}

	rawExists, found := result.Records[0].Get("exists")
	if !found {
		return false, nil
	}

	exists, ok := rawExists.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected return type for 'exists': %T", rawExists)
	}

	return exists, nil
}
