var setupLobbyBtn = document.querySelector("#setup-lobby-btn")
console.log(setupLobbyBtn)
setupLobbyBtn.addEventListener("click", function(){
    console.log("Hide homepage and reveal lobby name page.")
    homepage.setAttribute("hidden", true)
    lobbynamepage.removeAttribute("hidden")
})

var createLobbyNameBtn = document.querySelector("#create-lobby-name-btn")
console.log(createLobbyNameBtn)
createLobbyNameBtn.addEventListener("click", function(){
    console.log("Hide lobbynamepage and reveal user name page.")
    var lobbyNameInput = document.querySelector("#lobby_name")
    var lName = ""
    lName = lobbyNameInput.value
    console.log("User entered lobby name: " + lName)
    lobbynamepage.setAttribute("hidden", true)
    usernamepage.removeAttribute("hidden")
})

var createUserNameBtn = document.querySelector("#create-username-btn")
console.log(createUserNameBtn)
createUserNameBtn.addEventListener("click", function(){
    console.log("Hide usernamepage and reveal lobby page.")
    var userNameInput = document.querySelector("#user_name")
    var uName = ""
    uName = userNameInput.value
    console.log("User entered lobby name: " + uName)
    // send data to server and create a lobby
    usernamepage.setAttribute("hidden", true)
    lobbypage.removeAttribute("hidden")
})

var homepage = document.querySelector("#homepage")
var lobbynamepage = document.querySelector("#lobbyname")
var usernamepage = document.querySelector("#username")
var lobbypage = document.querySelector("#lobby")

// Create WebSocket connection.
const socket = new WebSocket('ws://localhost:80/socket')

// Connection opened
socket.addEventListener('open', function (event) {
    console.log("Socket connection established...")
    // Send LOBBY command to server and get list of lobbies
    var message = {
        command: "LOBBY"
    }
    socket.send(JSON.stringify(message))
})

// Listen for messages
socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);
    var message = JSON.parse(event.data)
    console.log("Command from server: " + message.command)
    var commands = (message.command.split(/(\s+)/)).filter(function(e) {
        return String(e).trim();
    })
    console.log(commands)
    switch (commands[0]) {
        case "SERVERLIST":
            console.log("Got SERVERLIST from server.")
            var serverList = commands.slice(1);
            console.log("List of servers: " + serverList)
            break;
        case "PLAY":
            console.log("Got PLAY from server.")
            break;
        case "PAUSE":
            console.log("Got PAUSE from server.")
            break;
        case "NOTIFY":
            console.log("Got NOTIFY from server.")
            break;
        case "UPDATE":
            console.log("Got UPDATE from server.")
            break;
        case "PING":
            console.log("Got PING from server.")
            break;
        case "SEND":
            console.log("Got SEND from server.")
            break;
        case "OKAY":
            console.log("Got OK from server.")
            break;
        default:
            console.error("Unknown command: " + command)
            break;
    }

})