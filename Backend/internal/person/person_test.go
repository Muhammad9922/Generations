package person

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func TestCreation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)
	personName, _, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  "Muhammad",
		Gender:      "Male",
		DateOfBirth: "22-10-2005",
		Alive:       false,
		DateOfDeath: "22-10-2008",
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
		t.Errorf("New User Not Created: %v", err)
	}
}

func TestCreationValidation(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
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
		{
			"No Gender Provided",
			NewPerson{
				PersonName:  "Person Name 1",
				DateOfBirth: "10-10-2010",
				Alive:       true,
			},
		},
		{
			"Invalid Date Of Death",
			NewPerson{
				PersonName:  "Person Name 5",
				DateOfBirth: "11-10-2003",
				Gender:      "Male",
				Alive:       false,
				DateOfDeath: "10-13-2029",
			},
		},
		{
			"Date Of Death In Past",
			NewPerson{
				PersonName:  "Person Name 6",
				DateOfBirth: "11-10-2003",
				Gender:      "Male",
				Alive:       false,
				DateOfDeath: "11-10-2000",
			},
		},
		{
			"Just Name",
			NewPerson{
				PersonName: "Person Name 4",
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
		Id:          uuid,
		PersonName:  "Some Person Name",
		Gender:      "Male",
		DateOfBirth: "20-10-2000",
	}

	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	_, id, _ := CreateNewPerson(ctx, driver, newPerson)

	if id != uuid {
		t.Errorf("The UUID Does Not Match %s != %s", id, uuid)
	}
}

func TestCheckingSimpleExistance(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
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
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
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
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
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
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")

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

// readPersonProperties fetches the stored properties of a Person node by its ID.
// It is a test helper used to verify that updates were actually persisted.
func readPersonProperties(t *testing.T, ctx context.Context, driver neo4j.Driver, id string) map[string]any {
	t.Helper()

	result, err := neo4j.ExecuteQuery(ctx, driver,
		`
		MATCH (p:Person {id: $id})
		RETURN p.name AS name, p.gender AS gender, p.date_of_birth AS date_of_birth, p.alive AS alive, p.date_of_death AS date_of_death
		`,
		map[string]any{"id": id},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("failed to read person %q from database: %v", id, err)
	}

	if len(result.Records) == 0 {
		t.Fatalf("person %q not found in database while verifying update", id)
	}

	record := result.Records[0]
	props := make(map[string]any, 4)
	for _, key := range []string{"name", "gender", "date_of_birth", "alive", "date_of_death"} {
		if value, found := record.Get(key); found {
			props[key] = value
		}
	}

	return props
}

func TestUpdatePerson(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	// A UUID-backed name ensures this test does not collide with data left
	// behind by the other tests that share the same database.
	originalName := "Update Test Person " + uuid.New().String()
	originalGender := Male
	originalDOB := DateProper("01-01-1990")
	originalDOD := DateProper("10-10-2003")
	originalAlive := false

	_, id, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  originalName,
		Gender:      originalGender,
		DateOfBirth: originalDOB,
		Alive:       originalAlive,
		DateOfDeath: originalDOD,
	})

	if err != nil {
		t.Fatalf("failed to create person for update test: %v", err)
	}
	if id == "" {
		t.Fatalf("expected a non-empty id when creating a person")
	}

	updatedName := "Updated Person Name"
	updatedGender := Female
	updatedDOB := DateProper("02-02-1992")
	updatedAlive := false
	updatedDOD := DateProper("02-11-2023")

	t.Run("Update All Fields", func(t *testing.T) {
		name, wasUpdated, err := UpdatePerson(ctx, driver, id, UpdateUser{
			Name:        &updatedName,
			Gender:      &updatedGender,
			DateOfBirth: &updatedDOB,
			Alive:       &updatedAlive,
			DateOfDeath: &updatedDOD,
		})

		if err != nil {
			t.Fatalf("expected no error updating all fields, got: %v", err)
		}
		if !wasUpdated {
			t.Errorf("expected wasUpdated to be true when updating all fields")
		}
		if name != updatedName {
			t.Errorf("expected returned name %q, got %q", updatedName, name)
		}

		props := readPersonProperties(t, ctx, driver, id)
		if got := props["name"]; got != updatedName {
			t.Errorf("persisted name = %v, want %q", got, updatedName)
		}
		if got := props["gender"]; got != string(updatedGender) {
			t.Errorf("persisted gender = %v, want %q", got, updatedGender)
		}
		if got := props["date_of_birth"]; got != string(updatedDOB) {
			t.Errorf("persisted date_of_birth = %v, want %q", got, updatedDOB)
		}
		if got := props["alive"]; got != updatedAlive {
			t.Errorf("persisted alive = %v, want %v", got, updatedAlive)
		}
		if got := props["date_of_death"]; got != string(updatedDOD) {
			t.Errorf("presisted death - %v, want %v", got, updatedDOD)
		}
	})

	updatedNameOnly := "Updated Name Only"
	t.Run("Update Only Name", func(t *testing.T) {
		name, wasUpdated, err := UpdatePerson(ctx, driver, id, UpdateUser{
			Name: &updatedNameOnly,
		})

		if err != nil {
			t.Fatalf("expected no error updating name only, got: %v", err)
		}
		if !wasUpdated {
			t.Errorf("expected wasUpdated to be true when updating name")
		}
		if name != updatedNameOnly {
			t.Errorf("expected returned name %q, got %q", updatedNameOnly, name)
		}

		props := readPersonProperties(t, ctx, driver, id)
		if got := props["name"]; got != updatedNameOnly {
			t.Errorf("persisted name = %v, want %q", got, updatedNameOnly)
		}
		if got := props["gender"]; got != string(updatedGender) {
			t.Errorf("gender should be untouched = %v, want %q", got, updatedGender)
		}
		if got := props["date_of_birth"]; got != string(updatedDOB) {
			t.Errorf("date_of_birth should be untouched = %v, want %q", got, updatedDOB)
		}
		if got := props["alive"]; got != updatedAlive {
			t.Errorf("alive should be untouched = %v, want %v", got, updatedAlive)
		}
	})

	updatedGenderAgain := Male
	t.Run("Update Only Gender", func(t *testing.T) {
		name, wasUpdated, err := UpdatePerson(ctx, driver, id, UpdateUser{
			Gender: &updatedGenderAgain,
		})

		if err != nil {
			t.Fatalf("expected no error updating gender only, got: %v", err)
		}
		if !wasUpdated {
			t.Errorf("expected wasUpdated to be true when updating gender")
		}
		if name != updatedNameOnly {
			t.Errorf("expected returned name %q, got %q", updatedNameOnly, name)
		}

		props := readPersonProperties(t, ctx, driver, id)
		if got := props["gender"]; got != string(updatedGenderAgain) {
			t.Errorf("persisted gender = %v, want %q", got, updatedGenderAgain)
		}
		if got := props["name"]; got != updatedNameOnly {
			t.Errorf("name should be untouched = %v, want %q", got, updatedNameOnly)
		}
		if got := props["date_of_birth"]; got != string(updatedDOB) {
			t.Errorf("date_of_birth should be untouched = %v, want %q", got, updatedDOB)
		}
		if got := props["alive"]; got != updatedAlive {
			t.Errorf("alive should be untouched = %v, want %v", got, updatedAlive)
		}
	})

	updatedDOBOnly := DateProper("03-03-1993")
	t.Run("Update Only Date Of Birth", func(t *testing.T) {
		name, wasUpdated, err := UpdatePerson(ctx, driver, id, UpdateUser{
			DateOfBirth: &updatedDOBOnly,
		})

		if err != nil {
			t.Fatalf("expected no error updating date of birth only, got: %v", err)
		}
		if !wasUpdated {
			t.Errorf("expected wasUpdated to be true when updating date of birth")
		}
		if name != updatedNameOnly {
			t.Errorf("expected returned name %q, got %q", updatedNameOnly, name)
		}

		props := readPersonProperties(t, ctx, driver, id)
		if got := props["date_of_birth"]; got != string(updatedDOBOnly) {
			t.Errorf("persisted date_of_birth = %v, want %q", got, updatedDOBOnly)
		}
		if got := props["name"]; got != updatedNameOnly {
			t.Errorf("name should be untouched = %v, want %q", got, updatedNameOnly)
		}
		if got := props["gender"]; got != string(updatedGenderAgain) {
			t.Errorf("gender should be untouched = %v, want %q", got, updatedGenderAgain)
		}
		if got := props["alive"]; got != updatedAlive {
			t.Errorf("alive should be untouched = %v, want %v", got, updatedAlive)
		}
	})

	updatedAliveAgain := true
	t.Run("Update Only Alive", func(t *testing.T) {
		name, wasUpdated, err := UpdatePerson(ctx, driver, id, UpdateUser{
			Alive: &updatedAliveAgain,
		})

		if err != nil {
			t.Fatalf("expected no error updating alive only, got: %v", err)
		}
		if !wasUpdated {
			t.Errorf("expected wasUpdated to be true when updating alive")
		}
		if name != updatedNameOnly {
			t.Errorf("expected returned name %q, got %q", updatedNameOnly, name)
		}

		props := readPersonProperties(t, ctx, driver, id)
		if got := props["alive"]; got != updatedAliveAgain {
			t.Errorf("persisted alive = %v, want %v", got, updatedAliveAgain)
		}
		if got := props["name"]; got != updatedNameOnly {
			t.Errorf("name should be untouched = %v, want %q", got, updatedNameOnly)
		}
		if got := props["gender"]; got != string(updatedGenderAgain) {
			t.Errorf("gender should be untouched = %v, want %q", got, updatedGenderAgain)
		}
		if got := props["date_of_birth"]; got != string(updatedDOBOnly) {
			t.Errorf("date_of_birth should be untouched = %v, want %q", got, updatedDOBOnly)
		}
	})

	t.Run("Update With No Fields", func(t *testing.T) {
		name, wasUpdated, err := UpdatePerson(ctx, driver, id, UpdateUser{})

		if err != nil {
			t.Fatalf("expected no error when no fields are supplied, got: %v", err)
		}
		// With an empty UpdateUser the generated Cypher is still `SET p += {}`.
		// Neo4j reports the SET clause as containing updates in its query
		// summary counters, so wasUpdated comes back true even though no
		// property value actually changed.
		if !wasUpdated {
			t.Errorf("expected wasUpdated to be true for an empty update")
		}
		if name != updatedNameOnly {
			t.Errorf("expected returned name %q, got %q", updatedNameOnly, name)
		}
	})

	t.Run("Update Non-Existent User", func(t *testing.T) {
		missingID := uuid.New().String()

		name, wasUpdated, err := UpdatePerson(ctx, driver, missingID, UpdateUser{
			Name: &updatedNameOnly,
		})

		if err == nil {
			t.Fatalf("expected an error when updating a non-existent user")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected error to mention the user does not exist, got: %v", err)
		}
		if wasUpdated {
			t.Errorf("expected wasUpdated to be false for a non-existent user")
		}
		if name != "" {
			t.Errorf("expected empty returned name for a non-existent user, got %q", name)
		}
	})

	t.Run("Person Still Exists After Updates", func(t *testing.T) {
		exists, err := CheckPersonExistence(ctx, driver, PersonQuery{ID: id})
		if err != nil {
			t.Fatalf("failed to check existence after updates: %v", err)
		}
		if !exists {
			t.Errorf("person %q should still exist after updates", id)
		}
	})
}

