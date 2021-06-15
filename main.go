package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
)

type Config struct {
	Host          string
	Port          int
	MaxBufferSize int // no use yet
}

var DefaultHost = "192.168.99.1"
var DefaultPort = 9099
var DefaultMaxBufferSize = 2 * 1024 * 1024
var configFile = "~/.spbcopy.ini"

func getStdin() (string, error) {
	var builder strings.Builder
	_, err := io.Copy(&builder, os.Stdin)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func genAPIAddr(host string, port int) string {
	var builder strings.Builder
	builder.WriteString("http://")
	builder.WriteString(host)
	builder.WriteString(":")
	builder.WriteString(strconv.Itoa(port))

	return builder.String()
}

func send(url string, content string) error {
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(content))
	if err != nil {
		return err
	}

	// no compress please
	request.Header.Set("Accept-Encoding", "identity")
	res, err := client.Do(request)
	if err != nil {
		return err
	}

	if res.StatusCode/100 != 2 {
		return errors.New("bad status code: " + strconv.Itoa(res.StatusCode))
	}

	return nil
}

// This client supply a cmd called spbcopy.
// Just use it as pbcopy in Mac.
//
// It looks like: cat xxx.txt | pbcopy, cat xxx.txt | spbcopy
// Which send a request to server, and server receive content and set it to Clipboard.
func main() {
	config := new(Config)

	_, err := os.Stat(configFile)
	if err == nil {
		// have config file
		cfg, err := ini.Load(configFile)
		if err != nil {
			log.Fatal("Failed to load config file: ", err)
		}

		config.Port, err = cfg.Section("base").Key("port").Int()
		if err != nil {
			log.Fatal("Failed to use config file port: ", err)
		}

		config.MaxBufferSize, err = cfg.Section("base").Key("maxbuffersize").Int()
		if err != nil {
			log.Fatal("Failed to use config file maxbuffersize: ", err)
		}

		config.Host = cfg.Section("base").Key("Host").String()
	} else {
		config.Host = DefaultHost
		config.Port = DefaultPort
		config.MaxBufferSize = DefaultMaxBufferSize
	}

	stdinContent, stdinErr := getStdin()
	if stdinErr != nil {
		log.Fatal("Failed to get stdin: ", stdinErr)
	}

	url := genAPIAddr(config.Host, config.Port)
	if err := send(url, string(stdinContent)); err != nil {
		log.Fatal("Failed to send spbcopy command: ", err)
	}
}
