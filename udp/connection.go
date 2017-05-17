package udp

import (
	"log"
	"net"

	"gitlab.com/vikingmakt/tyr/raid"
	utcode "gitlab.com/vikingmakt/go-utcode"
)

type Connection struct {
	conn        *net.UDPConn
	requests    map[string]chan *raid.Message
	writeToAddr *net.UDPAddr
}

func NewConnection(writeToAddr string) (*Connection, error) {
	addr, err := net.ResolveUDPAddr("udp", writeToAddr)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		requests: make(map[string]chan *raid.Message),
		writeToAddr: addr,
	}
	
	err = c.listen()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Connection) Send(msg *raid.Message) (chan *raid.Message, error) {
	encode, err := utcode.Encode(msg)
	if err != nil {
		return nil, err
	}

	_, err = c.conn.WriteToUDP(encode, c.writeToAddr)
	if err != nil {
		return nil, err
	}

	ch := make(chan *raid.Message, 1)
	c.requests[msg.Header.Etag] = ch

	return ch, nil
}

func (c *Connection) listen() error {
	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	c.conn, err = net.ListenUDP("udp", localAddr)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	go func() {
		for {
			size, addr, err := c.conn.ReadFromUDP(buf)
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