func TestUpdatePersonWithClosedDriver(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	driver.Close(ctx)

	newName := "Closed Driver Name"
	name, wasUpdated, err := UpdatePerson(ctx, driver, "any-id", UpdateUser{Name: &newName})

	if err == nil {
		t.Fatalf("expected an error when updating with a closed driver")
	}
	if wasUpdated {
		t.Errorf("expected wasUpdated to be false with a closed driver")
	}
	if name != "" {
		t.Errorf("expected empty returned name with a closed driver, got %q", name)
	}
}

func TestMoreUpdateData(t *testing.T) {
	defaultCreation := NewPerson{
		Alive:       false,
		DateOfBirth: "11-10-2000",
		DateOfDeath: "10-10-2003",
		PersonName:  "Some Name Here",
		Gender:      Male,
	}

	tests := []struct {
		name           string
		shouldComplete bool
		creation       NewPerson
		update         UpdateUser
	}{
		{
			name:           "Basic Test",
			shouldComplete: true,
			creation:       defaultCreation,
			update: UpdateUser{
				Name:        Ptr("New Name Here"),
				DateOfBirth: Ptr(DateProper("11-10-1999")),
				DateOfDeath: Ptr(DateProper("11-10-2009")),
				Gender:      Ptr(Female),
				Alive:       Ptr(false),
			},
		},
		{
			name:           "Invalid Date Of Birth",
			shouldComplete: false,
			creation:       defaultCreation,
			update: UpdateUser{
				DateOfBirth: Ptr(DateProper("11-32-1999")),
			},
		},
		{
			name:           "Invalid Date Of Death",
			shouldComplete: false,
			creation:       defaultCreation,
			update: UpdateUser{
				DateOfDeath: Ptr(DateProper("11-12-200")),
			},
		},
		{
			name:           "New Date Of Death Is Before Date Of Birth",
			shouldComplete: false,
			creation:       defaultCreation,
			update: UpdateUser{
				DateOfDeath: Ptr(DateProper("11-10-1993")),
			},
		},
		{
			name:           "New Date Of Birth Is After The Date Of Death",
			shouldComplete: false,
			creation:       defaultCreation,
			update: UpdateUser{
				DateOfBirth: Ptr(DateProper("10-10-2009")),
			},
		},
		// --- NEW TESTS ADDED BELOW ---
		{
			name:           "Partial Update - Only Name",
			shouldComplete: true,
			creation:       defaultCreation,
			update: UpdateUser{
				Name: Ptr("Updated Name Only"),
			},
		},
		{
			name:           "Partial Update - Only Gender",
			shouldComplete: true,
			creation:       defaultCreation,
			update: UpdateUser{
				Gender: Ptr(Female),
			},
		},
		{
			name:           "Empty Name Validation Should Fail",
			shouldComplete: false,
			creation:       defaultCreation,
			update: UpdateUser{
				Name: Ptr(""), // Assuming your validation rejects empty names
			},
		},
		{
			name:           "Update Both Dates Together - New DOD Before New DOB",
			shouldComplete: false,
			creation:       defaultCreation,
			update: UpdateUser{
				DateOfBirth: Ptr(DateProper("10-10-2020")),
				DateOfDeath: Ptr(DateProper("10-10-2015")),
			},
		},
		{
			name:           "DOB and DOD Are On The Same Day",
			shouldComplete: true,
			creation:       defaultCreation,
			update: UpdateUser{
				DateOfBirth: Ptr(DateProper("05-05-2015")),
				DateOfDeath: Ptr(DateProper("05-05-2015")),
			},
		},
		{
			name:           "Invalid Date Format - Wrong Separators",
			shouldComplete: false,
			creation:       defaultCreation,
			update: UpdateUser{
				DateOfBirth: Ptr(DateProper("11/10/1999")),
			},
		},
	}

	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, id, err := CreateNewPerson(ctx, driver, test.creation)

			if err != nil {
				t.Fatalf("Error Creating New User")
			}

			_, updated, err := UpdatePerson(ctx, driver, id, test.update)

			if test.shouldComplete {
				if err != nil {
					t.Fatalf("There Was An Error Updating The User: %q", err)
				}
			} else {
				if err == nil {
					t.Error("There should have been an error but wasn't")
				}
			}

			if test.shouldComplete {
				if !updated {
					t.Fatalf("Unable To Update The User For Some Reason")
				}
			} else {
				if updated {
					t.Errorf("Able To Update The User For Some Reason While It Should Have Failed")
				}
			}

		})
	}

	t.Run("Trying to update non existant id", func(t *testing.T) {
		_, _, err := UpdatePerson(ctx, driver, uuid.New().String(), UpdateUser{
			Name: Ptr("Hello"),
		})

		if err == nil {
			t.Errorf("Didn't Throw Error When Tring To Update A Non Existant User")
		}
	})
}

