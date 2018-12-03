package main

import (
	"github.com/nathan-osman/go-rpigpio"
	"time"
	"log"
	"net"
	"sync"
	"fmt"
)

var conn net.Conn
var connChan chan bool = make(chan bool)
var errorChan chan error = make(chan error)
var pingCmdChan chan bool = make(chan bool)
var receiceCmdChan chan bool = make(chan bool)
var pingOverChan chan bool = make(chan bool)
var receiveMsgOVerChan chan bool = make(chan bool)
var waitGroup=new(sync.WaitGroup)
var connStat bool


func init() {
	log.SetPrefix("[MyDoor]")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	log.Println("Init")
	initGpio()

	err := connectServer()
	if err != nil {
		log.Println(err)
		return
	}

}

func connectServer() (err error) {

	log.Println("starting connect...")
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:7777")
	if err != nil {
		return
	}


	go func() {
		waitGroup.Add(1)
		for {
			select {
			case err = <-errorChan:
				conn, err = net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					err = fmt.Errorf("dial remote error!,err=%s\n", err.Error())
					log.Println(err)
					time.Sleep(10 * time.Second)
					go func() {
						errorChan <- err
					}()
					continue
				}
				connChan <- true
			}

		}

	}()

	 go func() {
		 waitGroup.Add(1)
		for {
			log.Println("begin loop")
			select {
			case stat:= <-connChan: //开始发送数据和接收数据
				connStat=stat
				if stat{
					go func() {
						pingOverChan<-true
					}()
					go func() {
						receiveMsgOVerChan<-true
					}()
				}

			case <-pingOverChan:
				if connStat{
					pingCmdChan <- true
				}

			case <-receiveMsgOVerChan:
				if connStat{
					receiceCmdChan <- true
				}
			}
		}
	}()

	go func() {
		waitGroup.Add(1)
		for{
			select {
			case <-pingCmdChan:
				err = ping()
				if err != nil {
					errorChan <- err
					connChan <- false
					continue
				}
				pingOverChan<-true

			}
		}
	}()

	go func() {
		waitGroup.Add(1)
		for{
			select {
			case <-receiceCmdChan:
				err=receiveMSG()
				if err != nil {
					errorChan <- err
					connChan <- false
					continue
				}
				receiveMsgOVerChan<-true
			}
		}
	}()



	errorChan<-fmt.Errorf("begin to connect server")
	waitGroup.Wait()
	return
}

func receiveMSG() (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = err1.(error)
			log.Println("panic recover,err:", err)
			return
		}
	}()
	message := make([]byte, 20)
	n, err1 := conn.Read(message)
	if err1 != nil {
		return err1
	}
	msg := string(message[:n])
	log.Println(msg)
	if msg == "open" {
		err = openDoor()
		if err != nil {
			log.Println(err)
			err=nil
		}
	} else if msg == "pang" {
		log.Println("received pang")
	}
	return

}

func ping() (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = err1.(error)
			log.Println("panic recover,err:", err)
			return
		}
	}()

	_, err = conn.Write([]byte("ping"))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ping sended")
	time.Sleep(10*time.Second)
	return
}

func initGpio() (err error) {
	p1, err := rpi.OpenPin(15, rpi.OUT)
	if err != nil {
		return
	}
	defer p1.Close()
	p1.Write(rpi.HIGH)

	p2, err := rpi.OpenPin(14, rpi.IN)
	if err != nil {
		return
	}
	defer p2.Close()
	p2.Write(rpi.HIGH)
	return
}
func openDoor() (err error) {
	p1, err := rpi.OpenPin(15, rpi.OUT)
	if err != nil {
		return
	}
	defer p1.Close()
	p1.Write(rpi.LOW)
	time.Sleep(500 * time.Millisecond)
	p1.Write(rpi.HIGH)
	log.Println("door opend!")
	return
}















func inCheck() bool {
	var i int
	for {
		time.Sleep(100 * time.Millisecond)
		p1, err := rpi.OpenPin(14, rpi.IN)
		if err != nil {
			continue
		}
		defer p1.Close()
		value, err := p1.Read()

		if err != nil {
			continue
		}
		log.Println(value)
		if value == 0 {
			i++
		} else {
			i--
		}
		if i > 2 {
			return true
		} else if i < 0 {
			i = 0
		}
	}

}

func openMonitorAndDoor() (err error) {
	//开启屏幕
	writeValue(17)
	//开门
	time.Sleep(4 * time.Second)
	writeValue(18)

	//关闭屏幕
	writeValue(17)
	return
}

func writeValue(pin int) {
	p1, err := rpi.OpenPin(pin, rpi.OUT)
	if err != nil {
		return
	}
	p1.Write(rpi.LOW)
	p1.Close()
	time.Sleep(600 * time.Millisecond)

	p1, err = rpi.OpenPin(pin, rpi.OUT)
	if err != nil {
		return
	}
	p1.Write(rpi.HIGH)
	p1.Close()
}
