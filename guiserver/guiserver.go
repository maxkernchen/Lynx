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
	//"html/template"
	"net/http"
	"os"
	"io/ioutil"
	"net/url"
)
var INDEX_HTML []byte
var form url.Values

/** A struct that we combine with our Go template to produce desired HTML */
type UserInput struct {
	Name   string
	FavNum string
}
/** Struct specifically for adding html resources like css */
type HTMLFiles struct {
	fs http.FileSystem
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

	fs := HTMLFiles{http.Dir("js/")}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(fs)))

	css := HTMLFiles{http.Dir("css/")}
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(css)))

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/joinlynx", JoinHandler)
	http.HandleFunc("/createlynx", CreateHandler)
	http.HandleFunc("/removelynx", RemoveHandler)

	http.ListenAndServe(":"+port, nil)

}
/**
	Method which is called when a new HTMLFiles struct is created it simply opens the
	directory and returns the file and an error
	@returns http.File a file to be used for http
	@returns:error: an error is the file is not openable
 */
func (fs HTMLFiles) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}


/**
 * Function that handles requests on the index page: "/".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func IndexHandler(rw http.ResponseWriter, req *http.Request) {
	//usrIn := UserInput{Name: os.Args[2], FavNum: os.Args[3]} // This will change
	rw.Write(INDEX_HTML)
	/**t := template.New("cool template")
	t, _ = t.Parse("<h1>Hello {{.Name}}!</h1> <p>Your Favorite Number is {{.FavNum}}.</p>
	<input id=clickMe type=button value=clickme onclick=printSomething(); />")
	t.Execute(rw, usrIn)
	**/

}/**
 * Function that handles requests on the index page: "/createlynx".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func CreateHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	fmt.Println(form)
	//var formstr []string = form["Lynx_name"]
	//fmt.Println(formstr[0]) //returns an array of strings
	rw.Write(INDEX_HTML)

}
/**
 * Function that handles requests on the index page: "/joinlynx".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func JoinHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	fmt.Println(form)
	rw.Write(INDEX_HTML)

}/**
 * Function that handles requests on the index page: "/removelynx".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func RemoveHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	fmt.Println(form) //returns an array of strings
	rw.Write(INDEX_HTML)
}
/** Function INIT runs before main and allows us to load the index html before any operations
    are done on it
 */
func init(){
	INDEX_HTML, _ = ioutil.ReadFile("index.html")
}

