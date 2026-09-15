package integration

import (
	"context"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/children"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type FamilyDetail struct {
	marriageID  string
	spouseOneId string
	spouseTwoId string
	childrenIds []string
}

type MarriageOptionalParams struct {
	DateStart string
	DateEnd   string
	Id        string
}

func cleanup(ctx context.Context, driver neo4j.Driver, peopleToDelete []string, marriageId string) {
	if marriageId != "" {
		marriage.DeleteMarriage(ctx, driver, marriageId)
	}

	for _, id := range peopleToDelete {
		person.DeleteUser(ctx, driver, id)
	}
}

func CreateFamily(ctx context.Context, driver neo4j.Driver, spouseOne person.NewPerson, spouseTwo person.NewPerson, childrenItems []person.NewPerson, optionalMarriageParams MarriageOptionalParams) (*FamilyDetail, error) {
	peopleToDelete := []string{}
	childIds := []string{}
	marriageId := ""
	familyDetail := FamilyDetail{}

	// Create Persons
	spouseOneAlreadyExists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
		ID: spouseOne.Id,
	})

	if err != nil {
		cleanup(ctx, driver, peopleToDelete, marriageId)
		return nil, err
	}

	if !spouseOneAlreadyExists {
		_, spouseOneId, err := person.CreateNewPerson(ctx, driver, spouseOne)
		peopleToDelete = append(peopleToDelete, spouseOneId)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId)
			return nil, err
		}
		familyDetail.spouseOneId = spouseOneId
	} else {
		familyDetail.spouseOneId = spouseOne.Id
	}

	spouseTwoAlreadyExists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
		ID: spouseTwo.Id,
	})

	if err != nil {
		cleanup(ctx, driver, peopleToDelete, marriageId)
		return nil, err
	}

	if !spouseTwoAlreadyExists {
		_, spouseTwoId, err := person.CreateNewPerson(ctx, driver, spouseTwo)
		peopleToDelete = append(peopleToDelete, spouseTwoId)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId)
			return nil, err
		}
		familyDetail.spouseTwoId = spouseTwoId
	} else {
		familyDetail.spouseTwoId = spouseTwo.Id
	}

	for _, child := range childrenItems {
		_, childId, err := person.CreateNewPerson(ctx, driver, child)
		peopleToDelete = append(peopleToDelete, childId)
		childIds = append(childIds, childId)
		familyDetail.childrenIds = append(familyDetail.childrenIds, childId)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId)
			return nil, err
		}
	}

	// Create Marriage

	existingMarriage, err := marriage.GetMarriageFromMarriageId(ctx, driver, marriageId)
	if err != nil {
		return nil, err
	}

	if len(existingMarriage) != 0 {
		existingSpouseOneId, existingSpouseTwoId, existingMarriageId, err := marriage.GetSpouses(ctx, driver, marriage.SpouseQueryParams{
			MarriageId: marriageId,
		})

		if err != nil {
			return nil, err
		}

		if existingMarriageId != marriageId {
			return nil, fmt.Errorf("Marriage Id Does not match with database")
		}

		if spouseOne.Id != existingSpouseOneId && spouseOne.Id != existingSpouseTwoId {
			return nil, fmt.Errorf("The Spouses of the existing marriage are not the ones provided here Existing Spouse One and Two are %v and %v while Given spouses are %v and %v", existingSpouseOneId, existingSpouseTwoId, spouseOne.Id, spouseTwo.Id)
		}

		if spouseTwo.Id != existingSpouseOneId && spouseTwo.Id != existingSpouseTwoId {
			return nil, fmt.Errorf("The Spouses of the existing marriage are not the ones provided here Existing Spouse One and Two are %v and %v while Given spouses are %v and %v", existingSpouseOneId, existingSpouseTwoId, spouseOne.Id, spouseTwo.Id)
		}

		familyDetail.marriageID = existingMarriage[0].Id
	} else {
		marriageParams := marriage.NewMarriage{
			SpouseOne: familyDetail.spouseOneId,
			SpouseTwo: familyDetail.spouseTwoId,
			DateStart: person.DateProper(optionalMarriageParams.DateStart),
			DateEnd:   person.DateProper(optionalMarriageParams.DateEnd),
			ID:        optionalMarriageParams.Id,
		}

		createdMarriageId, err := marriage.CreateNewMarriage(ctx, driver, marriageParams)
		marriageId = createdMarriageId
		familyDetail.marriageID = createdMarriageId

		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId)
			return nil, err
		}
	}

	// Relate Children
	for _, childId := range childIds {
		ok, err := children.CreateNewChild(ctx, driver, marriageId, childId)
		if !ok {
			cleanup(ctx, driver, peopleToDelete, marriageId)
			return nil, fmt.Errorf("Unable to add child %v to marriage %v", childId, marriageId)
		}
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId)
			return nil, err
		}
	}

	// Return
	return &familyDetail, nil
}
