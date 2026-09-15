package integration

import (
	"github.com/Muhammad9922/Generations/internal/person"
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
