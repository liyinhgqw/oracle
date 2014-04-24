package main

import (
	"fmt"
	"github.com/liyinhgqw/oracle"
)

func main() {
	client := &oracle.Client{":7070"}
	fmt.Println(client.GetTS(5))
}
