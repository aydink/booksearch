package main

import (
	"fmt"
	"net/http"
)

func requireAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Our middleware logic goes here...
		r.Header.Set("userid", "aydink")
		fmt.Println("before auth")
		next(w, r)
		fmt.Println("after auth")
	})
}
