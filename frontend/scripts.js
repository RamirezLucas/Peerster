function sendNewMessage() {
    
    // Get message from text field
    var msg = document.getElementById("new_msg").value
    if (msg !== "") {
       
        // POST data
        var xhr = new XMLHttpRequest();
        var url = "/message";
        xhr.open("POST", url, true);
        xhr.setRequestHeader("Content-Type", "application/json");
        var data = JSON.stringify({"msg": String(msg)});
        xhr.send(data);

        document.getElementById("new_msg").value = ""

    }
}

function sendNewPeer() {

    // Get message from text field
    var peer = document.getElementById("new_peer").value
    if (peer !== "") {
       
        // POST data
        var xhr = new XMLHttpRequest();
        var url = "/node";
        xhr.open("POST", url, true);
        xhr.setRequestHeader("Content-Type", "application/json");
        var data = JSON.stringify({"peer": String(peer)});
        xhr.send(data);

        document.getElementById("new_peer").value = ""

    }
}

function refresh() {
    refreshMessages("/message")
    refreshPeers("/node")
    setTimeout(refresh, 3000);
}

function refreshMessages(request_path) {

    var xhr = new XMLHttpRequest();
    xhr.open("GET", request_path, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (xhr.responseText !== "") {
                var json = JSON.parse(xhr.responseText); // Parse JSON
                if(json.hasOwnProperty("messages")){ // Check that the key exists
                    for (var i = 0; i < json.messages.length; i++) {
                        var msg = json.messages[i]; // Get message
    
                        // Create new HTML element
                        var div = document.createElement("div");
                        if (messageID % 2 === 0) {
                            div.className = "message_box";
                        } else {
                            div.className = "message_box darker";
                        }
                        div.innerHTML = 
                            '<span style="font-weight:bold">' + msg.Name + '</span>\
                            <p>' + msg.Msg + '</p>';
    
                        // Append new element to list
                        document.getElementById('chat').appendChild(div);

                        // Increment the message counter
                        messageID += 1;
    
                    }
                }   
            }
        }
    };
    xhr.send();

}

function refreshPeers(request_path) {

    var xhr = new XMLHttpRequest();
    xhr.open("GET", request_path, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (xhr.responseText !== "") {
                var json = JSON.parse(xhr.responseText); // Parse JSON
                if(json.hasOwnProperty("peers")){ // Check that the key exists
                    for (var i = 0; i < json.peers.length; i++) {
                        var peer = json.peers[i]; // Get peer

                        // Create new HTML element
                        var div = document.createElement("div");
                        div.className = "peer_box";
                        div.innerHTML = 
                            ' <span>IP:   <b>' + peer.IP + '</b></span><br>\
                            <span>Port: <em>' + peer.Port + '</em></span>';

                        // Append new element to list
                        document.getElementById('peers').appendChild(div);

                    }
                }            
            }
        }
    };
    xhr.send();

}

function sayMyName() {
     
    // POST data
     var xhr = new XMLHttpRequest();
     var url = "/id";
     xhr.open("GET", url, true);
     xhr.setRequestHeader("Content-Type", "application/json");
     xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (xhr.responseText !== "") {
                var json = JSON.parse(xhr.responseText); // Parse JSON
                if(json.hasOwnProperty("name")){ // Check that the key exists
                    document.getElementById("myID").innerText = 'MESSAGES (gossiping as ' + json.name + ')'
                }
            }
        }
    };
     xhr.send();

}

var messageID = 0;

// Get our name
sayMyName()

// Get the current state
refreshMessages("/in_message")
refreshPeers("/in_node")

// Initial call to refresh the page
setTimeout(refresh, 3000);
