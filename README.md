Websocket Rooms
===============

# Installation
go get -u github.com/gorilla/mux
go get github.com/gorilla/websocket
go get gopkg.in/mgo.v2

# Running the Server
Because the tests files live with everything else, go run *.go will not run the program. You can build it first by executing "go build",
then modify the permissions to be executable "chmod +x" and execute it by running "./rooms".

# Connecting the Client
The endpoint for the websocket locally will be ws://localhost:8080/ws/{"a uuid that can uniquely represent the device"}.
The fastest way to connect here would be to use https://websocket.org/echo.html.

