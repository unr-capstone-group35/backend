package main

import (
	"net/http"

	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/user"
)

func main() {
	mapUser := user.NewMapStore()
	server := api.NewServer(mapUser)

	http.ListenAndServe(":3000", server.Mux)

}
