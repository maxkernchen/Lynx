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
	"../server"
	"../client"
	"../tracker"
	"net/http"
	"os"
	"io/ioutil"
	"net/url"
	"bufio"
	"strings"
	"html/template"
)
var INDEX_HTML []byte
var UPLOADS []byte
var DOWNLOADS []byte

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
	if err != nil{
		fmt.Println(err)
	}
	//t,_ = t.ParseFiles("index.html")
	tableEntries := TablePopulate("resources/lynks.txt")
	tableTemplate := template.HTML(tableEntries)
	removalEntries := RemoveListPopulate("resources/lynks.txt")
	removalTemplate := template.HTML(removalEntries)
	t.ExecuteTemplate(rw,"index.html", map[string] template.HTML {"Entries": tableTemplate,
		"RemovalList" : removalTemplate})



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
	var dir []string = form["DirectoryPath"]
	var name []string = form["Name"]

	client.CreateMeta(dir[0],name[0])

	tracker.CreateSwarm(dir[0],name[0])

	IndexHandler(rw,req)

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
	client.JoinLynk(metapath[0],downloads[0])

	IndexHandler(rw,req)

}/**
 * Function that handles requests on the index page: "/removelynx".
 * @param http.ResponseWriter rw - This is what we use to write our html back to
 * the web page.
 * @param *http.Request req - This is the http request sent to the server.
 */
func RemoveHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	var name []string = form["Lynks"]
	if name != nil{
		client.DeleteLynk(name[0])

		removeLynkFile("resources/lynks.txt")
		TablePopulate("resources/lynks.txt")
	}
	IndexHandler(rw,req)
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
	IndexHandler(rw,req)
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

	IndexHandler(rw,req)
}

func TablePopulate(pathtotable string) string {

	var tableEntries = ""
	lynksFile, err := os.Open(pathtotable)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(lynksFile)

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")
		tableEntries += "<tr class = settingrow > \n"
		tableEntries += "<td>"+ split[0]+ "</td>\n"
		tableEntries += "<td>"+ split[1]+ "</td>\n"
		tableEntries += "<td>"+ split[2]+ "</td>\n"
		tableEntries += "</tr>\n"

	}
	client.ParseLynks(pathtotable)
	return tableEntries
}

func RemoveListPopulate(pathtotable string) string {

	var tableEntries = ""
	lynksFile, err := os.Open(pathtotable)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(lynksFile)

	// Scan each line
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")
		tableEntries +=  "<option value=\"" + split[0] + "\">" + split[0] + "</option>"

	}
	return tableEntries
}

func removeLynkFile(path string){
	os.Remove(path)
}


/** Function INIT runs before main and allows us to load the index html before any operations
    are done on it
 */
func init(){
	INDEX_HTML, _ = ioutil.ReadFile("index.html")
	UPLOADS, _ = ioutil.ReadFile("uploads.html")
	DOWNLOADS, _ = ioutil.ReadFile("downloads.html")

}

