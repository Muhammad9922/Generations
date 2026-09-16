package integration

import (
	"context"
	"fmt"
	"uuid"

	"github.com/Muhammad9922/Generations/internal/children"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type FamilyDetail struct {
	MarriageID  string
	SpouseOneId string
	SpouseTwoId string
	ChildrenIds []string
}

type MarriageOptionalParams struct {
	DateStart string
	DateEnd   string
	Id        string
}

func cleanup(ctx context.Context, driver neo4j.Driver, peopleToDelete []string, marriageId string, shouldDeleteMarriage bool) {
	if shouldDeleteMarriage && marriageId != "" {
		marriage.DeleteMarriage(ctx, driver, marriageId)
	}

	for _, id := range peopleToDelete {
		person.DeleteUser(ctx, driver, id) // Ignore errors
	}
}

func CreateFamily(ctx context.Context, driver neo4j.Driver, spouseOne person.NewPerson, spouseTwo person.NewPerson, childrenItems []person.NewPerson, optionalMarriageParams MarriageOptionalParams) (*FamilyDetail, error) {
	peopleToDelete := []string{}
	childIds := []string{}
	marriageId := ""
	createdNewMarriage := false
	familyDetail := FamilyDetail{}

	// Create Persons
	spouseOneAlreadyExists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
		ID: spouseOne.Id,
	})

	if err != nil {
		cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
		return nil, err
	}

	if !spouseOneAlreadyExists {
		_, spouseOneId, err := person.CreateNewPerson(ctx, driver, spouseOne)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, err
		}
		peopleToDelete = append(peopleToDelete, spouseOneId)
		familyDetail.SpouseOneId = spouseOneId
	} else {
		familyDetail.SpouseOneId = spouseOne.Id
	}

	spouseTwoAlreadyExists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
		ID: spouseTwo.Id,
	})

	if err != nil {
		cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
		return nil, err
	}

	if !spouseTwoAlreadyExists {
		_, spouseTwoId, err := person.CreateNewPerson(ctx, driver, spouseTwo)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, err
		}
		peopleToDelete = append(peopleToDelete, spouseTwoId)
		familyDetail.SpouseTwoId = spouseTwoId
	} else {
		familyDetail.SpouseTwoId = spouseTwo.Id
	}

	for _, child := range childrenItems {
		childAlreadyExists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
			ID: child.Id,
		})
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, err
		}

		if !childAlreadyExists {
			_, childId, err := person.CreateNewPerson(ctx, driver, child)
			if err != nil {
				cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
				return nil, err
			}
			peopleToDelete = append(peopleToDelete, childId)
			childIds = append(childIds, childId)
			familyDetail.ChildrenIds = append(familyDetail.ChildrenIds, childId)
		} else {
			childIds = append(childIds, child.Id)
			familyDetail.ChildrenIds = append(familyDetail.ChildrenIds, child.Id)
		}

	}

	// Create Marriage
	if optionalMarriageParams.Id != "" {

		existingMarriage, err := marriage.GetMarriageFromMarriageId(ctx, driver, optionalMarriageParams.Id)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, err
		}

		if len(existingMarriage) == 0 {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, fmt.Errorf("marriage with id %v not found", optionalMarriageParams.Id)
		}

		existingSpouseOneId, existingSpouseTwoId, existingMarriageId, err := marriage.GetSpouses(ctx, driver, marriage.SpouseQueryParams{
			MarriageId: optionalMarriageParams.Id,
		})

		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, err
		}

		if existingMarriageId != optionalMarriageParams.Id {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, fmt.Errorf("Marriage Id Does not match with database")
		}

		s1 := familyDetail.SpouseOneId
		s2 := familyDetail.SpouseTwoId
		if !((s1 == existingSpouseOneId && s2 == existingSpouseTwoId) || (s1 == existingSpouseTwoId && s2 == existingSpouseOneId)) {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, fmt.Errorf("The Spouses of the existing marriage are not the ones provided here Existing Spouse One and Two are %v and %v while Given spouses are %v and %v", existingSpouseOneId, existingSpouseTwoId, s1, s2)
		}

		familyDetail.MarriageID = existingMarriage[0].Id
		marriageId = existingMarriage[0].Id
	} else {
		marriageParams := marriage.NewMarriage{
			SpouseOne: familyDetail.SpouseOneId,
			SpouseTwo: familyDetail.SpouseTwoId,
			DateStart: person.DateProper(optionalMarriageParams.DateStart),
			DateEnd:   person.DateProper(optionalMarriageParams.DateEnd),
			ID:        uuid.New().String(),
		}

		createdMarriageId, err := marriage.CreateNewMarriage(ctx, driver, marriageParams)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, err
		}

		marriageId = createdMarriageId
		createdNewMarriage = true
		familyDetail.MarriageID = createdMarriageId

	}

	// Relate Children
	for _, childId := range childIds {
		ok, err := children.CreateNewChild(ctx, driver, marriageId, childId)
		if err != nil {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, err
		}
		if !ok {
			cleanup(ctx, driver, peopleToDelete, marriageId, createdNewMarriage)
			return nil, fmt.Errorf("Unable to add child %v to marriage %v", childId, marriageId)
		}

	}

	// Return
	return &familyDetail, nil
}
