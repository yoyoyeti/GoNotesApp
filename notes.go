package main

import (
  "fmt"
  "time"
	"net/http"
  "math/rand"
  "encoding/json"

  "golang.org/x/crypto/bcrypt"
  "github.com/gocql/gocql"
  "github.com/rs/cors"
  "github.com/gorilla/mux"
)

type User struct {
	Username string `json:"username"`
	Password  string `json:"password"`
}

type NoteToken struct {
	AuthToken string `json:"token"`
	NoteText  string `json:"text"`
}

type GetNotes struct{
  Notes []string `json:"text"`
}

//connecting to cassandra
var cluster = gocql.NewCluster("127.0.0.1")

//hashes user's password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

//checks to see if the user's password matches the stored hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//attempts to log the user in and sends a token
//which will allow them to add and delete notes if login is successful
func logIn(w http.ResponseWriter, r *http.Request){
  //create session with cassandra
  CQLsession, _ := cluster.CreateSession()
  defer CQLsession.Close()
  //gets the json from the user's message and puts it into an object
  var user User
  json.NewDecoder(r.Body).Decode(&user)

  if len(user.Username) > 0 {
    var passHash string
    var authToken string

    //gets the hashed password and authentication token of the user
    if err := CQLsession.Query(`SELECT passhash, auth_token FROM user WHERE username = ?`,
  		user.Username).Consistency(gocql.One).Scan(&passHash, &authToken); err != nil {
  		fmt.Fprintf(w, "login failed")
      return
	  } else{
      if len(user.Password) > 0{
        //checks to see if the password matches the hash
        if success := checkPasswordHash(user.Password, passHash); !success {
          fmt.Fprintf(w, "login failed")
          return
        }
      } else{
        fmt.Fprintf(w, "Missing Password")
        return
      }
    }

    //if the authentication token exists send it
    if len(authToken) > 0{
      fmt.Fprintf(w, authToken)
    } else{
      //else create it
      token := generateToken()

      //update the user to have new token
      if err := CQLsession.Query(`UPDATE user SET auth_token = ? WHERE username = ?`,
        token, user.Username).Exec(); err != nil {
        fmt.Fprintf(w, "error")
      } else{
        //send new token to user
        fmt.Fprintf(w, token)
      }
    }
  } else{
    fmt.Fprintf(w, "Missing Username")
    return
  }
}

//returns a string of random letters and numbers
func generateToken() string {
  const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

  b := make([]byte, 50)
  for i := range b {
      b[i] = letterBytes[rand.Intn(len(letterBytes))]
  }
  return string(b)
}

//creates a new user if one with that username doesn't already exist
func signUp(w http.ResponseWriter, r *http.Request){
  //gets the json the user sent and puts it in an object
  var user User
  json.NewDecoder(r.Body).Decode(&user)

  //makes sure the user actually sent something
  if len(user.Password) > 0 && len(user.Username) > 0 {
    //opens session with database
    CQLsession, _ := cluster.CreateSession()
    defer CQLsession.Close()
    //hashes the password
    user.Password, _ = hashPassword(user.Password)
    var userCount int

    //makes sure the user doesn't already exist and lets the user know if it is
    if err := CQLsession.Query("SELECT count(username) FROM user WHERE username = ? LIMIT 1",
  		user.Username).Consistency(gocql.One).Scan(&userCount); err != nil {
      fmt.Fprintf(w, "error")
      return
	  } else{
      if userCount > 0 {
        fmt.Fprintf(w, "taken")
        return
      }
    }

    //add the user to the database
    if err := CQLsession.Query(`INSERT INTO user (username, passHash) VALUES (?, ?)`,
      user.Username, user.Password).Exec(); err != nil {
      fmt.Fprintf(w, "error")
    } else{
      fmt.Fprintf(w, "success")
    }
  } else{
    fmt.Fprintf(w, "missing data")
  }
}

