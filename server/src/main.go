package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

var port = 9325
var conn net.Conn

func main() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Unable to listen to port %d", port)
		return
	}

	http.HandleFunc("/", httpHandler)

	conn, _ = ln.Accept()
	if err != nil {
		// handle error
	}
	log.Printf("Client Connected %s\n", conn.RemoteAddr())

	log.Fatal(http.ListenAndServe(":9327", nil))

}

type request struct {
	Method  string
	Body    string
	Headers []string
}

type response struct {
	Status  int
	Body    string
	Headers []string
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	var sendingReq request

	buf := new(bytes.Buffer)

	buf.ReadFrom(r.Body)
	sendingReq.Body = buf.String()
	sendingReq.Method = r.Method

	buf.Reset()

	r.Header.Write(buf)

	sendingReq.Headers = strings.Split(buf.String(), "\n")

	sending, _ := json.Marshal(sendingReq)

	log.Printf("Request: %s\n", string(sending))

	send(sending, conn)

	var resp = recieve(conn)
	var sendingResp response
	json.Unmarshal(resp, &sendingResp)

	w.Write([]byte(sendingResp.Body))
	w.WriteHeader(sendingResp.Status)

	for _, header := range sendingResp.Headers {
		if header != "" {
			splitted := strings.Split(header, ":")
			w.Header().Add(
				strings.TrimSpace(splitted[0]),
				strings.TrimSpace(splitted[1]))
		}
	}

	log.Println("Response")
	log.Println(sendingResp.Body)

}

func send(bytes []byte, conn net.Conn) error {
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(bytes)))

	_, err := conn.Write(length)
	_, err = conn.Write(bytes)

	return err
}

func recieve(conn net.Conn) []byte {
	prefix := make([]byte, 4)
	_, err := io.ReadFull(conn, prefix)
	if err != nil {
		log.Fatal(err)
	}

	length := binary.BigEndian.Uint32(prefix)
	message := make([]byte, int(length))
	_, err = io.ReadFull(conn, message)

	return message
}
