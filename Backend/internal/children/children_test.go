package children

import (
	"context"
	"fmt"
	"testing"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

const testDatabaseURI = "bolt://192.168.0.133:7687"

// ---------------------------------------------------------------------------
// Test fixtures / helpers
// ---------------------------------------------------------------------------

// createTestPerson seeds a uniquely named person so that a re-run of the suite
// never collides with nodes left behind by earlier runs (person.CreateNewPerson
// MERGEs on every property, so identical fixtures would silently reuse a node).
// The person is intentionally left in the graph once the test finishes so the
// created data can be inspected afterwards.
func createTestPerson(t *testing.T, ctx context.Context, driver neo4j.Driver, gender person.Gender, label string) string {
	t.Helper()

	_, id, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  fmt.Sprintf("%s %s", label, uuid.New().String()),
		Gender:      gender,
		Alive:       true,
		DateOfBirth: "11-10-1990",
	})

	if err != nil {
		t.Fatalf("Failed To Create %s: %v", label, err)
	}

	t.Logf("%s ID: %v", label, id)

	return id
}

// createTestMarriage seeds a male/female couple and the marriage that binds
// them, returning the marriage ID together with both spouse IDs. Everything it
// creates is left in the graph after the test finishes.
func createTestMarriage(t *testing.T, ctx context.Context, driver neo4j.Driver) (string, string, string) {
	t.Helper()

	spouseOne := createTestPerson(t, ctx, driver, person.Male, "Spouse One")
	spouseTwo := createTestPerson(t, ctx, driver, person.Female, "Spouse Two")

	marriageID, err := marriage.CreateNewMarriage(ctx, driver, marriage.NewMarriage{
		SpouseOne: spouseOne,
		SpouseTwo: spouseTwo,
		DateStart: "11-10-2018",
	})

	if err != nil || marriageID == "" {
		t.Fatalf("Failed To Create Test Marriage: %v", err)
	}

	t.Logf("Test Marriage ID: %v", marriageID)

	return marriageID, spouseOne, spouseTwo
}

// countProducedRelationships counts the (Person)-[:PRODUCED]->(Marriage)
// relationships present in the graph. An empty childID or marriageID is treated
// as "any", allowing the same helper to answer "how many children does this
// marriage have?" and "how many marriages does this child belong to?".
func countProducedRelationships(t *testing.T, ctx context.Context, driver neo4j.Driver, childID string, marriageID string) int {
	t.Helper()

	const query = `
	MATCH (c:Person)<-[:PRODUCED]-(m:Marriage)
	WHERE ($cid = '' OR c.id = $cid) AND ($mid = '' OR m.id = $mid)
	RETURN count(*) AS total
	`

	result, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		query,
		map[string]any{
			"cid": childID,
			"mid": marriageID,
		},
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		t.Fatalf("Failed Counting PRODUCED Relationships: %v", err)
	}

	if len(result.Records) == 0 {
		return 0
	}

	rawTotal, found := result.Records[0].Get("total")
	if !found {
		t.Fatalf("No Count Was Returned By The PRODUCED Relationship Query")
	}

	total, ok := rawTotal.(int64)
	if !ok {
		t.Fatalf("Unexpected Type Returned For PRODUCED Count: %T", rawTotal)
	}

	return int(total)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestCreateChildren covers the happy path: a valid child attached to a valid
