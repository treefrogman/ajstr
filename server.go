// A simple WebSockets server per the tutorial here:
// https://tutorialedge.net/golang/go-websocket-tutorial/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var javaToStd chan string
var stdToJava chan string

func javaBox() {
	// check if "mc" directory exists
	if _, err := os.Stat("mc"); os.IsNotExist(err) {
		// if not, create it
		os.Mkdir("mc", 0777)
	}
	// cd mc
	os.Chdir("mc/")

	// set up the command to start the java process
	cmd := exec.Command("/usr/local/Cellar/openjdk/16.0.1/bin/java", "-Xmx3750M", "-Xms1G", "-jar", "server.jar", "nogui")

	// attach to the stdin of the java process
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		// watch the stdToJava channel and dump lines to inPipe (the stdin of the java process)
		for line := range stdToJava {
			fmt.Fprint(inPipe, line)
		}
	}()

	// attach to the stdout of the java process
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	outScanner := bufio.NewScanner(outPipe)
	go func() {
		// watch outPipe (the stdout of the java process) and dump lines to javaToStd
		for outScanner.Scan() {
			line := outScanner.Text()
			javaToStd <- line
		}
	}()

	// attach to the stderr of the java process
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	errScanner := bufio.NewScanner(errPipe)
	go func() {
		// watch errPipe (the stderr of the java process) and dump lines to javaToStd
		for errScanner.Scan() {
			line := errScanner.Text()
			javaToStd <- line
		}
	}()
	cmd.Stderr = os.Stderr

	// start the java process
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func stdIO() {

	reader := bufio.NewReader(os.Stdin)
	go func() {
		// watch stdin and dump lines to the stdToJava channel
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			stdToJava <- line
		}
	}()

	go func() {
		// watch the javaToStd channel and dump lines to stdout
		for line := range javaToStd {
			fmt.Println(line)
			if strings.Contains(line, "> $") {
				command := strings.Split(line, "> $")[1]
				if command == "" {
					continue
				}
				fmt.Println("happy moose")
				stdToJava <- "say " + command + "\n"
			}
		}
	}()
	fmt.Println()
}

func main() {

	fmt.Println("Go server started.")

	javaToStd = make(chan string)
	stdToJava = make(chan string)
	go javaBox()
	go stdIO()
	for {
		time.Sleep(1 * time.Second)
	}
}
