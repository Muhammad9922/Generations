package integration

import (
	"context"
	"fmt"
	"testing"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/children"
	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/davecgh/go-spew/spew"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

const testDatabaseURI = "bolt://192.168.0.133:7687"

// ===========================================================================
// Direct Neo4j DB Structure Verification & Fixture Helpers
// ===========================================================================

func verifyPersonInDB(t *testing.T, ctx context.Context, driver neo4j.Driver, id string, expected person.NewPerson) {
	t.Helper()
	res, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person {id: $id})
		 RETURN p.name AS name, p.gender AS gender, p.alive AS alive, p.date_of_birth AS dob, p.date_of_death AS dod`,
		map[string]any{"id": id},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("verifyPersonInDB query failed for %s: %v", id, err)
	}
	if len(res.Records) == 0 {
		t.Fatalf("Expected Person node with id %s in DB, but found none", id)
	}
	rec := res.Records[0]
	nameVal, _ := rec.Get("name")
	if nameVal != expected.PersonName {
		t.Errorf("Person name mismatch: expected %q, got %v", expected.PersonName, nameVal)
	}
	genderVal, _ := rec.Get("gender")
	if genderVal != string(expected.Gender) {
		t.Errorf("Person gender mismatch: expected %q, got %v", expected.Gender, genderVal)
	}
	aliveVal, _ := rec.Get("alive")
	if aliveVal != expected.Alive {
		t.Errorf("Person alive mismatch: expected %v, got %v", expected.Alive, aliveVal)
	}
	if expected.DateOfBirth != "" {
		dobVal, _ := rec.Get("dob")
		actualDOB := person.GetProperDate(dobVal)
		if actualDOB != expected.DateOfBirth {
			t.Errorf("Person DOB mismatch: expected %v, got %v", expected.DateOfBirth, actualDOB)
		}
	}
	if expected.DateOfDeath != "" {
		dodVal, _ := rec.Get("dod")
		actualDOD := person.GetProperDate(dodVal)
		if actualDOD != expected.DateOfDeath {
			t.Errorf("Person DOD mismatch: expected %v, got %v", expected.DateOfDeath, actualDOD)
		}
	}
}

func verifyMarriageInDB(t *testing.T, ctx context.Context, driver neo4j.Driver, marriageID, spouseOneID, spouseTwoID string, expectedStart, expectedEnd person.DateProper) {
	t.Helper()
	res, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (s1:Person {id: $s1})-[:MARRIED]->(m:Marriage {id: $mid})<-[:MARRIED]-(s2:Person {id: $s2})
		 RETURN m.id AS id, m.start AS start, m.end AS end`,
		map[string]any{
			"mid": marriageID,
			"s1":  spouseOneID,
			"s2":  spouseTwoID,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("verifyMarriageInDB query error: %v", err)
	}
	if len(res.Records) == 0 {
		resRev, errRev := neo4j.ExecuteQuery(ctx, driver,
			`MATCH (s2:Person {id: $s2})-[:MARRIED]->(m:Marriage {id: $mid})<-[:MARRIED]-(s1:Person {id: $s1})
			 RETURN m.id AS id, m.start AS start, m.end AS end`,
			map[string]any{
				"mid": marriageID,
				"s1":  spouseOneID,
				"s2":  spouseTwoID,
			},
			neo4j.EagerResultTransformer,
		)
		if errRev != nil || len(resRev.Records) == 0 {
			t.Fatalf("Marriage %s between %s and %s not found in DB with bidirectional [:MARRIED] relationships", marriageID, spouseOneID, spouseTwoID)
		}
		res = resRev
	}
	rec := res.Records[0]
	if expectedStart != "" {
		rawStart, _ := rec.Get("start")
		actualStart := person.GetProperDate(rawStart)
		if actualStart != expectedStart {
			t.Errorf("Marriage start mismatch: expected %v, got %v", expectedStart, actualStart)
		}
	}
	if expectedEnd != "" {
		rawEnd, _ := rec.Get("end")
		actualEnd := person.GetProperDate(rawEnd)
		if actualEnd != expectedEnd {
			t.Errorf("Marriage end mismatch: expected %v, got %v", expectedEnd, actualEnd)
		}
	}
}

