function getUpdates() {

    var xhr = new XMLHttpRequest();
    xhr.open("GET", "/updates", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (xhr.responseText !== "") {
                var json = JSON.parse(xhr.responseText); // Parse JSON
                if(json.hasOwnProperty("updates")){ // Check that the key exists
                    for (var i = 0; i < json.updates.length; i++) {
                        var update = json.updates[i]; // Get update
                        if (update.Rumor !== null) {
                            // This is a rumor
                            appendMessage("Global Channel", update.Rumor.Name, update.Rumor.Msg)
                        } else if (update.Peer !== null) {
                            // This is a peer
                            addPeer(update.Peer.IP + ":" + update.Peer.Port)
                        } else if (update.PrivateMessage !== null) {
                            // This is a private message
                            appendMessage(update.PrivateMessage.Name, update.PrivateMessage.Name, update.PrivateMessage.Msg)
                        } else if (update.PrivateContact !== null) {
                            // This is a private contact
                            addContact(update.PrivateContact.Name)
                        }
                    }
                }            
            }
        }
    };

    xhr.send();
}

function refresh() {
    getUpdates()
    setTimeout(refresh, 3000);
}
