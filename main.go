package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"project/middleware"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/request", handleRequest)
	handler := middleware.NewRateLimiter(mux).
		SetCapacity(10).
		SetRefillPeriod(60 * time.Second).
		SetRefillTokens(10).
		LimitByIp()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for {
			if <-c == os.Interrupt {
				_ = srv.Close()
				return
			}
		}
	}()

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Server stopped")
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	log.Printf("handleRequest url: %v\n", req.URL)
	w.WriteHeader(200)
	_, err := w.Write([]byte("Success"))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func sendRequest(URL string) *http.Response {
	log.Printf("Sending request to %s\n", URL)
	resp, err := http.Get(URL)
	if err != nil {
		log.Fatal("Error sending request: ", err)
		return nil
	}
	return resp
}
