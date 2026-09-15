package children

import (
	"context"
	"errors"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func DeleteChildren(ctx context.Context, driver neo4j.Driver, spouseID string, marriageID string) (bool, error) {
	query := `
	MATCH (m:Marriage {id: $mid})-[r:PRODUCED]->(p:Person {id: $pid})
	DELETE r
	`
	params := map[string]any{
		"pid": spouseID,
		"mid": marriageID,
	}

	records, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return false, err
	}

	if records.Summary.Counters().RelationshipsDeleted() == 0 {
		return false, errors.New("No Relationship Was Deleted")
	}

	return true, nil
}
