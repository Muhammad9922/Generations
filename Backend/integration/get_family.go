package integration

import (
	"context"
	"fmt"

	"github.com/Muhammad9922/Generations/internal/children"
	"github.com/Muhammad9922/Generations/internal/marriage"
	"github.com/Muhammad9922/Generations/internal/person"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

/*

Following data will be fetched based on a user id

- Parent
- Spouse
- Children

*/

type User struct {
	Id           string
	Name         string
	DateOfBirth  person.DateProper
	DeateOfDeath person.DateProper
	Gender       person.Gender
	Alive        bool
}

type FamilyCertificate struct {
	Id            string
	Spouse        *[]User
	Chidren       *[]User
	StartOfFamily *person.DateProper
	EndOfFamily   *person.DateProper
}

func GetFamilyCertificate(ctx context.Context, driver neo4j.Driver, userId string) (*FamilyCertificate, error) {
	userExists, err := person.CheckPersonExistence(ctx, driver, person.PersonQuery{
		ID: userId,
	})

	if err != nil {
		return nil, err
	}

	if !userExists {
		return nil, fmt.Errorf("User Does Not Exists")
	}

	marriageQueryResponse, err := marriage.GetMarriageFromSpouse(ctx, driver, userId)

	if err != nil {
		return nil, err
	}

	if marriageQueryResponse == nil {
		return nil, fmt.Errorf("Marriage Does Not Exist")
	}

	var familyCertificate FamilyCertificate
	familyCertificate.StartOfFamily = &marriageQueryResponse.Start
	familyCertificate.EndOfFamily = &marriageQueryResponse.End
	familyCertificate.Id = marriageQueryResponse.Id

	spouseOneId, spouseTwoId, _, err := marriage.GetSpouses(ctx, driver, marriage.SpouseQueryParams{
		MarriageId: marriageQueryResponse.Id,
	})

	if err != nil {
		return nil, err
	}

	spouseOne, err := person.GetPerson(ctx, driver, spouseOneId)
	if err != nil {
		return nil, err
	}

	userSpouseOne := User{
		Id:           spouseOneId,
		Name:         spouseOne.PersonName,
		DateOfBirth:  spouseOne.DateOfBirth,
		DeateOfDeath: spouseOne.DateOfDeath,
		Alive:        spouseOne.Alive,
		Gender:       spouseOne.Gender,
	}

	spouseTwo, err := person.GetPerson(ctx, driver, spouseTwoId)
	if err != nil {
		return nil, err
	}

	userSpouseTwo := User{
		Id:           spouseTwoId,
		Name:         spouseTwo.PersonName,
		DateOfBirth:  spouseTwo.DateOfBirth,
		DeateOfDeath: spouseTwo.DateOfDeath,
		Alive:        spouseTwo.Alive,
		Gender:       spouseTwo.Gender,
	}

	var spouses []User = []User{
		userSpouseOne,
		userSpouseTwo,
	}

	familyCertificate.Spouse = &spouses

	var childrens []User = []User{}

	childs, err := children.GetChildren(ctx, driver, marriageQueryResponse.Id)
	for _, child := range childs {
		childrens = append(childrens, User{
			Id:           child.Id,
			Name:         child.PersonName,
			DateOfBirth:  child.DateOfBirth,
			DeateOfDeath: child.DateOfDeath,
			Gender:       child.Gender,
			Alive:        child.Alive,
		})
	}

	familyCertificate.Chidren = &childrens

	return &familyCertificate, nil
}