func TestPersonQuery(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")

	tests := []struct {
		name   string
		person NewPerson
	}{
		{
			"No Date Of Birth Status",
			NewPerson{
				Id:         uuid.New().String(),
				PersonName: "Person Name 3",
				Gender:     "Male",
				Alive:      true,
			},
		},
		{
			"Full Account",
			NewPerson{
				Id:          uuid.New().String(),
				PersonName:  "Person Name 4",
				Gender:      "Male",
				Alive:       false,
				DateOfBirth: "11-10-2005",
				DateOfDeath: "11-10-2009",
			},
		},
	}

	for _, test := range tests {
		t.Run("Testing Creation Of "+test.name, func(t *testing.T) {
			personName, personID, err := CreateNewPerson(ctx, driver, test.person)

			if err != nil {
				t.Errorf("Error Creating New User ID %s and User Name %s", test.person.Id, test.person.PersonName)
			}

			if personName == "" || personID == "" {
				t.Errorf("Error Creating New User ID %s and User Name %s", test.person.Id, test.person.PersonName)
			}
		})

		t.Run("Testing Query "+test.name, func(t *testing.T) {
			person, err := GetPerson(ctx, driver, test.person.Id)
			t.Logf(
				"Person Detailed: %v",
				person,
			)
			if err != nil || person == nil {
				t.Errorf("Error While Querying User %v", err)
			}
			if len(strings.Split(person.PersonName, "")) < 4 {
				t.Errorf("Invalid User Name: %s", person.PersonName)
			}

			if test.person.DateOfBirth != "" && person.DateOfBirth != test.person.DateOfBirth {
				t.Errorf("Invalid Date Of Birth %v | %v", person.DateOfBirth, test.person.DateOfBirth)
			}

			if test.person.DateOfDeath != "" && person.DateOfDeath != test.person.DateOfDeath {
				t.Errorf("Invalid Date Of Death %v | %v", person.DateOfDeath, test.person.DateOfDeath)
			}
		})
	}

	t.Run("Querying Non Existant ID", func(t *testing.T) {
		_, err := GetPerson(ctx, driver, uuid.NewV4().String())
		if err == nil {
			t.Errorf("A Non Existant ID Should Give An Error")
		}
	})

	driver.Close(ctx)
	t.Run("Testing Closed Driver", func(t *testing.T) {
		_, err := GetPerson(ctx, driver, tests[0].person.Id)
		if err == nil {
			t.Errorf("A Non Existant ID Should Give An Error")
		}
	})

}

