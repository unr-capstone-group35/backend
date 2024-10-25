package main

import (
	"fmt"
	"net/http"

	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/user"
)

func main() {
	mapUser := user.NewMapStore()
	server := api.NewServer(mapUser)

	port := ":3000"

	fmt.Printf("Running server on %s\n", port)

	http.ListenAndServe(port, server.Mux)
}
