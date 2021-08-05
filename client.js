// Quick test code to paste in the browser

let webSocket = new WebSocket("ws://localhost:8080/ws");

webSocket.onmessage = function(event) {
	console.log(event.data);
}

// Paste this line after webSocket resolves
webSocket.send("moose")
