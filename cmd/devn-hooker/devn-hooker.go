package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// devn-hooker is an endpoint to send webhooks to and trigger
// commands (like builds), and

var flags = struct {
	bind        string
	scriptsRoot string
	secret      string
	middleware  string
}{}

func init() {
	flag.StringVar(&flags.bind, "bind", ":8080", "Server bind address")
	flag.StringVar(&flags.scriptsRoot, "scripts", "./scripts", "The root of the scripts, will be walked, or if a file, one script option per line")
	flag.StringVar(&flags.secret, "secret", "./secret", "File containing bytes to be added to the hash, will be automatically generated if missing or too short (20 bytes)")
	flag.StringVar(&flags.middleware, "middleware", "", "Run this command with the scripts as $1")
}

func main() {
	flag.Parse()

	err := loadScripts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loading scripts from %s: %s\n", flags.scriptsRoot, err.Error())
	}

	err = http.ListenAndServe(flags.bind, http.HandlerFunc(serve))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Starting the server: %s\n", err.Error())
	}
}

var scripts = map[string]string{}

func loadSecret() ([]byte, error) {
	secret, err := ioutil.ReadFile(flags.secret)
	if err != nil {
		secret = []byte{}
	}
	if len(secret) < 20 {
		fmt.Fprintf(os.Stderr, "Generating new secret")
		b := make([]byte, 256, 256)
		i, err := rand.Read(b)
		if err != nil {
			return nil, err
		}
		if i != 256 {
			return nil, fmt.Errorf("only got %d/256 bytes of random", i)
		}
		err = ioutil.WriteFile(flags.secret, b, 0500)
		if err != nil {
			return nil, err
		}
		secret = b
	}
	return secret, nil
}
func loadScripts() error {

	secret, err := loadSecret()
	if err != nil {
		return fmt.Errorf("loading secret: %s", err.Error())
	}

	rootFile, err := os.Open(flags.scriptsRoot)
	if err != nil {
		return err
	}
	defer rootFile.Close()
	stat, err := rootFile.Stat()
	if err != nil {
		return err
	}
	var scriptArray []string
	if stat.IsDir() {
		scriptArray, err = walkScripts(flags.scriptsRoot)
	} else {
		scriptArray, err = readScripts(rootFile)
	}

	for _, script := range scriptArray {
		h := sha512.New()
		io.WriteString(h, script)
		h.Write(secret)
		hashBytes := h.Sum(nil)
		hash := base64.URLEncoding.EncodeToString(hashBytes)
		fmt.Printf("%s -> %s\n", script, hash)
		scripts[hash] = script
	}
	return nil

}

func readScripts(from io.Reader) ([]string, error) {
	b, err := ioutil.ReadAll(from)
	if err != nil {
		return nil, err
	}
	in := strings.Split(string(b), "\n")
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			out = append(out, s)
		}
	}
	return out, nil
}

func walkScripts(root string) ([]string, error) {

	names := []string{}

	return names, filepath.Walk(root, func(fPath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Loading %s: %s", err.Error())
			return nil
		}
		if info.IsDir() {
			return nil
		}
		names = append(names, fPath)
		/* Not checking for exec, middleware may not require it, will bubble up
		* later
		if info.Mode()&0100 == 0 {
			return nil
		}*/
		return nil
	})
}

func serve(rw http.ResponseWriter, req *http.Request) {

	script, ok := scripts[req.URL.Path[1:]]
	if !ok {
		fmt.Printf("From %s, 404 %s\n", req.RemoteAddr, req.URL.Path)
		http.NotFound(rw, req)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK\n"))

	fmt.Printf("From %s, running %s\n", req.RemoteAddr, script)

	var cmd *exec.Cmd
	if len(flags.middleware) > 0 {
		cmd = exec.Command(flags.middleware, script)
	} else {
		cmd = exec.Command(script)
	}
	cmd.Start()
	cmd.Process.Release()
}
