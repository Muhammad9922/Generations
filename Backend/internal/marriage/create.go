package marriage

import (
	"context"

	"github.com/Muhammad9922/Generations/internal/person"
)

type NewMarriage struct {
	DateStart person.DateProper
	DateEnd   person.DateProper
	SpouseOne string
	SpouseTwo string
	id        string
}

func CreateNewMarriage(ctx context.Context) {

}
