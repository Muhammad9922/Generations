package personRouter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type wirePerson struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TestSinglesEndpoint drives the route rather than the query, so it also proves
// that /people/singles is not swallowed by the /people/{id} wildcard.
//
// The probe family is created and removed here because the single/married split
// cannot be asserted against a shared database that the other packages' tests
// are writing to at the same time: a global "everyone left out has a marriage"
// check passes alone and fails in `go test ./...`.
func TestSinglesEndpoint(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	// Registered before the deletes below so it runs after them: t.Cleanup is
	// last-in-first-out, and a closed driver cannot delete the probe family.
	t.Cleanup(func() { driver.Close(ctx) })

	// A per-run suffix keeps repeated runs from colliding.
	stamp := time.Now().UnixNano()
	single := createProbePerson(t, ctx, driver, fmt.Sprintf("Singles Probe %d", stamp), person.Male)
	spouseOne := createProbePerson(t, ctx, driver, fmt.Sprintf("Married Probe One %d", stamp), person.Male)
	spouseTwo := createProbePerson(t, ctx, driver, fmt.Sprintf("Married Probe Two %d", stamp), person.Female)

	marriageID, err := marriage.CreateNewMarriage(ctx, driver, marriage.NewMarriage{
		SpouseOne: spouseOne,
		SpouseTwo: spouseTwo,
	})
	if err != nil {
		t.Fatalf("CreateNewMarriage failed: %v", err)
	}

	t.Cleanup(func() {
		if _, err := marriage.DeleteMarriage(ctx, driver, marriageID); err != nil {
			t.Errorf("Cleanup could not delete the probe marriage: %v", err)
		}
		for _, id := range []string{single, spouseOne, spouseTwo} {
			if _, _, err := person.DeleteUser(ctx, driver, id); err != nil {
				t.Errorf("Cleanup could not delete probe person %s: %v", id, err)
			}
		}
	})

	mux := http.NewServeMux()
	RegisterPersonRoutes(mux, driver)
	server := httptest.NewServer(mux)
	defer server.Close()

	status, singles := getSingles(t, server.URL)
	if status != http.StatusOK {
		t.Fatalf("GET /people/singles answered %d, want 200", status)
	}
	if len(singles) == 0 {
		t.Fatal("Expected at least one single person")
	}

	found := make(map[string]bool, len(singles))
	for _, person := range singles {
		found[person.ID] = true
	}

	if !found[single] {
		t.Errorf("The unmarried probe person is missing from the singles list")
	}
	if found[spouseOne] || found[spouseTwo] {
		t.Errorf("A married probe person appeared in the singles list")
	}
}

func createProbePerson(t *testing.T, ctx context.Context, driver neo4j.Driver, name string, gender person.Gender) string {
	t.Helper()

	_, id, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName: name,
		Gender:     gender,
		Alive:      true,
	})
	if err != nil {
		t.Fatalf("Could not create %s: %v", name, err)
	}
	if id == "" {
		t.Fatalf("CreateNewPerson returned no id for %s", name)
	}

	return id
}

func getSingles(t *testing.T, baseURL string) (int, []wirePerson) {
	t.Helper()

	response, err := http.Get(baseURL + "/people/singles")
	if err != nil {
		t.Fatalf("GET /people/singles failed: %v", err)
	}
	defer response.Body.Close()

	var body struct {
		People []wirePerson `json:"people"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("GET /people/singles did not answer JSON: %v", err)
	}

	return response.StatusCode, body.People
}
