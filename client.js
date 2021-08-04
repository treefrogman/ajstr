// Quick test code to paste in the browser

let url = "http://localhost:8080/ws";
let protocols = ["stdio"];

let webSocket = new WebSocket(url, protocols);

webSocket.onmessage = function(event) {
	console.log(event.data);
}

// Paste this line after webSocket resolves
webSocket.send("moose")
