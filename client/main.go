package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

func download(finished chan bool, conn net.Conn) {
	gzipFile := "file.mp3.gz"
	fo, _ := os.Create(gzipFile)

	tmp := make([]byte, 1024)
	data := make([]byte, 0)

	for {
		n, _ := conn.Read(tmp)

		if strings.Contains(string(tmp[:n]), "stop!!!") {
			break
		}

		data = append(data, tmp[:n]...)
	}

	fo.Write(data)

	// conn.Close()
	fo.Close()
	finished <- true
}

func unzip() {
	gzipFile, _ := os.Open("file.mp3.gz")

	gzipReader, _ := gzip.NewReader(gzipFile)
	defer gzipReader.Close()

	outfileWriter, _ := os.Create("file.mp3")
	defer outfileWriter.Close()

	io.Copy(outfileWriter, gzipReader)

	os.Remove("file.mp3.gz")
}

func KeepAlive(conn net.Conn) {
	for {
		<-time.After(time.Duration(25) * time.Second)
		fmt.Fprintf(conn, "ok\n")
	}
}

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:9001")
	defer conn.Close()

	go func() {
		con, _ := net.Dial("tcp", "127.0.0.1:9002")
		// defer con.Close()
		KeepAlive(con)
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, text+"\n")

		if strings.Contains(text, "download") {
			finished := make(chan bool)

			go download(finished, conn)

			fmt.Println("Main: Waiting for worker to finish")
			<-finished
			fmt.Println("Main: Completed")
			unzip()
		}
	}
}
