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

func main() {
	var exchange, topic string
	flag.StringVar(&exchange, "exchange", "events", "AMQP exchange name")
	flag.StringVar(&topic, "topic", "astevents", "topic")
	flag.Parse()

	amiURI := os.Getenv("AMI_URI")
	amqpURI := os.Getenv("BROKER_URI")

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
		amqpPublisher.Publish(topic, []byte(jsonSerialized))
	}
}
