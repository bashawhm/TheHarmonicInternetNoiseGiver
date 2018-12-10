// Get pages
var homepage = document.querySelector("#homepage")
var lobbynamepage = document.querySelector("#lobbynamepage")
var usernamepagecreate = document.querySelector("#usernamepagecreate")
var usernamepagejoin = document.querySelector("#usernamepagejoin")
var lobbypage = document.querySelector("#lobbypage")
var waitingpage = document.querySelector("#waitingpage")
// Get inputs
var searchLobbyInput = document.querySelector("#searchLobbiesInput")
var lobbyNameInput = document.querySelector("#lobbyNameInput")
var userNameInput = document.querySelector("#userNameInput")
// Get buttons
var createLobbyBtn = document.querySelector("#createLobbyBtn")
var createLobbyNameBtn = document.querySelector("#createLobbyNameBtn")
var createUserNameCreateBtn = document.querySelector("#createUserNameCreateBtn")
var createUserNameJoinBtn = document.querySelector("#createUserNameJoinBtn")
// Get fields
var lobbyNameField = document.querySelector("#lobbyNameField")
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

function connectWebRTC(message) {
    // console.log("Connected! Makes webrtc connection with server. Data:\n")
    console.log(message)
    // // var config = {
    // //     iceServers: [{urls: 'stun:stun.l.google.com:19302'}]
    // // }
    // const pc = new RTCPeerConnection()

    // pc.createOffer()
    //     .then(offer => pc.setLocalDescription(new RTCSessionDescription(offer)))
    //     // .then(an => )

}

function setLobbyNameField() {
    lobbyNameField.innerHTML = lobbyName
}

function createLobbyName() {
    console.log("Hide homepage and reveal page to create lobby name.")
    homepage.setAttribute("hidden", true)
    lobbynamepage.removeAttribute("hidden")
}

function createUserName() {
    console.log("Hide lobbynamepage and reveal page to create user name.")
    lobbyName = lobbyNameInput.value
    lobbynamepage.setAttribute("hidden", true)
    usernamepagecreate.removeAttribute("hidden")
}

function createLobby() {
    console.log("Hide usernamepagecreate and reveal lobby page.")
    userName = userNameInput.value
    // send CREATE <lobby name> <username> to server and get lobby back then render lobby page
    var message = {
        command: "CREATE " + lobbyName + " " + userName 
    }
    socket.send(JSON.stringify(message))
    console.log("Sent CREATE <" + lobbyName + "> <" + userName + "> command to server.")
    usernamepagecreate.setAttribute("hidden", true)
    lobbypage.removeAttribute("hidden")
}

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

function enterLobby() {
    console.log("Trying to join " + lobbyName)
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