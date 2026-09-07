package marriage

import (
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
		DateEnd:   "11-10-2026",
	})

	if err != nil || id == "" {
		t.Fatalf("Error Creating New Marriage: %v", err)
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

	marriages, err := GetMarriages(ctx, driver, spouseOneId)
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
