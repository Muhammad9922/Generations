package marriage

import (
	"fmt"
	"testing"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/Muhammad9922/Generations/internal/person"
)

func TestCreation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://127.0.0.1:7687")
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
		DateEnd:   "11-10-2024",
	})

	if err != nil || id == "" {
		t.Fatalf("Error Creating New Marriage: %v", err)
	}
}

func TestCreationWithParams(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://127.0.0.1:7687")

	spouseOneName, personOneID, personOneErr := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Person One",
		Gender:      person.Male,
		Alive:       true,
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
	ctx, driver := db.ConnectDatabase("bolt://127.0.0.1:7687")
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

	marriages, err := GetMarriage(ctx, driver, spouseOneId)
	valueMarriages := *marriages

	if err != nil {
		t.Fatalf("Error Getting Marriages: %v", err)
	}

	if len(valueMarriages) != 1 {
		t.Errorf("Only One Marriage Should Have Been Here")
	}

	if valueMarriages[0].id != id {
		t.Error("Marriage ID Should Have been the same as returned when it was created")
	}

}

func TestQueryAllMarriages(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://127.0.0.1:7687")
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
	ctx, driver := db.ConnectDatabase("bolt://127.0.0.1:7687")
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
