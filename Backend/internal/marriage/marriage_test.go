package marriage

import (
	"context"
	"fmt"
	"testing"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/Muhammad9922/Generations/internal/person"
)

func TestCreation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	_, spouseOneId, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse One",
		Alive:       false,
		Gender:      person.Male,
		DateOfBirth: "11-10-2000",
		DateOfDeath: "11-10-2026",
	})

	if err != nil {
		t.Fatalf("Error Creating Spouse One %q", err)
	}

	_, spouseTwoId, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse Two",
		Alive:       false,
		Gender:      person.Female,
		DateOfBirth: "11-10-2000",
		DateOfDeath: "11-10-2026",
	})

	if err != nil {
		t.Fatalf("Error Creation Spouse Two %q", err)
	}

	id, err := CreateNewMarriage(ctx, driver, NewMarriage{
		SpouseOne: spouseOneId,
		SpouseTwo: spouseTwoId,
		DateStart: "11-10-2006",
		DateEnd:   "11-10-2024",
	})

	if err != nil {
		t.Fatalf("Error Creating New Marriage: %v", err)
	}

	if id == "" {
		t.Fatalf("No Marriage Was Created Without Errors")
	}
}

func TestCreationWithParams(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")

	spouseOneName, personOneID, personOneErr := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Person One",
		Gender:      person.Male,
		Alive:       false,
		DateOfBirth: "10-10-2000",
		DateOfDeath: "10-10-2020",
	})

	if personOneErr != nil {
		t.Errorf("Unable To Create %s Due To: %q", spouseOneName, personOneErr)
	}

	spouseTwoName, personTwoID, personTwoErr := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName: "Person Two",
		Gender:     person.Male,
		Alive:      true,
	})

	if personTwoErr != nil {
		t.Errorf("Unable To Create %s Due To: %q", spouseTwoName, personTwoErr)
	}

	spouseThreeName, personThreeID, personThreeErr := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName: "Person Three",
		Gender:     person.Female,
		Alive:      true,
	})

	if personThreeErr != nil {
		t.Errorf("Unable To Create %s Due To: %q", spouseThreeName, personThreeErr)
	}

	spouseFourName, personFourID, personFourErr := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName: "Person Four",
		Gender:     person.Female,
		Alive:      true,
	})

	if personFourErr != nil {
		t.Errorf("Unable To Create %s Due To: %q", spouseFourName, personFourErr)
	}

	tests := []NewMarriage{
		// No Spouses
		{
			ID: "Hello",
		},

		// One Spouse
		{
			SpouseOne: personOneID,
		},

		// Two Males
		{
			SpouseOne: personOneID,
			SpouseTwo: personTwoID,
		},

		// Two Females
		{
			SpouseOne: personThreeID,
			SpouseTwo: personThreeID,
		},

		// Invalid Start And End Date
		{
			SpouseOne: personOneID,
			SpouseTwo: personThreeID,
			DateStart: "33-22-2004",
			DateEnd:   "33-22-2004",
		},

		// Invalid Start Date
		{
			SpouseOne: personTwoID,
			SpouseTwo: personFourID,
			DateStart: "33-22-2004",
		},

		// Invalid End Date
		{
			SpouseOne: personOneID,
			SpouseTwo: personFourID,
			DateEnd:   "33-22-2003",
		},

		// Invalid Users
		{
			SpouseOne: personFourID,
			SpouseTwo: "invalid-id",
		},

		{
			SpouseOne: "invalid-id",
			SpouseTwo: personOneID,
		},

		// Marriage Before Birth
		{
			SpouseOne: personOneID,
			SpouseTwo: personThreeID,
			DateStart: "11-11-1990",
		},

		// Marriage After Birth
		{
			SpouseOne: personOneID,
			SpouseTwo: personThreeID,
			DateEnd:   "11-11-2025",
		},
	}

	for index, test := range tests {
		testName := fmt.Sprintf("Test Name: %v", index)
		t.Run(testName, func(t *testing.T) {
			_, err := CreateNewMarriage(ctx, driver, test)

			if err == nil {
				t.Errorf("Failed %q", err)
			}

			t.Logf("Index: %v", index)
		})
	}

	// Triggering Error
	driver.Close(ctx)
	_, err := CreateNewMarriage(
		ctx, driver, tests[0],
	)

	if err == nil {
		t.Errorf("Should Have Caused An Error")
	}

}

