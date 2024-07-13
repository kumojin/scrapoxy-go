package utils

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
)

func PipeSocket(dest, source net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	i, err := io.Copy(dest, source)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%d bytes copied\n", i)
}

func pipeClientToRemote(clientConn, remoteConn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	buff := make([]byte, 512)
	c := bufio.NewReader(clientConn)
	for {
		// read a single byte which contains the message length
		size, err := c.ReadByte()
		if err != nil {
			fmt.Println(err)
			return
		}

		// read the full message, or return an error
		_, err = io.ReadFull(c, buff[:int(size)])
		if err != nil {
			fmt.Println(err)
			return
		}

		remoteConn.Write(buff[:int(size)])
	}
}

func pipeRemoteToClient(remoteConn net.Conn, clientConn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	w := bufio.NewReader(remoteConn)
	wbuff := make([]byte, 256)
	for {
		size, err := w.ReadByte()
		if err != nil {
			return
		}

		// read the full message, or return an error
		_, err = io.ReadFull(w, wbuff[:int(size)])
		if err != nil {
			return
		}
		fmt.Println(string(wbuff))

		clientConn.Write(wbuff[:int(size)])
	}
}
