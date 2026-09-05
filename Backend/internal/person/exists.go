package person

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type PersonQuery struct {
	id   string
	name string
}

func CheckPersonExistance(ctx context.Context, driver neo4j.Driver, params PersonQuery) bool {
	query := `
		MATCH (p:Person)
		WHERE p.id = $id OR p.name = $name
		RETURN count(p) > 0 AS exists
	`
	query_params := map[string]any{
		"id":   params.id,
		"name": params.name,
	}

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		query_params,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return false
	}

	if len(result.Records) == 0 {
		return false
	}

	rawExists, _ := result.Records[0].Get("exists")

	return rawExists.(bool)
}
