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

var host = "localhost:9325"
var server = "http://localhost:9091/transmission/rpc"

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

func main() {
	conn, err := net.Dial("tcp", host)
	log.Printf("Server Connected %s\n", conn.RemoteAddr())
	if err != nil {
		fmt.Printf("Unable to listen to host %s", host)
		return
	}

	for {
		req := recieve(conn)
		fmt.Printf(string(req))
		resp := callHTTP(req)
		send(resp, conn)
	}
}

func send(message []byte, conn net.Conn) error {

	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(message)))

	_, err := conn.Write(length)
	_, err = conn.Write(message)

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

func callHTTP(input []byte) []byte {
	var recieveReq request
	json.Unmarshal(input, &recieveReq)

	forwardReq, _ := http.NewRequest(recieveReq.Method, server, strings.NewReader(recieveReq.Body))

	for _, header := range recieveReq.Headers {
		if header != "" {
			splitted := strings.Split(header, ":")
			forwardReq.Header.Add(
				strings.TrimSpace(splitted[0]),
				strings.TrimSpace(splitted[1]))
		}
	}

	client := http.Client{}
	resp, err := client.Do(forwardReq)

	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}

	var sendingResp response

	sendingResp.Status = resp.StatusCode

	var buffer bytes.Buffer
	buffer.ReadFrom(resp.Body)
	sendingResp.Body = buffer.String()
	buffer.Reset()
	resp.Header.Write(&buffer)
	sendingResp.Headers = strings.Split(buffer.String(), "\n")

	output, _ := json.Marshal(sendingResp)
	log.Printf("Response : %s\n", string(output))

	return output

}
