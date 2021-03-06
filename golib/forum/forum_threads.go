package forum

import (
  "fmt"
  "log"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "github.com/gorilla/sessions"  
  "strconv"  
  "time"
)


//struct containing a forum thread post's info
type ThreadPostInfoOutbound struct {
  Post_id int `json:"post_id"`
  Thread_id int `json:"thread_id"`
  User_id int `json:"user_id"`
  User_name string `json:"user_name"`  
  Contents string `json:"contents"`
  Rating int `json:"rating"`
  Creation_time time.Time `json:"creation_time"`
  Last_update_time time.Time `json:"last_update_time"`
}

//struct containing an array of forum thread posts
type ThreadPostCollectionOutbound struct {
  ThreadPosts []*ThreadPostInfoOutbound `json:"threadPosts"`
}

//struct containing popular topics
type PopularTopicCollectionOutbound struct {
  PopularTopics []*PopularTopicOutbound `json:"topics"`
}

//struct containing a popular tag and its count
type PopularTopicOutbound struct {
  Tag string `json:"tag"`
  Count int `json:"count"`
}


//struct containing a forum thread's info
type ForumThreadInfoOutbound struct {
  Thread_id int `json:"thread_id"`
  User_id int `json:"user_id"`
  User_name string `json:"user_name"`
  Title string `json:"title"`
  Body string `json:"body"` 
  Link string `json:"link"` 
  Tag string `json:"tag"` 
  Post_count int `json:"post_count"`
  Rating int `json:"rating"`
  Longitude float64 `json:"lng"`
  Latitude float64 `json:"lat"`    
  Creation_time time.Time `json:"creation_time"`
  Last_update_time time.Time `json:"last_update_time"`
  Last_post_time time.Time `json:"last_post_time"`
}

//struct containing an array of forum threads
type ForumThreadCollectionOutbound struct {
  ForumThreads []*ForumThreadInfoOutbound `json:"forumThreads"`
}


//TODO: handle panics/errors, as unhandled panics/errors will shut down the server
//TODO: make a checkerr function
//TODO: put duplicated code into functions


//TODO: Return correct status and message if session is invalid
//TODO: Return correct status and message if insert failed
//TODO: format retrieved datetime to javascript datetime
func CreateForumThread(w http.ResponseWriter, r *http.Request, db *sql.DB, store *sessions.CookieStore) { 

  fmt.Println("Creating forum thread...")

  //add headers to response
  w.Header()["access-control-allow-origin"] = []string{"http://localhost:8080"} //TODO: fix this?                                                           
  w.Header()["access-control-allow-methods"] = []string{"GET, POST, OPTIONS"}
  w.Header()["Content-Type"] = []string{"application/json"}

  //ignore options requests
  if r.Method == "OPTIONS" {
    fmt.Println("options request received")
    w.WriteHeader(http.StatusTemporaryRedirect)
    return
  }

  //check for session to see if client is authenticated
  session, err := store.Get(r, "flash-session")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fm := session.Flashes("message")
  if fm == nil {
    fmt.Println("Trying to create a thread as an invalid user")
    fmt.Fprint(w, "No flash messages")
    return
  }
  //session.Save(r, w)

  //get the user id and username from the cookie
  userid := session.Values["userid"].(int)
  //username := session.Values["username"].(string)

  //parse the body of the request into a string
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    panic(err)
  }
  //fmt.Println(string(body))
  
  //parse the JSON string body to get the forum thread info
  byt := body
  var dat map[string]interface{}
  if err := json.Unmarshal(byt, &dat); err != nil {
    panic(err)
  }
  thread_title := dat["title"].(string)
  thread_body := dat["body"].(string)

  var thread_link string
  if _, ok := dat["link"]; ok {
    thread_link = dat["link"].(string)
  }    

  var thread_tag string
  if _, ok := dat["tag"]; ok {
    thread_tag = dat["tag"].(string)
  }

  var thread_longitude float64
  if _, ok := dat["lng"]; ok {
    thread_longitude = dat["lng"].(float64)
  }

  var thread_latitude float64
  if _, ok := dat["lat"]; ok {
    thread_latitude = dat["lat"].(float64)
  }    

  //TODO: handle lat and long passed in as non float types

  //insert forum thread into database
  stmt, err := db.Prepare("insert into forum_threads (user_id, title, body, link, tag, longitude, latitude) values (?, ?, ?, ?, ?, ?, ?)")
  if err != nil {
    log.Fatal(err)
  }
  res, err := stmt.Exec(userid, thread_title, thread_body, thread_link, thread_tag, thread_longitude, thread_latitude)
  if err != nil {
    log.Fatal(err)
  }
  lastId, err := res.LastInsertId()
  if err != nil {
    log.Fatal(err)
  }
  rowCnt, err := res.RowsAffected()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Inserted thread " + thread_title + " into database. Last inserted ID = %d, rows affected = %d\n", lastId, rowCnt)

  //return 200 status to indicate success
  fmt.Println("about to write 200 header")
  w.Write([]byte("{\"thread_id\" : " + strconv.FormatInt(lastId, 10) + "}"))
}


