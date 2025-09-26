package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	listenAddress  string
	listener       net.Listener
	quitChannel    chan struct{}
	messageChannel chan Message
}

func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress:  listenAddress,
		quitChannel:    make(chan struct{}),
		messageChannel: make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}
	defer listener.Close()
	s.listener = listener
	fmt.Printf("Server started on %s\n", s.listenAddress)
	go s.Accept()
	<-s.quitChannel
	close(s.messageChannel)
	return nil
}

func (s *Server) Stop() {
	close(s.quitChannel)
}

func (s *Server) Accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Failed accepting connection: ", err)
			continue
		}
		fmt.Println("Connection accepted:", conn.RemoteAddr())
		go s.HandleConnection(conn)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by client")
				return
			}
			fmt.Println("Failed reading from connection: ", err)
			continue
		}
		s.messageChannel <- Message{
			from:    conn.RemoteAddr().String(),
			payload: bytes,
		}
		_, err = conn.Write([]byte("Message received\n"))
		if err != nil {
			fmt.Println("Failed sending response to client: ", err)
			return
		}
	}
}

func main() {
	server := NewServer(":3000")
	go func() {
		for msg := range server.messageChannel {
			fmt.Printf("Message received from %s: %s\n", msg.from, string(msg.payload))
		}
	}()
	server.Start()
}
