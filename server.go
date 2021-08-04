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
	// start the java process
	// check if "mc" directory exists
	// if not, create it
	if _, err := os.Stat("mc"); os.IsNotExist(err) {
		os.Mkdir("mc", 0777)
	}
	os.Chdir("mc/")
	cmd := exec.Command("/usr/local/Cellar/openjdk/16.0.1/bin/java", "-Xmx3750M", "-Xms1G", "-jar", "server.jar", "nogui")

	inPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		// watch the stdToJava channel and dump lines to inPipe
		for line := range stdToJava {
			fmt.Fprint(inPipe, line)
		}
	}()

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Stderr = os.Stderr

	go func() {
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(outPipe)
	for scanner.Scan() {
		line := scanner.Text()
		javaToStd <- line
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

	fmt.Println("Hello World")

	javaToStd = make(chan string)
	stdToJava = make(chan string)
	go javaBox()
	go stdIO()
	for {
		time.Sleep(1 * time.Second)
	}
}