//option: query by - 0) thread id, 1) user id, 2) all
//sortBy: sort by (does not apply to by thread id since by thread id is unique) - 0) rating, 1) datetime
//TODO: Return correct status and message if query failed
func GetForumThread(w http.ResponseWriter, r *http.Request, db *sql.DB, option int, sortBy int, pageNumber int, id int) {

  fmt.Println("Getting forum thread...")

  //add headers to response
  w.Header()["access-control-allow-origin"] = []string{"http://localhost:8080"} //TODO: fix this?                                                           
  w.Header()["access-control-allow-methods"] = []string{"GET, POST, OPTIONS"}
  w.Header()["Content-Type"] = []string{"application/json"}

  //ignore options requests
  if r.Method == "OPTIONS" {
    fmt.Println("options request received")
    w.WriteHeader(http.StatusTemporaryRedirect)
    return
  }

  //variable(s) to hold the returned values from the query
  var (
    queried_thread_id int
    queried_user_id int
    queried_user_name string
    queried_title string
    queried_body string    
    queried_link string
    queried_tag string
    queried_post_count int
    queried_rating int
    queried_longitude float64
    queried_latitude float64
    queried_creation_time time.Time
    queried_last_update_time time.Time
    queried_last_post_time time.Time
  )

  //change query based on option
  var dbQuery string 

  if option == 0 { //query by thread id
    dbQuery = "select thread_id, forum_threads.user_id, title, body, link, tag, post_count, rating, longitude, latitude, forum_threads.creation_time, forum_threads.last_update_time, last_post_time, user_name from forum_threads inner join users on forum_threads.user_id = users.user_id where thread_id = " + strconv.Itoa(id)   
  } else if option == 1 { //query by user id

    if sortBy == 0 { //get by rating

      dbQuery = "select thread_id, forum_threads.user_id, title, body, link, tag, post_count, rating, longitude, latitude, forum_threads.creation_time, forum_threads.last_update_time, last_post_time, user_name from forum_threads inner join users on forum_threads.user_id = users.user_id where forum_threads.user_id = " + strconv.Itoa(id) + " order by rating desc"

    } else { //get by creation time

      dbQuery = "select thread_id, forum_threads.user_id, title, body, link, tag, post_count, rating, longitude, latitude, forum_threads.creation_time, forum_threads.last_update_time, last_post_time, user_name from forum_threads inner join users on forum_threads.user_id = users.user_id where forum_threads.user_id = " + strconv.Itoa(id) + " order by creation_time desc"

    }

    if pageNumber >= 1 {
      //only get 25 threads per query, and get records based on page number
      limit := 25
      offset := (pageNumber - 1) * limit 

      dbQuery += " limit " + strconv.Itoa(limit) + " offset " + strconv.Itoa(offset)
    }    

  } else { //query all

    if sortBy == 0 { //get by rating

      dbQuery = "select thread_id, forum_threads.user_id, title, body, link, tag, post_count, rating, longitude, latitude, forum_threads.creation_time, forum_threads.last_update_time, last_post_time, user_name from forum_threads inner join users on forum_threads.user_id = users.user_id order by rating desc"

    } else { //get by creation time

      dbQuery = "select thread_id, forum_threads.user_id, title, body, link, tag, post_count, rating, longitude, latitude, forum_threads.creation_time, forum_threads.last_update_time, last_post_time, user_name from forum_threads inner join users on forum_threads.user_id = users.user_id order by creation_time desc"

    }

    if pageNumber >= 1 {
      //only get 25 threads per query, and get records based on page number
      limit := 25
      offset := (pageNumber - 1) * limit 

      dbQuery += " limit " + strconv.Itoa(limit) + " offset " + strconv.Itoa(offset)
    }    

  }

  //perform query and check for errors
  rows, err := db.Query(dbQuery)
  if err != nil {
    panic(err)
  } 

  //outbound object containing a collection of outbound objects for each forum thread
  forumThreadCollectionOutbound := ForumThreadCollectionOutbound{ForumThreads: make([]*ForumThreadInfoOutbound, 0)}

  //iterate through results of query
  for rows.Next() {
    //get the relevant information from the query results
    err = rows.Scan(&queried_thread_id, &queried_user_id, &queried_title, &queried_body, &queried_link, &queried_tag, &queried_post_count, &queried_rating, &queried_longitude, &queried_latitude, &queried_creation_time, &queried_last_update_time, &queried_last_post_time, &queried_user_name)
    if err != nil {
      panic(err)
    }

    //create outbound object for each row
    forumThreadInfoOutbound := ForumThreadInfoOutbound{Thread_id: queried_thread_id, User_id: queried_user_id, User_name: queried_user_name, Title: queried_title,
      Body: queried_body, Link: queried_link, Tag: queried_tag, Post_count: queried_post_count, Longitude: queried_longitude, Latitude: queried_latitude,
      Rating: queried_rating, Creation_time: queried_creation_time, Last_update_time: queried_last_update_time, Last_post_time: queried_last_post_time}

    //add each outbound object to the collection outbound object
    forumThreadCollectionOutbound.ForumThreads = append(forumThreadCollectionOutbound.ForumThreads, &forumThreadInfoOutbound)
  }

  //json stringify the data
  jsonString, err := json.Marshal(forumThreadCollectionOutbound)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(jsonString))      

  //return 200 status to indicate success
  fmt.Println("about to write 200 header")
  w.Write(jsonString)

}

