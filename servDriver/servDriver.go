/**
 *
 *	 This is just a simple driver for our server. Only calls Listen()
 *
 *	 @author: Michael Bruce
 *	 @author: Max Kernchen
 *
 *	 @verison: 2/17/2016
 */

package main

import "capstone/server"

/**
 * Function used to drive and test our server's functions
 */
func main() {
	server.Listen(server.HandleFileRequest)
}
