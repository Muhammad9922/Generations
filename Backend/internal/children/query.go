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
				personObject, err := person.GetPerson(ctx, driver, id)
				if err != nil {
					return nil, err
				}
				if personObject != nil {
					children = append(children, *personObject)
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
		MATCH (m:Marriage {id:$id})<-[:MARRIED]-(p:Person)
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
				personObject, err := person.GetPerson(ctx, driver, id)
				if err != nil {
					return nil, err
				}
				if personObject != nil {
					parents = append(parents, *personObject)
				} else {
					return nil, fmt.Errorf("Invalid Person Found With ID: %v", id)
				}
			}
		}
	}

	return parents, nil

}

func GetMarriageThatOfChild(ctx context.Context, driver neo4j.Driver, childId string) (*marriage.QueryResponse, error) {

	childOk, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
		ID: childId,
	})

	if err != nil {
		return nil, err
	}

	if !childOk {
		return nil, fmt.Errorf("Invalid Spouse Id Provided")
	}

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
				marriageObject, err := marriage.GetMarriageFromMarriageId(ctx, driver, id)
				if err != nil {
					return nil, err
				}

				if len(marriageObject) != 1 {
					return nil, fmt.Errorf("More Than One Marriages Were Found For This Id: %v", id)
				}

				itemMarriage = (marriageObject[0])

			}
		}
	}

	return &itemMarriage, nil
}

func GetMarriageThatOfSpouse(ctx context.Context, driver neo4j.Driver, spouseId string) (*[]marriage.QueryResponse, error) {

	spouseOk, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
		ID: spouseId,
	})

	if err != nil {
		return nil, err
	}

	if !spouseOk {
		return nil, fmt.Errorf("Invalid Spouse Id Provided")
	}

	const query = `
		MATCH (m:Marriage)<-[:MARRIED]-(p:Person {id: $id})
		RETURN m.id AS id
	`

	params := map[string]any{
		"id": spouseId,
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

	var itemMarriage []marriage.QueryResponse

	for _, marriageRecord := range records.Records {
		rawId, found := marriageRecord.Get("id")
		if found {
			id, ok := rawId.(string)
			if ok {
				marriageObject, err := marriage.GetMarriageFromMarriageId(ctx, driver, id)
				if err != nil {
					return nil, err
				}

				if len(marriageObject) != 1 {
					return nil, fmt.Errorf("More Than One Marriages Were Found For This Id: %v", id)
				}

				itemMarriage = append(itemMarriage, marriageObject[0])

			}
		}
	}

	return &itemMarriage, nil

}
