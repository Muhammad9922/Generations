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
	if err != nil || spouseOne == nil {
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

	// Parse every date that is present; empty dates stay as the zero time.
	spouseOneBirth, spouseOneBirthOK := spouseOne.DateOfBirth.ParseTime()
	spouseTwoBirth, spouseTwoBirthOK := spouseTwo.DateOfBirth.ParseTime()
	spouseOneDeath, spouseOneDeathOK := spouseOne.DateOfDeath.ParseTime()
	spouseTwoDeath, spouseTwoDeathOK := spouseTwo.DateOfDeath.ParseTime()
	dateStart, dateStartOK := marriage.DateStart.ParseTime()
	dateEnd, dateEndOK := marriage.DateEnd.ParseTime()

	// Validate Birth Dates (Birth cannot be AFTER Marriage Start)
	if spouseOne.DateOfBirth != "" && !spouseOneBirthOK {
		return "", fmt.Errorf("the birth date of spouse one is invalid: %s", spouseOne.DateOfBirth)
	}
	if spouseTwo.DateOfBirth != "" && !spouseTwoBirthOK {
		return "", fmt.Errorf("the birth date of spouse two is invalid: %s", spouseTwo.DateOfBirth)
	}
	if spouseOne.DateOfBirth != "" && marriage.DateStart != "" && dateStartOK && spouseOneBirth.After(dateStart) {
		return "", fmt.Errorf("the birth of spouse one (%v) is after the date of marriage (%v)", spouseOne.DateOfBirth, marriage.DateStart)
	}
	if spouseTwo.DateOfBirth != "" && marriage.DateStart != "" && dateStartOK && spouseTwoBirth.After(dateStart) {
		return "", fmt.Errorf("the birth of spouse two (%v) is after the date of marriage (%v)", spouseTwo.DateOfBirth, marriage.DateStart)
	}

	// Validate Death Dates (Marriage End cannot be AFTER Death)
	if spouseOne.DateOfDeath != "" && !spouseOneDeathOK {
		return "", fmt.Errorf("the death date of spouse one is invalid: %s", spouseOne.DateOfDeath)
	}
	if spouseTwo.DateOfDeath != "" && !spouseTwoDeathOK {
		return "", fmt.Errorf("the death date of spouse two is invalid: %s", spouseTwo.DateOfDeath)
	}
	if spouseOne.DateOfDeath != "" && marriage.DateEnd != "" && dateEndOK && spouseOneDeath.Before(dateEnd) {
		return "", fmt.Errorf("the end of marriage %v is after the death of spouse one %v", marriage.DateEnd, spouseOne.DateOfDeath)
	}
	if spouseTwo.DateOfDeath != "" && marriage.DateEnd != "" && dateEndOK && spouseTwoDeath.Before(dateEnd) {
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
	if err != nil {
		return "", err
	}
	marriageEnd, err := marriage.DateEnd.GetNeoDate()
	if err != nil {
		return "", err
	}

	if marriage.DateStart != "" && marriageStart == nil {
		return "", fmt.Errorf("Date Start Is Invalid: %v | %v", marriage.DateStart, marriageStart)
	}

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		map[string]any{
			"mid":   marriage.ID,
			"start": *marriageStart,
			"end":   *marriageEnd,
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
