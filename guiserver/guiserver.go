// The web server responsible for rendering our GUI.
// @author: Michael Bruce
// @author: Max Kernchen
//@verison: 2/17/2016
package main

import (
	"bufio"
	"capstone/client"
	"capstone/lynxutil"
	"capstone/server"
	"capstone/tracker"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jasonlvhit/gocron"
)

// Holds our uploads html page
var uploads []byte

// Holds our downloads html page
var downloads []byte

// current form data that was submitted
var form url.Values

// Holds the currentLynk being checked for changes
var currentLynk lynxutil.Lynk

// Tells us whether or not our lynk's files have changed
var changed = true

// UserInput - A struct that we combine with our Go template to produce desired HTML
type UserInput struct {
	Name   string
	FavNum string
}

// HTMLFiles - Struct specifically for adding html resources like css
type HTMLFiles struct {
	fs http.FileSystem
}

// Main simply calls our launch method which inits our web server
func main() {
	launch()
}

// Launches our web server
func launch() {
	// Checks to see if lynks.txt exists - if it doesn't it is created.
	if _, err := os.Stat(lynxutil.HomePath + "/lynks.txt"); os.IsNotExist(err) {
		os.Create(lynxutil.HomePath + "lynks.txt")
	}

	fmt.Println("Starting server on http://localhost:" + lynxutil.GUIPort)

	fs := HTMLFiles{http.Dir("js/")}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(fs)))

	css := HTMLFiles{http.Dir("css/")}
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(css)))

	http.Handle("/images/", http.StripPrefix("/images/",
		http.FileServer(http.Dir("images/"))))

	http.HandleFunc("/home", IndexHandler)
	http.HandleFunc("/joinlynx", JoinHandler)
	http.HandleFunc("/createlynx", CreateHandler)
	http.HandleFunc("/removelynx", RemoveHandler)
	http.HandleFunc("/settings", SettingsHandler)
	http.HandleFunc("/uploads", UploadHandler)
	http.HandleFunc("/downloads", DownloadHandler)
	http.HandleFunc("/", SplashHandler)
	http.HandleFunc("/files", FileHandler)
	http.HandleFunc("/removefile", RemoveFileHandler)

	// Do jobs with params
	//gocron.Every(30).Second().Do(checkLynks)
	//<-gocron.Start()

	go cronWrapper()

	go server.Listen()

	go tracker.Listen()

	http.ListenAndServe(":"+lynxutil.GUIPort, nil)
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
	t, err := t.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
	}
	//t,_ = t.ParseFiles("index.html")
	tableEntries := TablePopulate(lynxutil.HomePath + "/lynks.txt")
	//fmt.Println(client.GetFileTableIndex())

	if client.GetFileTableIndex() > -1 {
		fileEntry := FilePopulate(client.GetFileTableIndex())
		t.ExecuteTemplate(rw, "index.html", map[string]template.HTML{"Entries": template.HTML(tableEntries),
			"Files": template.HTML(fileEntry)})

	} else {
		t.ExecuteTemplate(rw, "index.html", map[string]template.HTML{"Entries": template.HTML(tableEntries)})

	}

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
	err := client.JoinLynk(metapath[0])
	if err != nil {
		fmt.Println(err.Error())
	}

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
		index, _ := strconv.Atoi(name[0])
		//fmt.Println("in here" + name[0])
		client.DeleteLynk(client.GetLynkNameFromIndex(index), false)
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
	//fmt.Println(form) //returns an array of strings
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
func SplashHandler(rw http.ResponseWriter, req *http.Request) {
	t := template.New("cool template")
	t, err := t.ParseFiles("splash.html")
	if err != nil {
		fmt.Println(err)
	}

	t.ExecuteTemplate(rw, "splash.html","");


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
		if i == client.GetFileTableIndex() {
			tableEntries += "<td><b style= \"color:blue;\">" + split[0] + "</b></td>\n"
		} else {
			tableEntries += "<td>" + split[0] + "</td>\n"
		}
		tableEntries += "<td><form id=\"remove\" method=\"POST\" action=\"/removelynx\"><button " +
			"type=\"submit\" class=\"transparent\" data-toggle=\"tooltip\" data-placement = \"bottom\" title = \"Delete this Lynk\" id=\"removelynk\"\n " +
			"input type=\"hidden\" name=\"index\" value=\"" + rowStringNum + "\"><img src=\"images/file-ex-red.png\"></button></form></td>\n"
		tableEntries += "<form name=\"row" + rowStringNum + "form\" id=\"row" + rowStringNum + "formid\" " +
			"method=\"POST\" action=\"/files\"><input type=\"hidden\" name=\"index\" value=\"" + rowStringNum + "\"></form>"
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
	if index < client.GetLynksLen() {

		client.PopulateFilesAndSize()
		lynks := client.GetLynks()
		fileEntries := ""
		tempLynk := lynks[index]
		//fmt.Println(tempLynk.Files)
		fmt.Println("file pop")
		fileNames := tempLynk.Files
		client.SetFileTableIndex(index)
		i := 0

		for i < len(fileNames) {
			fileEntries += "<tr> \n"
			fileEntries += "<td>" + fileNames[i].Name + "</td>\n"
			fileEntries += "<td>" + strconv.Itoa(fileNames[i].Length) + "</td>\n"
			fileEntries += "<td><form id=\"remove\" method=\"POST\" action=\"removefile\"> \n" +
				"<button type=\"submit\" class=\"transparent\" data-toggle=\"tooltip\"" +
				" data-placement=\"bottom\" \n" +
				"title=\"Delete this file\" input type=\"hidden\" name=\"index\" value=\"" +
				strconv.Itoa(i) + "\" ><img src=\"images/file-ex-red.png\"></button></form>"
			fileEntries += "</tr>\n"
			i++
		}
		//fmt.Println(fileEntries)

		return fileEntries
	}
	return ""

}

func FileHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	form = req.Form
	index := form["index"]

	t := template.New("cool template")
	t, err := t.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
	}

	indexInt, _ := strconv.Atoi(index[0])
	fileEntry := FilePopulate(indexInt)
	tableEntries := TablePopulate(lynxutil.HomePath + "/lynks.txt")
	tableTemplate := template.HTML(tableEntries)
	t.ExecuteTemplate(rw, "index.html", map[string]template.HTML{"Entries": tableTemplate,
		"Files": template.HTML(fileEntry)})

}

func RemoveFileHandler(rw http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	form = req.Form
	name := form["index"]
	if name != nil {
		index, _ := strconv.Atoi(name[0])
		//fmt.Println(index)
		client.DeleteFileIndex(index, client.GetFileTableIndex())
		lynk := client.GetLynkNameFromIndex(client.GetFileTableIndex())
		//client.DeleteLynk(client.GetLynkNameFromIndex(client.GetFileTableIndex()))
		client.CreateMeta(lynk)
		server.PushMeta(lynxutil.HomePath + lynk + "/meta.info")
		//tracker.CreateSwarm(lynk)
		//TablePopulate(lynxutil.HomePath + "/lynks.txt")
	}

	IndexHandler(rw, req)

}

// Function INIT runs before main and allows us to load the index html before any operations
// are done on it and find the root directory on the user's computer
func init() {
	uploads, _ = ioutil.ReadFile("uploads.html")
	downloads, _ = ioutil.ReadFile("downloads.html")
}

// Function which checks the files in a directory to see if any have been added / changed
// @param path string - the path where the root directory is located
// @param file os.FileInfo - each file within the root or inner directories
// @param err error - any error we way encoutner along the way
// @return error - An error can produced if we encounter an invalid file.
func checkFiles(path string, file os.FileInfo, err error) error {

	inMeta := false
	for _, f := range currentLynk.Files {
		// Checks that the file is in the meta.info and that it matches the size listed in the meta
		if f.Name == file.Name() && f.Length == int(file.Size()) {
			inMeta = true
		}
	}

	// Don't add directories, trackers, or a meta.info file to the new meta.info
	if !file.IsDir() && !strings.Contains(path, "_Tracker") && file.Name() != "meta.info" && !inMeta {
		fmt.Println("Changed ", file.Name())
		changed = true
	}

	return nil
}

// Helper function that we use to check to see if our Lynk has changed
func checkLynks() {
	// Loops through every Lynk to check to see if their files have changed
	for _, lynk := range client.GetLynks() {
		fmt.Println("Checking..." + lynk.Name)
		changed = false
		// Sets currentLynk so it can be used in checkFiles
		currentLynk = lynk
		// Sets changed to true if any files have been changed
		filepath.Walk(lynxutil.HomePath+lynk.Name, checkFiles)
		if changed {
			client.CreateMeta(lynk.Name)
			server.PushMeta(lynxutil.HomePath + lynk.Name + "/meta.info")
		}
	}
}

// Helper function that wraps around our cron call so we can call it in a goroutine
func cronWrapper() {
	s := gocron.NewScheduler()
	s.Every(10).Seconds().Do(checkLynks)
	<-s.Start()
}
