package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type middleware func(next http.HandlerFunc) http.HandlerFunc

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Since(start)
		log.Printf("%s %s %s", r.Method, r.URL.Path, elapsed)
	}
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header["Authorization"]
		if auth == nil {
			http.Error(w, "Unauthorized", 403)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func chainMiddleware(mw ...middleware) middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			last(w, r)
		}
	}
}

type User struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	user := User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("OK!"))
}

func main() {
	lt := chainMiddleware(withLogging)
	http.Handle("/", lt(handler))
	log.Fatal(http.ListenAndServe(":3000", nil))
}
