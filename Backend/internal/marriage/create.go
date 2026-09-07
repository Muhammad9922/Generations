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
	ID        string // Capitalized to export it, useful if you need the generated ID outside the package
}

func CreateNewMarriage(ctx context.Context, driver neo4j.Driver, marriage NewMarriage) (string, error) {
	if marriage.ID == "" {
		marriage.ID = uuid.New().String()
	}

	// Fetch Spouse One
	spouseOne, err := person.GetPerson(ctx, driver, marriage.SpouseOne)
	if err != nil {
		return "", fmt.Errorf("failed to get spouse one: %w", err)
	}
	if spouseOne == nil {
		return "", fmt.Errorf("user not found for ID: %s", marriage.SpouseOne)
	}

	// Fetch Spouse Two
	spouseTwo, err := person.GetPerson(ctx, driver, marriage.SpouseTwo)
	if err != nil {
		return "", fmt.Errorf("failed to get spouse two: %w", err)
	}
	if spouseTwo == nil {
		return "", fmt.Errorf("user not found for ID: %s", marriage.SpouseTwo)
	}

	// Validate Birth Dates (Birth cannot be AFTER Marriage Start)
	if spouseOne.DateOfBirth.GetMS() > marriage.DateStart.GetMS() {
		return "", fmt.Errorf("the birth of spouse one (%v) is after the date of marriage (%v)", spouseOne.DateOfBirth, marriage.DateStart)
	}
	if spouseTwo.DateOfBirth.GetMS() > marriage.DateStart.GetMS() {
		return "", fmt.Errorf("the birth of spouse two (%v) is after the date of marriage (%v)", spouseTwo.DateOfBirth, marriage.DateStart)
	}

	// Validate Death Dates (Marriage End cannot be AFTER Death)
	// Added a check assuming GetMS() == 0 means the person is still alive
	if spouseOne.DateOfDeath.GetMS() > 0 && spouseOne.DateOfDeath.GetMS() < marriage.DateEnd.GetMS() {
		return "", fmt.Errorf("the end of marriage %v is after the death of spouse one %v", marriage.DateEnd, spouseOne.DateOfDeath)
	}
	if spouseTwo.DateOfDeath.GetMS() > 0 && spouseTwo.DateOfDeath.GetMS() < marriage.DateEnd.GetMS() {
		return "", fmt.Errorf("the end of marriage %v is after the death of spouse two %v", marriage.DateEnd, spouseTwo.DateOfDeath)
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
