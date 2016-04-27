package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/daemonl/devn.go"
	"github.com/mattn/go-shellwords"
)

var flags = struct {
	script string
	ext    string
}{}

func init() {
	flag.StringVar(&flags.script, "script", "", "The bash-ish script to run")
	flag.StringVar(&flags.ext, "ext", "./docker-ext", "The root of the extension scripts")
}

func main() {
	flag.Parse()
	err := do()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
		return
	}
}

func run(command, args string) error {
	runScript := path.Clean(path.Join(flags.ext, command))
	if !strings.HasPrefix(runScript, flags.ext) {
		return fmt.Errorf("Script name %s not rooted at base", command)
	}
	argsArray, err := shellwords.Parse(args)
	if err != nil {
		return err
	}
	cmd := exec.Command(runScript, argsArray...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdout)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stderr, stderr)
	return cmd.Run()
}

func do() error {
	script, err := os.Open(flags.script)
	if err != nil {
		return err
	}
	defer script.Close()

	scanner := bufio.NewScanner(script)
	scanner.Split(devn.ScanEscapedLines)

	variables := map[string]string{}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		switch line[0:2] {
		case "#-":
			parts := strings.SplitN(line[2:], " ", 2)
			if len(parts) < 1 {
				parts = []string{"true"}
			}
			variables[parts[0]] = parts[1]
			fmt.Printf("Set %s to %s\n", parts[0], parts[1])
		case "#+":
			parts := strings.SplitN(line[2:], " ", 2)
			command := parts[0]
			args := ""
			if len(parts) > 0 {
				args = parts[1]
			}
			fmt.Printf("Run %s With %s\n", command, args)
			err = run(command, args)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
