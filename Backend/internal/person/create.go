package person

import (
	"context"
	"errors"
	"regexp"
	"uuid"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type Gender string
type DateOfBirth string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
)

type NewPerson struct {
	id          string
	PersonName  string
	Gender      Gender
	DateOfBirth DateOfBirth
	Alive       bool
}

var dateRegix = regexp.MustCompile(`^(0[1-9]|[12][0-9]|3[01])-(0[1-9]|1[0-2])-\d{4}$`)

func (g Gender) IsValid() bool {
	switch g {
	case Male, Female:
		return true
	default:
		return false
	}
}

func (d DateOfBirth) IsValid() bool {
	correct_date := dateRegix.MatchString(string(d))
	return correct_date

}

func CreateNewPerson(ctx context.Context, driver neo4j.Driver, params NewPerson) (string, string, error) {
	// Name Is Mandatory
	if params.PersonName == "" {
		return "", "", errors.New("The Person's Name Must Be Given")
	}

	if !params.Gender.IsValid() {
		return "", "", errors.New("A Gender Must Be Supplied For New Person")
	}

	if params.DateOfBirth != "" {
		if params.DateOfBirth.IsValid() {
			return "", "", errors.New("The Date Of Birth Must Be In DD-MM-YYYY Format")
		}
	}

	var uuid string = func(uid string) string {
		if uid != "" {
			return uid
		} else {
			return uuid.New().String()
		}
	}(params.id)

	neo4j.ExecuteQuery(
		ctx,
		driver,
		`MERGE (p:Person {name: $name, gender: $gender, date_of_birth: $dob, id: $id, alive: $alive})`,
		map[string]any{
			"name":   params.PersonName,
			"gender": params.Gender,
			"dob":    params.DateOfBirth,
			"id":     uuid,
			"alive":  params.Alive,
		},
		neo4j.EagerResultTransformer,
	)

	return params.PersonName, uuid, nil
}
