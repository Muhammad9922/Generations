package integration

import (
	"fmt"
	"testing"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func TestCreateComplexFamily_MuhammadNazir(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	var mToClean, pToClean []string
	defer func() {
		return
		for _, m := range mToClean {
			teardownTestNodes(ctx, driver, m)
		}
		for _, p := range pToClean {
			teardownTestNodes(ctx, driver, "", p)
		}
	}()

	nP := func(name string, g person.Gender, dob string) person.NewPerson {
		return person.NewPerson{
			PersonName:  fmt.Sprintf("%s %s", name, tag),
			Gender:      g,
			DateOfBirth: person.DateProper(dob),
			Alive:       true,
		}
	}

	nazir := nP("Muhammad Nazir", person.Male, "01-01-1950")
	khairan := nP("Khairan Bibi", person.Female, "01-01-1952")
	rKids := []person.NewPerson{
		nP("Muzaffar", person.Male, "01-01-1972"),
		nP("Zafar", person.Male, "01-01-1974"),
		nP("Saleem", person.Male, "01-01-1976"),
		nP("Yaseen", person.Male, "01-01-1978"),
		nP("Naseem", person.Female, "01-01-1980"),
	}

	root, err := CreateFamily(ctx, driver, nazir, khairan, rKids, MarriageOptionalParams{DateStart: "01-01-1971"})
	if err != nil {
		t.Fatalf("Root family failed: %v", err)
	}
	mToClean = append(mToClean, root.MarriageID)
	pToClean = append(pToClean, root.SpouseOneId, root.SpouseTwoId)
	pToClean = append(pToClean, root.ChildrenIds...)

	type famSpec struct {
		idx      int
		isFather bool
		sp       person.NewPerson
		kids     []person.NewPerson
		mDate    string
	}
	specs := []famSpec{
		{0, true, nP("Wife of Muzaffar", person.Female, "01-01-1975"), []person.NewPerson{
			nP("Mudasir", person.Male, "01-01-1996"),
			nP("Muzamil", person.Male, "01-01-1998"),
			nP("Mubashir", person.Male, "01-01-2000"),
			nP("Misbah", person.Female, "01-01-2002"),
			nP("Miftah", person.Female, "01-01-2004"),
		}, "01-01-1995"},
		{1, true, nP("Wife of Zafar", person.Female, "01-01-1976"), []person.NewPerson{
			nP("Sadia", person.Female, "01-01-1998"),
			nP("Amina", person.Female, "01-01-2000"),
			nP("Ruqaia", person.Female, "01-01-2002"),
			nP("Ali", person.Female, "01-01-2004"),
			nP("Mishi", person.Female, "01-01-2006"),
		}, "01-01-1997"},
		{2, true, nP("Wife of Saleem", person.Female, "01-01-1978"), []person.NewPerson{
			nP("Muhayodin", person.Male, "01-01-2001"),
			nP("Farid", person.Male, "01-01-2003"),
			nP("Faqih", person.Male, "01-01-2005"),
		}, "01-01-2000"},
		{3, true, nP("Wife of Yaseen", person.Female, "01-01-1980"), []person.NewPerson{
			nP("Basim", person.Male, "01-01-2003"),
			nP("Saira", person.Female, "01-01-2005"),
			nP("Sumaria", person.Female, "01-01-2007"),
		}, "01-01-2002"},
		{4, false, nP("Husband of Naseem", person.Male, "01-01-1978"), []person.NewPerson{
			nP("Arslan", person.Male, "01-01-2004"),
			nP("Faizan", person.Male, "01-01-2006"),
			nP("Usman", person.Male, "01-01-2008"),
			nP("Maryam", person.Female, "01-01-2010"),
		}, "01-01-2003"},
	}

	for _, s := range specs {
		c := rKids[s.idx]
		c.Id = root.ChildrenIds[s.idx]
		s1, s2 := c, s.sp
		if !s.isFather {
			s1, s2 = s.sp, c
		}
		sub, err := CreateFamily(ctx, driver, s1, s2, s.kids, MarriageOptionalParams{DateStart: s.mDate})
		if err != nil {
			t.Fatalf("Subfamily %d failed: %v", s.idx, err)
		}
		mToClean = append(mToClean, sub.MarriageID)
		pToClean = append(pToClean, sub.SpouseOneId, sub.SpouseTwoId)
		pToClean = append(pToClean, sub.ChildrenIds...)

		verifyPersonInDB(t, ctx, driver, sub.SpouseOneId, s1)
		verifyPersonInDB(t, ctx, driver, sub.SpouseTwoId, s2)
		verifyMarriageInDB(t, ctx, driver, sub.MarriageID, sub.SpouseOneId, sub.SpouseTwoId, person.DateProper(s.mDate), "")
		if cnt := countProducedRelationshipsForMarriage(t, ctx, driver, sub.MarriageID); cnt != int64(len(s.kids)) {
			t.Errorf("Expected %d kids, got %d", len(s.kids), cnt)
		}
		for i, k := range s.kids {
			verifyPersonInDB(t, ctx, driver, sub.ChildrenIds[i], k)
			verifyChildRelationshipInDB(t, ctx, driver, sub.MarriageID, sub.ChildrenIds[i])
		}
	}

	verifyPersonInDB(t, ctx, driver, root.SpouseOneId, nazir)
	verifyPersonInDB(t, ctx, driver, root.SpouseTwoId, khairan)
	verifyMarriageInDB(t, ctx, driver, root.MarriageID, root.SpouseOneId, root.SpouseTwoId, "01-01-1971", "")
	for i, k := range rKids {
		verifyPersonInDB(t, ctx, driver, root.ChildrenIds[i], k)
		verifyChildRelationshipInDB(t, ctx, driver, root.MarriageID, root.ChildrenIds[i])
	}
	if cnt := countProducedRelationshipsForMarriage(t, ctx, driver, root.MarriageID); cnt != 5 {
		t.Errorf("Expected 5 root kids, got %d", cnt)
	}

	res, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (p:Person {id: $id})-[:MARRIED]->(:Marriage)-[:PRODUCED]->(:Person)-[:MARRIED]->(:Marriage)-[:PRODUCED]->(gc:Person)
		 RETURN count(DISTINCT gc) AS total`,
		map[string]any{"id": root.SpouseOneId},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		t.Fatalf("Grandchildren query failed: %v", err)
	}
	tot, _ := res.Records[0].Get("total")
	if tot.(int64) != 20 {
		t.Errorf("Expected 20 grandchildren, got %d", tot.(int64))
	}
}