func TestQuery(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	_, spouseOneId, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse One",
		Alive:       false,
		Gender:      person.Male,
		DateOfBirth: "11-10-2000",
		DateOfDeath: "11-10-2026",
	})

	if err != nil {
		t.Fatalf("Error Creating Spouse One")
	}

	_, spouseTwoId, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse Two",
		Alive:       false,
		Gender:      person.Female,
		DateOfBirth: "11-10-2000",
		DateOfDeath: "11-10-2026",
	})

	if err != nil {
		t.Fatalf("Error Creation Spouse Two")
	}

	id, err := CreateNewMarriage(ctx, driver, NewMarriage{
		SpouseOne: spouseOneId,
		SpouseTwo: spouseTwoId,
		DateStart: "11-10-2006",
		DateEnd:   "11-10-2026",
	})

	if err != nil || id == "" {
		t.Fatalf("Error Creating New Marriage: %v", err)
	}

	marriages, err := GetMarriageFromMarriageId(ctx, driver, id)
	valueMarriages := marriages

	if err != nil {
		t.Fatalf("Error Getting Marriages: %v", err)
	}

	if len(valueMarriages) != 1 {
		t.Errorf("Only One Marriage Should Have Been Here")
	}

	if valueMarriages[0].Id != id {
		t.Error("Marriage ID Should Have been the same as returned when it was created")
	}

	spouseOneFromMarriageID, spouseTwoFromMarriageID, _, err := GetSpouses(ctx, driver, SpouseQueryParams{
		MarriageId: id,
	})

	if err != nil {
		t.Errorf("Error While Getting Spouses: %q", err)
	}

	if spouseOneFromMarriageID != spouseOneId && spouseOneFromMarriageID != spouseTwoId {
		t.Errorf("Spouse One Is Not Returned Properly: %v | %v || %v | %v", spouseOneFromMarriageID, spouseTwoFromMarriageID, spouseOneId, spouseTwoId)
	}

	if spouseTwoFromMarriageID != spouseOneId && spouseTwoFromMarriageID != spouseTwoId {
		t.Errorf("Spouse Two Is Not Returned Properly: %v | %v || %v | %v", spouseOneFromMarriageID, spouseTwoFromMarriageID, spouseOneId, spouseTwoId)
	}

	spouseOneFromSpouseOneID, spouseTwoFromSpouseOneId, marriageIDFromSpouseOneId, err := GetSpouses(ctx, driver, SpouseQueryParams{
		SpouseId: spouseOneId,
	})

	if err != nil {
		t.Errorf("Error While Getting Spouses: %q", err)
	}

	if spouseOneFromSpouseOneID != spouseOneId && spouseTwoFromSpouseOneId != spouseTwoId {
		t.Errorf("Spouse One Is Not Returned Properly: %v | %v || %v | %v", spouseOneFromSpouseOneID, spouseTwoFromSpouseOneId, spouseOneId, spouseTwoId)
	}

	if spouseTwoFromSpouseOneId != spouseOneId && spouseTwoFromSpouseOneId != spouseTwoId {
		t.Errorf("Spouse Two Is Not Returned Properly: %v | %v || %v | %v", spouseOneFromSpouseOneID, spouseTwoFromSpouseOneId, spouseOneId, spouseTwoId)
	}

	if marriageIDFromSpouseOneId != id {
		t.Errorf("Marriage ID from Spouse One Isn't Correct %v | %v", marriageIDFromSpouseOneId, id)
	}

}

func TestQueryAllMarriages(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)
	marriagesPointer, err := GetAllMarriages(ctx, driver)
	marriages := *marriagesPointer

	t.Log(marriages)

	if err != nil {
		t.Errorf("Error Getting Marriages: %q", err)
	}

	if len(marriages) == 0 {
		t.Errorf("Error Getting Marriages")
	}
}

func TestDelete(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	_, maleId, maleErr := person.CreateNewPerson(ctx, driver, person.NewPerson{
		Gender:      person.Male,
		Alive:       false,
		PersonName:  "Male Member",
		DateOfBirth: person.DateProper("11-11-2000"),
	})

	if maleErr != nil {
		t.Errorf("ERROR (ML): %q", maleErr)
	}

	_, femaleId, femaleErr := person.CreateNewPerson(ctx, driver, person.NewPerson{
		Gender:      person.Female,
		Alive:       false,
		PersonName:  "Female Member",
		DateOfBirth: person.DateProper("11-11-2000"),
	})

	if femaleErr != nil {
		t.Errorf("ERROR (FE): %q", femaleErr)
	}

	MarriageId, err := CreateNewMarriage(ctx, driver, NewMarriage{
		SpouseOne: maleId,
		SpouseTwo: femaleId,
	})

	if err != nil {
		t.Errorf("ERROR (M): %q", err)
	}

	t.Logf("Marriage ID: %v", MarriageId)

	deleted, deleteError := DeleteMarriage(ctx, driver, MarriageId)

	if deleteError != nil {
		t.Errorf("ERROR (DM): %q", deleteError)
	}

	t.Logf("Deleted: %v", deleted)
}

