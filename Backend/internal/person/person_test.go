package person

import (
	"testing"

	"github.com/Muhammad9922/Generations/internal/db"
)

func TestCreation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx)
	personName, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  "Muhammad",
		ParentID:    "",
		Gender:      "Male",
		DateOfBirth: "22-10-2005",
	})

	if personName == "" {
		t.Errorf("Person Name Should Have Been Defined!")
	}

	if err != nil {
		t.Errorf("Error Creation: %v", err)
	}

	personQuery := PersonQuery{
		Name: "Muhammad",
	}

	exists, err := CheckPersonExistence(ctx, driver, personQuery)

	if exists == false {
		t.Errorf("New User Still Not Found: %v", err)
	}

}

func TestCheckingSimpleExistance(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx)
	person := PersonQuery{
		ID:   "Anything",
		Name: "Anything",
	}
	exists, _ := CheckPersonExistence(ctx, driver, person)

	if exists {
		t.Errorf("Such A User Should Not Be Fount")
	}
}
