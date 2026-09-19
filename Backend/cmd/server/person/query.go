package personRouter

import (
	"context"
	"net/http"

	"github.com/Muhammad9922/Generations/cmd/server/apiresponse"
	"github.com/Muhammad9922/Generations/internal/children"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// finalPeopleDesign is the one person shape this API sends. The json tags make
// the keys lowercase — id, name, gender, alive, dateOfBirth, dateOfDeath — both
// in GET /people and inside a GET /people/{id} certificate, so the frontend reads
// one casing everywhere (API_SCOPE.md §2).
type finalPeopleDesign struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Alive       bool   `json:"alive"`
	DateOfBirth string `json:"dateOfBirth"`
	DateOfDeath string `json:"dateOfDeath"`
}

// marriageDetail is the container around those people. Only the person shape
// carries json tags, so this one keeps the Go field names — Id, Spouse,
// Children, StartOfFamily, EndOfFamily — which is what the frontend's
// FamilyCertificate expects.
type marriageDetail struct {
	Id            string              `json:"Id"`
	Spouse        []finalPeopleDesign `json:"Spouse"`
	Children      []finalPeopleDesign `json:"Children"`
	StartOfFamily string              `json:"StartOfFamily"`
	EndOfFamily   string              `json:"EndOfFamily"`
}

type personCertificate struct {
	Person          finalPeopleDesign `json:"Person"`
	ParentsMarriage *marriageDetail   `json:"ParentsMarriage"`
	Marriages       []marriageDetail  `json:"Marriages"`
}

// writeError answers with the JSON error envelope the frontend reads, {"error": "..."}
// (API_SCOPE.md §2). http.Error sends text/plain, which http.ts turns into
// "The API did not answer with JSON (...)" instead of showing the message.
//
// The envelope itself lives in apiresponse so this router and marriageRouter
// cannot answer with two slightly different shapes.
func writeError(w http.ResponseWriter, status int, message string) {
	apiresponse.Error(w, status, message)
}

// writeJSON is the mirror image of writeError: the one success envelope.
func writeJSON(w http.ResponseWriter, status int, body any) {
	apiresponse.Write(w, status, body)
}

// toWirePerson maps a stored person onto the shape the API sends.
func toWirePerson(p *person.NewPerson) finalPeopleDesign {
	return finalPeopleDesign{
		ID:          p.Id,
		Name:        p.PersonName,
		Gender:      string(p.Gender),
		Alive:       p.Alive,
		DateOfBirth: string(p.DateOfBirth),
		DateOfDeath: string(p.DateOfDeath),
	}
}

// buildMarriageDetail fills one certificate: both spouses, every child the
// marriage produced, and the dates it ran between.
//
// The people come from GetSpouses and GetChildren. They do NOT come from
// GetMarriageFromMarriageId: that query matches the Marriage node itself, so it
// returns one row holding the marriage's dates — never one row per spouse.
// Reading its rows as spouses is what used to answer 500 "Error While Getting
// Marriage Record For Parents" for every person who had parents.
func buildMarriageDetail(ctx context.Context, driver neo4j.Driver, id string, start, end person.DateProper) (*marriageDetail, error) {
	detail := &marriageDetail{
		Id:            id,
		StartOfFamily: string(start),
		EndOfFamily:   string(end),
	}

	spouseOneId, spouseTwoId, _, err := marriage.GetSpouses(ctx, driver, marriage.SpouseQueryParams{MarriageId: id})
	if err != nil {
		return nil, err
	}

	// A marriage that has lost a spouse still has its other one; an empty id
	// (or two of them) simply contributes nobody.
	for _, spouseId := range []string{spouseOneId, spouseTwoId} {
		if spouseId == "" {
			continue
		}

		spouse, err := person.GetPerson(ctx, driver, spouseId)
		if err != nil {
			return nil, err
		}

		detail.Spouse = append(detail.Spouse, toWirePerson(spouse))
	}

	kids, err := children.GetChildren(ctx, driver, id)
	if err != nil {
		return nil, err
	}

	for _, kid := range kids {
		child := kid
		detail.Children = append(detail.Children, toWirePerson(&child))
	}

	return detail, nil
}

func handleGetAllPeople(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		people, err := person.GetPersonList(ctx, driver)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to retrieve people")
			return
		}

		// Pre-allocate: the size is already known, and the palette renders the
		// whole list.
		allPeople := make([]finalPeopleDesign, 0, len(people))

		for _, p := range people {
			allPeople = append(allPeople, toWirePerson(&p))
		}

		// personFatherName is deliberately absent: the Person node stores no such
		// property (see the MERGE in internal/person/create.go), and the palette
		// hides the "Father: …" subtitle when it is missing.
		writeJSON(w, http.StatusOK, struct {
			People []finalPeopleDesign `json:"people"`
		}{People: allPeople})
	}
}

func handleGetPerson(driver neo4j.Driver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "Id Must Be Provided")
			return
		}

		ctx := r.Context()

		// An unknown person is a normal answer (404, and the page shows "Person
		// not found"). Any other failure is the server's, so it must not be
		// dressed up as "not found" — GetPerson returns the same nil on a lost
		// connection.
		exists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{ID: id})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error While Looking For The Person")
			return
		}
		if !exists {
			writeError(w, http.StatusNotFound, "Person not found")
			return
		}

		record, err := person.GetPerson(ctx, driver, id)
		if err != nil || record == nil {
			writeError(w, http.StatusInternalServerError, "Error While Getting Person")
			return
		}

		certificate := personCertificate{
			Person: toWirePerson(record),
			// Never null: the page maps over it on every render.
			Marriages: []marriageDetail{},
		}

		// The marriage this person is a child of, which is the only thing that
		// makes two spouses their parents.
		parentsMarriage, err := children.GetMarriageThatOfChild(ctx, driver, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error While Getting The Parents Of This Person")
			return
		}

		if parentsMarriage != nil && parentsMarriage.Id != "" {
			parents, err := buildMarriageDetail(ctx, driver, parentsMarriage.Id, parentsMarriage.Start, parentsMarriage.End)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Error While Getting The Marriage Of The Parents")
				return
			}
			certificate.ParentsMarriage = parents
		}

		// Every marriage this person is a spouse in. A person may have none, and
		// may have several.
		spouseMarriages, err := children.GetMarriageThatOfSpouse(ctx, driver, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error While Getting The Marriages Of This Person")
			return
		}

		if spouseMarriages != nil {
			for _, item := range *spouseMarriages {
				detail, err := buildMarriageDetail(ctx, driver, item.Id, item.Start, item.End)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "Error While Getting A Marriage Of This Person")
					return
				}
				certificate.Marriages = append(certificate.Marriages, *detail)
			}
		}

		writeJSON(w, http.StatusOK, certificate)
	}
}