//adds a note associated to a user
func addNote(w http.ResponseWriter, r *http.Request) {
  var userCount int
  //opens session with database
  CQLsession, _ := cluster.CreateSession()
  defer CQLsession.Close()
  //takes the user input and puts it in an object
  var noteInfo NoteToken
  json.NewDecoder(r.Body).Decode(&noteInfo)

  //make sure the note isn't too large
  if len(noteInfo.NoteText) < 1000 {
    //makes sure that there is actually a user with the provided authentication token
    if err := CQLsession.Query("SELECT count(username) FROM user WHERE auth_token = ? LIMIT 1 ALLOW FILTERING",
      noteInfo.AuthToken).Consistency(gocql.One).Scan(&userCount); err != nil {
      fmt.Fprintf(w, "error")
      return
    } else{
      if userCount > 0 {
        //if the token is valid then add the note
        if err := CQLsession.Query(`INSERT INTO note (user_token, note_text) VALUES (?, ?)`,
          noteInfo.AuthToken, noteInfo.NoteText).Exec(); err != nil {
          fmt.Fprintf(w, "error")
        } else{
          fmt.Fprintf(w, "success")
        }
      } else {
        fmt.Fprintf(w, "error")
      }
    }
  } else {
    fmt.Fprintf(w, "too long")
  }
}

//returns a list of all the notes associated with a given token
func getNotes(w http.ResponseWriter, r *http.Request){
  var note string
  notes := make([]string, 0)
  token := r.URL.Path[len("/notes/"):]  //gets the token from the url path

  if len(token) == 50{  //make sure the token is the proper length
    //opens session with database
    CQLsession, _ := cluster.CreateSession()
    defer CQLsession.Close()

    //select all the notes associated with the given token then add them to the notes variable
    iter := CQLsession.Query("SELECT note_text FROM note WHERE user_token = ? ALLOW FILTERING", token).Iter()
    for iter.Scan(&note) {
  		notes = append(notes, note)
  	}
    //convert notes to json then send it to the user
    json.NewEncoder(w).Encode(notes)
  } else {
    fmt.Fprintf(w, "error")
  }
}

//removes a note associated with a given token
func deleteNote(w http.ResponseWriter, r *http.Request){
  //takes the json sent by the user and puts it into an object
  var noteInfo NoteToken
  json.NewDecoder(r.Body).Decode(&noteInfo)

  if len(noteInfo.AuthToken) == 50{ //makes sure the token is the proper length
    //opens session with database
    CQLsession, _ := cluster.CreateSession()
    defer CQLsession.Close()

    //deletes the record from the database
    if err := CQLsession.Query("DELETE FROM note WHERE user_token = ? AND note_text = ?",
      noteInfo.AuthToken, noteInfo.NoteText).Exec(); err != nil {
      fmt.Fprintf(w, "error")
    } else{
      fmt.Fprintf(w, "success")
    }

  } else {
    fmt.Fprintf(w, "error")
  }
}

//handles the request based on the REST type
func handleNoteRequest(w http.ResponseWriter, r *http.Request){
  switch r.Method {
    case "GET":
      fmt.Printf("processing GET request\n")
      getNotes(w, r)
    case "POST":
      fmt.Printf("processing POST request\n")
      addNote(w, r)
    case "DELETE":
      fmt.Printf("processing DELETE request\n")
      deleteNote(w, r)
    default:
      fmt.Fprintf(w, "error1")
  }
}

func main() {
  fmt.Printf("running")

  //sets up cassandra
  cluster.Keyspace = "notesapp"
  cluster.Consistency = gocql.Quorum

  rand.Seed(time.Now().UnixNano())

  r := mux.NewRouter()
  r.HandleFunc("/signup", signUp)
  r.HandleFunc("/login", logIn)
  r.HandleFunc("/notes", handleNoteRequest)
  r.HandleFunc("/notes/{token}", handleNoteRequest)

  handler := cors.Default().Handler(r)
  //allows cross site requests
  c := cors.New(cors.Options{
    AllowedOrigins: []string{"http://firstruleoffight.club"},
    AllowCredentials: true,
    AllowedMethods: []string{"GET", "POST", "DELETE"},
  })

  handler = c.Handler(handler)

	http.ListenAndServe(":8080", handler)
}
