package integration

import (
	"fmt"
	"testing"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/db"
	"github.com/Muhammad9922/Generations/internal/person"
)

func TestCreate7GenComplexFamily(t *testing.T) {
	ctx, driver := db.ConnectDatabase(testDatabaseURI)
	defer driver.Close(ctx)

	tag := uuid.New().String()[:8]
	var mToClean, pToClean []string
	defer func() {
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

	// Helper to attach an existing DB ID to a person struct for cousin marriages
	withID := func(p person.NewPerson, id string) person.NewPerson {
		p.Id = id
		return p
	}

	// ==========================================
	// GENERATION 1 & 2: Founders and Children
	// ==========================================
	nazir := nP("Muhammad Nazir", person.Male, "01-01-1930")
	khairan := nP("Khairan Bibi", person.Female, "01-01-1935")
	gen2Kids := []person.NewPerson{
		nP("Muzaffar", person.Male, "01-01-1955"),
		nP("Zafar", person.Male, "01-01-1957"),
		nP("Saleem", person.Male, "01-01-1959"),
		nP("Yaseen", person.Male, "01-01-1961"),
		nP("Naseem", person.Female, "01-01-1963"),
	}

	root, err := CreateFamily(ctx, driver, nazir, khairan, gen2Kids, MarriageOptionalParams{DateStart: "01-01-1954"})
	if err != nil {
		t.Fatalf("Gen 1 family failed: %v", err)
	}
	mToClean = append(mToClean, root.MarriageID)
	pToClean = append(pToClean, root.SpouseOneId, root.SpouseTwoId)
	pToClean = append(pToClean, root.ChildrenIds...)

	// ==========================================
	// GENERATION 3: The First Cousins
	// ==========================================
	type famSpec struct {
		isFather bool
		sp       person.NewPerson
		kids     []person.NewPerson
		mDate    string
	}

	specs := []famSpec{
		{true, nP("Wife of Muzaffar", person.Female, "01-01-1958"), []person.NewPerson{
			nP("Mudasir", person.Male, "01-01-1978"), nP("Muzamil", person.Male, "01-01-1980"), nP("Mubashir", person.Male, "01-01-1982"), nP("Misbah", person.Female, "01-01-1984"), nP("Miftah", person.Female, "01-01-1986"),
		}, "01-01-1977"},
		{true, nP("Wife of Zafar", person.Female, "01-01-1960"), []person.NewPerson{
			nP("Sadia", person.Female, "01-01-1980"), nP("Amina", person.Female, "01-01-1982"), nP("Ruqaia", person.Female, "01-01-1984"), nP("Ali", person.Male, "01-01-1986"), nP("Mishi", person.Female, "01-01-1988"),
		}, "01-01-1979"},
		{true, nP("Wife of Saleem", person.Female, "01-01-1962"), []person.NewPerson{
			nP("Muhayodin", person.Male, "01-01-1982"), nP("Farid", person.Male, "01-01-1984"), nP("Faqih", person.Male, "01-01-1986"),
		}, "01-01-1981"},
		{true, nP("Wife of Yaseen", person.Female, "01-01-1964"), []person.NewPerson{
			nP("Basim", person.Male, "01-01-1985"), nP("Saira", person.Female, "01-01-1987"), nP("Sumaria", person.Female, "01-01-1989"),
		}, "01-01-1984"},
		{false, nP("Husband of Naseem", person.Male, "01-01-1960"), []person.NewPerson{
			nP("Arslan", person.Male, "01-01-1984"), nP("Faizan", person.Male, "01-01-1986"), nP("Usman", person.Male, "01-01-1988"), nP("Maryam", person.Female, "01-01-1990"),
		}, "01-01-1983"},
	}

	// Store ONLY the slice of child IDs for each of the 5 subfamilies
	gen3ChildrenIds := make([][]string, 5)

	for idx, s := range specs {
		c := gen2Kids[idx]
		c.Id = root.ChildrenIds[idx]
		s1, s2 := c, s.sp
		if !s.isFather {
			s1, s2 = s.sp, c
		}

		sub, err := CreateFamily(ctx, driver, s1, s2, s.kids, MarriageOptionalParams{DateStart: s.mDate})
		if err != nil {
			t.Fatalf("Gen 3 Subfamily %d failed: %v", idx, err)
		}

		// Save the IDs so we can link them in Gen 4
		gen3ChildrenIds[idx] = sub.ChildrenIds

		mToClean = append(mToClean, sub.MarriageID)
		pToClean = append(pToClean, sub.SpouseOneId, sub.SpouseTwoId)
		pToClean = append(pToClean, sub.ChildrenIds...)
	}

	// ==========================================
	// GENERATION 4: The First Cousin Marriages
	// ==========================================

	// Map the specific Gen 3 cousins we need by injecting their saved DB IDs
	mudasir := withID(specs[0].kids[0], gen3ChildrenIds[0][0])
	misbah := withID(specs[0].kids[3], gen3ChildrenIds[0][3])

	sadia := withID(specs[1].kids[0], gen3ChildrenIds[1][0])
	mishi := withID(specs[1].kids[4], gen3ChildrenIds[1][4])

	muhayodin := withID(specs[2].kids[0], gen3ChildrenIds[2][0])
	saira := withID(specs[3].kids[1], gen3ChildrenIds[3][1])
	arslan := withID(specs[4].kids[0], gen3ChildrenIds[4][0])
	faizan := withID(specs[4].kids[1], gen3ChildrenIds[4][1])

	// 1. Mudasir marries Sadia -> Tariq, Hina
	tariqSpec := nP("Tariq", person.Male, "01-01-2000")
	hinaSpec := nP("Hina", person.Female, "01-01-2002")
	gen4_1, err := CreateFamily(ctx, driver, mudasir, sadia, []person.NewPerson{tariqSpec, hinaSpec}, MarriageOptionalParams{DateStart: "01-01-1999"})
	if err != nil {
		t.Fatalf("Gen 4.1 failed: %v", err)
	}

	// 2. Misbah marries Muhayodin -> Bilal
	bilalSpec := nP("Bilal", person.Male, "01-01-2005")
	gen4_2, err := CreateFamily(ctx, driver, muhayodin, misbah, []person.NewPerson{bilalSpec}, MarriageOptionalParams{DateStart: "01-01-2004"})
	if err != nil {
		t.Fatalf("Gen 4.2 failed: %v", err)
	}

	// 3. Arslan marries Saira -> Fatima
	fatimaSpec := nP("Fatima", person.Female, "01-01-2008")
	gen4_3, err := CreateFamily(ctx, driver, arslan, saira, []person.NewPerson{fatimaSpec}, MarriageOptionalParams{DateStart: "01-01-2007"})
	if err != nil {
		t.Fatalf("Gen 4.3 failed: %v", err)
	}

	// 4. Faizan marries Mishi -> Zeeshan
	zeeshanSpec := nP("Zeeshan", person.Male, "01-01-2010")
	gen4_4, err := CreateFamily(ctx, driver, faizan, mishi, []person.NewPerson{zeeshanSpec}, MarriageOptionalParams{DateStart: "01-01-2009"})
	if err != nil {
		t.Fatalf("Gen 4.4 failed: %v", err)
	}

	// ==========================================
	// GENERATION 5: Deepening the Roots
	// ==========================================

	tariq := withID(tariqSpec, gen4_1.ChildrenIds[0])
	hina := withID(hinaSpec, gen4_1.ChildrenIds[1])
	bilal := withID(bilalSpec, gen4_2.ChildrenIds[0])
	fatima := withID(fatimaSpec, gen4_3.ChildrenIds[0])
	zeeshan := withID(zeeshanSpec, gen4_4.ChildrenIds[0])

	// 1. Tariq marries Fatima -> Omar, Zara
	omarSpec := nP("Omar", person.Male, "01-01-2025")
	zaraSpec := nP("Zara", person.Female, "01-01-2027")
	gen5_1, err := CreateFamily(ctx, driver, tariq, fatima, []person.NewPerson{omarSpec, zaraSpec}, MarriageOptionalParams{DateStart: "01-01-2024"})
	if err != nil {
		t.Fatalf("Gen 5.1 failed: %v", err)
	}

	// 2. Bilal marries Hina -> Zainab
	zainabSpec := nP("Zainab", person.Female, "01-01-2026")
	gen5_2, err := CreateFamily(ctx, driver, bilal, hina, []person.NewPerson{zainabSpec}, MarriageOptionalParams{DateStart: "01-01-2025"})
	if err != nil {
		t.Fatalf("Gen 5.2 failed: %v", err)
	}

	// 3. Zeeshan marries Outsider -> Rayan
	outsiderWife := nP("Outsider Wife", person.Female, "01-01-2012")
	rayanSpec := nP("Rayan", person.Male, "01-01-2032")
	gen5_3, err := CreateFamily(ctx, driver, zeeshan, outsiderWife, []person.NewPerson{rayanSpec}, MarriageOptionalParams{DateStart: "01-01-2031"})
	if err != nil {
		t.Fatalf("Gen 5.3 failed: %v", err)
	}

	// ==========================================
	// GENERATION 6: Convergence
	// ==========================================

	omar := withID(omarSpec, gen5_1.ChildrenIds[0])
	zara := withID(zaraSpec, gen5_1.ChildrenIds[1])
	zainab := withID(zainabSpec, gen5_2.ChildrenIds[0])
	rayan := withID(rayanSpec, gen5_3.ChildrenIds[0])

	// 1. Omar marries Zainab -> Hamza
	hamzaSpec := nP("Hamza", person.Male, "01-01-2050")
	gen6_1, err := CreateFamily(ctx, driver, omar, zainab, []person.NewPerson{hamzaSpec}, MarriageOptionalParams{DateStart: "01-01-2048"})
	if err != nil {
		t.Fatalf("Gen 6.1 failed: %v", err)
	}

	// 2. Zara marries Rayan -> Aisha
	aishaSpec := nP("Aisha", person.Female, "01-01-2052")
	gen6_2, err := CreateFamily(ctx, driver, rayan, zara, []person.NewPerson{aishaSpec}, MarriageOptionalParams{DateStart: "01-01-2050"})
	if err != nil {
		t.Fatalf("Gen 6.2 failed: %v", err)
	}

	// ==========================================
	// GENERATION 7: The Apex
	// ==========================================

	hamza := withID(hamzaSpec, gen6_1.ChildrenIds[0])
	aisha := withID(aishaSpec, gen6_2.ChildrenIds[0])

	// Hamza marries Aisha -> Ibrahim
	ibrahimSpec := nP("Ibrahim", person.Male, "01-01-2075")
	_, err = CreateFamily(ctx, driver, hamza, aisha, []person.NewPerson{ibrahimSpec}, MarriageOptionalParams{DateStart: "01-01-2074"})
	if err != nil {
		t.Fatalf("Gen 7 failed: %v", err)
	}
}
