#!/bin/bash

go get github.com/jasonlvhit/gocron
go get github.com/skratchdot/open-golang/open
go get golang.org/x/crypto/openpgp
echo Downloaded Required Packages

cd client
go install
echo Client Installed
cd ..

cd server
go install
echo Server Installed
cd ..

cd tracker
go install
echo Tracker Installed
cd ..

cd servDriver
go install
echo ServDriver Installed
cd ..

cd trackDriver
go install
echo TrackDriver Installed
cd ..

cd guiserver
go install
echo GUIServer Installed
cd ..

cd lynxutil
go install
echo Lynxutil Installed
cd ..

cd mycrypt
go install
echo Mycrypt Installed
cd ..

cd mypgp
go install
echo Mypgp Installed
cd ..

cd guiserver
echo Starting Lynx...
go run guiserver.go
exit
