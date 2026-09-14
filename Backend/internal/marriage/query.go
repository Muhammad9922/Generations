package marriage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type QueryResponse struct {
	Id    string
	Start person.DateProper
	End   person.DateProper
}

func GetMarriage(ctx context.Context, driver neo4j.Driver, id string) ([]QueryResponse, error) {
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

	records := result.Records
	responses := make([]QueryResponse, 0, len(records))

	if len(responses) == 0 {
		return nil, fmt.Errorf("No Record Found: %v || ID: %v", responses, id)
	}

	for _, record := range records {
		idRaw, found := record.Get("id")
		if !found || idRaw == nil {
			continue
		}

		marriageID, ok := idRaw.(string)
		if !ok {
			continue
		}

		var thisResponse QueryResponse
		thisResponse.Id = marriageID

		// Safely handle optional start date
		if startDateRaw, found := record.Get("start"); found && startDateRaw != nil {
			if startDate, ok := startDateRaw.(string); ok {
				thisResponse.Start = person.DateProper(startDate)
			} else {
				return nil, fmt.Errorf("start date expected string, got %T (%v)", startDateRaw, startDateRaw)
			}
		}

		// Safely handle optional end date
		if endDateRaw, found := record.Get("end"); found && endDateRaw != nil {
			if endDate, ok := endDateRaw.(string); ok {
				thisResponse.End = person.DateProper(endDate)
			} else {
				return nil, fmt.Errorf("end date expected string, got %T (%v)", endDateRaw, endDateRaw)
			}
		}

		responses = append(responses, thisResponse)
	}

	return responses, nil
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
