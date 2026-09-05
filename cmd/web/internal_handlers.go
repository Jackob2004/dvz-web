package main

import "net/http"

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}
