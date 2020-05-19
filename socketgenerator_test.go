package main

import (
	"fmt"
	"testing"
)

func TestGenerateSockets(t *testing.T) {
	out := make(chan Socket)
	go func() {
		err := GenerateSockets(out, "192.168.1.0/24", []int{20, 40})
		if err != nil {
			close(out)
			t.Error(err)
		}
	}()

	for v := range out {
		fmt.Println(v)
	}

}
