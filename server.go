package main

import (
  "fmt"
  "net/http"
  "flag"    
  "log"
  "database/sql"
  _ "github.com/go-sql-driver/mysql" 
  "github.com/gorilla/sessions"   
  "net/url"
  "strconv"
  "strings"
  "docker_app/golib/friend"
  "docker_app/golib/chat"
  "docker_app/golib/forum"
  "docker_app/golib/message"
  "docker_app/golib/search"
  "docker_app/golib/user"
  "docker_app/golib/auth"
)

//"constant" variables to be used throughout the program
const (
  //database configuration information
  DB_USER = "root"
  DB_PASSWORD = ""
  DB_NAME = "virtual_arm"

  //int to represent an invalid selection
  INVALID_INT = -9999
)

//variables to be used throughout the program
var (
  //cookie information
  store = sessions.NewCookieStore([]byte("a-secret-string"))

  //server address information
  addr = flag.String("addr", ":8080", "http service address")  
)






type FriendInfoOutbound struct {
  User_id int `json:"id"`
  User_name string `json:"username"`
  First_name string `json:"first"`
  Last_name string `json:"last"`
}

type FriendsListOutbound struct {
  Friends []*FriendInfoOutbound `json:"friends"`
}

//struct containing properties needed to create a chatter
type ChatterHandler struct {
  Id int `json:"id"`
  Username string `json:"username"`
  Room *chat.ChatRoom `json:"room"`
}



func main() {
  flag.Parse()

  //open the database connection
  var db = initializeDB()
  defer db.Close() //defer closing the connection

  //create the game room
  //var room = createGameRoom(1)
  //go room.run()


  //serve static assets
  http.Handle("/", http.FileServer(http.Dir("./build")))


  //TODO: make urls more RESTful


  //routes in auth.go

//GET:
//user.GetUserInfo                                 profile/
//POST:
//updateUserInfo                              profile/                       body: {"bio" : bio, "avatar_link" : avatar_link}
  http.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        auth.GetUserInfoHandler(w, r, db, store)

        break 
      case "POST":

        auth.UpdateUserInfoHandler(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

//POST:
//createUser                                   users/                         body: {"username" : username, "password" : password, "firstname" : firstname, "lastname" : lastname}
  http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":
        break 
      case "POST":

        auth.CreateUserHandler(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

//POST:
//authenticate                                 authenticate/                  body: {"username" : username, "password" : password}  
  http.HandleFunc("/authenticate/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":
        break 
      case "POST":

        auth.LoginHandler(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  }) 


  //routes in forum_threads.go

//GET:
//forum.GetForumThreadByThreadId                     thread/ id
//POST:
//upvoteForumThread                            thread/XidX/?upvote=true       //id in url so don't need body
//downvoteForumThread                          thread/XidX/?downvote=true     //id in url so don't need body
//PUT:
//forum.EditForumThread                              thread/                        body: {"title" : title, "body" : body, "link" : link, "tag" : tag}
//DELETE:
//forum.DeleteForumThread                            thread/ id
  http.HandleFunc("/thread/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          threadId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          forum.GetForumThread(w, r, db, 0, INVALID_INT, INVALID_INT, threadId)
        } else {
          //error
        }
        
        break 
      case "POST":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for userid parameter in url
        option := int(0)
        if val, ok := m["upvote"]; ok { //TODO: case insensitve match
          if(val[0] == "true") {
            option = 1
          }
        } else if val, ok := m["downvote"]; ok {
          if(val[0] == "true") {
            option = -1
          }
        }

        if option == 1 || option == -1 {

          s := strings.Split(r.URL.Path, "/")
          if len(s) == 4 {
            threadId, err := strconv.Atoi(s[2])

            if err != nil {
              //error
            }

            forum.ScoreForumThread(w, r, db, store, option, threadId)
  
          } else {
            //error
          }

        }

        break  
      case "PUT":

        forum.EditForumThread(w, r, db, store)

        break  
      case "DELETE":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          threadId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          forum.DeleteForumThread(w, r, db, store, threadId)

        } else {
          //error
        }

        break  
      default:
        break
    }
  })