//sortBy: sort by - 0) rating, 1) datetime
//TODO: Return correct status and message if session is invalid
//TODO: Return correct status and message if query failed
func GetForumThreadProtected(w http.ResponseWriter, r *http.Request, db *sql.DB, store *sessions.CookieStore, sortBy int, pageNumber int) {

  fmt.Println("Getting forum thread...")

  //add headers to response
  w.Header()["access-control-allow-origin"] = []string{"http://localhost:8080"} //TODO: fix this?                                                           
  w.Header()["access-control-allow-methods"] = []string{"GET, POST, OPTIONS"}
  w.Header()["Content-Type"] = []string{"application/json"}

  //ignore options requests
  if r.Method == "OPTIONS" {
    fmt.Println("options request received")
    w.WriteHeader(http.StatusTemporaryRedirect)
    return
  }

  //check for session to see if client is authenticated
  session, err := store.Get(r, "flash-session")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fm := session.Flashes("message")
  if fm == nil {
    fmt.Println("Trying to get forum thread info as an invalid user")
    fmt.Fprint(w, "No flash messages")
    return
  }
  //session.Save(r, w)

  //get the user from the cookie
  userid := session.Values["userid"].(int)  

  //variable(s) to hold the returned values from the query
  var (
    queried_thread_id int
    queried_user_id int
    queried_user_name string
    queried_title string
    queried_body string    
    queried_link string
    queried_tag string
    queried_post_count int
    queried_rating int
    queried_longitude float64
    queried_latitude float64   
    queried_creation_time time.Time
    queried_last_update_time time.Time
    queried_last_post_time time.Time
  )

  //change query based on option
  var dbQuery string 
  if sortBy == 0 { 

    dbQuery = "select thread_id, forum_threads.user_id, title, body, link, tag, post_count, rating, longitude, latitude, forum_threads.creation_time, forum_threads.last_update_time, last_post_time, user_name from forum_threads inner join users on forum_threads.user_id = users.user_id where forum_threads.user_id = " + strconv.Itoa(userid) + " order by rating desc"
  
  } else {

    dbQuery = "select thread_id, forum_threads.user_id, title, body, link, tag, post_count, rating, longitude, latitude, forum_threads.creation_time, forum_threads.last_update_time, last_post_time, user_name from forum_threads inner join users on forum_threads.user_id = users.user_id where forum_threads.user_id = " + strconv.Itoa(userid) + " order by creation_time desc" 

  }

  if pageNumber >= 1 {
    //only get 25 threads per query, and get records based on page number
    limit := 25
    offset := (pageNumber - 1) * limit 

    dbQuery += " limit " + strconv.Itoa(limit) + " offset " + strconv.Itoa(offset)
  }  

  //perform query and check for errors
  rows, err := db.Query(dbQuery)
  if err != nil {
    panic(err)
  } 

  //outbound object containing a collection of outbound objects for each forum thread
  forumThreadCollectionOutbound := ForumThreadCollectionOutbound{ForumThreads: make([]*ForumThreadInfoOutbound, 0)}

  //iterate through results of query
  for rows.Next() {
    //get the relevant information from the query results
    err = rows.Scan(&queried_thread_id, &queried_user_id, &queried_title, &queried_body, &queried_link, &queried_tag, &queried_post_count, &queried_rating, &queried_longitude, &queried_latitude, &queried_creation_time, &queried_last_update_time, &queried_last_post_time, &queried_user_name)
    if err != nil {
      panic(err)
    }

    //create outbound object for each row
    forumThreadInfoOutbound := ForumThreadInfoOutbound{Thread_id: queried_thread_id, User_id: queried_user_id, User_name: queried_user_name, Title: queried_title,
      Body: queried_body, Link: queried_link, Tag: queried_tag, Post_count: queried_post_count, Longitude: queried_longitude, Latitude: queried_latitude,
      Rating: queried_rating, Creation_time: queried_creation_time, Last_update_time: queried_last_update_time, Last_post_time: queried_last_post_time}

    //add each outbound object to the collection outbound object
    forumThreadCollectionOutbound.ForumThreads = append(forumThreadCollectionOutbound.ForumThreads, &forumThreadInfoOutbound)
  }

  //json stringify the data
  jsonString, err := json.Marshal(forumThreadCollectionOutbound)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(jsonString))      

  //return 200 status to indicate success
  fmt.Println("about to write 200 header")
  w.Write(jsonString)

}


