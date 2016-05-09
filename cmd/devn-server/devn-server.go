package main

import (
	"net/http"
	"os"
	"path"
)

/* devn-server
a web api and websocket server
* to monitor them. This might be uncomfortably large compared
* to the other components.
*/

var flags = struct {
	builds string
	bind   string
}{}

func HandleBuildInfo(rw http.ResponseWriter, req *http.Request) {
	file, err := os.Open(path.Join(flags.builds, req))
	if err != nil {
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s\n", err.Error())
		return
	}
	defer file.Close()
	io.Copy(rw, file)
}

func main() {
	http.Handle("/builds/", http.StripPrefix("/builds", http.HandleFunc(HandleBuildInfo)))
	http.ListenAndServe(bind, nil)
}

// Hook Endpoint

// Log Streaming Endpoint

// List builds

// List projects?`
