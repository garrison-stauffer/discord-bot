package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("Hello World")
	mx := http.NewServeMux()
	mx.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received message lul")
		io.WriteString(w, "Doo doo")
	})

	err := http.ListenAndServe(":8080", mx)

	fmt.Printf("Shutting down server %v", err)
}
