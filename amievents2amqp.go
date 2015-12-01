package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"encoding/json"
	"net/url"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
	"github.com/aleasoluciones/simpleamqp"
	"github.com/cenkalti/backoff"
)

const (
	AMI_TIMEOUT = 5 * time.Second
)

func connectToManager(nivisAmiURI string) (net.Conn, error) {
	u, err := url.Parse(nivisAmiURI)
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

// Event AMI event type, including timestamp and all the event data as a string to string map
type Event struct {
	Timestamp int64
	Data      map[string]string
}

func receiveEvents(nivisAmiURI string, events chan (Event)) error {
	c, err := connectToManager(nivisAmiURI)
	if err != nil {
		log.Println("Error", err)
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
				key := strings.TrimSpace(r[0])
				value := strings.TrimSpace(r[1])
				data[key] = value
			}
		} else {
			events <- Event{time.Now().Unix(), data}
			data = make(map[string]string)
		}
		if err != nil {
			return err
		}
	}
}

func sipShowPeers(nivisAmiURI string) (int, error) {
	u, err := url.Parse(nivisAmiURI)
	if err != nil {
		return 0, err
	}
	password, _ := u.User.Password()
	user := u.User.Username()
	hostPort := fmt.Sprintf("%s:%d", u.Host, 5038)
	c, err := net.Dial("tcp", hostPort)
	if err != nil {
		return 0, err
	}
	defer c.Close()
	fmt.Fprintf(c, "Action: login\r\n")
	fmt.Fprintf(c, "Username: %s\r\n", user)
	fmt.Fprintf(c, "Secret: %s\r\n", password)
	fmt.Fprintf(c, "Events: off\r\n")
	fmt.Fprintf(c, "\r\n")

	fmt.Fprintf(c, "Action: command\r\n")
	fmt.Fprintf(c, "Command: sip show peers\r\n")
	fmt.Fprintf(c, "\r\n")

	c.SetReadDeadline(time.Now().Add(AMI_TIMEOUT))
	onlineRegexp := regexp.MustCompile(`monitored: ([\d]+) online`)
	connbuf := bufio.NewReader(c)
	sipUsers := false
	for {
		str, err := connbuf.ReadString('\n')
		if len(str) > 0 {
			if strings.Contains(str, "Name/username") {
				sipUsers = true
			} else if len(onlineRegexp.FindStringSubmatch(str)) == 2 {
				continue
			} else if strings.Contains(str, "--END COMMAND--") {
				return 0, nil
			} else if sipUsers {
				str = strings.TrimSpace(str)
				log.Println("Sip peer", str)
			}
		}
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func main() {
	var exchange, topic, amiURI, amqpURI string
	flag.StringVar(&exchange, "exchange", "events", "AMQP exchange name")
	flag.StringVar(&topic, "topic", "astevents", "topic")
	flag.StringVar(&amiURI, "amiURI", "", "AMI connection URI (use AMI_URI env var as default)")
	flag.StringVar(&amqpURI, "amqpURI", "", "AMQP connection URI (use BROKER_URI env var as default)")
	flag.Parse()

	if amiURI == "" {
		amiURI = os.Getenv("AMI_URI")
	}
	if amqpURI == "" {
		amqpURI = os.Getenv("BROKER_URI")
	}
	fmt.Println("AMI_URI", amiURI)
	fmt.Println("amqpURI", amqpURI)

	if amqpURI == "" || amiURI == "" {
		log.Panic("No amiURI or amqpURI provided")
	}

	events := make(chan (Event), 1)
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

	scheduledtask.NewScheduledTask(func() { sipShowPeers(amiURI) }, 60*time.Second, 0)

	amqpPublisher := simpleamqp.NewAmqpPublisher(amqpURI, exchange)
	for e := range events {
		jsonSerialized, _ := json.Marshal(e)
		amqpPublisher.Publish(topic, []byte(jsonSerialized))
	}
}
