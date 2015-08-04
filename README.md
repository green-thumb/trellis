![ScreenShot](http://i.imgur.com/BfbiWio.png)

Easy package to deploy an intuitive community forum.
Powered with Golang server and MySQL database, Trellis aims to help you bring together and manage your communities

##Features
* Threads / Comments
* Search
* Upvotes & Downvotes
* Direct / Global Chat System using WebSockets
* Signup / Login
* User Authentication
* User Profiles
* Friends
* Geolocation

##To Develop:

1. Go to root directory (/) in terminal
2. Start mySQL server by typing 'mysql.server start'
3. Set up schema by typing 'mysql -u root < schema.sql'
4. Run "npm install" at root
4. Build the client files with "gulp" on root directory
5. Start the server by running "go install", "go build", and "./trellis"


##To Test:

1. Run 'gulp test'.
2. Spec files are inside /specs.

##Example:
http://54.149.59.170:8080/














