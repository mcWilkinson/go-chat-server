package main

import (
	"flag"
	"fmt"
  "net"
  "bufio"
  "strings"
  "sync"
  "encoding/json"
)

type User struct {
  Id int
  Connection net.Conn
  Name string
}

type Server struct {
  ip string
  port string
}

var server Server = Server{
  ip: "",
  port: "",
}

func main() {
	var (
		ip = flag.String("b", "127.0.0.1", "Base IP Address")
		port = flag.String("p", "8080", "port")
    server Server
	)
	flag.Parse()

  var full_address = fmt.Sprintf("%s:%s", *ip, *port)
	fmt.Printf("Attempting to start server on %s...\n", full_address)
	fmt.Println("Hello, world!")

  server = Server{
    ip: *ip,
    port: *port,
  }

  startServer(server)
}

func startServer(server Server) {
  // var users = make([]User, 0)
  ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.ip, server.port))

  if err != nil {
    fmt.Printf("Error: %s", err)
  }

  for {
    conn, err := ln.Accept()

    if err != nil {
      fmt.Printf("Error: %s", err)
    }

    user := initializeConnection(conn)
    userPtr := &user
    go handleConnection(conn, userPtr)
  }
}

var mu sync.Mutex
var users []User

func createUser(conn net.Conn) User {
  mu.Lock()
  defer mu.Unlock()

  var (
    lastId int = 0
    user User
  )

  if len(users) > 0 {
    lastUser := users[len(users) - 1]
    lastId = lastUser.Id
  }

  user = User{
    Connection: conn,
    Id: lastId + 1,
  }

  users = append(users, user)
  userString, _ := json.Marshal(users)
  fmt.Printf("Users: %s\n", userString)
  return user
}

func getUsers() []User {
  mu.Lock()
  defer mu.Unlock()

  return users
}

func initializeConnection(conn net.Conn) User {
  // server.users = append(server.users, user)
  fmt.Printf("Looks like someone is trying to connect...\n Connection details: %s\n", conn)
  return createUser(conn)
}

func handleConnection(conn net.Conn, user *User) {
  defer closeConnection(conn)

  fmt.Fprintf(conn, "%d\n", user.Id)
  for {
      response, err := bufio.NewReader(conn).ReadString('\n')
      response = strings.TrimSuffix(response, "\n")

      if err != nil {
        if err.Error() == "Error: EOF" {
          fmt.Println("Error: Client Disconnected")
        }
        fmt.Printf("Error: %s\n", err)
        return
      }

      fmt.Printf("Recieved message from user#%d: \"%s\"\n", user.Id, response)
      success := processMessage(response, user)
      if success == 1 {
        fmt.Println("Client Manually Disconnected")
        return
      }
  }
}

func closeConnection(conn net.Conn) {
  conn.Close()
}

func processMessage(message string, sender *User) int {
  if message == "\\quit" {
    return 1
  }

  for _, user := range getUsers() {
    if user.Id != sender.Id {
      fmt.Fprintf(user.Connection, "%s#%d: %s\n", sender.Name, sender.Id, message)
    }
  }

  if strings.HasPrefix(message, "\\name") {
    fmt.Printf("Someone tried to change their name...\n")
    name := strings.Trim(message[5:], " \r\n")
    updateName(name, sender)
    fmt.Printf("New name: %s\n", sender.Name)
  }

  return 0
}

func updateName(name string, user *User) {
  *user = User{
    Connection: user.Connection,
    Id: user.Id,
    Name: name,
  }
}
