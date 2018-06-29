# Lynx
Peer-to-Peer File Synchronization Application


Lynx was written for my college capstone project over two semesters, it was a partner project. 
My main responsibilities were the design and functionality of the UI and file parsing into struts and meta files. 
Currently I am making enhancements and bug fixes to all parts of Lynx which can be viewed here.

Some terms and processes I have defined below for clarification:

Lynx — the name of the application

Lynk — the name of a directory which has been either created or joined

Join a Lynk — By entering the location of the meta file the user joins the Lynk which was created on another machine. 
Files are then synchronized between all peers which have joined the Lynk.

Create a Lynk — By entering the exact name of the folder within the Links folder in the root of the user,
this directory can be shared amongst other machines which are running Lynx and have a copy of the meta.info file 
which is created when this process is completed. This file is within the same folder and must be shared to other
users to begin the 'Join a Lynk' process.

To install Lynx the user needs to create a folder named 'Lynx' at the root of the users folder.

Windows — C:\Users<UserName>\Lynx

Linux — /home//Lynx/

I am working on making this directory created automatically currently, but still need to do some testing to 
confirm this is safe to do on both Linux and Windows due to different directory structures.

Any folders created inside this folder can be turned into a Lynk by following the 'Create a Lynk' process.

On Windows, you should be able to run Lynx with the already compiled Lynx.exe within the guiserver package. 
Please do not move the location of this file else it will not be able to open the HTML and other files needed to start the application.

On Linux you will have to run the install.sh bash script to get all required packages and compile the code. 
Afterwards run the run.sh bash script to start the application. These steps assume you have installed golang and set
the correct environment variables as defined in the documentation at https://golang.org/doc/install

Below is a video which shows a working example of Lynx which should contain most information needed to run and use Lynx.