func verifyChildRelationshipInDB(t *testing.T, ctx context.Context, driver neo4j.Driver, marriageID, childID string) {
	t.Helper()
	res, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (m:Marriage {id: $mid})-[:PRODUCED]->(c:Person {id: $cid})
		 RETURN count(*) AS total`,
		map[string]any{
			"mid": marriageID,
			"cid": childID,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("verifyChildRelationshipInDB query failed: %v", err)
	}
	rawTotal, found := res.Records[0].Get("total")
	if !found || rawTotal.(int64) != 1 {
		t.Fatalf("Expected exactly 1 [:PRODUCED] relationship from marriage %s to child %s, got %v", marriageID, childID, rawTotal)
	}
}

func countProducedRelationshipsForMarriage(t *testing.T, ctx context.Context, driver neo4j.Driver, marriageID string) int64 {
	t.Helper()
	res, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (m:Marriage {id: $mid})-[:PRODUCED]->(:Person)
		 RETURN count(*) AS total`,
		map[string]any{"mid": marriageID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("countProducedRelationshipsForMarriage failed: %v", err)
	}
	rawTotal, _ := res.Records[0].Get("total")
	return rawTotal.(int64)
}

func verifyNodeDoesNotExist(t *testing.T, ctx context.Context, driver neo4j.Driver, id string) {
	t.Helper()
	if id == "" {
		return
	}
	res, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (n {id: $id}) RETURN count(n) AS total`,
		map[string]any{"id": id},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("verifyNodeDoesNotExist failed query: %v", err)
	}
	rawTotal, _ := res.Records[0].Get("total")
	if rawTotal.(int64) != 0 {
		t.Fatalf("Expected node %s to NOT exist in DB, but found %v matching nodes", id, rawTotal)
	}
}

func teardownTestNodes(ctx context.Context, driver neo4j.Driver, marriageID string, personIDs ...string) {
	if marriageID != "" {
		marriage.DeleteMarriage(ctx, driver, marriageID)
	}
	for _, pid := range personIDs {
		if pid != "" {
			person.DeleteUser(ctx, driver, pid)
		}
	}
}

// ===========================================================================
// Happy Path Tests
// ===========================================================================

func TestCreateFamily_AllNewEntities_Success(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("Father %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("Mother %s", tag),
		Gender:      person.Female,
		DateOfBirth: "05-05-1982",
		Alive:       true,
	}
	child1 := person.NewPerson{
		PersonName:  fmt.Sprintf("Child One %s", tag),
		Gender:      person.Male,
		DateOfBirth: "10-10-2010",
		Alive:       true,
	}
	child2 := person.NewPerson{
		PersonName:  fmt.Sprintf("Child Two %s", tag),
		Gender:      person.Female,
		DateOfBirth: "12-12-2012",
		Alive:       true,
	}

	marriageParams := MarriageOptionalParams{
		DateStart: "15-06-2005",
		DateEnd:   "20-08-2020",
	}

	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, []person.NewPerson{child1, child2}, marriageParams)
	if err != nil {
		t.Fatalf("CreateFamily failed: %v", err)
	}

	defer teardownTestNodes(ctx, driver, family.MarriageID, append([]string{family.SpouseOneId, family.SpouseTwoId}, family.ChildrenIds...)...)

	if family.MarriageID == "" {
		t.Errorf("Expected non-empty MarriageID")
	}
	if family.SpouseOneId == "" || family.SpouseTwoId == "" {
		t.Errorf("Expected non-empty spouse IDs")
	}
	if len(family.ChildrenIds) != 2 {
		t.Fatalf("Expected 2 children IDs, got %d", len(family.ChildrenIds))
	}

	// Direct DB checks
	verifyPersonInDB(t, ctx, driver, family.SpouseOneId, spouseOne)
	verifyPersonInDB(t, ctx, driver, family.SpouseTwoId, spouseTwo)
	verifyPersonInDB(t, ctx, driver, family.ChildrenIds[0], child1)
	verifyPersonInDB(t, ctx, driver, family.ChildrenIds[1], child2)

	verifyMarriageInDB(t, ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId, "15-06-2005", "20-08-2020")
	verifyChildRelationshipInDB(t, ctx, driver, family.MarriageID, family.ChildrenIds[0])
	verifyChildRelationshipInDB(t, ctx, driver, family.MarriageID, family.ChildrenIds[1])
}

func TestCreateFamily_NoChildren_Success(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("Husband %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1990",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("Wife %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1992",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, []person.NewPerson{}, MarriageOptionalParams{})
	if err != nil {
		t.Fatalf("CreateFamily failed without children: %v", err)
	}

	defer teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)

	if len(family.ChildrenIds) != 0 {
		t.Errorf("Expected 0 children IDs, got %d", len(family.ChildrenIds))
	}

	verifyPersonInDB(t, ctx, driver, family.SpouseOneId, spouseOne)
	verifyPersonInDB(t, ctx, driver, family.SpouseTwoId, spouseTwo)
	verifyMarriageInDB(t, ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId, "", "")

	childCount := countProducedRelationshipsForMarriage(t, ctx, driver, family.MarriageID)
	if childCount != 0 {
		t.Errorf("Expected 0 PRODUCED relationships, found %d", childCount)
	}
}

func TestCreateFamily_ExistingSpouses_Success(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	s1Input := person.NewPerson{
		PersonName:  fmt.Sprintf("PreSpouseOne %s", tag),
		Gender:      person.Male,
		DateOfBirth: "10-10-1985",
		Alive:       true,
	}
	s2Input := person.NewPerson{
		PersonName:  fmt.Sprintf("PreSpouseTwo %s", tag),
		Gender:      person.Female,
		DateOfBirth: "12-12-1987",
		Alive:       true,
	}

	_, s1ID, err := person.CreateNewPerson(ctx, driver, s1Input)
	if err != nil {
		t.Fatalf("Failed pre-creating spouse 1: %v", err)
	}
	_, s2ID, err := person.CreateNewPerson(ctx, driver, s2Input)
	if err != nil {
		person.DeleteUser(ctx, driver, s1ID)
		t.Fatalf("Failed pre-creating spouse 2: %v", err)
	}

	s1Input.Id = s1ID
	s2Input.Id = s2ID

	child := person.NewPerson{
		PersonName:  fmt.Sprintf("Child %s", tag),
		Gender:      person.Female,
		DateOfBirth: "15-05-2015",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, s1Input, s2Input, []person.NewPerson{child}, MarriageOptionalParams{
		DateStart: "10-10-2010",
	})
	if err != nil {
		teardownTestNodes(ctx, driver, "", s1ID, s2ID)
		t.Fatalf("CreateFamily failed with existing spouses: %v", err)
	}

	defer teardownTestNodes(ctx, driver, family.MarriageID, append([]string{s1ID, s2ID}, family.ChildrenIds...)...)

	if family.SpouseOneId != s1ID {
		t.Errorf("Expected SpouseOneId %s, got %s", s1ID, family.SpouseOneId)
	}
	if family.SpouseTwoId != s2ID {
		t.Errorf("Expected SpouseTwoId %s, got %s", s2ID, family.SpouseTwoId)
	}

	verifyMarriageInDB(t, ctx, driver, family.MarriageID, s1ID, s2ID, "10-10-2010", "")
	verifyChildRelationshipInDB(t, ctx, driver, family.MarriageID, family.ChildrenIds[0])
}

func TestCreateFamily_PartiallyExistingSpouses_Success(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	s1Input := person.NewPerson{
		PersonName:  fmt.Sprintf("PreSpouseOnly %s", tag),
		Gender:      person.Male,
		DateOfBirth: "03-03-1988",
		Alive:       true,
	}
	_, s1ID, err := person.CreateNewPerson(ctx, driver, s1Input)
	if err != nil {
		t.Fatalf("Failed pre-creating spouse 1: %v", err)
	}
	s1Input.Id = s1ID

	s2New := person.NewPerson{
		PersonName:  fmt.Sprintf("BrandNewSpouseTwo %s", tag),
		Gender:      person.Female,
		DateOfBirth: "04-04-1990",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, s1Input, s2New, nil, MarriageOptionalParams{})
	if err != nil {
		teardownTestNodes(ctx, driver, "", s1ID)
		t.Fatalf("CreateFamily failed with partially existing spouses: %v", err)
	}

	defer teardownTestNodes(ctx, driver, family.MarriageID, s1ID, family.SpouseTwoId)

	if family.SpouseOneId != s1ID {
		t.Errorf("Expected SpouseOneId to be %s, got %s", s1ID, family.SpouseOneId)
	}
	if family.SpouseTwoId == "" || family.SpouseTwoId == s1ID {
		t.Errorf("Expected valid new SpouseTwoId, got %s", family.SpouseTwoId)
	}

	verifyPersonInDB(t, ctx, driver, family.SpouseTwoId, s2New)
	verifyMarriageInDB(t, ctx, driver, family.MarriageID, s1ID, family.SpouseTwoId, "", "")
}

func TestCreateFamily_ExistingAndMixedChildren_Success(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	// Pre-create child 1
	c1Input := person.NewPerson{
		PersonName:  fmt.Sprintf("PreChildOne %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-2012",
		Alive:       true,
	}
	_, c1ID, err := person.CreateNewPerson(ctx, driver, c1Input)
	if err != nil {
		t.Fatalf("Failed pre-creating child 1: %v", err)
	}
	c1Input.Id = c1ID

	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("ParentOne %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1985",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("ParentTwo %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1987",
		Alive:       true,
	}
	c2New := person.NewPerson{
		PersonName:  fmt.Sprintf("NewChildTwo %s", tag),
		Gender:      person.Female,
		DateOfBirth: "05-05-2016",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, []person.NewPerson{c1Input, c2New}, MarriageOptionalParams{})
	if err != nil {
		person.DeleteUser(ctx, driver, c1ID)
		t.Fatalf("CreateFamily failed with mixed children: %v", err)
	}

	defer teardownTestNodes(ctx, driver, family.MarriageID, append([]string{family.SpouseOneId, family.SpouseTwoId, c1ID}, family.ChildrenIds...)...)

	if len(family.ChildrenIds) != 2 {
		t.Fatalf("Expected 2 children IDs, got %d", len(family.ChildrenIds))
	}
	if family.ChildrenIds[0] != c1ID {
		t.Errorf("Expected first child to be pre-existing ID %s, got %s", c1ID, family.ChildrenIds[0])
	}

	verifyPersonInDB(t, ctx, driver, c1ID, c1Input)
	verifyPersonInDB(t, ctx, driver, family.ChildrenIds[1], c2New)
	verifyChildRelationshipInDB(t, ctx, driver, family.MarriageID, c1ID)
	verifyChildRelationshipInDB(t, ctx, driver, family.MarriageID, family.ChildrenIds[1])
}

func TestCreateFamily_AttachToExistingMarriage_Success(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	s1Input := person.NewPerson{
		PersonName:  fmt.Sprintf("OrigSpouseOne %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	s2Input := person.NewPerson{
		PersonName:  fmt.Sprintf("OrigSpouseTwo %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}
	_, s1ID, _ := person.CreateNewPerson(ctx, driver, s1Input)
	_, s2ID, _ := person.CreateNewPerson(ctx, driver, s2Input)
	s1Input.Id = s1ID
	s2Input.Id = s2ID

	existingMID, err := marriage.CreateNewMarriage(ctx, driver, marriage.NewMarriage{
		SpouseOne: s1ID,
		SpouseTwo: s2ID,
		DateStart: "01-01-2005",
		DateEnd:   "01-01-2025",
	})
	if err != nil {
		teardownTestNodes(ctx, driver, "", s1ID, s2ID)
		t.Fatalf("Failed creating test marriage: %v", err)
	}

	child := person.NewPerson{
		PersonName:  fmt.Sprintf("ChildOfExistingMarriage %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-2010",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, s1Input, s2Input, []person.NewPerson{child}, MarriageOptionalParams{
		Id: existingMID,
	})
	if err != nil {
		teardownTestNodes(ctx, driver, existingMID, s1ID, s2ID)
		t.Fatalf("CreateFamily failed when attaching to existing marriage: %v", err)
	}

	defer teardownTestNodes(ctx, driver, existingMID, append([]string{s1ID, s2ID}, family.ChildrenIds...)...)

	if family.MarriageID != existingMID {
		t.Errorf("Expected returned MarriageID %s, got %s", existingMID, family.MarriageID)
	}

	verifyChildRelationshipInDB(t, ctx, driver, existingMID, family.ChildrenIds[0])
}

func TestCreateFamily_AttachToExistingMarriage_ReversedSpouses_Success(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	s1Input := person.NewPerson{
		PersonName:  fmt.Sprintf("ReverseSpouseOne %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	s2Input := person.NewPerson{
		PersonName:  fmt.Sprintf("ReverseSpouseTwo %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}
	_, s1ID, _ := person.CreateNewPerson(ctx, driver, s1Input)
	_, s2ID, _ := person.CreateNewPerson(ctx, driver, s2Input)
	s1Input.Id = s1ID
	s2Input.Id = s2ID

	existingMID, err := marriage.CreateNewMarriage(ctx, driver, marriage.NewMarriage{
		SpouseOne: s1ID,
		SpouseTwo: s2ID,
		DateStart: "01-01-2005",
	})
	if err != nil {
		teardownTestNodes(ctx, driver, "", s1ID, s2ID)
		t.Fatalf("Failed creating test marriage: %v", err)
	}

	child := person.NewPerson{
		PersonName:  fmt.Sprintf("ReverseChild %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-2012",
		Alive:       true,
	}

	// Pass s2 as spouseOne and s1 as spouseTwo
	family, err := CreateFamily(ctx, driver, s2Input, s1Input, []person.NewPerson{child}, MarriageOptionalParams{
		Id: existingMID,
	})
	if err != nil {
		teardownTestNodes(ctx, driver, existingMID, s1ID, s2ID)
		t.Fatalf("CreateFamily failed with reversed spouses on existing marriage: %v", err)
	}

	defer teardownTestNodes(ctx, driver, existingMID, append([]string{s1ID, s2ID}, family.ChildrenIds...)...)

	if family.MarriageID != existingMID {
		t.Errorf("Expected MarriageID %s, got %s", existingMID, family.MarriageID)
	}
	verifyChildRelationshipInDB(t, ctx, driver, existingMID, family.ChildrenIds[0])
}

// ===========================================================================
// Validation & Failure Tests (Data Validation & Direct DB Rollback Verification)
// ===========================================================================

func TestCreateFamily_Validation_InvalidSpouseOne_Error(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	// Missing PersonName
	invalidSpouseOne := person.NewPerson{
		PersonName:  "",
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("Mother %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, invalidSpouseOne, spouseTwo, nil, MarriageOptionalParams{})
	if err == nil {
		if family != nil {
			teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)
		}
		t.Fatalf("Expected error due to empty spouseOne name, but got nil")
	}
	if family != nil {
		t.Errorf("Expected family to be nil on error, got %v", family)
	}
}

func TestCreateFamily_Validation_InvalidSpouseTwo_Rollback(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("FatherToRollback %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	// Invalid gender for spouse 2
	invalidSpouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("MotherInvalid %s", tag),
		Gender:      person.Gender("UnknownGender"),
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, spouseOne, invalidSpouseTwo, nil, MarriageOptionalParams{})
	if err == nil {
		if family != nil {
			teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)
		}
		t.Fatalf("Expected error due to invalid spouseTwo gender, but got nil")
	}

	// Direct DB check: spouseOne must have been cleaned up and deleted from DB
	res, errQuery := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person {name: $name}) RETURN p.id AS id`,
		map[string]any{"name": spouseOne.PersonName},
		neo4j.EagerResultTransformer,
	)
	if errQuery != nil {
		t.Fatalf("Query failed: %v", errQuery)
	}
	if len(res.Records) > 0 {
		rolledBackID, _ := res.Records[0].Get("id")
		person.DeleteUser(ctx, driver, rolledBackID.(string))
		t.Fatalf("SpouseOne was NOT cleaned up from DB after spouseTwo validation failure")
	}
}