//GET:
//forum.GetForumThreadsByLoggedInUserIdByRating      profilethreads/ ? sortby = XXX & pagenumber = XXX
//forum.GetForumThreadsByLoggedInUserIdByTime        profilethreads/ ? sortby = XXX & pagenumber = XXX  
  http.HandleFunc("/profilethreads/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for sortby parameter in url
        sortBy := 0
        if val, ok := m["sortby"]; ok {
          if val[0] == "creationtime" {
            sortBy = 1
          }
        }

        //look for pagenumber parameter in url
        pageNumber := 0
        var err error
        if val, ok := m["pagenumber"]; ok {
          pageNumber, err = strconv.Atoi(val[0])
          if err != nil {
            //error
          }
        }

        forum.GetForumThreadProtected(w, r, db, store, sortBy, pageNumber)

        break 
      case "POST":
        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

//GET:  
//forum.GetForumThreadsByUserIdByRating              threads/ ? userid = XXX & sortby = XXX & pagenumber = XXX
//forum.GetForumThreadsByUserIdByTime                threads/ ? userid = XXX & sortby = XXX & pagenumber = XXX
//forum.GetForumThreadsByRating                      threads/ ? sortby = XXX & pagenumber = XXX
//forum.GetForumThreadsByTime                        threads/ ? sortby = XXX & pagenumber = XXX
//POST:
//forum.CreateForumThread                            threads/                       body: {"title" : title, "body" : body, "link" : link, "tag" : tag}
  http.HandleFunc("/threads/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for sortby parameter in url
        sortBy := 0
        if val, ok := m["sortby"]; ok {
          if val[0] == "creationtime" {
            sortBy = 1
          }
        }

        //look for pagenumber parameter in url
        pageNumber := 0
        var err error
        if val, ok := m["pagenumber"]; ok {
          pageNumber, err = strconv.Atoi(val[0])
          if err != nil {
            //error
          }
        }

        //select what to query by
        if val, ok := m["userid"]; ok { //by userid
          userId, err := strconv.Atoi(val[0])
          if err != nil {
            //error
          }

          forum.GetForumThread(w, r, db, 1, sortBy, pageNumber, userId)

        } else { //by all

          forum.GetForumThread(w, r, db, 2, sortBy, pageNumber, INVALID_INT)

        }

        break 
      case "POST":

        forum.CreateForumThread(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })

//GET:
//trending       trending/
  http.HandleFunc("/trending/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      forum.PopularThreads(w, r, db)

      break 
    case "POST":
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})


  //routes in thread_posts.go

//GET:
//getThreadPostByPostId                        post/ id
//POST:
//upvoteThreadPost                             post/XidX/?upvote=true         //id in url so don't need body
//downvoteThreadPost                           post/XidX/?downvote=true       //id in url so don't need body
//PUT:
//forum.EditThreadPost                               post/                          body: {"thread_id" : threadId, "contents" : contents}
//DELETE:
//forum.DeleteThreadPost                             post/ id
  http.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          postId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          forum.GetThreadPost(w, r, db, 0, INVALID_INT, INVALID_INT, postId)

        } else {
          //error
        }

        break 
      case "POST":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for userid parameter in url
        option := 0
        if val, ok := m["upvote"]; ok {
          if(val[0] == "true") {
            option = 1
          }
        } else if val, ok := m["downvote"]; ok {
          if(val[0] == "true") {
            option = -1
          }
        }

        if option == 1 || option == -1 {

          s := strings.Split(r.URL.Path, "/")
          if len(s) == 4 {
            threadId, err := strconv.Atoi(s[2])

            if err != nil {
              //error
            }

            forum.ScoreThreadPost(w, r, db, store, option, threadId)
  
          } else {
            //error
          }

        }

        break  
      case "PUT":

        forum.EditThreadPost(w, r, db, store)

        break  
      case "DELETE":

        s := strings.Split(r.URL.Path, "/")
        if len(s) == 3 {
          postId, err := strconv.Atoi(s[2])

          if err != nil {
            //error
          }

          forum.DeleteThreadPost(w, r, db, store, postId)

        } else {
          //error
        }

        break  
      default:
        break
    }
  })

