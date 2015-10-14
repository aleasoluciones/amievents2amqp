package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"net/url"
)

var (
	host, user, password, endpoint string
	port                           int
	debug                          bool = false
)

func executeCommand(nivis_ami_uri, command string) ([]string, error) {
	result := []string{}

	fmt.Println("C1")

	u, err := url.Parse(nivis_ami_uri)
	if err != nil {
		return result, err
	}
	fmt.Println("C2")
	password, _ := u.User.Password()
	user := u.User.Username()
	hostPort := fmt.Sprintf("%s:%d", u.Host, 5038)

	c, err := net.Dial("tcp", hostPort)
	if err != nil {
		return result, err
	}
	defer c.Close()
	fmt.Println("C3")

	fmt.Fprintf(c, "Action: login\r\n")
	fmt.Fprintf(c, "Username: %s\r\n", user)
	fmt.Fprintf(c, "Secret: %s\r\n", password)
	fmt.Fprintf(c, "\r\n")
	fmt.Println("C4")

	fmt.Fprintf(c, "Action: command\r\n")
	fmt.Fprintf(c, "Command: %s\r\n", command)
	fmt.Fprintf(c, "\r\n")
	fmt.Println("C5")

	connbuf := bufio.NewReader(c)
	for {
		str, err := connbuf.ReadString('\n')
		fmt.Println("C6", str)
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

func receiveEvents(nivis_ami_uri string) error {
	result := []string{}

	u, err := url.Parse(nivis_ami_uri)
	if err != nil {
		return err
	}
	password, _ := u.User.Password()
	user := u.User.Username()
	hostPort := fmt.Sprintf("%s:%d", u.Host, 5038)

	c, err := net.Dial("tcp", hostPort)
	if err != nil {
		return err
	}
	defer c.Close()

	fmt.Fprintf(c, "Action: login\r\n")
	fmt.Fprintf(c, "Username: %s\r\n", user)
	fmt.Fprintf(c, "Secret: %s\r\n", password)
	fmt.Fprintf(c, "\r\n")

	connbuf := bufio.NewReader(c)
	for {
		str, err := connbuf.ReadString('\n')
		if len(str) > 0 {
			if str == "--END COMMAND--\r\n" {
				break
			}
			result = append(result, str)
			fmt.Println("EFA", strings.TrimSpace(str))
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

	//commandResult, err := executeCommand(amiURI, command)
	err := receiveEvents(amiURI)
	fmt.Println("E3")
	if err != nil {
		log.Panic("Error", err)
		return
	}
	fmt.Println("E4")

}
