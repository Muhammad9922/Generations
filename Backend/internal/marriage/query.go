package marriage

import (
	"context"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type QueryResponse struct {
	id    string
	start person.DateProper
	end   person.DateProper
}

func GetMarriage(ctx context.Context, driver neo4j.Driver, id string) (*[]QueryResponse, error) {
	const query = `
	MATCH (p:Person {id: $id})-->(m: Marriage)
	RETURN m.start AS start, m.end AS end, m.id AS id
	`

	result, err := neo4j.ExecuteQuery(ctx, driver, query, map[string]any{
		"id": id,
	}, neo4j.EagerResultTransformer)

	if err != nil {
		return nil, err
	}

	var response []QueryResponse

	records := result.Records

	for _, record := range records {
		startDateRaw, startDateFound := record.Get("start")
		endDateRaw, endDateFound := record.Get("end")
		idRaw, idFound := record.Get("id")

		if !idFound {
			continue
		}

		var thisResponse QueryResponse

		id, ok := idRaw.(string)
		if ok {
			thisResponse.id = id
		}

		if startDateFound {
			startDate, ok := startDateRaw.(string)
			if ok {
				thisResponse.start = person.DateProper(startDate)
			}
		}

		if endDateFound {
			endDate, ok := endDateRaw.(string)
			if ok {
				thisResponse.end = person.DateProper(endDate)
			}
		}

		response = append(response, thisResponse)
	}

	return &response, nil
}

func GetAllMarriages(ctx context.Context, driver neo4j.Driver) (*[]QueryResponse, error) {
	const query = `
	MATCH (p:Person)-->(m: Marriage)
	RETURN m.start AS start, m.end AS end, m.id AS id
	`

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		map[string]any{},
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return nil, err
	}

	var response []QueryResponse

	for _, record := range result.Records {
		var start person.DateProper
		var end person.DateProper
		var id string
		rawStart, found := record.Get("start")
		if found {
			parsedStart, ok := rawStart.(string)
			if ok {
				start = person.DateProper(parsedStart)
			}
		}

		rawEnd, found := record.Get("end")
		if found {
			parsedEnd, ok := rawEnd.(string)
			if ok {
				end = person.DateProper(parsedEnd)
			}
		}

		rawId, found := record.Get("id")
		if found {
			parsedId, ok := rawId.(string)
			if ok {
				id = parsedId
			}
		}

		response = append(response, QueryResponse{
			start: start,
			end:   end,
			id:    id,
		})
	}

	return &response, nil

}