// marriage must report success, leave exactly one PRODUCED relationship behind
// and must not disturb either the child or the marriage.
func TestCreateChildren(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	marriageID, spouseOneID, spouseTwoID := createTestMarriage(t, ctx, driver)
	childID := createTestPerson(t, ctx, driver, person.Male, "Child")

	created, err := CreateNewChildren(ctx, driver, marriageID, childID)

	if err != nil {
		t.Fatalf("Expected Child Creation To Succeed, Got Error: %v", err)
	}

	if !created {
		t.Errorf("Expected The Child Creation To Be Reported As Successful")
	}

	if total := countProducedRelationships(t, ctx, driver, childID, marriageID); total != 1 {
		t.Errorf("Expected Exactly One PRODUCED Relationship, Found %v", total)
	}

	// The child itself must be untouched by the link.
	child, err := person.GetPerson(ctx, driver, childID)
	if err != nil || child == nil {
		t.Fatalf("Child Should Still Exist After Being Linked: %v", err)
	}

	if child.Id != childID {
		t.Errorf("Child ID Should Have Remained %v, Got %v", childID, child.Id)
	}

	// The marriage must be untouched as well.
	marriages, err := marriage.GetMarriageFromMarriageId(ctx, driver, marriageID)
	if err != nil || len(marriages) != 1 {
		t.Fatalf("Marriage Should Still Exist After Linking A Child: %v", err)
	}

	if marriages[0].Id != marriageID {
		t.Errorf("Marriage ID Should Not Change When A Child Is Linked: %v != %v", marriages[0].Id, marriageID)
	}

	if marriages[0].Start != "11-10-2018" {
		t.Errorf("Marriage Start Date Should Not Change When A Child Is Linked: %v", marriages[0].Start)
	}

	t.Logf("Linked Child %v Into Marriage %v Of %v & %v", childID, marriageID, spouseOneID, spouseTwoID)
}

// TestCreateMultipleChildren checks that a single marriage can receive several
// children and that every child keeps its own PRODUCED relationship.
func TestCreateMultipleChildren(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	marriageID, _, _ := createTestMarriage(t, ctx, driver)

	childIDs := []string{
		createTestPerson(t, ctx, driver, person.Male, "First Child"),
		createTestPerson(t, ctx, driver, person.Female, "Second Child"),
		createTestPerson(t, ctx, driver, person.Male, "Third Child"),
	}

	for index, childID := range childIDs {
		created, err := CreateNewChildren(ctx, driver, marriageID, childID)

		if err != nil {
			t.Fatalf("Child %v Should Have Been Linked Without Error: %v", index, err)
		}

		if !created {
			t.Errorf("Child %v Should Have Been Reported As Created", index)
		}
	}

	if total := countProducedRelationships(t, ctx, driver, "", marriageID); total != len(childIDs) {
		t.Errorf("Expected %v Children On The Marriage, Found %v", len(childIDs), total)
	}

	for index, childID := range childIDs {
		if total := countProducedRelationships(t, ctx, driver, childID, marriageID); total != 1 {
			t.Errorf("Child %v Should Hold Exactly One PRODUCED Relationship, Found %v", index, total)
		}
	}
}

// TestSameChildOnMultipleMarriages verifies that one person can be linked as a
// child of more than one marriage, and that each link stays independent.
func TestSameChildOnMultipleMarriages(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	firstMarriage, _, _ := createTestMarriage(t, ctx, driver)
	secondMarriage, _, _ := createTestMarriage(t, ctx, driver)
	childID := createTestPerson(t, ctx, driver, person.Female, "Shared Child")

	for _, marriageID := range []string{firstMarriage, secondMarriage} {
		created, err := CreateNewChildren(ctx, driver, marriageID, childID)

		if err != nil || !created {
			t.Fatalf("Child Should Have Been Linked To Marriage %v (created: %v, err: %v)", marriageID, created, err)
		}
	}

	if total := countProducedRelationships(t, ctx, driver, childID, firstMarriage); total != 1 {
		t.Errorf("Expected One Link To The First Marriage, Found %v", total)
	}

	if total := countProducedRelationships(t, ctx, driver, childID, secondMarriage); total != 1 {
		t.Errorf("Expected One Link To The Second Marriage, Found %v", total)
	}

	if total := countProducedRelationships(t, ctx, driver, childID, ""); total != 2 {
		t.Errorf("Expected The Child To Belong To Two Marriages, Found %v", total)
	}
}

