package person

import (
	"context"
	"errors"
	"regexp"
	"uuid"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type Gender string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
)

type NewPerson struct {
	PersonName  string
	Gender      Gender
	DateOfBirth string
	MarriageID  string
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

func CreateNewPerson(ctx context.Context, driver neo4j.Driver, params NewPerson) (string, string, error) {
	// Name Is Mandatory
	if params.PersonName == "" {
		return "", "", errors.New("The Person's Name Must Be Given")
	}

	if !params.Gender.IsValid() {
		return "", "", errors.New("A Gender Must Be Supplied For New Person")
	}

	if params.DateOfBirth != "" {
		correct_date := dateRegix.MatchString(params.DateOfBirth)
		if correct_date == false {
			return "", "", errors.New("The Date Of Birth Must Be In DD-MM-YYYY Format")
		}
	}

	// TODO(): Add Marriage ID Validation Logic

	var uuid string = uuid.New().String()

	neo4j.ExecuteQuery(
		ctx,
		driver,
		`MERGE (p:Person {name: $name, gender: $gender, DateOfBirth: $dob})`,
		map[string]any{
			"name":   params.PersonName,
			"gender": params.Gender,
			"dob":    params.DateOfBirth,
		},
		neo4j.EagerResultTransformer,
	)

	// TODO: Implement Marriage ID Logic

	return params.PersonName, uuid, nil
}
