package person

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func DeleteUser(ctx context.Context, driver neo4j.Driver, userID string) (string, bool, error) {
	query := `
		MATCH (p:Person {id: $id})
		WITH p, p.name AS name
		DETACH DELETE p
		RETURN name
	`

	// TODO: Automatically Delete Marriage And Children Upon Deletation

	params := map[string]any{
		"id": userID,
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

	counters := result.Summary.Counters()
	var deleted bool = counters.NodesDeleted() > 0

	if !deleted {
		return "", deleted, err
	}

	// Safely unpack the record value
	rawName, found := result.Records[0].Get("name")
	if !found {
		return "", true, nil
	}

	name, ok := rawName.(string)
	if !ok {
		return "", true, fmt.Errorf("unexpected type for 'name': %T", rawName)
	}

	return name, true, nil
}
