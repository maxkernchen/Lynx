Currently if client.go is run it will parse the meta.info file and print out the list of structs
that are retrieved from the meta.info. client_test.go can be run simply by typing "go test".
There are some more methods that are commented out in main in client.go, they have been tested
to work correctly but they depend upon the meta.info file being correct (i.e. if we remove
test.txt it should be in the meta.info and the list of structs).

This is a WIP as of 2/17/2016

-Michael Bruce
-Max Kernchen