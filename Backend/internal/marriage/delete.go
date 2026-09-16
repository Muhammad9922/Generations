package marriage

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func DeleteMarriage(ctx context.Context, driver neo4j.Driver, id string) (bool, error) {
	const query = `
	MATCH (m:Marriage {id: $id})
	DETACH DELETE (m)
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
		return false, err
	}

	if result.Summary.Counters().NodesDeleted() == 0 {
		return false, fmt.Errorf("Without Error, No Nodes Were Deleted")
	}

	return true, nil
}