// TestDuplicateChildLink pins down the current MERGE behaviour: re-linking an
// already linked child does not duplicate the relationship, but because MERGE
// reports zero created relationships the function still answers with an error
// ("No Children Were Created") even though the link is perfectly intact.
// If the function ever grows an idempotent "already linked" path, this test is
// the one that should be flipped to expect success instead.
func TestDuplicateChildLink(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	marriageID, _, _ := createTestMarriage(t, ctx, driver)
	childID := createTestPerson(t, ctx, driver, person.Female, "Only Child")

	if _, err := CreateNewChildren(ctx, driver, marriageID, childID); err != nil {
		t.Fatalf("First Link Should Have Succeeded, Got: %v", err)
	}

	repeatCreated, repeatErr := CreateNewChildren(ctx, driver, marriageID, childID)

	if repeatCreated {
		t.Errorf("Repeating An Existing Link Must Not Report A Newly Created Relationship")
	}

	if repeatErr == nil {
		t.Errorf("Expected The Current Implementation To Error On An Existing Link, Got None")
	} else {
		t.Logf("Repeat Link Reported (Expected With The Current Implementation): %v", repeatErr)
	}

	if total := countProducedRelationships(t, ctx, driver, childID, marriageID); total != 1 {
		t.Errorf("Expected The Duplicate Call To Leave A Single PRODUCED Relationship, Found %v", total)
	}
}

// TestCreateChildrenInvalidInput walks through every way the arguments can be
// wrong. CreateNewChildren must report "nothing was created" plus an error for
// all of them, and none of the attempts may leave a relationship behind.
func TestCreateChildrenInvalidInput(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	marriageID, _, _ := createTestMarriage(t, ctx, driver)
	childID := createTestPerson(t, ctx, driver, person.Male, "Valid Child")

	tests := []struct {
		name       string
		marriageID string
		childID    string
	}{
		{
			name:       "Non Existent Child",
			marriageID: marriageID,
			childID:    uuid.New().String(),
		},
		{
			name:       "Non Existent Marriage",
			marriageID: uuid.New().String(),
			childID:    childID,
		},
		{
			name:       "Non Existent Child And Marriage",
			marriageID: uuid.New().String(),
			childID:    uuid.New().String(),
		},
		{
			name:       "Empty Child ID",
			marriageID: marriageID,
			childID:    "",
		},
		{
			name:       "Empty Marriage ID",
			marriageID: "",
			childID:    childID,
		},
		{
			name:       "Empty Child And Marriage ID",
			marriageID: "",
			childID:    "",
		},
		{
			name:       "Marriage ID Used As Child ID",
			marriageID: marriageID,
			childID:    marriageID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			created, err := CreateNewChildren(ctx, driver, test.marriageID, test.childID)

			if err == nil {
				t.Errorf("Expected An Error For %q, Got None", test.name)
			} else {
				t.Logf("Rejected %q As Expected: %v", test.name, err)
			}

			if created {
				t.Errorf("%q Must Not Report A Created Relationship", test.name)
			}
		})
	}

	// None of the failed attempts may have touched the valid fixture pair.
	if total := countProducedRelationships(t, ctx, driver, childID, marriageID); total != 0 {
		t.Errorf("Invalid Attempts Should Not Have Linked The Fixtures, Found %v Relationship(s)", total)
	}
}

// TestCreateChildrenSpouseLinkedAsOwnChild pins down a gap in the current
// validation: CreateNewChildren only checks that the child and the marriage
// exist, so a spouse can be attached to their own marriage as a child of it.
// Once that guard is added, this test should be flipped to expect an error.
func TestCreateChildrenSpouseLinkedAsOwnChild(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	marriageID, spouseOneID, _ := createTestMarriage(t, ctx, driver)

	created, err := CreateNewChildren(ctx, driver, marriageID, spouseOneID)

	if err != nil {
		t.Fatalf("Current Implementation Should Allow This, Got Error: %v", err)
	}

	if !created {
		t.Errorf("Expected The Relationship To Be Reported As Created")
	}

	if total := countProducedRelationships(t, ctx, driver, spouseOneID, marriageID); total != 1 {
		t.Errorf("Expected The Spouse To Be Linked To Their Own Marriage Today, Found %v", total)
	}

	t.Log("Known Gap: A Spouse Was Accepted As A Child Of Their Own Marriage")
}

// TestCreateChildrenClosedDriver makes sure a broken connection is surfaced
// instead of being swallowed into a silent "nothing happened" response.
func TestCreateChildrenClosedDriver(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	driver.Close(ctx)

	created, err := CreateNewChildren(ctx, driver, uuid.New().String(), uuid.New().String())

	if err == nil {
		t.Errorf("A Closed Driver Should Have Caused An Error")
	}

	if created {
		t.Errorf("A Closed Driver Should Never Report A Created Relationship")
	}

	t.Logf("Closed Driver Error: %v", err)
}
