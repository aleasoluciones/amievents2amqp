package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	gami "code.google.com/p/gami"
)

var (
	host, user, password, endpoint string
	port                           int
	debug                          bool = false
)

func init() {
	flag.IntVar(&port, "port", 5038, "AMI port")
	flag.StringVar(&host, "host", "localhost", "AMI host")
	flag.StringVar(&user, "user", "admin", "AMI user")
	flag.StringVar(&password, "password", "admin", "AMI secret")
	flag.StringVar(&endpoint, "endpoint", "http://localhost:3000/calls", "Endpoint issues API for calls")
	flag.Parse()
}

func printEvents(messages []gami.Message) {
	for index, m := range messages {
		event, _ := m["Event"]
		if event == "CoreShowChannel" {
			fmt.Println(fmt.Sprintf("Index %3d Event %10s Extension %10s ConnectedLineNum %10s CallerIDnum %10s ChannelStateDesc %10s ID %15s BridguedID %15s %s",
				index,
				m["Event"],
				m["Extension"],
				m["ConnectedLineNum"],
				m["CallerIDnum"],
				m["ChannelStateDesc"],
				m["UniqueID"],
				m["BridgedUniqueID"],
				m["Channel"]))
		}
		// fmt.Println(m)

	}

}

func main() {
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	g := gami.NewAsterisk(&c, nil)
	err = g.Login(user, password)
	if err != nil {
		log.Fatal(err)
	}

	lastId := ""
	buffer := []gami.Message{}

	handler := func(m gami.Message) {
		actionId, ok := m["ActionID"]
		if ok {

			if actionId == lastId {
				buffer = append(buffer, m)
			} else {
				printEvents(buffer)
				lastId = actionId
				buffer = []gami.Message{}
				buffer = append(buffer, m)
			}
		}

	}

	g.DefaultHandler(&handler)

	//---------
	go func() {
		for {
			ping(g)
			time.Sleep(10 * time.Second)
		}
	}()

	for {
		m := gami.Message{"Action": "CoreShowChannels"} // gami.Message simple alias on map[string]string
		g.SendAction(m, nil)
		time.Sleep(1 * time.Second)
	}

	g.Logoff()
}

func ping(a *gami.Asterisk) {
	m := gami.Message{"Action": "Ping"}
	cb := func(m gami.Message) {
		fmt.Println(m)
	}
	err := a.SendAction(m, &cb)
	if err != nil {
		log.Fatal(err)
	}
}
