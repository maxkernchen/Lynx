// The web server responsible for rendering our GUI.
// @author: Michael Bruce
// @author: Max Kernchen
//@verison: 2/17/2016
package main

import (
	"bufio"
	"../client"
	"../lynxutil"
	"../server"
	"../tracker"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Holds our uploads html page
var uploads []byte

// Holds our downloads html page
var downloads []byte

// current form data that was submitted
var form url.Values

// UserInput - A struct that we combine with our Go template to produce desired HTML
type UserInput struct {
	Name   string
	FavNum string
}

// HTMLFiles - Struct specifically for adding html resources like css
type HTMLFiles struct {
	fs http.FileSystem
}

// Launches our web server using the port specified by commandline argument 1.
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], " <port>")
		os.Exit(1)
	}

	// Checks to see if lynks.txt exists - if it doesn't it is created.
	if _, err := os.Stat(lynxutil.HomePath + "/lynks.txt"); os.IsNotExist(err) {
		os.Create(lynxutil.HomePath + "/lynks.txt")
		fmt.Println(lynxutil.HomePath)
	}

	port := os.Args[1]
	fmt.Println("Starting server on http://localhost:" + port)

	fs := HTMLFiles{http.Dir("C:/Projects/Go/src/capstone2/guiserver/js")}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(fs)))

	css := HTMLFiles{http.Dir("C:/Projects/Go/src/capstone2/guiserver/css")}
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(css)))

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/joinlynx", JoinHandler)
	http.HandleFunc("/createlynx", CreateHandler)
	http.HandleFunc("/removelynx", RemoveHandler)
	http.HandleFunc("/settings", SettingsHandler)
	http.HandleFunc("/uploads", UploadHandler)
	http.HandleFunc("/downloads", DownloadHandler)
	http.HandleFunc("/home", HomeHandler)
	http.HandleFunc("/files", FileHandler)

	go server.Listen(server.HandleFileRequest)

	go tracker.Listen()

	http.ListenAndServe(":"+port, nil)
}

// Open - Method which is called when a new HTMLFiles struct is created it simply opens the
// directory and returns the file and an error
// @returns http.File a file to be used for http
// @returns:error: an error is the file is not openable
func (fs HTMLFiles) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// IndexHandler - Function that handles requests on the index page: "/".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func IndexHandler(rw http.ResponseWriter, req *http.Request) {
	t := template.New("cool template")
	t, err := t.ParseFiles("C:/Projects/Go/src/capstone2/guiserver/index2.html")
	if err != nil {
		fmt.Println(err)
	}
	tableEntries := TablePopulate(lynxutil.HomePath + "/lynks.txt")
	tableTemplate := template.HTML(tableEntries)

	t.ExecuteTemplate(rw, "index2.html", map[string]template.HTML{"Entries": tableTemplate})
}



// CreateHandler - Function that handles requests on the index page: "/createlynx".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func CreateHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	name := form["Name"]

	client.CreateMeta(name[0])
	tracker.CreateSwarm(name[0])

	IndexHandler(rw, req)
}

// JoinHandler - Function that handles requests on the index page: "/joinlynx".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func JoinHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	metapath := form["MetaPath"]
	client.JoinLynk(metapath[0])

	IndexHandler(rw, req)

}

// RemoveHandler - Function that handles requests on the index page: "/removelynx".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func RemoveHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	name := form["index"]
	if name != nil {
		index,_ := strconv.Atoi(name[0])
		client.DeleteLynk(client.GetLynkNameFromIndex(index))
		TablePopulate(lynxutil.HomePath + "/lynks.txt")
	}
	IndexHandler(rw, req)
}

// SettingsHandler - Function that handles requests on the index page: "/settings".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func SettingsHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	fmt.Println(form) //returns an array of strings
	IndexHandler(rw, req)
}

// UploadHandler - Function that handles requests on the index page: "/uploads".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func UploadHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write(uploads)
}

// DownloadHandler - Function that handles requests on the index page: "/downloads".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func DownloadHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write(downloads)
}

