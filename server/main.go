package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	HOST = "127.0.0.1"
	PORT = "9001"
	TYPE = "tcp"
)

func KeepAlive(conn net.Conn) {
	fmt.Println("NEW keep alive connection:", conn.RemoteAddr())
	for {
		netData, _ := bufio.NewReader(conn).ReadString('\n')
		<-time.After(time.Duration(5) * time.Second)

		if !strings.Contains(netData, "ok") {
			fmt.Println("Close:", conn.RemoteAddr())
			conn.Close()
			break
		}
	}
}

func HandleConnection(conn net.Conn) {
	fmt.Println("NEW P2P connection:", conn.RemoteAddr())

	for {
		netData, _ := bufio.NewReader(conn).ReadString('\n')

		if strings.Contains(netData, "download") {
			nameFile := "./file.mp3"

			file, _ := os.Open(nameFile)
			pr, pw := io.Pipe()
			w, _ := gzip.NewWriterLevel(pw, 7)

			go func() {
				n, _ := io.Copy(w, file)

				w.Close()
				pw.Close()
				log.Printf("copied to piped writer via the compressed writer: %d", n)
			}()

			n, _ := io.Copy(conn, pr)

			log.Printf("copied to connection: %d", n)

			fmt.Fprintf(conn, "stop!!!\n")
		}
	}
}

func main() {
	go func() {
		fmt.Println("** Keep Alive Server **")
		liste, _ := net.Listen(TYPE, HOST+":9002")
		defer liste.Close()

		for {
			conn, _ := liste.Accept()

			go KeepAlive(conn)
		}
	}()

	func () {
		fmt.Println("** P2P Server **")
		listen, _ := net.Listen(TYPE, HOST+":"+PORT)

		defer listen.Close()

		for {
			conn, _ := listen.Accept()

			go HandleConnection(conn)
		}
	}()
}
