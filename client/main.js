var setupLobbyBtn = document.querySelector("#setup-lobby-btn")
console.log(setupLobbyBtn)
setupLobbyBtn.addEventListener("click", function(){
    console.log("Button pressed.")
})

// Create WebSocket connection.
const socket = new WebSocket('ws://localhost:80/socket')

// Connection opened
socket.addEventListener('open', function (event) {
    console.log("Socket connection established...")
    var message = {
        command: "HELLO SERVER"
    }
    socket.send(JSON.stringify(message))
})

// Listen for messages
socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);
})