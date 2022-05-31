package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "ERROR: Invalid number of arguments")
		fmt.Fprintf(os.Stderr, "USAGE: %s <service-name>\n", os.Args[0])
		os.Exit(1)
	}

	probeResults := make(chan probeServiceResult)
	go probeServiceOpenPorts(probeResults, os.Args[1])

	for result := range probeResults {
		if result.err != nil {
			// TODO: Send a sigterm to all child processes?
			log.Panicf("ERROR: minikube process died: %s", *result.err)
		}

		go tunnelPort(result.value) // TODO: Keep track of the child process started here
	}
}

// ServiceAddress represents an open port for a given service
type openPort struct {
	host string // minikube address
	port int    // port service is running on
}

type probeServiceResult struct {
	err   *string   // Set if an error occured
	value *openPort // Set if all is good
}

// probeServiceOpenPorts listens for open ports exposed by a k8s service
// running in minikube and writes them to the given channel.
func probeServiceOpenPorts(ch chan probeServiceResult, serviceName string) {
	log.Printf("Starting process: minikube service --url %s", serviceName)
	cmd := exec.Command("minikube", "service", "--url", serviceName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		message := fmt.Sprintf("Could not attach to child process: %s", err.Error())
		ch <- probeServiceResult{err: &message}
		return
	}

	reader := bufio.NewReader(stdout)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			message := fmt.Sprintf("Could not read output from child process: %s", err.Error())
			ch <- probeServiceResult{err: &message}
			return
		}

		port := extractOpenPort(line)
		if port == nil {
			time.Sleep(1)
			continue
		}

		log.Printf("Found open port: %s:%d", port.host, port.port)
		ch <- probeServiceResult{value: port}
	}
}

func extractOpenPort(line string) *openPort {
	regex, err := regexp.Compile(`^\s*http://(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d+)`)
	if err != nil {
		log.Panicf("Failed to extract open port information: %s", err.Error())
	}

	matches := regex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	port, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Printf("WARNING: Failed to parse port: %s", line)
		return nil
	}

	return &openPort{host: matches[1], port: port}
}

// Creates a ssh tunnel to localhost:`port.port` from `port.host`:`port.port`
func tunnelPort(port *openPort) {
	address := fmt.Sprintf("docker@%s", port.host)
	keyPath := getSshKeyPath()
	portMapping := fmt.Sprintf("%d:localhost:%d", port.port, port.port)

	log.Printf("Creating tunnel from %s:%d", port.host, port.port)
	cmd := exec.Command("ssh", address, "-i", keyPath, "-L", portMapping)
	err := cmd.Run()
	if err != nil {
		log.Printf("WARNING: Could not create tunnel from %s:%d", port.host, port.port)
	}
}

func getSshKeyPath() string {
	// TODO: Override minikubeHome path with user specified path
	return path.Join(os.Getenv("HOME"), ".minikube/machines.minikube/id_rsa")
}
