package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"encoding/json"
	"net/url"

	"github.com/cenkalti/backoff"
)

var (
	host, user, password, endpoint string
	port                           int
	debug                          bool = false
)

func connectToManager(nivis_ami_uri string) (net.Conn, error) {
	fmt.Println("C1")
	u, err := url.Parse(nivis_ami_uri)
	if err != nil {
		return nil, err
	}
	password, _ := u.User.Password()
	user := u.User.Username()
	hostPort := fmt.Sprintf("%s:%d", u.Host, 5038)

	c, err := net.Dial("tcp", hostPort)
	if err != nil {
		return c, err
	}
	fmt.Fprintf(c, "Action: login\r\n")
	fmt.Fprintf(c, "Username: %s\r\n", user)
	fmt.Fprintf(c, "Secret: %s\r\n", password)
	fmt.Fprintf(c, "\r\n")

	return c, err
}

func executeCommand(nivis_ami_uri, command string) ([]string, error) {
	result := []string{}
	c, err := connectToManager(nivis_ami_uri)
	if err != nil {
		return result, err
	}
	defer c.Close()

	fmt.Fprintf(c, "Action: command\r\n")
	fmt.Fprintf(c, "Command: %s\r\n", command)
	fmt.Fprintf(c, "\r\n")

	connbuf := bufio.NewReader(c)
	for {
		str, err := connbuf.ReadString('\n')
		if len(str) > 0 {
			if str == "--END COMMAND--\r\n" {
				break
			}
			result = append(result, str)
		}
		if err != nil {
			return []string{}, err
		}
	}
	return result, nil
}

type Event struct {
	Timestamp int64
	Data      map[string]string
}

func receiveEvents(nivis_ami_uri string, events chan (Event)) error {
	c, err := connectToManager(nivis_ami_uri)
	if err != nil {
		return err
	}
	defer c.Close()

	connbuf := bufio.NewReader(c)
	data := make(map[string]string)
	for {
		str, err := connbuf.ReadString('\n')
		str = strings.TrimSpace(str)
		if len(str) > 0 {
			r := strings.SplitN(str, ":", 2)
			if len(r) == 2 {
				data[r[0]] = r[1]
			}
		} else {
			events <- Event{time.Now().Unix(), data}
			data = make(map[string]string)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var amiURI, command string

	fmt.Println("E1")
	flag.StringVar(&amiURI, "amiuri", "ami://user:pass@host", "AMI uri")
	flag.StringVar(&command, "command", "sip show peers", "command to execute")
	flag.Parse()
	fmt.Println("E2")

	events := make(chan (Event), 1)

	//commandResult, err := executeCommand(amiURI, command)

	go func() {
		operation := func() error {
			return receiveEvents(amiURI, events)
		}

		err := backoff.Retry(operation, backoff.NewExponentialBackOff())
		if err != nil {
			log.Panic("Error", err)
			return
		}
	}()

	for e := range events {
		jsonSerialized, _ := json.Marshal(e)
		fmt.Println("Event", string(jsonSerialized))
	}

	fmt.Println("E4")

}
