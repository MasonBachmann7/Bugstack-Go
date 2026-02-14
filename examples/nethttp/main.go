// Example: BugStack with net/http
package main

import (
	"fmt"
	"net/http"

	bugstack "github.com/MasonBachmann7/bugstack-go"
	"github.com/MasonBachmann7/bugstack-go/middleware"
)

func main() {
	bugstack.Init(bugstack.Config{
		APIKey:  "bs_live_your_api_key_here",
		AutoFix: true,
		Debug:   true,
	})
	defer bugstack.Flush()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong!")
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", middleware.NetHTTP(mux))
}
