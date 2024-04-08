package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

type Server struct {
	URL               *url.URL // URL of the backend server.
	Healthy           bool
}
type ClientResponse struct {
	Header http.Header
	Body []byte
	statusCode int
}
func (s *Server) Proxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

var servers []*Server
var lastRun int

func nextServerLeastActive(servers []*Server) *Server {
	
	lastRun = (lastRun + 1) % len(servers)
	for ; !servers[lastRun].Healthy;{
		lastRun = (lastRun + 1) % len(servers)
	}
	return servers[lastRun]
}


func healthyCheck(servers []*Server){
	log.Println("healthyChecking")
	for _, server := range servers{
		go func(s *Server) {
			timeBetweenRequest,_ := time.ParseDuration("10s")
			ticker:= time.NewTicker(timeBetweenRequest)
			for ;; <- ticker.C{
				res, err := http.Get(s.URL.String())
				if err != nil || res.StatusCode >= 500 {
					log.Printf("Server %s is down", s.URL.Host)
					s.Healthy = false
				} else {
					s.Healthy = true
				}
			}
		}(server)
	}
}

func forwardRequest(r *http.Request, server *Server) (ClientResponse, error) {
	proxyReq, err := http.NewRequest(r.Method, fmt.Sprintf("%s://%s%s",server.URL.Scheme,server.URL.Host, server.URL.Path), r.Body)
	if err != nil {
		return ClientResponse{
			Header: proxyReq.Header,
			Body: []byte(fmt.Sprintf("Error creating request to the forwarded server: %v", err)),
			statusCode: 500,
		}, err
	}
	proxyReq.Header.Set("Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}
	log.Println("Headers set %v", proxyReq)

	client := &http.Client{}
	proxyRes, err := client.Do(proxyReq)
	if err != nil {
		return ClientResponse{
			Header: proxyReq.Header,
			Body: []byte(fmt.Sprintf("Error comunicating with the client: %v", err)),
			statusCode: 500,
		}, err
	}
	defer proxyRes.Body.Close()
	body, err := io.ReadAll(proxyRes.Body)
	
	if err != nil {
		return ClientResponse{
			Header: proxyReq.Header,
			Body: []byte(fmt.Sprintf("Error parsing the body from the forwarded server: %v", err)),
			statusCode: 500,
		}, err
	}
	return ClientResponse{
		Header: proxyRes.Header,
		Body: body,
		statusCode: proxyRes.StatusCode,
	}, nil
}

func handleLoadBalanceRequest(w http.ResponseWriter, r *http.Request) {
	server := nextServerLeastActive(servers)
	log.Printf("Trying connection in the server: %s", server.URL.Path)
	response, err := forwardRequest(r, server)
	if err != nil{
		log.Fatalf(string(response.Body))
		return	
	}
	for header, values := range response.Header {
		for _, value := range values {
			w.Header().Set(header, value)
		}
	}
	w.WriteHeader(response.statusCode)
	w.Write(response.Body)

}

func main() {
	godotenv.Load(".env")

	url,err := url.Parse("http://localhost:8001/")
	if err != nil{
		log.Printf("Error parsing the url: %v", err)
		return
	}
	servers = append(servers, &Server{URL:url})
	url, err = url.Parse("http://localhost:8002/")
	if err != nil{
		log.Printf("Error parsing the url: %v", err)
		return
	}
	servers = append(servers, &Server{URL:url})
	healthyCheck(servers)
	lastRun = 0	
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("Port is not found in the environment")
	}
	fmt.Println("Port: ", portString)
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	v1Router := chi.NewRouter()
	v1Router.Handle("/", http.HandlerFunc(handleLoadBalanceRequest))
	router.Mount("/v1", v1Router)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}
	log.Printf("Server starting on port: %v", portString)
	err = server.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}
