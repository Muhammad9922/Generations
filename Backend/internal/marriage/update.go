package marriage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type MarriageUpdate struct {
	DateStart person.DateProper
	DateEnd   person.DateProper
}

func UpdateMarriageDates(ctx context.Context, driver neo4j.Driver, id string, update MarriageUpdate) (string, error) {

	if update.DateEnd == "" && update.DateStart == "" {
		return "", fmt.Errorf("No Updates Were Requested: %v", update)
	}

	existingRecord, err := GetMarriageFromMarriageId(ctx, driver, id)

	if err != nil {
		return "", err
	}

	if existingRecord == nil {
		return "", errors.New("No Existing Record Found")
	}

	if len(existingRecord) != 1 {
		return "", fmt.Errorf("Unable To Properly Query Existing Marriage: %v", existingRecord)
	}

	existingMarriage := (existingRecord)[0]

	if update.DateStart == "" {
		update.DateStart = existingMarriage.Start
	}

	if update.DateEnd == "" {
		update.DateEnd = existingMarriage.End
	}

	if update.DateStart != "" && !update.DateStart.IsValid() {
		return "", fmt.Errorf("New DateStart Is Invalid %v", update.DateStart)
	}

	if update.DateEnd != "" && !update.DateEnd.IsValid() {
		return "", fmt.Errorf("New DateEnd Is Invalid %v", update.DateEnd)
	}

	if update.DateStart != "" && update.DateEnd != "" {
		start, startOK := update.DateStart.ParseTime()
		end, endOK := update.DateEnd.ParseTime()
		if startOK && endOK && start.After(end) {
			return "", fmt.Errorf("Date Start %v Is After Date End %v", update.DateStart, update.DateEnd)
		}
	}

	spouseOneId, spouseTwoId, marriageId, err := getSpousesFromMarriageId(ctx, driver, id)

	if err != nil {
		return "", err
	}

	if marriageId != id {
		return "", errors.New("Somehow the new marriage id is different than the one provided")
	}

	spouseOne, err := person.GetPerson(ctx, driver, spouseOneId)

	if err != nil {
		return "", fmt.Errorf("Error Getting Spouse One: %v", err)
	}

	spouseTwo, err := person.GetPerson(ctx, driver, spouseTwoId)

	if err != nil {
		return "", fmt.Errorf("Error Getting Spouse Two: %q", err)
	}

	// 1. Validate that start date is not after end date
	if update.DateEnd.IsValid() && update.DateStart > update.DateEnd {
		return "", fmt.Errorf("marriage start date (%v) cannot be after end date (%v)", update.DateStart, update.DateEnd)
	}

	// Helper to validate a spouse's timeline against marriage dates
	validateSpouseDates := func(spouseLabel string, dob, dod person.DateProper) error {
		// Birth validation
		if update.DateStart < dob {
			return fmt.Errorf("marriage start date (%v) cannot be before %s's birth date (%v)", update.DateStart, spouseLabel, dob)
		}
		if update.DateEnd.IsValid() && update.DateEnd < dob {
			return fmt.Errorf("marriage end date (%v) cannot be before %s's birth date (%v)", update.DateEnd, spouseLabel, dob)
		}

		// Death validation (only evaluate if DateOfDeath is set / non-zero)
		if dod.IsValid() {
			if update.DateStart > dod {
				return fmt.Errorf("marriage start date (%v) cannot be after %s's death date (%v)", update.DateStart, spouseLabel, dod)
			}
			if update.DateEnd != "" && update.DateEnd > dod {
				return fmt.Errorf("marriage end date (%v) cannot be after %s's death date (%v)", update.DateEnd, spouseLabel, dod)
			}
		}
		return nil
	}

	if err := validateSpouseDates("spouse one", spouseOne.DateOfBirth, spouseOne.DateOfDeath); err != nil {
		return "", err
	}

	if err := validateSpouseDates("spouse two", spouseTwo.DateOfBirth, spouseTwo.DateOfDeath); err != nil {
		return "", err
	}

	query := `
	MATCH (m:Marriage {id: $id})
	SET m.start = $start, m.end = $end
	`

	params := map[string]any{
		"id":    id,
		"start": nil,
		"end":   nil,
	}

	if update.DateStart != "" {
		params["start"] = update.DateStart
	}

	if update.DateEnd != "" {
		params["end"] = update.DateEnd
	}

	records, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return "", err
	}

	if !records.Summary.Counters().ContainsUpdates() {
		return "", fmt.Errorf("No Updates Were Done Even When Requested")
	}

	return id, nil
}
