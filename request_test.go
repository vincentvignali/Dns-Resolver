// request_test.go
package main

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestMain(t *testing.T) {
	fmt.Println(" ####### START TESTING #######")
	fmt.Println(" ")
	fmt.Println(" ============ SERVER ============")
	go main()
	// cmd := exec.Command("dig", "@localhost -p 8090 google.com")
	out, _ := exec.Command("dig", "@127.0.0.1", "googleads.g.doubleclick.net").Output()
	// out, _ := exec.Command("curl", "situation.sh").Output()
	fmt.Println("\n ================================")
	fmt.Println(" ")
	fmt.Println(" ")
	fmt.Println(" ")
	fmt.Println(" ============ CLIENT ============")
	fmt.Println(string(out))
	fmt.Println(" ================================")
	fmt.Println(" ")
	fmt.Println(" ")
	fmt.Println(" ")
	fmt.Println(" ####### END TESTING #######")
	// send DNS packet to this server
}