//GET:
//getThreadPostsByThreadIdByRating             posts / ? threadid = XXX & sortby = rating & pagenumber = XXX
//getThreadPostsByThreadIdByTime               posts / ? threadid = XXX & sortby = creationtime & pagenumber = XXX
//getThreadPostsByUserIdByRating               posts / ? userid = XXX & sortby = rating & pagenumber = XXX
//getThreadPostsByUserIdByTime                 posts / ? userid = XXX & sortby = creationtime & pagenumber = XXX
//getThreadPostsByRating                       posts / ? sortby = rating & pagenumber = XXX
//getThreadPostsByTime                         posts / ? sortby = creationtime & pagenumber = XXX
//POST:
//forum.CreateThreadPost                             posts/                         body: {"thread_id" : threadId, "contents" : contents}
  http.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        //look for sortby parameter in url
        sortBy := 0
        if val, ok := m["sortby"]; ok {
          if val[0] == "creationtime" {
            sortBy = 1
          }
        }

        //look for pagenumber parameter in url
        pageNumber := 0
        var err error
        if val, ok := m["pagenumber"]; ok {
          pageNumber, err = strconv.Atoi(val[0])
          if err != nil {
            //error
          }
        }

        //select what to query by
        if val, ok := m["threadid"]; ok { //by threadid
          threadId, err := strconv.Atoi(val[0])
          if err != nil {
            //error
          }

          forum.GetThreadPost(w, r, db, 1, sortBy, pageNumber, threadId)
        } else if val, ok := m["userid"]; ok { //by userid
          userId, err := strconv.Atoi(val[0])
          if err != nil {
            //error
          }

          forum.GetThreadPost(w, r, db, 2, sortBy, pageNumber, userId)
        } else { //by all

          forum.GetThreadPost(w, r, db, 3, sortBy, pageNumber, INVALID_INT)
        }

        break 
      case "POST":

        forum.CreateThreadPost(w, r, db, store)

        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })


  //routes in users.go

//GET:
//getUserInfoByUserId                          user / 1
//getUserInfoByUsername                        user / ? username = XXX
  http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
      case "GET":

        m, _ := url.ParseQuery(r.URL.RawQuery)

        if val, ok := m["username"]; ok { //if username can be parsed from url

          username := val[0]

          user.GetUserInfo(w, r, db, 1, INVALID_INT, username)

        } else { //else if username cannot be parsed from url

          s := strings.Split(r.URL.Path, "/")
          if len(s) == 3 { //if path can be divided into 3 parts

            userId, err := strconv.Atoi(s[2]) //convert parsed user id into an int

            if err != nil { //if parsed user id could not be converted into an int
              //error
            }
            user.GetUserInfo(w, r, db, 0, userId, "")
          } else {
            //error
          }
        }

        break 
      case "POST":
        break  
      case "PUT":
        break  
      case "DELETE":
        break  
      default:
        break
    }
  })


  //routes in friend.go

  http.HandleFunc("/friend/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      friend.GetFriendsList(w, r, db, store)
      break 
    case "POST":
      m, _ := url.ParseQuery(r.URL.RawQuery)
      if val, ok := m["action"]; ok {
        if val[0] == "add" {
          friend.AddFriend(w, r, db, store)
        } else if val[0] == "remove" {
          friend.RemoveFriend(w, r, db, store)
        } else {
          //error
        }
      }
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})

//routes in search.go

//GET:
//search      search/?title=hello&sortby=rating&pagenumber=1
http.HandleFunc("/search/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      m, _ := url.ParseQuery(r.URL.RawQuery)

      //look for title parameter in url
      var title string
      if val, ok := m["title"]; ok {
        if len(val[0]) > 2 {
          title = val[0]
        } else {
          //error
          return
        }
      }

      //look for sortby parameter in url
      sortBy := 0
      if val, ok := m["sortby"]; ok {
        if val[0] == "creationtime" {
          sortBy = 1
        }
      }

      //look for pagenumber parameter in url
      pageNumber := 0
      var err error
      if val, ok := m["pagenumber"]; ok {
        pageNumber, err = strconv.Atoi(val[0])
        if err != nil {
          //error
        }
      }

      search.SearchForForumThreads(w, r, db, sortBy, pageNumber, title)

      break 
    case "POST":
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})


//routes in messages.go

