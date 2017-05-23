package udp

import (
	"fmt"
	"log"
	"net"

	"gitlab.com/vikingmakt/tyr/raid"
	utcode "gitlab.com/vikingmakt/go-utcode"
)

type Connection struct {
	conn    net.Conn
	requests    map[raid.Etag]chan *raid.Message
}

func NewConnection(writeToAddr string) (*Connection, error) {
	conn, err := net.Dial("udp", writeToAddr)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		requests: make(map[raid.Etag]chan *raid.Message),
		conn: conn,
	}

	err = c.listen()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Connection) Conn() net.Conn {
	return c.conn
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

	fmt.Fprintf(c.conn, string(encode))
	return nil
}

func (c *Connection) listen() error {
	buf := make([]byte, 1024)
	go func() {
		for {
			size, err := c.conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			content := buf[0:size]

			c.onMessage(content)
		}
	}()

	return nil
}

func (c *Connection) onMessage(content []byte) {
	var msg raid.Message
	err := utcode.Decode(content, &msg)
	if err != nil {
		log.Println(err)
		return
	}

	if ch, ok := c.requests[msg.Header.Etag]; ok {
		ch <- &msg
	} else {
		log.Println("unknown etag ", msg.Header.Etag)
	}
}