func TestCreateFamily_Validation_SameGender_Rollback(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	// Both spouses Male
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("MaleOne %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("MaleTwo %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, nil, MarriageOptionalParams{})
	if err == nil {
		if family != nil {
			teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)
		}
		t.Fatalf("Expected error for same-sex marriage creation, got nil")
	}

	// Verify both newly created spouses were rolled back
	res, _ := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person) WHERE p.name IN [$n1, $n2] RETURN count(p) AS total`,
		map[string]any{"n1": spouseOne.PersonName, "n2": spouseTwo.PersonName},
		neo4j.EagerResultTransformer,
	)
	total, _ := res.Records[0].Get("total")
	if total.(int64) != 0 {
		t.Errorf("Expected both spouses to be deleted upon gender conflict, but found %d", total.(int64))
	}
}

func TestCreateFamily_Validation_MarriageBeforeBirth_Rollback(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	// Spouse one birth in 2010, marriage in 2000
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("LateBornFather %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-2010",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("NormalMother %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1985",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, nil, MarriageOptionalParams{
		DateStart: "01-01-2000",
	})
	if err == nil {
		if family != nil {
			teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)
		}
		t.Fatalf("Expected error when marriage date is before birth date, got nil")
	}

	// Verify both spouses were rolled back
	res, _ := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person) WHERE p.name IN [$n1, $n2] RETURN count(p) AS total`,
		map[string]any{"n1": spouseOne.PersonName, "n2": spouseTwo.PersonName},
		neo4j.EagerResultTransformer,
	)
	total, _ := res.Records[0].Get("total")
	if total.(int64) != 0 {
		t.Errorf("Expected both spouses to be deleted upon date conflict, but found %d", total.(int64))
	}
}