//GET:
//getMessage                               message/ id 
//DELETE:
//deleteMessage                            message/ id
http.HandleFunc("/message/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      s := strings.Split(r.URL.Path, "/")
      if len(s) == 3 {
        messageId, err := strconv.Atoi(s[2])

        if err != nil {
          //error
        }

        message.RecvMessages(w, r, db, store, 0, INVALID_INT, INVALID_INT, messageId)

      } else {
        //error
      }

      break 
    case "POST":
      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})

//GET:
//getMessage                               messages/?q=sender&sortby=desc&pagenumber=1
//getMessage                               messages/?q=recipient&sortby=desc&pagenumber=1
//POST:
//sendMessage                              messages/                         body: { ... }    
http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":

      m, _ := url.ParseQuery(r.URL.RawQuery)

      //look for query type parameter in url
      option := 0
      if val, ok := m["q"]; ok {
        if(val[0] == "sender") {
          option = 1
        } else {
          option = 2
        }
      }

      //look for sortby parameter in url
      sortBy := 0
      if val, ok := m["sortby"]; ok {
        if val[0] == "asc" {
          sortBy = 1
        }
      }

      //look for pagenumber parameter in url
      pageNumber := 0
      var err error
      if val, ok := m["pagenumber"]; ok {
        pageNumber, err = strconv.Atoi(val[0])
        if err != nil {
          //error
        }
      }

      message.RecvMessages(w, r, db, store, option, sortBy, pageNumber, INVALID_INT)

      break 
    case "POST":

      message.CreateMessage(w, r, db, store)

      break  
    case "PUT":
      break  
    case "DELETE":
      break  
    default:
      break
  }
})

  // route for friend_list

  go friend.RunH() 
  http.HandleFunc("/friendlist/", func(w http.ResponseWriter, r *http.Request ) {
    currUserId := friend.CheckSession(w, r, store)  
   friend.ServeWs(w, r, db, currUserId)
  })


  var room = chat.CreateChatRoom(1)
  go room.Run()


  //listen for user chat
  http.HandleFunc("/chat/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("trying to initiate chat")
    chattingRoom(w, r, store, room)
  })


  //listen on specified port
  fmt.Println("Server starting")
  err := http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Fatal("ListenAndServe:", err)
  }

  // err := http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil)
  // if err != nil {
  //   log.Fatal("ListenAndServeTLS: ", err)
  // }
}

//function to open connection with database
func initializeDB() *sql.DB {
  db, err := sql.Open("mysql",  DB_USER + ":" + DB_PASSWORD + "@/" + DB_NAME + "?parseTime=true")
  if err != nil {
    panic(err)
  } 

  return db
}

//handle the chat event which checks if the cookie corresponds to a logged in user and adds the user to the chat room
//TODO: Return correct status and message if session is invalid
func chattingRoom(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, room *chat.ChatRoom) {

  //check for session to see if client is authenticated
  session, err := store.Get(r, "flash-session")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fm := session.Flashes("message")
  if fm == nil {
    fmt.Println("Trying to log in as invalid user")
    fmt.Fprint(w, "No flash messages")
    return
  }
  //session.Save(r, w)

  fmt.Println("New user connected to chat")

  //use the id and username attached to the session to create the player
  // chatterHandler := ChatterHandler{Id: session.Values["userid"].(int), Username: session.Values["username"].(string), Room: room}

  chat.CreateChatter(w, r, session.Values["userid"].(int), session.Values["username"].(string), room)
}



/*

//handle the connect event which checks if the cookie corresponds to a logged in user
//and creates the player in the game
//TODO: Return correct status and message if session is invalid
func connect(w http.ResponseWriter, r *http.Request, room *GameRoom, store *sessions.CookieStore) {

  //check for session to see if client is authenticated
  session, err := store.Get(r, "flash-session")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fm := session.Flashes("message")
  if fm == nil {
    fmt.Println("Trying to log in as invalid user")
    fmt.Fprint(w, "No flash messages")
    return
  }
  //session.Save(r, w)

  fmt.Println("New user connected")

  //use the id and username attached to the session to create the player
  playerHandler := PlayerHandler{Id: session.Values["userid"].(int), Username: session.Values["username"].(string), Room: room}

  playerHandler.createPlayer(w, r)
}

*/



