package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"encoding/json"
	"net/url"

	"github.com/aleasoluciones/simpleamqp"
	"github.com/cenkalti/backoff"
)

func connectToManager(nivis_ami_uri string) (net.Conn, error) {
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

type Event struct {
	Timestamp int64
	Data      map[string]string
}

func receiveEvents(nivis_ami_uri string, events chan (Event)) error {
	c, err := connectToManager(nivis_ami_uri)
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

func defaultBrokerUri() string {
	brokerUri := os.Getenv("BROKER_URI")

	if len(brokerUri) == 0 {
		brokerUri = "amqp://guest:guest@localhost/"
	}

	return brokerUri
}

func main() {
	var amiURI, amqpURI, exchange, topic string

	flag.StringVar(&amiURI, "amiuri", "ami://user:pass@host", "AMI uri")
	flag.StringVar(&amqpURI, "amqpuri", defaultBrokerUri(), "AMQP connection uri")
	flag.StringVar(&exchange, "exchange", "events", "AMQP exchange name")
	flag.StringVar(&topic, "topic", "astevents", "topic")
	flag.Parse()

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

	amqpPublisher := simpleamqp.NewAmqpPublisher(amqpURI, exchange)
	for e := range events {
		jsonSerialized, _ := json.Marshal(e)
		fmt.Println("Event", string(jsonSerialized))
		amqpPublisher.Publish(topic, []byte(jsonSerialized))
	}

	fmt.Println("E4")

}
