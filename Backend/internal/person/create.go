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
	Id          string
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
	correctDate := dateRegix.MatchString(string(d))
	return correctDate
}

func (d DateProper) GetMS() int {
	const layout = "02-01-2006"
	dObject, err := time.Parse(layout, string(d))
	if err != nil {
		return -1
	}
	return int(dObject.UnixMilli())
}

func (d DateProper) GetNeoDate() (*neo4j.Date, error) {
	if d == "" {
		return nil, nil
	}
	const dateLayout = "02-01-2006"
	if d.IsValid() {
		parsed, err := time.Parse(dateLayout, string(d))
		if err != nil {
			return nil, err
		}
		neoDate := neo4j.DateOf(parsed)
		return &neoDate, nil
	} else {
		return nil, errors.New("Invalid Date")
	}
}

func GetProperDate(val any) DateProper {
	const dateLayout = "02-01-2006"
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case neo4j.Date:
		return DateProper(v.Time().Format(dateLayout))
	case time.Time:
		return DateProper(v.Format(dateLayout))
	case string:
		// Fallback for raw ISO string "YYYY-MM-DD"
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return DateProper(t.Format(dateLayout))
		}
		return DateProper(v)
	default:
		return ""
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
			msBirth := params.DateOfBirth.GetMS()

			if msBirth == -1 {
				return "", "", errors.New("Improper Date Of Birth")
			}

			msDeath := params.DateOfDeath.GetMS()
			if msDeath == -1 {
				return "", "", errors.New("Improper Date Of Death")
			}

			if msDeath < msBirth {
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
	}(params.Id)

	var neoDateBirthFinal neo4j.Date
	var neoDateDeathFinal neo4j.Date

	if params.DateOfBirth != "" {
		neoDate, err := params.DateOfBirth.GetNeoDate()
		if err != nil {
			return "", "", err
		} else {
			neoDateBirthFinal = *neoDate
		}
	}

	if params.DateOfDeath != "" {
		neoDate, err := params.DateOfDeath.GetNeoDate()
		if err != nil {
			return "", "", err
		} else {
			neoDateDeathFinal = *neoDate
		}
	}

	neo4j.ExecuteQuery(
		ctx,
		driver,
		`MERGE (p:Person {name: $name, gender: $gender, date_of_birth: $dob, id: $id, alive: $alive, date_of_death: $dod})`,
		map[string]any{
			"name":   params.PersonName,
			"gender": params.Gender,
			"dob":    neoDateBirthFinal,
			"id":     uuid,
			"alive":  params.Alive,
			"dod":    neoDateDeathFinal,
		},
		neo4j.EagerResultTransformer,
	)

	return params.PersonName, uuid, nil
}