func TestCreateFamily_Validation_MarriageEndAfterDeath_Rollback(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	// Spouse one died in 2015, marriage end in 2020
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("DeceasedFather %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1970",
		DateOfDeath: "01-01-2015",
		Alive:       false,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("LivingMother %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1975",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, nil, MarriageOptionalParams{
		DateStart: "01-01-1995",
		DateEnd:   "01-01-2020",
	})
	if err == nil {
		if family != nil {
			teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)
		}
		t.Fatalf("Expected error when marriage end is after spouse death date, got nil")
	}

	// Verify both spouses were rolled back
	res, _ := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person) WHERE p.name IN [$n1, $n2] RETURN count(p) AS total`,
		map[string]any{"n1": spouseOne.PersonName, "n2": spouseTwo.PersonName},
		neo4j.EagerResultTransformer,
	)
	total, _ := res.Records[0].Get("total")
	if total.(int64) != 0 {
		t.Errorf("Expected both spouses to be deleted upon death date conflict, but found %d", total.(int64))
	}
}

func TestCreateFamily_Validation_NonExistentExistingMarriage_Rollback(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("GhostHusband %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("GhostWife %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}

	invalidMarriageID := uuid.New().String()
	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, nil, MarriageOptionalParams{
		Id: invalidMarriageID,
	})
	if err == nil {
		if family != nil {
			teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)
		}
		t.Fatalf("Expected error for non-existent marriage ID, got nil")
	}

	// Verify both newly created spouses were rolled back
	res, _ := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person) WHERE p.name IN [$n1, $n2] RETURN count(p) AS total`,
		map[string]any{"n1": spouseOne.PersonName, "n2": spouseTwo.PersonName},
		neo4j.EagerResultTransformer,
	)
	total, _ := res.Records[0].Get("total")
	if total.(int64) != 0 {
		t.Errorf("Expected both spouses to be deleted when marriage ID is not found, but found %d", total.(int64))
	}
}

