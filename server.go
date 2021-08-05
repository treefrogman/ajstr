// A simple WebSockets server per the tutorial here:
// https://tutorialedge.net/golang/go-websocket-tutorial/

package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var wsmc chan string
var mcws chan string

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// helpful log statement to show connections
	log.Println("WebSockets Client Connected")
	err = ws.WriteMessage(1, []byte("Go says Hello!"))
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	go webSocketListener(ws)
	// listen indefinitely for new messages coming
	// from the Minecraft server
	mcwsChannelListener(ws)
	log.Fatal("mcwsChannelListener broke")
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	// say hello
	fmt.Println("Go Minecraft WebSocket Server")

	// setup our channels
	wsmc = make(chan string)
	mcws = make(chan string)

	// setup our routes
	fmt.Println("setupRoutes()")
	setupRoutes()
	fmt.Println("javaBox()")
	go javaBox()
	fmt.Println("log.Fatal(http.ListenAndServe(\":8080\", nil))")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webSocketListener(conn *websocket.Conn) {
	for {
		// listen for messages from the websocket connection and send them to wsmc (the websocket-to-minecraft channel)
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// binary? what is this?
		if messageType == websocket.BinaryMessage {
			log.Println("Binary message received?!?")
			continue
		}

		// log the message
		log.Printf("wsmc <- %s", p)

		// push the message to the wsmc channel
		wsmc <- string(p)
	}
}

func mcwsChannelListener(conn *websocket.Conn) {
	// watch mcws (the minecraft-to-websocket channel) for messages and send them to the websocket connection
	for {
		// wait for a message on mcws
		message := <-mcws
		// log the message
		log.Printf("mcws -> %s", message)
		// send the message to the websocket connection
		err := conn.WriteMessage(1, []byte(message))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func javaBox() {
	const serverDirPath = "mc/"
	commandAndArgs := []string{"/usr/local/Cellar/openjdk/16.0.1/bin/java", "-Xmx3750M", "-Xms1G", "-jar", "server.jar", "nogui"}
	// cd to the server directory
	os.Chdir(serverDirPath)

	// set up the command to start the java process
	cmd := exec.Command(commandAndArgs[0], commandAndArgs[1:]...)

	// attach to the stdin of the java process
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	// attach to the stdout of the java process
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	// attach to the stderr of the java process
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			// watch wsmc (the websocket-to-minecraft channel) and dump lines to inPipe (the stdin of the java process)
			for line := range wsmc {
				fmt.Fprintln(inPipe, line)
				log.Printf("wsmc -> %s", line)
			}
		}
	}()

	outScanner := bufio.NewScanner(outPipe)
	go func() {
		// watch outPipe (the stdout of the java process) and dump lines to mcws (the minecraft-to-websocket channel)
		for outScanner.Scan() {
			line := outScanner.Text()
			mcws <- line
			log.Printf("mcws <- %s", line)
		}
	}()

	errScanner := bufio.NewScanner(errPipe)
	go func() {
		// watch errPipe (the stderr of the java process) and dump lines to mcws (the minecraft-to-websocket channel)
		for errScanner.Scan() {
			line := errScanner.Text()
			mcws <- line
			log.Printf("mcws <- (err) %s", line)
		}
	}()

	// start the java process
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
