package main

import (
	"fmt"
	"net/http"

	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/lessons"
	"github.com/tylerolson/capstone-backend/user"
)

func main() {
	mapUser := user.NewMapStore()
	lessons := lessons.NewMapStore()

	if err := lessons.LoadPath("lessons/programming_basics.json"); err != nil {
		fmt.Println(err)
		return
	}

	paths, _ := lessons.ListPathNames()

	fmt.Println(paths)

	questions, _ := lessons.ListQuestionsInPath(paths[0])

	for _, question := range questions {
		fmt.Printf("%+v\n", question)

	}

	server := api.NewServer(mapUser)

	port := ":3000"

	fmt.Printf("Running server on %s\n", port)

	http.ListenAndServe(port, server.Mux)
}
