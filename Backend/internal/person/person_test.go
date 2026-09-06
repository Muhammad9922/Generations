package person

import (
	"strconv"
	"testing"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/db"
)

func TestCreation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx)
	personName, _, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  "Muhammad",
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
				DateOfBirth: "2023-11-10",
				Gender:      "Male",
				Alive:       true,
			},
		},
		{
			"Empty Name",
			NewPerson{
				PersonName:  "",
				DateOfBirth: "2023-11-10",
				Gender:      "Male",
				Alive:       true,
			},
		},
		{
			"Wrong Gender Content",
			NewPerson{
				PersonName:  "Person Name 2",
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

func TestCreationWithUID(t *testing.T) {
	uuid := uuid.New().String()
	newPerson := NewPerson{
		id:          uuid,
		PersonName:  "Some Person Name",
		Gender:      "Male",
		DateOfBirth: "20-10-2000",
	}

	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx)

	_, id, _ := CreateNewPerson(ctx, driver, newPerson)

	if id != uuid {
		t.Errorf("The UUID Does Not Match %s != %s", id, uuid)
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

func TestCheckSameNameUsers(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx) // 1. Defer closure so driver stays open during tests

	users := []NewPerson{
		{
			PersonName:  "Same Name",
			Gender:      Male,
			DateOfBirth: "09-11-2000",
			Alive:       true,
		},
		{
			PersonName:  "Same Name",
			Gender:      Female,
			DateOfBirth: "09-11-2000",
			Alive:       true,
		},
	}

	createdIDs := make([]string, len(users))

	for index, user := range users {
		t.Run(strconv.Itoa(index), func(tx *testing.T) {
			_, id, err := CreateNewPerson(ctx, driver, user)
			if err != nil { // 3. Assert creation success
				tx.Fatalf("failed to create person %d: %v", index, err)
			}

			createdIDs[index] = id
		})
	}

	t.Run("Checking All Exist", func(tx *testing.T) {
		for _, id := range createdIDs {
			exists, err := CheckPersonExistence(ctx, driver, PersonQuery{
				ID: id,
			})
			if err != nil {
				tx.Fatalf("failed to check existence for ID %q: %v", id, err)
			}
			if !exists {
				tx.Errorf("user with ID %q does not exist in database", id)
			}
		}
	})
}

func TestDeleteUser(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")

	var id string

	t.Run("Creating Sample User", func(t *testing.T) {
		_, personID, err := CreateNewPerson(ctx, driver, NewPerson{
			PersonName:  "New Person Name",
			Gender:      Male,
			DateOfBirth: "10-11-2005",
		})

		if err != nil {
			t.Fatalf("There Was An Error Creating New User %v", err)
		}

		id = personID
	})

	t.Run("Delete New User", func(t *testing.T) {
		_, success, err := DeleteUser(ctx, driver, id)
		if !success {
			t.Fatalf("Unsuccessful Attempt To Delete User")
		}
		if err != nil {
			t.Fatalf("Error While Deleting: %v", err)
		}
	})

	t.Run("Check Existance", func(t *testing.T) {
		exists, err := CheckPersonExistence(ctx, driver, PersonQuery{
			ID: id,
		})

		if exists {
			t.Errorf("User Exists Even After Deleting")
		}

		if err != nil {
			t.Errorf("Error While Checking For User Existance: %v", err)
		}
	})

	t.Run("Delete To Non Existing User ", func(t *testing.T) {
		_, success, err := DeleteUser(ctx, driver, id)
		if success {
			t.Fatalf("Somehow Attempt To Delete User Succeded")
		}
		if err != nil {
			t.Fatalf("Error While Deleting: %v", err)
		}
	})

	driver.Close(ctx)
	t.Run("Deleting New User After Closing The Driver", func(t *testing.T) {
		_, success, err := DeleteUser(ctx, driver, id)
		if success {
			t.Fatalf("Somehow Closed Driver Deleted A Non Existing User")
		}

		if err == nil {
			t.Fatalf("Somehow Closed Driver Didn't Throw Error")
		}
	})

}
