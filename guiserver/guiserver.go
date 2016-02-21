/**
 *
 *	The web server resposible for rendering our GUI.
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

/** A struct that we combine with our Go template to produce desired HTML */
type UserInput struct {
	Name   string
	FavNum string
}

/**
 * Launches our web server using the port specified by commandline argument 1.
 */
func main() {

	if len(os.Args) != 4 {
		fmt.Println("Usage: ", os.Args[0], " <port>")
		os.Exit(1)
	}

	port := os.Args[1]
	fmt.Println("Starting server on http://localhost:" + port)

	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe(":" + port, nil)
}

/**
 * Function that handles requests on the index page: "/".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func IndexHandler(rw http.ResponseWriter, req *http.Request) {
	usrIn := UserInput{Name: os.Args[2], FavNum: os.Args[3]} // This will change

	t := template.New("cool template")
	t, _ = t.Parse("<h1>Hello {{.Name}}!</h1> <p>Your Favorite Number is {{.FavNum}}.</p>")

	t.Execute(rw, usrIn)
}
