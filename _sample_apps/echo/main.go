package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func echo(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	fmt.Println(string(data))
	io.WriteString(w, string(data))
}

func main() {
	http.HandleFunc("/", echo)
	http.ListenAndServe("0.0.0.0:8888", nil)
}
