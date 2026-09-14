package children

import (
	"context"
	"errors"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func CreateNewChildren(ctx context.Context, driver neo4j.Driver, marriageId string, childrenId string) (bool, error) {

	child, err := person.GetPerson(ctx, driver, childrenId)

	if err != nil {
		return false, err
	}

	if child == nil {
		return false, fmt.Errorf("Child not found")
	}

	marriages, err := marriage.GetMarriageFromMarriageId(ctx, driver, marriageId)

	if err != nil {
		return false, err
	}

	if len(marriages) == 0 {
		return false, errors.New("Marriage not found")
	}

	query := `
		MATCH (m:Marriage {id: $mid})
		MATCH (c:Person {id: $cid})
		MERGE (c)-[:PRODUCED]->(m)
	`

	params := map[string]any{
		"mid": marriageId,
		"cid": childrenId,
	}

	record, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return false, err
	}

	if record.Summary.Counters().RelationshipsCreated() == 0 {
		return false, fmt.Errorf("No Children Were Created")
	}

	return true, nil
}