//TODO: Return correct status and message if session is invalid
//TODO: Return correct status and message if query failed
func ScoreForumThread(w http.ResponseWriter, r *http.Request, db *sql.DB, store *sessions.CookieStore, option int, thread_id int) {

  fmt.Println("Score forum thread...")

  //add headers to response
  w.Header()["access-control-allow-origin"] = []string{"http://localhost:8080"} //TODO: fix this?                                                           
  w.Header()["access-control-allow-methods"] = []string{"GET, POST, OPTIONS"}
  w.Header()["Content-Type"] = []string{"application/json"}

  //ignore options requests
  if r.Method == "OPTIONS" {
    fmt.Println("options request received")
    w.WriteHeader(http.StatusTemporaryRedirect)
    return
  }

  //check for session to see if client is authenticated
  session, err := store.Get(r, "flash-session")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fm := session.Flashes("message")
  if fm == nil {
    fmt.Println("Trying to vote on forum thread as an invalid user")
    fmt.Fprint(w, "No flash messages")
    return
  }
  //session.Save(r, w)

  //get the user id and username from the cookie
  userid := session.Values["userid"].(int)
  //username := session.Values["username"].(string)  

  //variable(s) to hold the returned values from the query
  var (
    queried_score int
  )

  //query the thread_votes table for the thread id and user id
  err = db.QueryRow("select score from thread_votes where thread_id = ? and user_id = ?", thread_id, userid).Scan(&queried_score)
  switch {

    //if record doesn't exist   
    case err == sql.ErrNoRows:
      //insert a new row to indicate that the user has voted for the thread
      stmt, err := db.Prepare("insert into thread_votes (thread_id, user_id, score) values (?, ?, ?)")
      if err != nil {
        log.Fatal(err)
      }
      res, err := stmt.Exec(thread_id, userid, option)
      if err != nil {
        log.Fatal(err)
      }
      fmt.Printf("Inserted record into thread_votes table.\n")

      //update the forum thread by the score
      stmt, err = db.Prepare("update forum_threads set rating = rating + ? where thread_id = ?")
      if err != nil {
        log.Fatal(err)
      }
      res, err = stmt.Exec(option, thread_id)
      if err != nil {
        log.Fatal(err)
      }
      rowCnt, err := res.RowsAffected()
      if err != nil {
        log.Fatal(err)
      }
      fmt.Printf("Updated score of thread " + strconv.Itoa(thread_id) + ". Rows affected = %d\n", rowCnt)     

      //update rep of user who created forum thread
      stmt, err = db.Prepare("update users inner join forum_threads on users.user_id = forum_threads.user_id set rep = rep + 2 * ? where forum_threads.thread_id = ?")
      if err != nil {
        log.Fatal(err)
      }
      res, err = stmt.Exec(option, thread_id)
      if err != nil {
        log.Fatal(err)
      }
      rowCnt, err = res.RowsAffected()
      if err != nil {
        log.Fatal(err)
      }
      fmt.Printf("Updated rep of forum thread creator. Rows affected = %d\n", rowCnt)              

      //return 200 status to indicate success
      fmt.Println("about to write 200 header")
      w.WriteHeader(http.StatusOK)

      break

    //if error querying database  
    case err != nil:
      log.Fatal(err)
      //return 400 status to indicate error
      fmt.Println("about to write 400 header")
      w.Write([]byte(fmt.Sprintf("Error querying database")))  
      break

    //if record exists
    default:
      if queried_score == -1 && option == 1 || queried_score == 0 && option == 1 || queried_score == 0 && option == -1  || queried_score == 1 && option == -1 {
        //update thread_votes table for the thread id and user id
        stmt, err := db.Prepare("update thread_votes set score = ? where thread_id = ? and user_id = ?")
        if err != nil {
          log.Fatal(err)
        }
        _, err = stmt.Exec(queried_score + option, thread_id, userid)
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Updated record in thread_votes table.\n")
      } else
      {
        //return 400 status to indicate error
        fmt.Println("about to write 400 header")
        fmt.Println("Cannot upvote twice or downvote twice")     
        w.Write([]byte(fmt.Sprintf("Cannot upvote twice or downvote twice")))   
        return
      }

      //update the forum thread by the score
      stmt, err := db.Prepare("update forum_threads set rating = rating + ? where thread_id = ?")
      if err != nil {
        log.Fatal(err)
      }
      res, err := stmt.Exec(option, thread_id)
      if err != nil {
        log.Fatal(err)
      }
      rowCnt, err := res.RowsAffected()
      if err != nil {
        log.Fatal(err)
      }
      fmt.Printf("Updated score of thread " + strconv.Itoa(thread_id) + ". Rows affected = %d\n", rowCnt)

      //update rep of user who created forum thread
      stmt, err = db.Prepare("update users inner join forum_threads on users.user_id = forum_threads.user_id set rep = rep + 2 * ? where forum_threads.thread_id = ?")
      if err != nil {
        log.Fatal(err)
      }
      res, err = stmt.Exec(option, thread_id)
      if err != nil {
        log.Fatal(err)
      }
      rowCnt, err = res.RowsAffected()
      if err != nil {
        log.Fatal(err)
      }
      fmt.Printf("Updated rep of forum thread creator. Rows affected = %d\n", rowCnt)               

      //return 200 status to indicate success
      fmt.Println("about to write 200 header")
      w.WriteHeader(http.StatusOK)
      
      break
  }

}

