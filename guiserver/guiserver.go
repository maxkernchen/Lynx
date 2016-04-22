/**
 *
 *	The web server responsible for rendering our GUI.
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package main

import (
	"bufio"
	"capstone/client"
	"capstone/server"
	"capstone/tracker"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
)

/** Holds our uploads html page */
var UPLOADS []byte

/** Holds our downloads html page */
var DOWNLOADS []byte

/** current form data that was submitted */
var form url.Values

/** The location of the user's root directory */
var homePath string

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

	if len(os.Args) != 2 {
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
	http.HandleFunc("/settings", SettingsHandler)
	http.HandleFunc("/uploads", UploadHandler)
	http.HandleFunc("/downloads", DownloadHandler)
	http.HandleFunc("/home", HomeHandler)

	go server.Listen(server.HandleFileRequest)

	go tracker.Listen()

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
	//rw.Write(INDEX_HTML)
	t := template.New("cool template")
	t, err := t.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
	}
	//t,_ = t.ParseFiles("index.html")
	tableEntries := TablePopulate(homePath + "/lynks.txt")
	tableTemplate := template.HTML(tableEntries)
	removalEntries := RemoveListPopulate(homePath + "/lynks.txt")
	removalTemplate := template.HTML(removalEntries)
	t.ExecuteTemplate(rw, "index.html", map[string]template.HTML{"Entries": tableTemplate,
		"RemovalList": removalTemplate})

}

/**
 * Function that handles requests on the index page: "/createlynx".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func CreateHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	//var dir []string = form["DirectoryPath"]
	var name []string = form["Name"]

	//client.CreateMeta(dir[0], name[0])
	client.CreateMeta(name[0])

	//tracker.CreateSwarm(dir[0], name[0])
	tracker.CreateSwarm(name[0])

	IndexHandler(rw, req)
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
	var metapath []string = form["MetaPath"]
	var downloads []string = form["downloadlocation"]
	client.JoinLynk(metapath[0], downloads[0])

	IndexHandler(rw, req)

} /**
 * Function that handles requests on the index page: "/removelynx".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func RemoveHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	var name []string = form["Lynks"]
	if name != nil {

		client.DeleteLynk(name[0])
		//removeLynkFile("resources/lynks.txt")
		TablePopulate(homePath + "/lynks.txt")
	}
	IndexHandler(rw, req)
}

/**
 * Function that handles requests on the index page: "/settings".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */

func SettingsHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	fmt.Println(form) //returns an array of strings
	IndexHandler(rw, req)
}

/**
 * Function that handles requests on the index page: "/uploads".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func UploadHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write(UPLOADS)
}

/**
 * Function that handles requests on the index page: "/downloads".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func DownloadHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write(DOWNLOADS)
}

/**
 * Function that handles requests on the index page: "/home".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */

func HomeHandler(rw http.ResponseWriter, req *http.Request) {
	IndexHandler(rw, req)
}

/**
  Function which will replace an element in the table in order to popluate it within the html
  file
  @param: pathToTable - the location of the lynks table .txt file
  @returns: the string which cotains the correct html table tags to be added to the html file
*/

func TablePopulate(pathToTable string) string {
	var tableEntries = ""
	lynksFile, err := os.Open(pathToTable)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(lynksFile)

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")
		tableEntries += "<tr class = settingrow > \n"
		tableEntries += "<td>" + split[0] + "</td>\n"
		tableEntries += "<td>" + split[1] + "</td>\n"
		tableEntries += "<td>" + split[2] + "</td>\n"
		tableEntries += "</tr>\n"

	}
	client.ParseLynks(pathToTable)
	return tableEntries
}

/**
  Function which creates string that contains the html for populating the dropdown list in
  the remove button

  @param: pathToTable - the path that the lynks table lives in
  @returns - a string that can be used with the html file to populate the dropdown list

*/

func RemoveListPopulate(pathToTable string) string {
	var tableEntries = ""
	lynksFile, err := os.Open(pathToTable)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(lynksFile)

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")
		tableEntries += "<option value=\"" + split[0] + "\">" + split[0] + "</option>"

	}
	return tableEntries
}

/** Function INIT runs before main and allows us to load the index html before any operations
  are done on it and find the root directory on the user's computer
*/
func init() {
	UPLOADS, _ = ioutil.ReadFile("uploads.html")
	DOWNLOADS, _ = ioutil.ReadFile("downloads.html")
	currentusr, _ := user.Current()
	homePath = currentusr.HomeDir + "/Lynx"
}
