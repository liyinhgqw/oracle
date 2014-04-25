package main

import (
	"fmt"
	"github.com/liyinhgqw/oracle"
)

func main() {
	client := oracle.NewClient(":7070")
	if ts, err := client.TS(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ts)
	}

}
