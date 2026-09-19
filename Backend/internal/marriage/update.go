package marriage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// MarriageUpdate holds the dates to change. Every field is a pointer, because a
// bare DateProper cannot say both "leave this date alone" and "remove it":
//
//	nil   → the date was not sent, keep the stored one
//	&""   → remove the stored date
//	&date → store the new date
//
// The distinction matters on both ends. A partial update sends only one of the
// two fields and must not wipe the other, while the frontend's "clear a field to
// remove its date" sends an empty string for the date it wants gone.
type MarriageUpdate struct {
	DateStart *person.DateProper
	DateEnd   *person.DateProper
}

func UpdateMarriageDates(ctx context.Context, driver neo4j.Driver, id string, update MarriageUpdate) (string, error) {

	if update.DateStart == nil && update.DateEnd == nil {
		return "", fmt.Errorf("%w: no dates were sent", ErrInvalidUpdate)
	}

	existingRecord, err := GetMarriageFromMarriageId(ctx, driver, id)

	if err != nil {
		return "", err
	}

	if existingRecord == nil {
		return "", fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	if len(existingRecord) != 1 {
		return "", fmt.Errorf("Unable To Properly Query Existing Marriage: %v", existingRecord)
	}

	existingMarriage := (existingRecord)[0]

	// Resolve each date to the value it will hold after the update, so every
	// validation below runs against the marriage's real timeline rather than
	// against a request field that may have been left out or cleared.
	start := existingMarriage.Start
	end := existingMarriage.End

	if update.DateStart != nil {
		start = *update.DateStart
	}

	if update.DateEnd != nil {
		end = *update.DateEnd
	}

	if start != "" && !start.IsValid() {
		return "", fmt.Errorf("%w: New DateStart Is Invalid %v", ErrInvalidUpdate, start)
	}

	if end != "" && !end.IsValid() {
		return "", fmt.Errorf("%w: New DateEnd Is Invalid %v", ErrInvalidUpdate, end)
	}

	if start != "" && end != "" {
		startTime, startOK := start.ParseTime()
		endTime, endOK := end.ParseTime()
		if startOK && endOK && startTime.After(endTime) {
			return "", fmt.Errorf("%w: Date Start %v Is After Date End %v", ErrInvalidUpdate, start, end)
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

	// Helper to validate a spouse's timeline against the marriage dates. The
	// comparisons go through ParseTime: DD-MM-YYYY strings sort by day first, so
	// comparing them directly would call 01-02-2000 later than 31-01-2000.
	validateSpouseDates := func(spouseLabel string, dob, dod person.DateProper) error {
		startTime, startOK := start.ParseTime()
		endTime, endOK := end.ParseTime()

		if birthTime, ok := dob.ParseTime(); ok {
			if startOK && startTime.Before(birthTime) {
				return fmt.Errorf("%w: marriage start date (%v) cannot be before %s's birth date (%v)", ErrInvalidUpdate, start, spouseLabel, dob)
			}
			if endOK && endTime.Before(birthTime) {
				return fmt.Errorf("%w: marriage end date (%v) cannot be before %s's birth date (%v)", ErrInvalidUpdate, end, spouseLabel, dob)
			}
		}

		// Death validation (only evaluated when the spouse has a death date).
		if deathTime, ok := dod.ParseTime(); ok {
			if startOK && startTime.After(deathTime) {
				return fmt.Errorf("%w: marriage start date (%v) cannot be after %s's death date (%v)", ErrInvalidUpdate, start, spouseLabel, dod)
			}
			if endOK && endTime.After(deathTime) {
				return fmt.Errorf("%w: marriage end date (%v) cannot be after %s's death date (%v)", ErrInvalidUpdate, end, spouseLabel, dod)
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

	// The query writes both properties, so each parameter holds the value the
	// property should end up with. `start` and `end` were resolved above, which
	// gives all three cases their correct spelling here:
	//
	//	not sent → the stored value, written back unchanged
	//	cleared  → null, which removes the property instead of storing a null
	//	replaced → the new date, as a native Neo4j date like CreateNewMarriage
	query := `
	MATCH (m:Marriage {id: $id})
	SET m.start = $start, m.end = $end
	`

	params := map[string]any{
		"id":    id,
		"start": nil,
		"end":   nil,
	}

	for _, date := range []struct {
		key   string
		value person.DateProper
	}{
		{"start", start},
		{"end", end},
	} {
		if date.value == "" {
			continue
		}

		neoDate, err := date.value.GetNeoDate()
		if err != nil {
			return "", fmt.Errorf("%w: invalid %s date: %v", ErrInvalidUpdate, date.key, err)
		}
		params[date.key] = *neoDate
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