func TestCreateFamily_Validation_MismatchedSpousesForExistingMarriage_Rollback(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	s1Input := person.NewPerson{
		PersonName:  fmt.Sprintf("RealSpouseOne %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	s2Input := person.NewPerson{
		PersonName:  fmt.Sprintf("RealSpouseTwo %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}
	_, s1ID, _ := person.CreateNewPerson(ctx, driver, s1Input)
	_, s2ID, _ := person.CreateNewPerson(ctx, driver, s2Input)

	existingMID, err := marriage.CreateNewMarriage(ctx, driver, marriage.NewMarriage{
		SpouseOne: s1ID,
		SpouseTwo: s2ID,
		DateStart: "01-01-2005",
	})
	if err != nil {
		teardownTestNodes(ctx, driver, "", s1ID, s2ID)
		t.Fatalf("Failed creating marriage: %v", err)
	}
	defer teardownTestNodes(ctx, driver, existingMID, s1ID, s2ID)

	s1Input.Id = s1ID
	// Fake spouse 3 (newly created)
	fakeSpouse := person.NewPerson{
		PersonName:  fmt.Sprintf("FakeSpouseThree %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1985",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, s1Input, fakeSpouse, nil, MarriageOptionalParams{
		Id: existingMID,
	})

	if err == nil || family != nil {
		t.Fatalf("Expected error when providing mismatched spouse for existing marriage, got nil")
	}

	// Verify newly created fakeSpouse was deleted during cleanup
	res, _ := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person {name: $name}) RETURN count(p) AS total`,
		map[string]any{"name": fakeSpouse.PersonName},
		neo4j.EagerResultTransformer,
	)
	total, _ := res.Records[0].Get("total")
	if total.(int64) != 0 {
		t.Errorf("Expected fakeSpouse to be rolled back, but found in DB")
	}
}

func TestCreateFamily_Validation_InvalidChild_Rollback(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	spouseOne := person.NewPerson{
		PersonName:  fmt.Sprintf("FatherRollbackChild %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-1980",
		Alive:       true,
	}
	spouseTwo := person.NewPerson{
		PersonName:  fmt.Sprintf("MotherRollbackChild %s", tag),
		Gender:      person.Female,
		DateOfBirth: "01-01-1982",
		Alive:       true,
	}

	// Child 1 is valid, Child 2 is invalid (empty name)
	validChild := person.NewPerson{
		PersonName:  fmt.Sprintf("GoodChild %s", tag),
		Gender:      person.Male,
		DateOfBirth: "01-01-2010",
		Alive:       true,
	}
	invalidChild := person.NewPerson{
		PersonName:  "", // invalid!
		Gender:      person.Female,
		DateOfBirth: "01-01-2012",
		Alive:       true,
	}

	family, err := CreateFamily(ctx, driver, spouseOne, spouseTwo, []person.NewPerson{validChild, invalidChild}, MarriageOptionalParams{})
	if err == nil {
		if family != nil {
			teardownTestNodes(ctx, driver, family.MarriageID, family.SpouseOneId, family.SpouseTwoId)
		}
		t.Fatalf("Expected error when one child is invalid, got nil")
	}

	// Verify all newly created people (spouse 1, spouse 2, child 1) were cleaned up and rolled back
	res, _ := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person) WHERE p.name IN [$n1, $n2, $n3] RETURN count(p) AS total`,
		map[string]any{
			"n1": spouseOne.PersonName,
			"n2": spouseTwo.PersonName,
			"n3": validChild.PersonName,
		},
		neo4j.EagerResultTransformer,
	)
	total, _ := res.Records[0].Get("total")
	if total.(int64) != 0 {
		t.Errorf("Expected all created persons to be rolled back on child failure, but %d remain", total.(int64))
	}
}

func TestGetFamilyCertificate(t *testing.T) {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	// 1. Setup: Create Spouse One
	_, spouseOneId, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse One",
		Alive:       true,
		Gender:      person.Male,
		DateOfBirth: "11-10-1990",
	})
	if err != nil {
		t.Fatalf("Error Creating Spouse One: %v", err)
	}

	// 2. Setup: Create Spouse Two
	_, spouseTwoId, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Spouse Two",
		Alive:       true,
		Gender:      person.Female,
		DateOfBirth: "15-05-1992",
	})
	if err != nil {
		t.Fatalf("Error Creating Spouse Two: %v", err)
	}

	// 3. Setup: Create the Marriage
	marriageId, err := marriage.CreateNewMarriage(ctx, driver, marriage.NewMarriage{
		SpouseOne: spouseOneId,
		SpouseTwo: spouseTwoId,
		DateStart: "10-10-2015",
	})
	if err != nil || marriageId == "" {
		t.Fatalf("Error Creating New Marriage: %v", err)
	}

	// 4. Setup: Create a Child
	_, childId, err := person.CreateNewPerson(ctx, driver, person.NewPerson{
		PersonName:  "Child One",
		Alive:       true,
		Gender:      person.Male,
		DateOfBirth: "01-01-2018",
	})
	if err != nil {
		t.Fatalf("Error Creating Child: %v", err)
	}

	// 5. Setup: Link the Child to the Marriage
	// Assuming CreateNewChild is inside the children package based on your main code
	created, err := children.CreateNewChild(ctx, driver, marriageId, childId)
	if err != nil {
		t.Fatalf("Expected Child Creation To Succeed, Got Error: %v", err)
	}
	if !created {
		t.Fatalf("Expected The Child Linking To Be Reported As Successful")
	}

	// 6. Test: Retrieve the Family Certificate
	cert, err := GetFamilyCertificate(ctx, driver, spouseOneId)
	if err != nil {
		t.Fatalf("Failed to GetFamilyCertificate: %v", err)
	}

	// 7. Assertions
	if cert == nil {
		t.Fatalf("Expected a valid FamilyCertificate, but got nil")
	}

	if cert.Id != marriageId {
		t.Errorf("Expected Marriage ID %q, but got %q", marriageId, cert.Id)
	}

	// Verify Spouses
	if cert.Spouse == nil || len(*cert.Spouse) != 2 {
		t.Fatalf("Expected exactly 2 spouses, got nil or wrong length")
	}

	spouses := *cert.Spouse
	hasSpouseOne := false
	hasSpouseTwo := false
	for _, s := range spouses {
		if s.Id == spouseOneId {
			hasSpouseOne = true
		}
		if s.Id == spouseTwoId {
			hasSpouseTwo = true
		}
	}
	if !hasSpouseOne || !hasSpouseTwo {
		t.Errorf("Missing expected spouse IDs in certificate. Got IDs: %v, %v", spouses[0].Id, spouses[1].Id)
	}

	// Verify Children (Using your exact struct spelling 'Chidren')
	if cert.Chidren == nil {
		t.Fatalf("Expected children slice to be initialized, but got nil")
	}

	childrenList := *cert.Chidren
	if len(childrenList) != 1 {
		t.Fatalf("Expected exactly 1 child in the certificate, got %d", len(childrenList))
	}

	if childrenList[0].Id != childId {
		t.Errorf("Expected Child ID %q in certificate, but got %q", childId, childrenList[0].Id)
	}

	printedSpew := spew.Sdump(cert)
	t.Logf("CERT: %v", printedSpew)
}
