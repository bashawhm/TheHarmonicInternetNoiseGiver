// Get pages
var homepage = document.querySelector("#homepage")
var lobbynamepage = document.querySelector("#lobbynamepage")
var usernamepagecreate = document.querySelector("#usernamepagecreate")
var usernamepagejoin = document.querySelector("#usernamepagejoin")
var lobbypage = document.querySelector("#lobbypage")
var waitingpage = document.querySelector("#waitingpage")
var membersModal = document.querySelector("#memberspage")
var joinRequestsModal = document.querySelector("#joinrequestspage")
var settingsModal = document.querySelector("#settingspage")
// Get inputs
var searchLobbyInput = document.querySelector("#searchLobbiesInput")
var lobbyNameInput = document.querySelector("#lobbyNameInput")
var userNameInput = document.querySelector("#userNameInput")
var userNameInputCreate = document.querySelector("#userNameInputCreate")
var songUploadId = document.getElementById('songUploadId')
// Get buttons
var createLobbyBtn = document.querySelector("#createLobbyBtn")
var createLobbyNameBtn = document.querySelector("#createLobbyNameBtn")
var createUserNameCreateBtn = document.querySelector("#createUserNameCreateBtn")
var createUserNameJoinBtn = document.querySelector("#createUserNameJoinBtn")
var memberListBtn = document.querySelector("#memberListBtn")
var uploadSongBtn = document.querySelector("#uploadSongBtn")
var inviteBtn = document.querySelector("#inviteBtn")
var inviteListBtn = document.querySelector("#inviteListBtn")
var settingsBtn = document.querySelector("#settingsBtn")
var playBtn = document.querySelector("#playBtn")
var wildCardBtn = document.querySelector("#wildCardBtn")
var stopBtn = document.querySelector("#stopBtn")
// Get fields
var lobbyNameField = document.querySelector("#lobbyNameField")
var lobbyNameFieldWait = document.querySelector("#lobbyNameFieldWait")
// Get audio
var sound = document.querySelector("#audio")
// Data structures
var lobbies = []
var lobbyName = ""
var userName = ""
var lobbyName = ""
var audioFile = {}
var songnames = []
var artists = []
var tags = []
var clients = []
var offer = ""

// navigator.getUserMedia = navigator.getUserMedia || navigator.mozGetUserMedia || navigator.webkitGetUserMedia;
window.RTCPeerConnection = window.RTCPeerConnection || window.mozRTCPeerConnection || window.webkitRTCPeerConnection;
window.RTCIceCandidate = window.RTCIceCandidate || window.mozRTCIceCandidate || window.webkitRTCIceCandidate;
window.RTCSessionDescription = window.RTCSessionDescription || window.mozRTCSessionDescription || window.webkitRTCSessionDescription;

function connectWebRTC(message) {
    console.log("Received offer.\n")
    // const mediaStreamConstraints = {
    //     audio: true,
    // }
    // const localAudio = document.querySelector('#audio')
    // console.log("Connected! Makes webrtc connection with server. Data:\n")
    // console.log(message.command)
    offer = message.command
    // // var config = {
    // //     iceServers: [{urls: 'stun:stun.l.google.com:19302'}]
    // // }
    // const pc = new RTCPeerConnection()
    // pc.createOffer()
    //     .then(offer => pc.setLocalDescription(new RTCSessionDescription(offer)))
    //     // .then(an => )
    // var message = {
    //     command: offer
    // }
    // console.log("Sending offer.\n")
    // socket.send(JSON.stringify(message))
    let log = msg => {
        console.log(msg)
    }
    let pcConfig = {iceServers: [{urls: 'stun:stun.l.google.com:19302'}]}
    const pc = new RTCPeerConnection(pcConfig)
    let sendChannel = pc.createDataChannel('audio')
    sendChannel.onclose = () => console.log('sendChannel has closed')
    sendChannel.onopen = () => console.log('sendChannel has opened')
    sendChannel.onmessage = e => log(`Message from DataChannel '${sendChannel.label}' payload '${e.data}'`)
    pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
    pc.onnegotiationneeded = e => pc.createOffer().then(offer => pc.setLocalDescription(offer)).catch(log)
    socket.send(JSON.stringify(message))
    
    // The server is the initiator of the connection and created an offer for 
    // us that we received (message), now we need to send an answer back
    // let pcConfig = {iceServers: [{urls: 'stun:stun.l.google.com:19302'}]}
    // const pc = new RTCPeerConnection(pcConfig)
    // pc.setRemoteDescription(message).then(function () {
    //     return navigator.mediaDevices.getUserMedia(null);
    // }).then(function(stream) {
    //     document.getElementById("audio").srcObject = stream;
    //     return pc.addStream(stream);
    // }).then(function() {
    //     return pc.createAnswer();
    // }).then(function(answer) {
    //     return pc.setLocalDescription(answer);
    // }).then(function() {
    //     // Send the answer to the remote peer using the signaling server
    // }).catch(function (e) { console.log(e) })    


}