func TestFullAccountPerson(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	// Setup input with all fields populated
	expected := NewPerson{
		Id:          uuid.New().String(),
		PersonName:  "Person Name 4",
		Gender:      "Male",
		Alive:       false,
		DateOfBirth: "11-10-2005",
		DateOfDeath: "11-10-2009",
	}

	// Step 1: Create Person
	t.Run("Create Full Account", func(t *testing.T) {
		createdName, createdID, err := CreateNewPerson(ctx, driver, expected)
		if err != nil {
			t.Fatalf("Failed to create full account person: %v", err)
		}
		if createdID != expected.Id {
			t.Errorf("ID mismatch on creation: got %q, want %q", createdID, expected.Id)
		}
		if createdName != expected.PersonName {
			t.Errorf("Name mismatch on creation: got %q, want %q", createdName, expected.PersonName)
		}
	})

	// Step 2: Query & Validate Every Field
	t.Run("Query and Verify All Fields", func(t *testing.T) {
		got, err := GetPerson(ctx, driver, expected.Id)
		if err != nil {
			t.Fatalf("Unexpected error querying person %s: %v", expected.Id, err)
		}
		if got == nil {
			t.Fatalf("Expected valid person record for ID %s, got nil", expected.Id)
		}

		// Strict assertions for all attributes
		if got.Id != expected.Id {
			t.Errorf("Id = %q; want %q", got.Id, expected.Id)
		}
		if got.PersonName != expected.PersonName {
			t.Errorf("PersonName = %q; want %q", got.PersonName, expected.PersonName)
		}
		if got.Gender != expected.Gender {
			t.Errorf("Gender = %q; want %q", got.Gender, expected.Gender)
		}
		if got.Alive != expected.Alive {
			t.Errorf("Alive = %t; want %t", got.Alive, expected.Alive)
		}
		if got.DateOfBirth != expected.DateOfBirth {
			t.Errorf("DateOfBirth = %q; want %q", got.DateOfBirth, expected.DateOfBirth)
		}
		if got.DateOfDeath != expected.DateOfDeath {
			t.Errorf("DateOfDeath = %q; want %q", got.DateOfDeath, expected.DateOfDeath)
		}
	})
}

