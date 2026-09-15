package children

import (
	"context"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func GetChildren(ctx context.Context, driver neo4j.Driver, marriageId string) ([]person.NewPerson, error) {
	const query = `
		MATCH (m:Marriage {id:$id})-[:PRODUCED]->(p:Person)
		RETURN p.id AS id
	`

	params := map[string]any{
		"id": marriageId,
	}

	records, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return nil, err
	}

	if len(records.Records) == 0 {
		return nil, nil
	}

	var children []person.NewPerson

	for _, record := range records.Records {
		idRaw, found := record.Get("id")
		if found {
			id, ok := idRaw.(string)
			if ok {
				person, err := person.GetPerson(ctx, driver, id)
				if err != nil {
					return nil, err
				}
				if person != nil {
					children = append(children, *person)
				} else {
					return nil, fmt.Errorf("Invalid Person Found With ID: %v", id)
				}
			}
		}
	}

	return children, nil
}

func GetParents(ctx context.Context, driver neo4j.Driver, marriageId string) ([]person.NewPerson, error) {
	const query = `
		MATCH (m:Marriage)<-[:MARRIED]-(p:Person {id:$id})
		RETURN p.id AS id
	`

	params := map[string]any{
		"id": marriageId,
	}

	records, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return nil, err
	}

	if len(records.Records) == 0 {
		return nil, nil
	}

	var parents []person.NewPerson

	for _, record := range records.Records {
		idRaw, found := record.Get("id")
		if found {
			id, ok := idRaw.(string)
			if ok {
				person, err := person.GetPerson(ctx, driver, id)
				if err != nil {
					return nil, err
				}
				if person != nil {
					parents = append(parents, *person)
				} else {
					return nil, fmt.Errorf("Invalid Person Found With ID: %v", id)
				}
			}
		}
	}

	return parents, nil

}

func GetMarriageThatOfChild(ctx context.Context, driver neo4j.Driver, childId string) (*marriage.QueryResponse, error) {
	const query = `
	MATCH (m:Marriage)-[:PRODUCED]->(p:Person {id: $id})
	RETURN m.id AS id
	`

	params := map[string]any{
		"id": childId,
	}

	records, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return nil, err
	}

	if len(records.Records) == 0 {
		return nil, nil
	}

	var itemMarriage marriage.QueryResponse

	for _, marriageRecord := range records.Records {
		rawId, found := marriageRecord.Get("id")
		if found {
			id, ok := rawId.(string)
			if ok {
				marriage, err := marriage.GetMarriageFromMarriageId(ctx, driver, id)
				if err != nil {
					return nil, err
				}

				if len(marriage) != 1 {
					return nil, fmt.Errorf("More Than One Marriages Were Found For This Id: %v", id)
				}

				itemMarriage = (marriage[0])

			}
		}
	}

	return &itemMarriage, nil
}

func GetMarriageThatOfSpouse(ctx context.Context, driver neo4j.Driver, spouseId string) (marriage.NewMarriage, error)