func TestUpdateMarriage(t *testing.T) {
	ctx := context.Background()
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	// --- Test Fixture Setup ---
	// Spouse 1: DOB = 11-10-2000, DOD = 11-10-2026
	_, spouseOneID, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse One",
		Alive:       false,
		Gender:      person.Male,
		DateOfBirth: "11-10-2000",
		DateOfDeath: "11-10-2026",
	})
	if err != nil {
		t.Fatalf("Failed to create Spouse One: %v", err)
	}

	// Spouse 2: DOB = 11-10-2000, DOD = 11-10-2026
	_, spouseTwoID, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse Two",
		Alive:       false,
		Gender:      person.Female,
		DateOfBirth: "11-10-2000",
		DateOfDeath: "11-10-2026",
	})
	if err != nil {
		t.Fatalf("Failed to create Spouse Two: %v", err)
	}

	// Helper to seed a standard marriage record for testing updates
	createFixtureMarriage := func(t *testing.T) string {
		id, err := CreateNewMarriage(ctx, driver, NewMarriage{
			SpouseOne: spouseOneID,
			SpouseTwo: spouseTwoID,
			DateStart: "11-10-2018",
			DateEnd:   "11-10-2022",
		})
		if err != nil || id == "" {
			t.Fatalf("Failed setting up baseline marriage fixture: %v", err)
		}
		return id
	}

	// --- 1. Successful Update Scenarios ---

	t.Run("Full Update (Start and End Dates)", func(t *testing.T) {
		marriageID := createFixtureMarriage(t)

		updatePayload := MarriageUpdate{
			DateStart: person.DateProper("11-10-2019"),
			DateEnd:   person.DateProper("11-10-2023"),
		}

		t.Logf("ID For Marrige: %v | Spouse One: %v -- Spouse Two: %v", marriageID, spouseOneID, spouseTwoID)

		updatedID, err := UpdateMarriageDates(ctx, driver, marriageID, updatePayload)
		if err != nil {
			t.Fatalf("Expected update to succeed, got error: %v", err)
		}
		if updatedID != marriageID {
			t.Errorf("Expected updated ID to be %s, got %s", marriageID, updatedID)
		}

		// Verify state persistence
		fetched, err := GetMarriageFromMarriageId(ctx, driver, marriageID)
		if err != nil || fetched == nil || len(fetched) == 0 {
			t.Fatalf("Failed fetching updated marriage record: %v", err)
		}

		record := (fetched)[0]
		if record.Start != updatePayload.DateStart {
			t.Errorf("Expected start date %v, got %v", updatePayload.DateStart, record.Start)
		}
		if record.End != updatePayload.DateEnd {
			t.Errorf("Expected end date %v, got %v", updatePayload.DateEnd, record.End)
		}
	})

	t.Run("Partial Update (Start Date Only)", func(t *testing.T) {
		marriageID := createFixtureMarriage(t)

		updatePayload := MarriageUpdate{
			DateStart: "11-10-2020",
		}

		_, err := UpdateMarriageDates(ctx, driver, marriageID, updatePayload)
		if err != nil {
			t.Fatalf("Expected partial start date update to succeed, got: %v", err)
		}

		fetched, _ := GetMarriageFromMarriageId(ctx, driver, marriageID)
		record := (fetched)[0]

		if record.Start != "11-10-2020" {
			t.Errorf("Expected updated start date 11-10-2020, got %v", record.Start)
		}
		if record.End != "11-10-2022" { // Must preserve original end date
			t.Errorf("Expected original end date 11-10-2022 to remain, got %v", record.End)
		}
	})

	t.Run("Partial Update (End Date Only)", func(t *testing.T) {
		marriageID := createFixtureMarriage(t)

		updatePayload := MarriageUpdate{
			DateEnd: "11-10-2024",
		}

		_, err := UpdateMarriageDates(ctx, driver, marriageID, updatePayload)
		if err != nil {
			t.Fatalf("Expected partial end date update to succeed, got: %v", err)
		}

		fetched, _ := GetMarriageFromMarriageId(ctx, driver, marriageID)
		record := (fetched)[0]

		if record.Start != "11-10-2018" { // Must preserve original start date
			t.Errorf("Expected original start date 11-10-2018 to remain, got %v", record.Start)
		}
		if record.End != "11-10-2024" {
			t.Errorf("Expected updated end date 11-10-2024, got %v", record.End)
		}
	})

	// --- 2. Failure & Validation Scenarios ---

	testCases := []struct {
		name          string
		marriageID    func() string
		updatePayload MarriageUpdate
	}{
		{
			name:       "Invalid Marriage ID",
			marriageID: func() string { return "invalid-uuid-9999" },
			updatePayload: MarriageUpdate{
				DateStart: "11-10-2019",
			},
		},
		{
			name:       "Malformed Start Date",
			marriageID: func() string { return createFixtureMarriage(t) },
			updatePayload: MarriageUpdate{
				DateStart: "invalid-date-format",
			},
		},
		{
			name:       "Malformed End Date",
			marriageID: func() string { return createFixtureMarriage(t) },
			updatePayload: MarriageUpdate{
				DateEnd: "32-13-2020",
			},
		},
		{
			name:       "Marriage Start Before Spouse Birth Date",
			marriageID: func() string { return createFixtureMarriage(t) },
			updatePayload: MarriageUpdate{
				DateStart: "11-10-1995", // Spouse DOB is 11-10-2000
			},
		},
		{
			name:       "Marriage End After Spouse Death Date",
			marriageID: func() string { return createFixtureMarriage(t) },
			updatePayload: MarriageUpdate{
				DateEnd: "11-10-2030", // Spouse DOD is 11-10-2026
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.marriageID()
			_, err := UpdateMarriageDates(ctx, driver, id, tc.updatePayload)
			if err == nil {
				t.Fatalf("Expected validation error for '%s', but operation succeeded", tc.name)
			}
		})
	}
}