// ============================================================================
// Regression tests for bugs previously found in the person package.
// These are intended to FAIL (with a clear message) whenever the underlying
// bug reappears, and PASS once the bug is fixed.
// ============================================================================

// TestCreateNoDatesReadsBackEmpty guards against storing a zero-value
// neo4j.Date (e.g. 0001-01-01) when a person is created without dates.
func TestCreateNoDatesReadsBackEmpty(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	t.Cleanup(func() { driver.Close(ctx) })

	id := uuid.New().String()
	_, gotID, err := CreateNewPerson(ctx, driver, NewPerson{
		Id:         id,
		PersonName: "No Dates Person " + id,
		Gender:     Male,
		Alive:      true,
	})
	if err != nil {
		t.Fatalf("failed to create person without dates: %v", err)
	}
	if gotID != id {
		t.Fatalf("expected id %q, got %q", id, gotID)
	}
	t.Cleanup(func() { _, _, _ = DeleteUser(ctx, driver, id) })

	person, err := GetPerson(ctx, driver, id)
	if err != nil {
		t.Fatalf("failed to read person back: %v", err)
	}
	if person == nil {
		t.Fatalf("expected a person record, got nil")
	}
	if person.DateOfBirth != "" {
		t.Errorf("expected empty DateOfBirth for a person created without a DOB, got %q", person.DateOfBirth)
	}
	if person.DateOfDeath != "" {
		t.Errorf("expected empty DateOfDeath for a person created without a DOD, got %q", person.DateOfDeath)
	}
}

