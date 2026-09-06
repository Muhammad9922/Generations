package person

import (
	"testing"

	"github.com/Muhammad9922/Generations/internal/db"
)

func TestCreation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx)
	personName, _, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  "Muhammad",
		MarriageID:  "",
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

func TestCreationValidation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx)

	tests := []struct {
		name   string
		person NewPerson
	}{
		{
			"Wrong Date Of Birth",
			NewPerson{
				PersonName:  "Person Name 1",
				MarriageID:  "",
				DateOfBirth: "2023-11-10",
				Gender:      "Male",
				Alive:       true,
			},
		},
		{
			"Empty Name",
			NewPerson{
				PersonName:  "",
				MarriageID:  "",
				DateOfBirth: "2023-11-10",
				Gender:      "Male",
				Alive:       true,
			},
		},
		{
			"Wrong Gender Content",
			NewPerson{
				PersonName:  "Person Name 2",
				MarriageID:  "",
				DateOfBirth: "11-10-2003",
				Gender:      "Invalid Gender",
				Alive:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(tx *testing.T) {
			personName, _, err := CreateNewPerson(ctx, driver, tt.person)

			if personName != "" || err == nil {
				tx.Errorf("The Account Should Have Not Been Created")
			}
		})
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

func TestCheckSimpleExistanceWithClosedDriver(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	driver.Close(ctx)

	person := PersonQuery{
		ID:   "Anything",
		Name: "Anything",
	}

	_, err := CheckPersonExistence(ctx, driver, person)

	if err == nil {
		t.Errorf("Error Should Have Been Thrown Due To Closed Driver")
	}

}
