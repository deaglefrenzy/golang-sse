package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	resp, err := http.Get("http://localhost:8080/events")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	fmt.Println("Listening for streamed numbers...")

	for {
		// read the line until the \n limiter
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed:", err)
			return
		}

		// trim the line
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// convert to int so can be detected as prime or not
		num, err := strconv.Atoi(line)
		if err != nil {
			continue
		}

		fmt.Println("Received:", num)

		if isPrime(num) {
			fmt.Println("Connection cut because a prime number was detected")
			return
		}
	}
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}