// TestCreateReturnsErrorOnClosedDriver guards against create.go silently
// swallowing the error returned by neo4j.ExecuteQuery.
func TestCreateReturnsErrorOnClosedDriver(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	driver.Close(ctx)

	name, id, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  "Closed Driver Create",
		Gender:      Male,
		DateOfBirth: "01-01-1990",
		Alive:       true,
	})
	if err == nil {
		t.Errorf("expected an error when creating with a closed driver, got nil (name=%q id=%q)", name, id)
	}
}

// TestCreateWithSameIDDoesNotDuplicate guards against the MERGE statement
// matching on every property, which creates duplicate nodes for the same id.
func TestCreateWithSameIDDoesNotDuplicate(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	t.Cleanup(func() { driver.Close(ctx) })

	id := uuid.New().String()

	if _, gotID, err := CreateNewPerson(ctx, driver, NewPerson{
		Id:          id,
		PersonName:  "Duplicate One " + id,
		Gender:      Male,
		DateOfBirth: "01-01-1990",
		Alive:       false,
		DateOfDeath: "01-01-2000",
	}); err != nil {
		t.Fatalf("failed to create first person: %v", err)
	} else if gotID != id {
		t.Fatalf("expected id %q, got %q", id, gotID)
	}

	// Create a second person with the SAME id but different data.
	_, _, err := CreateNewPerson(ctx, driver, NewPerson{
		Id:          id,
		PersonName:  "Duplicate Two " + id,
		Gender:      Female,
		DateOfBirth: "02-02-1992",
		Alive:       true,
	})

	t.Cleanup(func() { _, _, _ = DeleteUser(ctx, driver, id) })

	result, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person {id: $id}) RETURN count(p) AS cnt`,
		map[string]any{"id": id},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("failed to count nodes with id %q: %v", id, err)
	}
	if len(result.Records) == 0 {
		t.Fatalf("no count record returned for id %q", id)
	}
	rawCnt, found := result.Records[0].Get("cnt")
	if !found {
		t.Fatalf("count field missing for id %q", id)
	}
	cnt, ok := rawCnt.(int64)
	if !ok {
		t.Fatalf("unexpected type for count: %T", rawCnt)
	}
	if cnt != 1 {
		t.Errorf("expected exactly 1 node for id %q, found %d (duplicate node created by MERGE)", id, cnt)
	}
}

// TestCreateRejectsAliveWithDeathDate guards the invariant that a person
// cannot be both alive and have a date of death.
func TestCreateRejectsAliveWithDeathDate(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	t.Cleanup(func() { driver.Close(ctx) })

	_, _, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  "Alive With Death Date",
		Gender:      Male,
		DateOfBirth: "01-01-1990",
		Alive:       true,
		DateOfDeath: "01-01-2000",
	})
	if err == nil {
		t.Errorf("expected an error when creating a person that is both alive and has a death date")
	}
}

// TestCreateRejectsImpossibleDates guards against regex-only date validation
// that accepts non-existent dates such as 31 February.
func TestCreateRejectsImpossibleDates(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	t.Cleanup(func() { driver.Close(ctx) })

	_, _, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  "Impossible Date",
		Gender:      Male,
		DateOfBirth: "31-02-2023",
		Alive:       true,
	})
	if err == nil {
		t.Errorf("expected an error for impossible date of birth 31-02-2023")
	}
}

// TestUpdatePre1970Dates guards against Unix-millisecond date comparison, which
// returns negative values for pre-1970 dates that the update code then
// misinterpreted as invalid.
func TestUpdatePre1970Dates(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	t.Cleanup(func() { driver.Close(ctx) })

	id := uuid.New().String()
	if _, _, err := CreateNewPerson(ctx, driver, NewPerson{
		Id:          id,
		PersonName:  "Pre 1970 " + id,
		Gender:      Male,
		DateOfBirth: "01-01-1900",
		Alive:       false,
		DateOfDeath: "01-01-1950",
	}); err != nil {
		t.Fatalf("failed to create pre-1970 person: %v", err)
	}
	t.Cleanup(func() { _, _, _ = DeleteUser(ctx, driver, id) })

	newDOB := DateProper("02-02-1901")
	if _, updated, err := UpdatePerson(ctx, driver, id, UpdateUser{DateOfBirth: &newDOB}); err != nil {
		t.Errorf("expected to be able to update a pre-1970 date, got error: %v", err)
	} else if !updated {
		t.Errorf("expected update to be reported as successful")
	}
}

// TestGetPersonMalformedDeathDateError guards against the copy-paste bug in
// query.go that reports an invalid date of death as "Invalid Date Of Birth".
func TestGetPersonMalformedDeathDateError(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	t.Cleanup(func() { driver.Close(ctx) })

	id := uuid.New().String()
	if _, _, err := CreateNewPerson(ctx, driver, NewPerson{
		Id:          id,
		PersonName:  "Bad DOD " + id,
		Gender:      Male,
		DateOfBirth: "01-01-1990",
		Alive:       false,
		DateOfDeath: "01-01-2000",
	}); err != nil {
		t.Fatalf("failed to create person: %v", err)
	}
	t.Cleanup(func() { _, _, _ = DeleteUser(ctx, driver, id) })

	// Corrupt the stored death date directly.
	if _, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person {id: $id}) SET p.date_of_death = $dod`,
		map[string]any{"id": id, "dod": "9999-99-99"},
		neo4j.EagerResultTransformer,
	); err != nil {
		t.Fatalf("failed to corrupt stored death date: %v", err)
	}

	_, err := GetPerson(ctx, driver, id)
	if err == nil {
		t.Fatalf("expected GetPerson to return an error for a malformed death date")
	}
	if !strings.Contains(err.Error(), "Death") {
		t.Errorf("expected error message to reference the date of death, got: %v", err)
	}
}

// TestGetPersonEmptyDateDoesNotPanic guards against query.go panicking with an
// index-out-of-range error when a stored date is an empty string.
func TestGetPersonEmptyDateDoesNotPanic(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	t.Cleanup(func() { driver.Close(ctx) })

	id := uuid.New().String()
	if _, _, err := CreateNewPerson(ctx, driver, NewPerson{
		Id:          id,
		PersonName:  "Empty Date " + id,
		Gender:      Male,
		DateOfBirth: "01-01-1990",
		Alive:       true,
	}); err != nil {
		t.Fatalf("failed to create person: %v", err)
	}
	t.Cleanup(func() { _, _, _ = DeleteUser(ctx, driver, id) })

	// Store an empty-string birth date directly (malformed state).
	if _, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person {id: $id}) SET p.date_of_birth = $dob`,
		map[string]any{"id": id, "dob": ""},
		neo4j.EagerResultTransformer,
	); err != nil {
		t.Fatalf("failed to corrupt stored birth date: %v", err)
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetPerson panicked on empty-string date: %v", r)
			}
		}()
		_, _ = GetPerson(ctx, driver, id)
	}()
}
