package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

// Overridden response writer for HEAD requests
// Suppresses writing body in the response
type headResponseWriter struct {
	http.ResponseWriter
}

func (writer headResponseWriter) Write(body []byte) (int, error) {
	return 0, nil
}

func getClientIP(request *http.Request) string {
	xff := request.Header.Get("X-Forwarded-For")
	if xff != "" {
		// the value of the XFF header can be a list of IP addresses, separated by comma
		return strings.SplitN(xff, ",", 2)[0]
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return ""
}

func handleGet(writer http.ResponseWriter, request *http.Request) {
	ip := getClientIP(request)
	switch format := request.FormValue("format"); format {
	case "":
		writer.Write([]byte(ip))
	case "json":
		writer.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(map[string]string{"ip": ip})
		writer.Write(b)
	default:
		http.NotFound(writer, request)
	}
}

func handler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		handleGet(writer, request)
	case "HEAD":
		handleGet(headResponseWriter{writer}, request)
	default:
		writer.Header().Set("Allow", "GET, HEAD")
		http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc("/ipaddress", handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
