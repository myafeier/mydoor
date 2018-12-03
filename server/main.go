package main

import "os"
import (
	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"net"
	"fmt"
	"time"
)

var messageChan chan string
var commandChan chan string

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example

	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	log.WithFields(log.Fields{})

	commandChan=make(chan string)
}

func main() {
	go socketServer()
	Server := gin.Default()
	Server.GET("/open", func(context *gin.Context) {
		log.Println("received cmd")
		commandChan <- "open"
		//context.Status(200)
		return
	})
	Server.Run(":7778")
}

func socketServer() (err error) {

	service := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		return
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}

}

func handleConn(conn net.Conn)  {
	fmt.Println("client connected:",conn.RemoteAddr().String())
	defer conn.Close()

	go deal(conn)

	fmt.Println("begin handle message")
	message := make([]byte,20)
	for {

		n, err := conn.Read(message)
		if err != nil {
			fmt.Println(err)
			return
		}

		if string(message[:n])=="ping"{
			_, err = conn.Write([]byte("pang"))
			fmt.Println("send message: pang")
			if err != nil {
				fmt.Println(err)
				return
			}

		}

		message=make([]byte,20)
		time.Sleep(10*time.Millisecond)
	}
}




func deal(conn net.Conn) (err error) {
	fmt.Println("wait for cmd")
	for {
		select {
		case <-commandChan:
			fmt.Println("begin to open door")
			_, err = conn.Write([]byte("open"))
			if err != nil {
				conn.Close()
				return
			}
		}
	}
}
