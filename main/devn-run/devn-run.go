package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/daemonl/devn.go"
)

type Proc struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	done        chan error
	log         io.Writer
	InjectAfter []string
}

func GetProc(workdir string, env []string, envScript string, log io.Writer) (*Proc, error) {

	cmd := exec.Command("/bin/bash", envScript)
	cmd.Env = env
	cmd.Dir = workdir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	go io.Copy(log, stdout)
	go io.Copy(log, stderr)

	p := &Proc{
		cmd:   cmd,
		stdin: stdin,
		log:   log,
		done:  make(chan error),
	}

	return p, nil
}

func (p *Proc) Start() {
	go func() {
		p.done <- p.cmd.Run()
	}()
}

func (p *Proc) Wait() error {
	return <-p.done
}

func (p *Proc) Writeln(line string) {
	fmt.Fprintf(p.log, "> %s\n", line)
	_, err := p.stdin.Write([]byte(line + "\n"))
	if err != nil {
		go func() {
			p.done <- err
		}()
	}
}

func (p *Proc) End() error {
	for _, l := range p.InjectAfter {
		p.Writeln(l)
	}
	p.stdin.Close()
	p.Start()
	return p.Wait()
}

func (p *Proc) Scan(scanner *bufio.Scanner) ([]string, error) {
	p.Writeln("#!/bin/bash")
	p.Writeln("set -ue")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 1 {
			continue
		}

		if line[0:2] == "#+" {
			// Close off previous command
			err := p.End()
			if err != nil {
				return nil, err
			}

			line = line[2:] // Drop the #+ prefix
			return strings.SplitN(line, " ", 2), nil
		}
		p.Writeln(line)
	}

	return nil, p.End()
}

var flags = struct {
	script string
	envs   string
}{}

func init() {
	flag.StringVar(&flags.script, "script", "", "The bash-ish script to run")
	flag.StringVar(&flags.envs, "envs", "./environments", "The root of the environment scripts")
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

func do() error {

	if !path.IsAbs(flags.envs) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		flags.envs = path.Clean(path.Join(wd, flags.envs))
	}

	workdir := path.Dir(flags.script)

	script, err := os.Open(flags.script)
	if err != nil {
		return err
	}
	defer script.Close()

	scanner := bufio.NewScanner(script)
	scanner.Split(devn.ScanEscapedLines)

	var proc *Proc

	// skip shebang
	if !scanner.Scan() {
		return fmt.Errorf("No content in file")
	}

	delim := `183271489437584397580329`
	buff := &bytes.Buffer{}
	preamble, err := GetProc(workdir, os.Environ(), "-s", buff)
	// -s is basically a noop
	if err != nil {
		return err
	}
	preamble.InjectAfter = []string{
		`echo "` + delim + `"`,
		"env",
	}

	next, err := preamble.Scan(scanner)
	if err != nil {
		fmt.Fprintln(os.Stderr, buff.String())
		return err
	}

	lines := strings.Split(buff.String(), "\n")
	baseEnv := []string{}
	after := false
	for _, l := range lines {
		if l == delim {
			after = true
			continue
		}
		if !after {
			continue
		}
		baseEnv = append(baseEnv, l)
	}

	for {
		// loop over lines
		fmt.Printf("-- Run in %s\n", next[0])
		envScript := path.Clean(path.Join(flags.envs, next[0]))
		if !strings.HasPrefix(envScript, flags.envs) {
			return fmt.Errorf("Script name %s not rooted at base", next[0])
		}

		next[0] = envScript
		proc, err = GetProc(workdir, baseEnv, strings.Join(next, " "), os.Stdout)
		if err != nil {
			return err
		}

		next, err = proc.Scan(scanner)
		if err != nil {
			return err
		}
		if len(next) < 1 {
			return nil
		}
	}

	return nil
}
