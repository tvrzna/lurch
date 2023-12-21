package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

type socketAction byte
type socketResponse byte

const (
	socketActionStart socketAction = iota + 1
)

const (
	socketResponseOk socketResponse = iota
	socketResponseFail
)

const (
	envSocketPort  = "LURCH_SOCKET_PORT"
	envSocketToken = "LURCH_SOCKET_TOKEN"
)

type socketServerContext struct {
	port  int
	c     *Context
	j     *Job
	l     net.Listener
	token string
}

type socketData struct {
	token  string
	action socketAction
	length int
	data   string
}

func (s *socketServerContext) parseSocketData(str string) *socketData {
	result := &socketData{}
	_, err := fmt.Sscanf(str, "%s %d %d %s", &result.token, &result.action, &result.length, &result.data)

	if err != nil || len(result.data) != result.length || s.token != result.token {
		return nil
	}
	return result
}

func (s *socketData) String() string {
	s.length = len(s.data)
	return fmt.Sprintf("%s %d %d %s", s.token, s.action, s.length, s.data)
}

func startServerSocket(c *Context, j *Job) *socketServerContext {
	s := &socketServerContext{c: c, j: j, token: randomToken(32)}

	var err error
	s.l, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Print("could not start server socket for ", j.p.name, " #", j.name)
		return nil
	}

	address := s.l.Addr().(*net.TCPAddr)
	s.port = address.Port

	go s.handle()

	return s
}

func (s *socketServerContext) handle() {
	for {
		con, err := s.l.Accept()
		if err != nil {
			break
		}

		msg, err := bufio.NewReader(con).ReadString('\n')
		if err != nil {
			break
		}
		d := s.parseSocketData(msg)
		if d == nil {
			log.Print("incomming data are not valid")
			con.Write([]byte{byte(socketResponseFail), '\n'})
			con.Close()
			continue
		}

		response := socketResponseFail

		switch d.action {
		case socketActionStart:
			p := s.c.OpenProject(d.data)
			if p != nil {
				if r := s.c.StartJob(p, nil); r != "" {
					response = socketResponseOk
				}
			}
		}

		con.Write([]byte{byte(response), '\n'})
		con.Close()
	}
}

func (s *socketServerContext) stop() {
	s.l.Close()
}

func makeSocketClientAction(action socketAction, data string) int {
	s := &socketData{action: action, data: data}
	s.token = os.Getenv(envSocketToken)

	con, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", os.Getenv(envSocketPort)))
	if err != nil {
		log.Println(err)
		return int(socketResponseFail)
	}
	defer con.Close()

	con.Write([]byte(s.String() + "\n"))
	res, err := bufio.NewReader(con).ReadString('\n')
	if err != nil {
		log.Println(err)
	}

	log.Print("result code: ", int(res[0]))

	return int(res[0])
}
