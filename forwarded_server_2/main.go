package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

func handleLoadBalanceRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Hello")
	responseText := fmt.Sprintf("<html lang=\"en\"><head><meta charset=\"utf-8\"><title>Index Page</title></head><body>Hello from the web server running on port 8082.</body></html>")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseText))
}

func main() {
	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("Port is not found in the environment")
	}
	fmt.Println("Port: ", portString)
	router := chi.NewRouter()
	router.HandleFunc("/", handleLoadBalanceRequest)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}
	log.Printf("Server starting on port: %v", portString)
	err := server.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}
