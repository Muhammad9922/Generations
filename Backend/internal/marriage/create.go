package marriage

import (
	"context"
	"fmt"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type NewMarriage struct {
	DateStart person.DateProper
	DateEnd   person.DateProper
	SpouseOne string
	SpouseTwo string
	ID        string
}

func CreateNewMarriage(ctx context.Context, driver neo4j.Driver, marriage NewMarriage) (string, error) {
	if marriage.ID == "" {
		marriage.ID = uuid.New().String()
	}

	if marriage.SpouseOne == "" || marriage.SpouseTwo == "" {
		return "", fmt.Errorf("both spouses must be provided")
	}

	// Fetch Spouse One
	spouseOne, err := person.GetPerson(ctx, driver, marriage.SpouseOne)
	if err != nil {
		return "", fmt.Errorf("failed to get spouse one: %w", err)
	}

	// Fetch Spouse Two
	spouseTwo, err := person.GetPerson(ctx, driver, marriage.SpouseTwo)
	if err != nil {
		return "", fmt.Errorf("failed to get spouse two: %w", err)
	}

	if marriage.DateStart != "" && !marriage.DateStart.IsValid() {
		return "", fmt.Errorf("date start is invalid: %s", marriage.DateStart)
	}

	if marriage.DateEnd != "" && !marriage.DateEnd.IsValid() {
		return "", fmt.Errorf("date end is invalid: %s", marriage.DateEnd)
	}

	// Validate Birth Dates (Birth cannot be AFTER Marriage Start)
	if spouseOne.DateOfBirth.GetMS() > marriage.DateStart.GetMS() {
		return "", fmt.Errorf("the birth of spouse one (%v) is after the date of marriage (%v)", spouseOne.DateOfBirth, marriage.DateStart)
	}
	if spouseTwo.DateOfBirth.GetMS() > marriage.DateStart.GetMS() {
		return "", fmt.Errorf("the birth of spouse two (%v) is after the date of marriage (%v)", spouseTwo.DateOfBirth, marriage.DateStart)
	}

	// Validate Death Dates (Marriage End cannot be AFTER Death)
	if spouseOne.DateOfDeath.GetMS() < marriage.DateEnd.GetMS() {
		return "", fmt.Errorf("the end of marriage %v is after the death of spouse one %v", marriage.DateEnd, spouseOne.DateOfDeath)
	}
	if spouseTwo.DateOfDeath.GetMS() < marriage.DateEnd.GetMS() {
		return "", fmt.Errorf("the end of marriage %v is after the death of spouse two %v", marriage.DateEnd, spouseTwo.DateOfDeath)
	}

	// Both Genders Should Be Provided
	if spouseOne.Gender == person.Male && spouseTwo.Gender == person.Male {
		return "", fmt.Errorf("both spouses are male")
	}

	if spouseOne.Gender == person.Female && spouseTwo.Gender == person.Female {
		return "", fmt.Errorf("both spouses are female")
	}

	// Combined Cypher Query guarantees atomicity.
	// If the relationships fail, the marriage node isn't created.
	const query = `
		MATCH (spa:Person {id: $said})
		MATCH (spb:Person {id: $sbid})
		CREATE (spa)-[:married]->(m:Marriage {id: $mid, start: $start, end: $end})<-[:married]-(spb)
	`

	marriageStart, err := marriage.DateStart.GetNeoDate()
	marriageEnd, err := marriage.DateEnd.GetNeoDate()

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		map[string]any{
			"mid":   marriage.ID,
			"start": marriageStart,
			"end":   marriageEnd,
			"said":  spouseOne.Id,
			"sbid":  spouseTwo.Id,
		},
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return "", fmt.Errorf("database execution failed: %w", err)
	}

	if result.Summary.Counters().NodesCreated() == 0 {
		return "", fmt.Errorf("no new marriage node was created")
	}

	if result.Summary.Counters().RelationshipsCreated() == 0 {
		return "", fmt.Errorf("no new relationships were created")
	}

	return marriage.ID, nil
}
