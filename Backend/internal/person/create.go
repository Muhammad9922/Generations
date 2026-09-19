package person

import (
	"context"
	"errors"
	"fmt"
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

// ParseTime parses the DateProper (DD-MM-YYYY) into a time.Time and reports
// whether it is a valid calendar date. Unlike the old Unix-millisecond based
// GetMS, this works for every date Go's time package can represent — including
// all pre-1970 dates (which previously produced negative millis that callers
// mistook for the "-1" invalid sentinel) — and never risks integer overflow.
func (d DateProper) ParseTime() (time.Time, bool) {
	const layout = "02-01-2006"
	t, err := time.Parse(layout, string(d))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
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
		// A create takes dates in DD-MM-YYYY only. IsValid checks that shape but
		// is regex-only, so 31-02-2000 used to slip through and fail later inside
		// time.Parse with "day out of range"; ParseTime rejects a day that does
		// not exist in that month.
		if !params.DateOfBirth.IsValid() {
			return "", "", errors.New("The Date Of Birth Must Be In DD-MM-YYYY Format")
		}

		if _, ok := params.DateOfBirth.ParseTime(); !ok {
			return "", "", fmt.Errorf("The Date Of Birth Is Not A Day That Exists: %v", params.DateOfBirth)
		}
	}

	if params.DateOfDeath != "" {

		if params.Alive {
			return "", "", errors.New("The Person Can't Both Be Alive And Have A Death Date!")
		}

		if !params.DateOfDeath.IsValid() {
			return "", "", errors.New("The Date Of Death Must Be In DD-MM-YYYY Format")
		}

		if _, ok := params.DateOfDeath.ParseTime(); !ok {
			return "", "", fmt.Errorf("The Date Of Death Is Not A Day That Exists: %v", params.DateOfDeath)
		}

		// Date Of Death Must Be After Date Of Birth
		if params.DateOfBirth != "" {
			birthTime, ok := params.DateOfBirth.ParseTime()
			if !ok {
				return "", "", errors.New("Improper Date Of Birth")
			}

			deathTime, ok := params.DateOfDeath.ParseTime()
			if !ok {
				return "", "", errors.New("Improper Date Of Death")
			}

			if deathTime.Before(birthTime) {
				return "", "", errors.New("Time Of Death Must Be After Time Of Birth")
			}
		}
	}

	var uid string

	if params.Id != "" {
		exists, err := CheckPersonExistence(ctx, driver, PersonQuery{
			ID: params.Id,
		})

		if err != nil {
			return "", "", fmt.Errorf("Error While Check For Existing Users With Same UUID: %q", err)
		}

		if exists {
			return "", "", fmt.Errorf("Duplicate UID Provided: %v", params.Id)
		}

		uid = params.Id

	} else {
		uid = uuid.New().String()
	}

	var queryParams map[string]any

	queryParams = map[string]any{
		"name":   params.PersonName,
		"gender": params.Gender,
		"id":     uid,
		"alive":  params.Alive,
		"dob":    nil,
		"dod":    nil,
	}

	if params.DateOfBirth != "" {
		neoDate, err := params.DateOfBirth.GetNeoDate()
		if err != nil {
			return "", "", err
		} else {
			queryParams["dob"] = *neoDate
		}
	}

	if params.DateOfDeath != "" {
		neoDate, err := params.DateOfDeath.GetNeoDate()
		if err != nil {
			return "", "", err
		} else {
			queryParams["dod"] = *neoDate
		}
	}

	_, err := neo4j.ExecuteQuery(
		ctx,
		driver,
		`MERGE (p:Person {name: $name, gender: $gender, date_of_birth: $dob, id: $id, alive: $alive, date_of_death: $dod})`,
		queryParams,
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return "", "", err
	}

	return params.PersonName, uid, nil
}
