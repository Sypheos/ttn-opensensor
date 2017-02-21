package main

import (
	"net/http"
	"fmt"
)

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	r.Body.Close()
}

func main() {

	http.HandleFunc("/", handle)
	http.ListenAndServe(":3000", nil)
}
