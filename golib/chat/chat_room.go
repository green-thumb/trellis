package chat


//struct needed to help broadcast a message to selected clients


//function to create a room with the passed in id
func CreateChatRoom(id int) *ChatRoom {
  return &ChatRoom{
    Id: id,
    Broadcast:   make(chan *BroadcastStruct),
    Register:    make(chan *Chatter),
    Unregister:  make(chan *Chatter),
    Chatters: make(map[int]*Chatter),
  }
}

//function that monitors channels
func (room *ChatRoom) Run() {
  //run indefinitely and look for values in the channels
  for {
    select {

      //put chatter in chat room
      case p := <-room.Register:
        room.Chatters[p.Id] = p

      //remove chatter from chat room and close the socket
      case p := <-room.Unregister:
        if _, ok := room.Chatters[p.Id]; ok {
          delete(room.Chatters, p.Id)
          close(p.Send)
        }

      //send messages to selected clients
      case m := <-room.Broadcast:

        //send to specified targets
        if m.BroadcastType == 0 { 

          //loop through the TargetChatters map/hash table to get the chatters to send to
          for _, p := range m.TargetChatters {
            select {
            case p.Send <- m.Message:
            default:
              delete(room.Chatters, p.Id)
              close(p.Send)
            }
          }   

        //send to everyone in the room
        } else if m.BroadcastType == 1 { 

          //loop through the chatters map/hash table of the room to get the chatters to send to
          for _, p := range room.Chatters {
            select {
            case p.Send <- m.Message:
            default:
              delete(room.Chatters, p.Id)
              close(p.Send)
            }
          }  

        //send to everyone in the room except for the specified targets
        } else if(m.BroadcastType == 2) {

          //loop through the chatters map/hash table of the room to get the chatters to send to
          for _, p := range room.Chatters {
            _, ok := m.TargetChatters[p.Id]
            if ok {
              //don't do anything if chatter is in the TargetChatters map/hash table
            } else {
              //else send a message to the chatter
              select {
              case p.Send <- m.Message:
              default:
                delete(room.Chatters, p.Id)
                close(p.Send)
              }
            }
          }   

        }

    } //end select statement
  } //end for loop
} //end run function
