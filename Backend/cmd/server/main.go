package main

import (
	"net/http"

	personRouter "github.com/Muhammad9922/Generations/cmd/server/person"
	"github.com/Muhammad9922/Generations/internal/db"
)

func main() {
	ctx, driver := db.ConnectDatabase("bolt://192.168.0.133:7687")
	defer driver.Close(ctx)

	mux := http.NewServeMux()

	personRouter.RegisterPersonRoutes(mux, driver)

	http.ListenAndServe(":8080", mux)
}
