package friend

import (
	"fmt"
)

// Hub maintains set of active connections, and "broadcasts" message to connecitons

type Hub struct {
	// register conneciton
	connections map[*friend_connection]bool /// this is from friend_connection.go
	connectionsId map[*friend_connection]string  // passed from friend_conn

	// inbound message from connections
	broadcast chan []byte

	// register requests from friend_connection.  
	register chan *friend_connection

	// Unregister requestsf from friend_connection
	unregister chan *friend_connection

	broadcastUsers chan string
}

var H = Hub{
	broadcast: make(chan []byte),
	register: make(chan *friend_connection),
	unregister: make(chan *friend_connection),
	connections: make(map[*friend_connection]bool),  // map of current connections?
	connectionsId: make(map[*friend_connection]string),  // change this to array?
	broadcastUsers: make(chan string,20),
}

func (H *Hub) run(){
	for {
		fmt.Println("H.run running...")
		select {
		case c := <- H.register:  // Receive from conn.go at serveWs.  Receive user friend_connection
			fmt.Println("registering c.userId...", c.userId)
			H.connections[c] = true
			H.connectionsId[c] = c.userId
			// need to send updated "connections" to broadcast
			H.broadcastUsers <- c.userId

		case c := <- H.unregister:  
			if _, ok := H.connections[c]; ok {
				fmt.Println("Unregistering...", c)

				delete(H.connections, c)
				close(c.send)

				H.broadcastUsers <- c.userId
				delete(H.connectionsId, c)
				close(c.sendUsers)
			}

		case <- H.broadcastUsers:
			// when new users registers, send new "userId connection list" to all members
			for c := range H.connections {
				fmt.Println("range of H.connectionsId, ", c.userId)
				select {
				case c.sendUsers <- H.connectionsId:  // for each of connection, send in connectionId map
				default:
					// close(c.sendUsers)
					// delete(H.connectionsId, c)
				}
			}

		case m := <- H.broadcast:  // received from conn.go at readPump (read from user to Hub)
			// BRYAN: might not need this since we are not getting any message from users
			fmt.Println("Broadcasting message: ", m)

			for c := range H.connections {
				select {
					case c.send <- m:   // when a message is received from a user, "send" that message to all connections. At writePump 
					default:
						close(c.send)
						delete(H.connections, c)
				}
			}
		}
	}
}

func RunH () {
	H.run()
}