// Sets the lobby name field to the lobby name the user entered
function setLobbyNameField() {
    lobbyNameField.innerHTML = lobbyName
}

// Hides the homepage and unhides the page for entering the lobby name
function createLobbyName() {
    console.log("Hide homepage and reveal page to create lobby name.")
    homepage.setAttribute("hidden", true)
    lobbynamepage.removeAttribute("hidden")
}

// Hides the lobby name page and unhides the page for entering the user name
function createUserName() {
    console.log("Hide lobbynamepage and reveal page to create user name.")
    lobbyName = lobbyNameInput.value
    lobbynamepage.setAttribute("hidden", true)
    usernamepagecreate.removeAttribute("hidden")
}

// Hides the user name page and unhides the lobby page
// Sends the server a command to create a new lobby
function createLobby() {
    console.log("Hide usernamepagecreate and reveal lobby page.")
    userName = userNameInputCreate.value
    // send CREATE <lobby name> <username> to server and get lobby back then render lobby page
    var message = {
        command: "CREATE " + lobbyName + " " + userName 
    }
    socket.send(JSON.stringify(message))
    console.log("Sent CREATE <" + lobbyName + "> <" + userName + "> command to server.")
    usernamepagecreate.setAttribute("hidden", true)
    setLobbyNameField()
    lobbypage.removeAttribute("hidden")
}

// Hide user name page and unhide the waiting to join lobby page
// Sends the server a command to join a lobby
function joinLobby() {
    console.log("Hide usernamepagejoin and reveal waiting page.")
    userName = userNameInput.value
    // send JOIN <lobby name> <username> to server and get lobby back then render lobby page
    var message = {
        command: "JOIN " + lobbyName + " " + userName 
    }
    socket.send(JSON.stringify(message))
    console.log("Sent JOIN <" + lobbyName + "> <" + userName + "> command to server.")
    usernamepagejoin.setAttribute("hidden", true)
    setLobbyNameField()
    waitingpage.removeAttribute("hidden")
}

// Show waiting page
function enterLobby() {
    console.log("Trying to join " + lobbyName)
    lobbyNameFieldWait.innerHTML = lobbyName
    homepage.setAttribute("hidden", true)
    usernamepagejoin.removeAttribute("hidden")
}

createLobbyBtn.addEventListener("click", function(){
    createLobbyName()
})

createLobbyNameBtn.addEventListener("click", function(){
    createUserName()
})

createUserNameCreateBtn.addEventListener("click", function(){
    createLobby()
})

createUserNameJoinBtn.addEventListener("click", function(){
    joinLobby()
})

uploadSongBtn.addEventListener("click", function() {
    songUploadId.click()
})

songUploadId.onchange = function() {
    var reader = new FileReader()
    reader.onload = function(e) {
        sound.src = this.result
        // sound.play()
    }
    reader.readAsDataURL(this.files[0])
}

playBtn.addEventListener("click", function() {
    sound.play()
})

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
    // console.log('Message from server ', event.data);
    var message = JSON.parse(event.data)
    // console.log("Command from server: " + message.command)
    var commands = (message.command.split(/(\s+)/)).filter(function(e) {
        return String(e).trim();
    })
    console.log(commands)
    switch (commands[0]) {
        case "SERVERLIST":
            console.log("Got SERVERLIST from server.")
            var serverList = commands.slice(1);
            console.log("List of servers: " + serverList)
            autocomplete(searchLobbyInput, serverList)
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
        case "v=0":
            console.log("Needs to make WebRTC connection here.")
            connectWebRTC(message)
            break;
        default:
            console.error("Unknown command: " + command)
            break;
    }

})

memberListBtn.addEventListener('click', function() {
    membersModal.setAttribute("open",true)
})

inviteListBtn.addEventListener('click', function() {
    joinRequestsModal.setAttribute("open",true)
})

