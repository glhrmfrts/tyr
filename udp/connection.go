package udp

import (
	"fmt"
	"log"
	"net"

	"gitlab.com/vikingmakt/tyr/raid"
	utcode "gitlab.com/vikingmakt/go-utcode"
)

type Connection struct {
	dialConn    net.Conn
	listenConn  *net.UDPConn
	requests    map[string]chan *raid.Message
}

func NewConnection(writeToAddr string) (*Connection, error) {
	dialConn, err := net.Dial("udp", writeToAddr)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		requests: make(map[string]chan *raid.Message),
		dialConn: dialConn,
	}

	err = c.listen()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Connection) Send(msg *raid.Message) (chan *raid.Message, error) {
	err := c.SendNoAck(msg)
	if err != nil {
		return nil, err
	}

	ch := make(chan *raid.Message, 1)
	c.requests[msg.Header.Etag] = ch

	return ch, nil
}

func (c *Connection) SendNoAck(msg *raid.Message) error {
	encode, err := utcode.Encode(msg)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.dialConn, string(encode))
	return nil
}

func (c *Connection) listen() error {
	localAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:0")
	if err != nil {
		return err
	}

	c.listenConn, err = net.ListenUDP("udp4", localAddr)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	go func() {
		for {
			size, addr, err := c.listenConn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}

			content := buf[0:size]

			log.Printf("received %v from %v", string(content), addr)
			c.onMessage(content)
		}
	}()

	return nil
}

func (c *Connection) onMessage(content []byte) {
	var msg raid.Message
	err := utcode.Decode(content, &msg)
	if err != nil {
		log.Fatal(err)
		return
	}

	if ch, ok := c.requests[msg.Header.Etag]; ok {
		ch <- &msg
	} else {
		log.Fatal("unknown etag ", msg.Header.Etag)
	}
}