//TODO: Return correct status and message if session is invalid
//TODO: Return correct status and message if query failed
func EditForumThread(w http.ResponseWriter, r *http.Request, db *sql.DB, store *sessions.CookieStore) {

  fmt.Println("Edit forum thread...")

  //add headers to response
  w.Header()["access-control-allow-origin"] = []string{"http://localhost:8080"} //TODO: fix this?                                                           
  w.Header()["access-control-allow-methods"] = []string{"GET, POST, OPTIONS"}
  w.Header()["Content-Type"] = []string{"application/json"}

  //ignore options requests
  if r.Method == "OPTIONS" {
    fmt.Println("options request received")
    w.WriteHeader(http.StatusTemporaryRedirect)
    return
  }

  //check for session to see if client is authenticated
  session, err := store.Get(r, "flash-session")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fm := session.Flashes("message")
  if fm == nil {
    fmt.Println("Trying to edit forum thread as an invalid user")
    fmt.Fprint(w, "No flash messages")
    return
  }
  //session.Save(r, w)

  //get the user id and username from the cookie
  userid := session.Values["userid"].(int)    

  //parse the body of the request into a string
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    panic(err)
  }
  //fmt.Println(string(body))

  //parse the JSON string body to get the forum thread info
  byt := body
  var dat map[string]interface{}
  if err := json.Unmarshal(byt, &dat); err != nil {
    panic(err)
  }
  thread_id := int(dat["thread_id"].(float64))
  thread_title := dat["title"].(string)
  thread_body := dat["body"].(string)
  thread_link := dat["link"].(string)
  thread_tag := dat["tag"].(string)
  thread_longitude:= dat["lng"].(float64)
  thread_latitude := dat["lat"].(float64)  

  var (
    queried_user_id int
  )  

  //don't edit the thread if the user was not the one who created it
  err = db.QueryRow("select user_id from forum_threads where thread_id = ?", thread_id).Scan(&queried_user_id)
  switch {

    //if thread doesn't exist 
    case err == sql.ErrNoRows:
      //return 400 status to indicate error
      fmt.Println("about to write 400 header")
      fmt.Println("Thread cannot be found")     
      w.Write([]byte(fmt.Sprintf("Thread cannot be found"))) 
      return
      //break

    //if error querying database  
    case err != nil:
      log.Fatal(err)
      //return 400 status to indicate error
      fmt.Println("about to write 400 header")
      w.Write([]byte(fmt.Sprintf("Error querying database")))  
      return
      //break

    //if thread exists
    default:
      if queried_user_id != userid {
        fmt.Println("about to write 400 header")
        fmt.Println("Cannot edit another user's thread") 
        w.Write([]byte(fmt.Sprintf("Cannot edit another user's thread")))  
        return
      }
      break

  }    

  //TODO: return error if thread id is blank/nan

  //update the forum thread post
  stmt, err := db.Prepare("update forum_threads set title = ?, body = ?, link = ?, tag = ?, longitude = ?, latitude = ? where thread_id = ?")
  if err != nil {
    log.Fatal(err)
  }
  res, err := stmt.Exec(thread_title, thread_body, thread_link, thread_tag, thread_longitude, thread_latitude, thread_id)
  if err != nil {
    log.Fatal(err)
  }
  rowCnt, err := res.RowsAffected()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Updated contents of forum thread " + strconv.Itoa(thread_id) + ". Rows affected = %d\n", rowCnt)      

  //return 200 status to indicate success
  fmt.Println("about to write 200 header")
  w.WriteHeader(http.StatusOK)

}

