package main

import (
	"fmt"
	"net/http"
	"os"
	"html/template"
)

type UserInput struct {
	Name string
	FavNum string
}

func main() {

	if len(os.Args) != 4 {
		fmt.Println("Usage: ", os.Args[0], " <port>")
		os.Exit(1)
	}

	port := os.Args[1]
	fmt.Println("Starting server on http://localhost:" + port)


	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe(":"+port, nil)
}

func IndexHandler(rw http.ResponseWriter, req *http.Request) {
	usrIn := UserInput{Name: os.Args[2], FavNum: os.Args[3]}

	t := template.New("cool template")
	t, _ = t.Parse("<h1>Hello {{.Name}}!</h1> <p>Your Favorite Number is {{.FavNum}}.</p>")

	t.Execute(rw, usrIn)
}