// HomeHandler - Function that handles requests on the index page: "/home".
// @param http.ResponseWriter rw - This is what we use to write our html back to
// the web page.
// @param *http.Request req - This is the http request sent to the server.
func HomeHandler(rw http.ResponseWriter, req *http.Request) {
	IndexHandler(rw, req)
}

// TablePopulate - Function which will replace an element in the table in order to popluate it
// within the html file
// @param pathToTable - the location of the lynks table .txt file
// @returns the string which cotains the correct html table tags to be added to the html file
func TablePopulate(pathToTable string) string {
	var tableEntries = ""
	lynksFile, err := os.Open(pathToTable)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(lynksFile)
	i := 0
	// Scan each line
	for scanner.Scan() {

		rowStringNum := strconv.Itoa(i)
		line := strings.TrimSpace(scanner.Text()) // Trim helps with errors in \n
		split := strings.Split(line, ":::")
		tableEntries += "<tr id= row" + rowStringNum + " > \n"
		tableEntries += "<td>" + split[0] + "</td>\n"
		tableEntries += "<td><form id=\"remove\" method=\"POST\" action=\"/removelynx\"><input type=\"submit\" class=\"btn btn-danger \"id=\"removelynk\" value=\"Remove Lynk\">\n <input type=\"hidden\" name=\"index\" value=\"" + rowStringNum + "\"></form></td>\n"
		tableEntries += "<form name=\"row" + rowStringNum + "form\" id=\"row" + rowStringNum + "formid\" method=\"POST\" action=\"/files\"><input type=\"hidden\" name=\"index\" value=\"" + rowStringNum + "\"></form>"
		tableEntries += "</tr>\n"
		i++
	}
	//client.ParseLynks(pathToTable)
	return tableEntries
}

// RemoveListPopulate - Function which creates string that contains the html for populating the
// dropdown list in the remove button
// @param pathToTable - the path that the lynks table lives in
// @returns - a string that can be used with the html file to populate the dropdown list
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

// FilePopulate - Function which populates a lynk with their files and filesizes so we
// display them.
// @param pathToTable - the lynk whose files we want to populate
// @returns - a string containing all the file entries
func FilePopulate(index int) string {
	client.PopulateFilesAndSize()
	lynks := client.GetLynks()
	fileEntries := ""
	tempLynk := lynks[index]
	//fmt.Println(tempLynk.Name)
	fileNames := tempLynk.FileNames
	fileSizes := tempLynk.FileSize
	i := 0
	for i < len(fileNames) {
		fileEntries += "<tr> \n"
		fileEntries += "<td>" + fileNames[i] + "</td>\n"
		fileEntries += "<td>" + strconv.Itoa(fileSizes[i]) + "</td>\n"
		fileEntries += "<td><form id=\"remove\" method=\"POST\" action=\"/removelynx\"><input type=\"button\" class=\"btn btn-danger \"id=\"removelynk\" value=\"Remove File\"></form></td>\n"
		fileEntries += "</tr>\n"
		i++
	}

	return fileEntries
}
func FileHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	index := form["index"]

	t := template.New("cool template")
	t, err := t.ParseFiles("C:/Projects/Go/src/capstone2/guiserver/index2.html")
	if err != nil {
		fmt.Println(err)
	}
	tableEntries := TablePopulate(lynxutil.HomePath + "/lynks.txt")
	tableTemplate := template.HTML(tableEntries)
	indexInt,err := strconv.Atoi(index[0]);
	fileEntry := FilePopulate(indexInt)

	t.ExecuteTemplate(rw, "index2.html", map[string]template.HTML{"Entries": tableTemplate,
		"Files": template.HTML(fileEntry)})


}


// Function INIT runs before main and allows us to load the index html before any operations
// are done on it and find the root directory on the user's computer
func init() {
	uploads,_  = ioutil.ReadFile("uploads.html")
	downloads, _ = ioutil.ReadFile("downloads.html")
}
