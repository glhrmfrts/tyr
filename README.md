# Tyr

Go network messaging framework:

- RabbitMQ (based on Streadway's [AMQP](https://github.com/streadway/amqp) library).
- UDP

## Raid protocol

A protocol to send and receive messages in a request/response fashion. Example:

```go
import (
       "log"

       "github.com/glhrmfrts/tyr/raid"
       "github.com/glhrmfrts/tyr/udp"
)

func makeMessage() *raid.Message {
     header := raid.Header{
            Action: "ping",
            Etag: raid.NewEtag(),
     }

     return &raid.Message{
            Header: header,
            Body: "ping",
     }
}

func main() {
     connection := udp.NewConnection("my-udp-server:31337")
     ch, err := connection.Send(makeMessage())
     if err != nil {
        log.Fatal(err)
     }

     // response has type *raid.Message
     response := <-ch
     log.Println(response.Body) // pong
}
```

Example server:

```go
package main

import (
       "log"
       "net"

       utcode "github.com/glhrmfrts/go-utcode"
)

func main() {
    serverAddr,err := net.ResolveUDPAddr("udp",":10001")
    if err != nil {
       log.Fatal(err)
    }

    serverConn, err := net.ListenUDP("udp", ServerAddr)
    if err != nil {
       log.Fatal(err)
    }
    defer serverConn.Close()

    buf := make([]byte, 1024)

    for {
        n,addr,err := serverConn.ReadFromUDP(buf)
        if err != nil {
            log.Fatal(err)
        }

        var msg raid.Message
        err = utcode.Decode(buf[0:n], &msg)
        if err != nil {
           log.Fatal(err)
        }

        if msg.Header.Action == "ping" {
           msg.Body = "pong"

           encodedMsg, err := utcode.Encode(&msg)
           if err != nil {
              log.Fatal(err)
           }

           serverConn.WriteToUDP(encodedMsg, addr)
        } else {
           log.Println("unknown action ", msg.Header.Action)
        }
    }
}
```