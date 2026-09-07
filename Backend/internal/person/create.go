package person

import (
	"context"
	"errors"
	"regexp"
	"time"
	"uuid"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type Gender string
type DateProper string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
)

type NewPerson struct {
	id          string
	PersonName  string
	Gender      Gender
	DateOfBirth DateProper
	Alive       bool
	DateOfDeath DateProper
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

func (d DateProper) IsValid() bool {
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
		if !params.DateOfBirth.IsValid() {
			return "", "", errors.New("The Date Of Birth Must Be In DD-MM-YYYY Format")
		}
	}

	if params.DateOfDeath != "" {
		if !params.DateOfDeath.IsValid() {
			return "", "", errors.New("The Date Of Death Must Be In DD-MM-YYYY Format")
		}

		// Date Of Death Must Be After Date Of Birth
		if params.DateOfBirth != "" {
			const layout = "02-01-2006"
			time_birth_object, err := time.Parse(layout, string(params.DateOfBirth))
			if err != nil {
				return "", "", errors.New("The Date Of Birth Is Not Proper")
			}
			time_birth := time_birth_object.UnixMilli()

			time_death_object, err := time.Parse(layout, string(params.DateOfDeath))

			if time_death := time_death_object.UnixMilli(); time_death < time_birth {
				return "", "", errors.New("Time Of Death Must Be After Time Of Birth")
			}
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
		`MERGE (p:Person {name: $name, gender: $gender, date_of_birth: $dob, id: $id, alive: $alive, date_of_death: $dod})`,
		map[string]any{
			"name":   params.PersonName,
			"gender": params.Gender,
			"dob":    params.DateOfBirth,
			"id":     uuid,
			"alive":  params.Alive,
			"dod":    params.DateOfDeath,
		},
		neo4j.EagerResultTransformer,
	)

	return params.PersonName, uuid, nil
}
