package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Server struct {
	Addr    string
	Handler http.Handler
}
type responseWriter struct {
	conn net.Conn
}

func (s *Server) ListenAndServe(c chan bool) error {
	if s.Handler == nil {
		panic("http server started without a handler.")
	}
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		c <- true
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.HandleConnection(conn)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {

	defer conn.Close()
	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		return
	}
	fields := strings.Fields(string(line))
	if len(fields) < 2 {
		return
	}
	r := &http.Request{
		Method:     fields[0],
		URL:        &url.URL{Scheme: "http", Path: fields[1]},
		Proto:      "HTTP/0.9",
		ProtoMajor: 0,
		ProtoMinor: 9,
		RemoteAddr: conn.RemoteAddr().String(),
	}
	s.Handler.ServeHTTP(newWriter(conn), r)
}

func (r *responseWriter) Header() http.Header {
	// unsupported with HTTP/0.9
	return nil
}

func (r *responseWriter) Write(b []byte) (int, error) {
	return r.conn.Write(b)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	// unsupported with HTTP/0.9
}

func newWriter(c net.Conn) http.ResponseWriter {
	return &responseWriter{
		conn: c,
	}
}

func client() {

	conn, err := net.Dial("tcp", "127.0.0.1:5675")
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	if _, err := conn.Write([]byte("GET /this/is/a/test\r\n")); err != nil {
		log.Fatalf("err: %s", err)
	}

	body, err := io.ReadAll(conn)
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	fmt.Println(string(body))

}

func main() {
	addr := "127.0.0.1:5675"
	s := Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World!\n"))
		}),
	}
	log.Printf("Listening on %s", addr)
	c := make(chan bool)
	go func() {
		if err := s.ListenAndServe(c); err != nil {
			log.Fatal(err)
		}
	}()

	<-c
	// for the sake of testing with go
	client()

}
