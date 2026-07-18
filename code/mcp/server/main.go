package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	serverName    = "rekindled"
	serverTitle   = "ReKindled Display"
	serverVersion = "0.1.0"
)

func main() {
	rootDefault := discoverRoot()
	root := flag.String("root", rootDefault, "ReKindled project root")
	sshConfig := flag.String("ssh-config", "", "SSH config (default: ROOT/device-config/ssh_config)")
	host := flag.String("host", "rekindled-pw5", "SSH host alias")
	timeout := flag.Duration("timeout", 25*time.Second, "device operation timeout")
	manual := flag.Bool("manual", false, "print the complete human-readable manual and exit")
	version := flag.Bool("version", false, "print the version and exit")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "ReKindled MCP — a small stdio bridge for the Kindle display.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "\nRun with no options to speak MCP over stdin/stdout. Diagnostics use stderr.")
		fmt.Fprintln(flag.CommandLine.Output(), "Use -manual for tools, resources, gestures, examples, and recovery help.")
	}
	flag.Parse()

	if *version {
		fmt.Printf("%s %s\n", serverName, serverVersion)
		return
	}
	if *manual {
		fmt.Print(manualMarkdown)
		return
	}
	if *root == "" {
		log.Fatal("cannot locate project root; pass -root /absolute/path/to/Project Re-Kindled")
	}
	if *sshConfig == "" {
		*sshConfig = filepath.Join(*root, "device-config", "ssh_config")
	}

	device := &SSHDevice{
		Root:      *root,
		SSHConfig: *sshConfig,
		Host:      *host,
		Timeout:   *timeout,
	}
	server := NewServer(device)
	if err := server.Serve(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func discoverRoot() string {
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if executable, err := os.Executable(); err == nil {
		dir := filepath.Dir(executable)
		candidates = append(candidates, dir, filepath.Dir(dir), filepath.Dir(filepath.Dir(dir)))
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(candidate, "device-config", "ssh_config")); err == nil {
			absolute, err := filepath.Abs(candidate)
			if err == nil {
				return absolute
			}
		}
	}
	return ""
}