//TODO: Return correct status and message if session is invalid
//TODO: Return correct status and message if query failed
func DeleteForumThread(w http.ResponseWriter, r *http.Request, db *sql.DB, store *sessions.CookieStore, id int) {

  fmt.Println("Delete forum thread...")

  //add headers to response
  w.Header()["access-control-allow-origin"] = []string{"http://localhost:8080"} //TODO: fix this?                                                           
  w.Header()["access-control-allow-methods"] = []string{"GET, POST, OPTIONS"}
  w.Header()["Content-Type"] = []string{"application/json"}

  //ignore options requests
  if r.Method == "OPTIONS" {
    fmt.Println("options request received")
    w.WriteHeader(http.StatusTemporaryRedirect)
    return
  }

  //check for session to see if client is authenticated
  session, err := store.Get(r, "flash-session")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fm := session.Flashes("message")
  if fm == nil {
    fmt.Println("Trying to delete forum thread as an invalid user")
    fmt.Fprint(w, "No flash messages")
    return
  }
  //session.Save(r, w)

  //get the user id and username from the cookie
  userid := session.Values["userid"].(int)

  var (
    queried_user_id int
  )  

  //don't delete the thread if the user was not the one who created it
  err = db.QueryRow("select user_id from forum_threads where thread_id = ?", id).Scan(&queried_user_id)
  switch {

    //if thread doesn't exist 
    case err == sql.ErrNoRows:
      //return 400 status to indicate error
      fmt.Println("about to write 400 header")
      fmt.Println("Thread cannot be found")     
      w.Write([]byte(fmt.Sprintf("Thread cannot be found"))) 
      return
      //break

    //if error querying database  
    case err != nil:
      log.Fatal(err)
      //return 400 status to indicate error
      fmt.Println("about to write 400 header")
      w.Write([]byte(fmt.Sprintf("Error querying database")))  
      return
      //break

    //if thread exists
    default:
      if queried_user_id != userid {
        fmt.Println("about to write 400 header")
        fmt.Println("Cannot delete another user's thread") 
        w.Write([]byte(fmt.Sprintf("Cannot delete another user's thread")))  
        return
      }
      break

  }  

  //TODO: return error if thread id is blank/nan

  //delete all votes related to forum posts
  stmt, err := db.Prepare("delete post_votes from post_votes inner join thread_posts on post_votes.post_id = thread_posts.post_id where thread_posts.thread_id = ?")

  if err != nil {
    log.Fatal(err)
  }
  res, err := stmt.Exec(id)
  if err != nil {
    log.Fatal(err)
  }
  rowCnt, err := res.RowsAffected()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Deleted votes for forum posts with forum thread id " + strconv.Itoa(id) + ". Rows affected = %d\n", rowCnt)    

  //delete all forum posts related to the forum thread
  stmt, err = db.Prepare("delete from thread_posts where thread_id = ?")
  if err != nil {
    log.Fatal(err)
  }
  res, err = stmt.Exec(id)
  if err != nil {
    log.Fatal(err)
  }
  rowCnt, err = res.RowsAffected()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Deleted forum posts with forum thread id " + strconv.Itoa(id) + ". Rows affected = %d\n", rowCnt)  

  //delete all votes related to the forum thread
  stmt, err = db.Prepare("delete from thread_votes where thread_id = ?")

  if err != nil {
    log.Fatal(err)
  }
  res, err = stmt.Exec(id)
  if err != nil {
    log.Fatal(err)
  }
  rowCnt, err = res.RowsAffected()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Deleted votes for forum thread with id " + strconv.Itoa(id) + ". Rows affected = %d\n", rowCnt)      

  //delete the forum thread
  stmt, err = db.Prepare("delete from forum_threads where thread_id = ?")
  if err != nil {
    log.Fatal(err)
  }
  res, err = stmt.Exec(id)
  if err != nil {
    log.Fatal(err)
  }
  rowCnt, err = res.RowsAffected()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Deleted forum thread " + strconv.Itoa(id) + ". Rows affected = %d\n", rowCnt)      

  //return 200 status to indicate success
  fmt.Println("about to write 200 header")
  w.WriteHeader(http.StatusOK)

}

