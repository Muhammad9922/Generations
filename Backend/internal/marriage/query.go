package marriage

import (
	"context"
	"errors"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type QueryResponse struct {
	Id    string
	Start person.DateProper
	End   person.DateProper
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
			thisResponse.Id = id
		}

		if startDateFound {
			startDate, ok := startDateRaw.(string)
			if ok {
				thisResponse.Start = person.DateProper(startDate)
			}
		}

		if endDateFound {
			endDate, ok := endDateRaw.(string)
			if ok {
				thisResponse.End = person.DateProper(endDate)
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
			Start: start,
			End:   end,
			Id:    id,
		})
	}

	return &response, nil

}

type SpouseQueryParams struct {
	MarriageId string
	SpouseId   string
}

func getSpousesFromMarriageId(ctx context.Context, driver neo4j.Driver, id string) (string, string, string, error) {

	const query = `
		MATCH (p:Person)-[:married]->(m:Marriage {id: $id})
		RETURN p.id as id
		`

	queryParams := map[string]any{
		"id": id,
	}

	results, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		queryParams,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return "", "", "", err
	}

	if len(results.Records) > 2 {
		return "", "", "", errors.New("More Than Two Spouses Found For This Marriage")
	}

	spouseOneId := ""
	spouseTwoId := ""

	for index, record := range results.Records {
		rawID, found := record.Get("id")
		if !found {
			return "", "", "", errors.New("ID not found for some spouse")
		}

		id, ok := rawID.(string)

		if !ok {
			return "", "", "", errors.New("Invalid ID found for some spouse")
		}

		if index == 0 {
			spouseOneId = id
		} else {
			spouseTwoId = id
		}

	}

	return spouseOneId, spouseTwoId, id, nil

}

func GetSpouses(ctx context.Context, driver neo4j.Driver, params SpouseQueryParams) (string, string, string, error) {

	if params.MarriageId != "" {
		return getSpousesFromMarriageId(ctx, driver, params.MarriageId)
	} else if params.SpouseId != "" {
		const query = `
		MATCH (p:Person {id: $id})-[:married]->(m:Marriage)
		return m.id as id
		`

		queryParams := map[string]any{
			"id": params.SpouseId,
		}

		records, err := neo4j.ExecuteQuery(ctx, driver, query, queryParams, neo4j.EagerResultTransformer)

		if err != nil {
			return "", "", "", err
		}

		marriage := records.Records[0]
		marriageID, found := marriage.Get("id")
		if !found {
			return "", "", "", errors.New("Invalid Marriage Was Returned")
		}

		marriageId, ok := marriageID.(string)
		if ok {
			return getSpousesFromMarriageId(ctx, driver, marriageId)
		} else {
			return "", "", "", errors.New("Invalid Marriage ID Was Returned")
		}

	} else {
		return "", "", "", errors.New("no valid query parameter provided")
	}

}
