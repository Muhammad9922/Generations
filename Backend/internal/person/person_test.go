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
		{
			"No Gender Provided",
			NewPerson{
				PersonName:  "Person Name 1",
				DateOfBirth: "10-10-2010",
				Alive:       true,
			},
		},
		{
			"No Alive Status",
			NewPerson{
				PersonName:  "Person Name 2",
				DateOfBirth: "2023-11-10",
				Gender:      "Male",
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

// readPersonProperties fetches the stored properties of a Person node by its ID.
// It is a test helper used to verify that updates were actually persisted.
func readPersonProperties(t *testing.T, ctx context.Context, driver neo4j.Driver, id string) map[string]any {
	t.Helper()

	result, err := neo4j.ExecuteQuery(ctx, driver,
		`
		MATCH (p:Person {id: $id})
		RETURN p.name AS name, p.gender AS gender, p.date_of_birth AS date_of_birth, p.alive AS alive
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
	for _, key := range []string{"name", "gender", "date_of_birth", "alive"} {
		if value, found := record.Get(key); found {
			props[key] = value
		}
	}

	return props
}

func TestUpdatePerson(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
	defer driver.Close(ctx)

	// A UUID-backed name ensures this test does not collide with data left
	// behind by the other tests that share the same database.
	originalName := "Update Test Person " + uuid.New().String()
	originalGender := Male
	originalDOB := DateOfBirth("01-01-1990")
	originalAlive := true

	_, id, err := CreateNewPerson(ctx, driver, NewPerson{
		PersonName:  originalName,
		Gender:      originalGender,
		DateOfBirth: originalDOB,
		Alive:       originalAlive,
	})
	if err != nil {
		t.Fatalf("failed to create person for update test: %v", err)
	}
	if id == "" {
		t.Fatalf("expected a non-empty id when creating a person")
	}

	updatedName := "Updated Person Name"
	updatedGender := Female
	updatedDOB := DateOfBirth("02-02-1992")
	updatedAlive := false

	t.Run("Update All Fields", func(t *testing.T) {
		name, wasUpdated, err := UpdatePerson(ctx, driver, id, UpdateUser{
			Name:        &updatedName,
			Gender:      &updatedGender,
			DateOfBirth: &updatedDOB,
			Alive:       &updatedAlive,
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

	updatedDOBOnly := DateOfBirth("03-03-1993")
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
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")
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

func TestPersonQuery(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://localhost:7687")

	tests := []struct {
		name   string
		person NewPerson
	}{
		{
			"No Date Of Birth Status",
			NewPerson{
				id:         uuid.New().String(),
				PersonName: "Person Name 3",
				Gender:     "Male",
				Alive:      true,
			},
		},
		{
			"Full Account",
			NewPerson{
				id:          uuid.New().String(),
				PersonName:  "Person Name 4",
				Gender:      "Male",
				Alive:       false,
				DateOfBirth: "11-10-2005",
			},
		},
	}

	for _, test := range tests {
		t.Run("Testing Creation Of "+test.name, func(t *testing.T) {
			personName, personID, err := CreateNewPerson(ctx, driver, test.person)

			if err != nil {
				t.Errorf("Error Creating New User ID %s and User Name %s", test.person.id, test.person.PersonName)
			}

			if personName == "" || personID == "" {
				t.Errorf("Error Creating New User ID %s and User Name %s", test.person.id, test.person.PersonName)
			}
		})

		t.Run("Testing Query "+test.name, func(t *testing.T) {
			person, err := GetPerson(ctx, driver, test.person.id)
			if err != nil || person == nil {
				t.Errorf("Error While Querying User %v", err)
			}
			if len(strings.Split(person.PersonName, "")) < 4 {
				t.Errorf("Invalid User Name: %s", person.PersonName)
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
		_, err := GetPerson(ctx, driver, tests[0].person.id)
		if err == nil {
			t.Errorf("A Non Existant ID Should Give An Error")
		}
	})

}
