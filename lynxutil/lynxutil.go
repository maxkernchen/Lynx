// Package lynxutil contains all the common functions / structs Lynx uses throughout it's classes.
// @author: Michael Bruce
// @author: Max Kernchen
// @verison: 4/30/2016
package lynxutil

import (
	"capstone/mypgp"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"strings"
	"time"
)

// ServerPort - The Default Port For The Lynx Server
const ServerPort = "8080"

// TrackerPort - The Default Port For The Lynx Tracker
const TrackerPort = "9000"

// SockErr - Represents A Welcome Socket Error
const SockErr = -1

// ReconnAttempts - Represents The Maximum Numbers Of Reconnection Attempts Lynx Will Make
const ReconnAttempts = 3

// HomePath - The absolute path of the user's Lynx directory
var HomePath string

// PrivateKey - This is the armored string that represents our private OpenPGP Key.
var PrivateKey string

// PublicKey - This is the armored string that represents our public OpenPGP Key.
var PublicKey string

// Peer - A struct which represents a Peer of the client
type Peer struct {
	IP   string
	Port string
	Key  string
}

// Lynk - A struct which holds all the information about a specific Lynk.
type Lynk struct {
	Name      string
	Owner     string
	Synced    string
	Tracker   string
	Files     []File
	Peers     []Peer
	FileNames []string
	FileSize  []int
	DLing     bool
}

// File - A struct based which represents a File in a Lynk's directory. It is based
// upon BitTorrent protocol dictionaries
type File struct {
	Length      int
	Path        string // Might not need path
	Name        string
	Chunks      string
	ChunkLength int
}

// FileCopy - Copies a file from src to dst
// @param string src - the file that will be copied
// @param string dst - the destination of the file to be copied
// @return error - An error can be produced when issues arise from trying to access,
// create, and write from either the src or dst files - otherwise error will be nil.
func FileCopy(src, dst string) error {
	in, err := os.Open(src) // Opens input
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst) // Opens output
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in) // Copies the file contents
	if err != nil {
		return err
	}

	return out.Close() // Checks for close error
}

// GetIP - Finds the ip of the current pc
// @return error - The single string ip
func GetIP() string {
	var onlyfirstip = false //only need first ip address
	var ipstring = ""
	ifaces, err := net.Interfaces()
	for _, i := range ifaces {
		addrs, errI := i.Addrs()
		if errI != nil {
			fmt.Println(err)
		}
		for _, addrs := range addrs {
			if ipnet, ok := addrs.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					if !onlyfirstip {
						onlyfirstip = true
						ipstring = ipnet.IP.String()
					}

				}
			}

		}

	}
	if err != nil {
		fmt.Println(err)
	}
	return ipstring
}

// GetLynk - Simple helper method that checks a lynks array for specific lynk.
// @param l []Lynk - The lynks array
// @param lynkName string - The lynk we are checking for
func GetLynk(l []Lynk, lynkName string) *Lynk {
	for i, a := range l {
		if a.Name == lynkName {
			return &l[i]
		}
	}
	return nil // Don't have Lynk
}

// Listen - Creates a welcomeSocket that listens for TCP connections - once someone connects a
// goroutine is spawned to handle the request
// @param handler func(net.Conn) err - This is the function we want to use to handle a new
// connection
// @param func(net.Conn) error handler - This is the function we use to handle the requests we get
// @param func(net.Conn) error handler - This is the port we will listen on.
func Listen(handler func(net.Conn) error, port string) {
	fmt.Println("Listening on Port: " + port)

	welcomeSocket, wErr := net.Listen("tcp", ":"+port)
	if wErr != nil {
		fmt.Println("Could Not Create Server Welcome Socket - Aborting.")
		os.Exit(SockErr) // Cannot recover from not being able to generate welcomeSocket
	}

	var cErr error
	for cErr == nil {
		conn, cErr := welcomeSocket.Accept()
		if cErr != nil {
			// If a connection error occurs
			continue // To avoid calling handler
		}
		go handler(conn)
	}

}

// Function init runs as soon as this class is imported and allows us to setup our HomePath
func init() {
	currentusr, _ := user.Current()
	HomePath = currentusr.HomeDir + "/Lynx/"
	HomePath = strings.Replace(HomePath, "\\", "/", -1) // Replaces Windows "\" With Unix "/" in path
	config := mypgp.Config{Expiry: 365 * 24 * time.Hour}
	key, _ := mypgp.CreateKey(currentusr.Name, "openpgp:lynxkeys", currentusr.Name+"@lynx.com", &config)
	PublicKey = key.Public
	PrivateKey = key.Private
}