//TODO: Return correct status and message if query failed
func PopularThreads(w http.ResponseWriter, r *http.Request, db *sql.DB) {

  fmt.Println("Getting popular forum threads...")

  //add headers to response
  w.Header()["access-control-allow-origin"] = []string{"http://localhost:8080"} //TODO: fix this?                                                           
  w.Header()["access-control-allow-methods"] = []string{"GET, POST, OPTIONS"}
  w.Header()["Content-Type"] = []string{"application/json"}

  //ignore options requests
  if r.Method == "OPTIONS" {
    fmt.Println("options request received")
    w.WriteHeader(http.StatusTemporaryRedirect)
    return
  }

  //variable(s) to hold the returned values from the query
  var (
    queried_tag string
    queried_count int
  )  

  //get the 5 most popular tags within the past 24 hours
  //perform query and check for errors
  rows, err := db.Query("select tag, count(tag) from forum_threads where creation_time > DATE_SUB( NOW(), INTERVAL 24 HOUR) group by tag order by count(tag) desc limit 5")
  if err != nil {
    panic(err)
  } 

  //outbound object containing a collection of outbound objects for each forum thread
  popularTopicCollectionOutbound := PopularTopicCollectionOutbound{PopularTopics: make([]*PopularTopicOutbound, 0)}

  //iterate through results of query
  for rows.Next() {
    //get the relevant information from the query results
    err = rows.Scan(&queried_tag, &queried_count)
    if err != nil {
      panic(err)
    }

    //create outbound object for each row
    popularTopicOutbound := PopularTopicOutbound{Tag: queried_tag, Count: queried_count}

    //add each outbound object to the collection outbound object
    popularTopicCollectionOutbound.PopularTopics = append(popularTopicCollectionOutbound.PopularTopics, &popularTopicOutbound)
  }

  //json stringify the data
  jsonString, err := json.Marshal(popularTopicCollectionOutbound)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(jsonString))      

  //return 200 status to indicate success
  fmt.Println("about to write 200 header")
  w.Write(jsonString)


}


