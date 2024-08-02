package utils

import (
	"fmt"
	"io"
	"net"
	"sync"
)

func PipeSocket(dest, source net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := io.Copy(dest, source)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//fmt.Printf("%d bytes copied\n", i)
}
