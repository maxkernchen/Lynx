package main

import (
"fmt"
"io"
"os"

	"strconv"
)

var metamap map[string]map[string]string
var infodict map[string]string

func metainfo(src string){
	metamap = make(map[string]map[string]string)
	infodict = make(map[string]string)

	fileInfo, err := os.Stat(src)
	if err != nil {
	}
	metamap["announce"] = map[string]string{"announceInner": "url of tracker annouce"}
	temp := fileInfo.ModTime().String()
	metamap["creation_date"] =  map[string]string{"creation_dateInner": temp}

	infodict["length"] = strconv.FormatInt(fileInfo.Size(), 10) //covert to string
	infodict["name"] = fileInfo.Name()
	infodict["piece_length"] = "fileInfo.Size()/constant"
	infodict["pieces"] = "sha1.Sum(chunckBuffer)"
	metamap["info"] = infodict

	fmt.Println(metamap["announce"])
	fmt.Println(metamap["creation_date"])

	fmt.Println(metamap["info"])


}

func fileCopy(src, dst string) error {
	in, err := os.Open(src) // Opens input
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst) // Opens output
	if err != nil {
		return err
	}
	//defer out.Close()

	_, err = io.Copy(out, in) // Copies the file contents
	if err != nil {
		return err
	}

	cerr := out.Close() // Checks for close error
	return cerr
}

func main() {
	/*fmt.Println("Hello World!")
	fmt.Println("Cool Beans!")
	err := fileCopy(os.Args[1], os.Args[2])
	
	if err != nil {
		fmt.Println("You suck")
	} else {
		fmt.Println(os.Args[1] + " copied to " + os.Args[2])
	}
	*/

	metainfo(os.Args[1]);
}
