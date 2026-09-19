package personRouter

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Muhammad9922/Generations/internal/children"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type finalPeopleDesign struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Alive       bool   `json:"alive"`
	DateOfBirth string `json:"dateOfBirth"`
	DateOfDeath string `json:"dateOfDeath"`
}

func handleGetAllPeople(driver neo4j.Driver) http.HandlerFunc {

	// 1. Capitalize fields so the JSON encoder can read them.
	// Use `json:"..."` tags to dictate exactly how they appear in the JSON output.

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		people, err := person.GetPersonList(ctx, driver)

		// 2. Move error handling UP. Check this immediately after the database call.
		if err != nil {
			http.Error(w, "Failed to retrieve people", http.StatusInternalServerError)
			return
		}

		// 3. Performance optimization: Pre-allocate the slice's capacity
		// since you already know how many items are in 'people'.
		allPeople := make([]finalPeopleDesign, 0, len(people))

		for _, p := range people {

			// 4. Simplified mapping. If DateOfBirth is "", converting it to a string
			// safely results in "", so you don't need the extra if-statements.
			allPeople = append(allPeople, finalPeopleDesign{
				ID:          p.Id,
				Name:        p.PersonName,
				Gender:      string(p.Gender),
				Alive:       p.Alive,
				DateOfBirth: string(p.DateOfBirth),
				DateOfDeath: string(p.DateOfDeath),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allPeople)
	}
}

func handleGetPerson(driver neo4j.Driver) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.PathValue("id") == "" {
			http.Error(w, "Id Must Be Provided", http.StatusForbidden)
			return
		}

		type marriageDetail struct {
			Id            string              `json:"Id"`
			Spouse        []finalPeopleDesign `json:"Spouse"`
			Children      []finalPeopleDesign `json:"Children"`
			StartOfFamily string              `json:"StartOfFamily"`
			EndOfFamily   string              `json:"EndOfFamily"`
		}

		type PersonCertificate struct {
			Person          finalPeopleDesign
			ParentsMarriage *marriageDetail
			Marriages       []marriageDetail
		}

		var finalPersonDetail finalPeopleDesign
		var finalParents *marriageDetail
		var finalMarriages []marriageDetail

		ctx := r.Context()

		// Getting Person
		personResponse, err := person.GetPerson(ctx, driver, r.PathValue("id"))

		if err != nil || personResponse == nil {
			http.Error(w, "Error While Getting Person", http.StatusBadRequest)
			return
		} else {
			finalPersonDetail.Alive = personResponse.Alive
			finalPersonDetail.DateOfBirth = string(personResponse.DateOfBirth)
			finalPersonDetail.DateOfDeath = string(personResponse.DateOfDeath)
			finalPersonDetail.ID = personResponse.Id
			finalPersonDetail.Name = personResponse.PersonName
			finalPersonDetail.Gender = string(personResponse.Gender)
		}

		// Parents Of The Person
		parentsResponse, err := children.GetMarriageThatOfChild(ctx, driver, personResponse.Id)
		if err != nil { // Because Parent Can Be Nil
			http.Error(w, "Error While Getting List Of Parent", http.StatusBadRequest)
			return
		} else {
			if parentsResponse == nil {
				finalParents = nil
			} else if parentsResponse.Id != "" {
				finalParents = &marriageDetail{}
				finalParents.Id = parentsResponse.Id
				parentsMarriageDetail, err := marriage.GetMarriageFromMarriageId(ctx, driver, parentsResponse.Id)
				if err != nil || len(parentsMarriageDetail) != 2 {
					http.Error(w, "Error While Getting Marriage Record For Parents", http.StatusBadRequest)
					return
				}
				spouseOneDetail := parentsMarriageDetail[0]
				spouseTwoDetail := parentsMarriageDetail[1]

				finalParents.StartOfFamily = string(parentsResponse.Start)
				finalParents.EndOfFamily = string(parentsResponse.End)

				// Spouse One
				spouseOnePerson, err := person.GetPerson(ctx, driver, spouseOneDetail.Id)
				if err != nil {
					http.Error(w, "Details For Spouse One Not Found", http.StatusBadRequest)
					return
				}
				finalPeopleDesignSpouseOne := finalPeopleDesign{
					ID:          spouseOneDetail.Id,
					DateOfBirth: string(spouseOnePerson.DateOfBirth),
					DateOfDeath: string(spouseOnePerson.DateOfDeath),
					Alive:       spouseOnePerson.Alive,
					Gender:      string(spouseOnePerson.Gender),
					Name:        string(spouseOnePerson.PersonName),
				}
				finalParents.Spouse = append(finalParents.Spouse, finalPeopleDesignSpouseOne)

				// Spouse Two
				spouseTwoPerson, err := person.GetPerson(ctx, driver, spouseTwoDetail.Id)
				if err != nil {
					http.Error(w, "Details For Spouse Two Not Found", http.StatusBadRequest)
					return
				}
				finalPersonDesignSpouseTwo := finalPeopleDesign{
					ID:          spouseTwoPerson.Id,
					Name:        spouseTwoPerson.PersonName,
					DateOfBirth: string(spouseTwoPerson.DateOfBirth),
					DateOfDeath: string(spouseTwoPerson.DateOfDeath),
					Gender:      string(spouseTwoPerson.Gender),
					Alive:       spouseTwoPerson.Alive,
				}
				finalParents.Spouse = append(finalParents.Spouse, finalPersonDesignSpouseTwo)

				// Children
				childrenForMarriage, err := children.GetChildren(ctx, driver, parentsResponse.Id)
				if err != nil {
					http.Error(w, "Unable To Find Siblings For Person", http.StatusBadRequest)
					return
				}

				for _, child := range childrenForMarriage {
					finalChild := finalPeopleDesign{
						ID:          child.Id,
						Name:        child.PersonName,
						Gender:      string(child.Gender),
						Alive:       child.Alive,
						DateOfBirth: string(child.DateOfBirth),
						DateOfDeath: string(child.DateOfDeath),
					}
					finalParents.Spouse = append(finalParents.Spouse, finalChild)
				}
			} else {
				finalParents = nil
			}
		}

		// Get Marriages Of The Person
		marriagesPerson, err := children.GetMarriageThatOfSpouse(ctx, driver, finalPersonDetail.ID)
		if err != nil {
			http.Error(w, "Error While Getting Marriages For The Person", http.StatusBadRequest)
			return
		}

		for _, singleMarriage := range *marriagesPerson {
			finalMarriageDetail := marriageDetail{
				Id:            singleMarriage.Id,
				StartOfFamily: string(singleMarriage.Start),
				EndOfFamily:   string(singleMarriage.End),
			}

			childrenDetail, err := children.GetChildren(ctx, driver, finalMarriageDetail.Id)
			if err != nil {
				errorMessage := fmt.Sprintf("Error While Getting The Marriage Detail For ID: %v", finalMarriageDetail.Id)
				http.Error(w, errorMessage, http.StatusBadRequest)
				return
			}

			for _, child := range childrenDetail {
				childDesign := finalPeopleDesign{
					ID:          child.Id,
					Name:        child.PersonName,
					DateOfBirth: string(child.DateOfBirth),
					DateOfDeath: string(child.DateOfDeath),
					Gender:      string(child.Gender),
					Alive:       child.Alive,
				}

				finalMarriageDetail.Children = append(finalMarriageDetail.Children, childDesign)
			}

			spouseOneId, spouseTwoId, _, err := marriage.GetSpouses(ctx, driver, marriage.SpouseQueryParams{
				MarriageId: finalMarriageDetail.Id,
			})

			if err != nil {
				errorString := fmt.Sprintf("Error While Getting Spouses Of Marriage: %v", finalMarriageDetail.Id)
				http.Error(w, errorString, http.StatusBadRequest)
				return
			}

			spouseOneDetail, err := person.GetPerson(ctx, driver, spouseOneId)
			if err != nil {
				errorString := fmt.Sprintf("Error While Getting Detail Of Person: %v", spouseOneId)
				http.Error(w, errorString, http.StatusBadRequest)
				return
			}

			finalMarriageDetail.Spouse = append(finalMarriageDetail.Spouse, finalPeopleDesign{
				ID:          spouseOneDetail.Id,
				Name:        spouseOneDetail.PersonName,
				Gender:      string(spouseOneDetail.Gender),
				Alive:       spouseOneDetail.Alive,
				DateOfBirth: string(spouseOneDetail.DateOfBirth),
				DateOfDeath: string(spouseOneDetail.DateOfDeath),
			})

			spouseTwoDetail, err := person.GetPerson(ctx, driver, spouseTwoId)
			if err != nil {
				errorString := fmt.Sprintf("Error While Getting Detail Of Person: %v", spouseTwoId)
				http.Error(w, errorString, http.StatusBadRequest)
				return
			}

			finalMarriageDetail.Spouse = append(finalMarriageDetail.Spouse, finalPeopleDesign{
				ID:          spouseTwoDetail.Id,
				Name:        spouseTwoDetail.PersonName,
				DateOfBirth: string(spouseTwoDetail.DateOfBirth),
				DateOfDeath: string(spouseTwoDetail.DateOfDeath),
				Gender:      string(spouseTwoDetail.Gender),
				Alive:       spouseTwoDetail.Alive,
			})

			finalMarriages = append(finalMarriages, finalMarriageDetail)
		}

		finalPersonCertificate := PersonCertificate{
			Person:          finalPersonDetail,
			ParentsMarriage: finalParents,
			Marriages:       finalMarriages,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(finalPersonCertificate)
	}
}
