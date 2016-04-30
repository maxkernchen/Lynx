#!/usr/bin/env bash

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

exit
