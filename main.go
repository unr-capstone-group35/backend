package main

import (
	"fmt"
	"net/http"

	"github.com/tylerolson/capstone-backend/api"
	"github.com/tylerolson/capstone-backend/course"
	"github.com/tylerolson/capstone-backend/user"
)

func main() {
	userMapStore := user.NewMapStore()
	coursesMapStore := course.NewMapStore()

	// if err := coursesMapStore.LoadCourse("data/programming_basics/0.json"); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	server := api.NewServer(userMapStore, coursesMapStore)

	port := ":3000"

	fmt.Printf("Running server on %s\n", port)

	http.ListenAndServe(port, server.Mux)
}