settingsBtn.addEventListener('click', function() {
    settingsModal.setAttribute("open",true)
})

membersModal.addEventListener('click', event => {
    if (event.target === membersModal) {
        membersModal.removeAttribute("open")
    }
})

joinRequestsModal.addEventListener('click', event => {
    if (event.target === joinRequestsModal) {
        joinRequestsModal.removeAttribute("open")
    }
})

settingsModal.addEventListener('click', event => {
    if (event.target === settingsModal) {
        settingsModal.removeAttribute("open")
    }
})

/* renders an autocomplete form given an element to append form to and the array
of form fields */
function autocomplete(inp, arr) {
    /*the autocomplete function takes two arguments,
    the text field element and an array of possible autocompleted values:*/
    var currentFocus;
    /*execute a function when someone writes in the text field:*/
    inp.addEventListener("input", function(e) {
        var a, b, i, val = this.value;
        /*close any already open lists of autocompleted values*/
        closeAllLists();
        if (!val) { return false;}
        currentFocus = -1;
        /*create a DIV element that will contain the items (values):*/
        a = document.createElement("DIV");
        a.setAttribute("id", this.id + "autocomplete-list");
        a.setAttribute("class", "autocomplete-items");
        /*append the DIV element as a child of the autocomplete container:*/
        this.parentNode.appendChild(a);
        /*for each item in the array...*/
        for (i = 0; i < arr.length; i++) {
          /*check if the item starts with the same letters as the text field value:*/
          if (arr[i].substr(0, val.length).toUpperCase() == val.toUpperCase()) {
            /*create a DIV element for each matching element:*/
            b = document.createElement("DIV");
            /*make the matching letters bold:*/
            b.innerHTML = "<strong>" + arr[i].substr(0, val.length) + "</strong>";
            b.innerHTML += arr[i].substr(val.length);
            /*insert a input field that will hold the current array item's value:*/
            b.innerHTML += "<input type='hidden' value='" + arr[i] + "'>";
            /*execute a function when someone clicks on the item value (DIV element):*/
                b.addEventListener("click", function(e) {
                /*insert the value for the autocomplete text field:*/
                inp.value = this.getElementsByTagName("input")[0].value;
                /*close the list of autocompleted values,
                (or any other open lists of autocompleted values:*/
                closeAllLists();
            });
            a.appendChild(b);
          }
        }
    });
    /*execute a function presses a key on the keyboard:*/
    inp.addEventListener("keydown", function(e) {
        var x = document.getElementById(this.id + "autocomplete-list");
        if (x) x = x.getElementsByTagName("div");
        if (e.keyCode == 40) {
          /*If the arrow DOWN key is pressed,
          increase the currentFocus variable:*/
          currentFocus++;
          /*and and make the current item more visible:*/
          addActive(x);
        } else if (e.keyCode == 38) { //up
          /*If the arrow UP key is pressed,
          decrease the currentFocus variable:*/
          currentFocus--;
          /*and and make the current item more visible:*/
          addActive(x);
        } else if (e.keyCode == 13) {
          /*If the ENTER key is pressed, prevent the form from being submitted,*/
          e.preventDefault();
          if (currentFocus > -1) {
            /*and simulate a click on the "active" item:*/
            if (x) x[currentFocus].click();
            // Get text from field and submit
            lobbyName = x[currentFocus].textContent || ""
            enterLobby()
          }
        }
    });
    function addActive(x) {
      /*a function to classify an item as "active":*/
      if (!x) return false;
      /*start by removing the "active" class on all items:*/
      removeActive(x);
      if (currentFocus >= x.length) currentFocus = 0;
      if (currentFocus < 0) currentFocus = (x.length - 1);
      /*add class "autocomplete-active":*/
      x[currentFocus].classList.add("autocomplete-active");
    }
    function removeActive(x) {
      /*a function to remove the "active" class from all autocomplete items:*/
      for (var i = 0; i < x.length; i++) {
        x[i].classList.remove("autocomplete-active");
      }
    }
    function closeAllLists(elmnt) {
      /*close all autocomplete lists in the document,
      except the one passed as an argument:*/
      var x = document.getElementsByClassName("autocomplete-items");
      for (var i = 0; i < x.length; i++) {
        if (elmnt != x[i] && elmnt != inp) {
        x[i].parentNode.removeChild(x[i]);
      }
    }
  }
  /*execute a function when someone clicks in the document:*/
  document.addEventListener("click", function (e) {
      closeAllLists(e.target);
  });
} 