function getUpdates() {

    let xhr = new XMLHttpRequest();
    xhr.open("GET", "/updates", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (xhr.responseText !== "") {
                let json = JSON.parse(xhr.responseText); // Parse JSON
                if(json.hasOwnProperty("updates")){ // Check that the key exists
                    for (let i = 0; i < json.updates.length; i++) {
                        let update = json.updates[i]; // Get update
                        if (update.Rumor !== null) {
                            // This is a rumor
                            appendMessage("Global Channel", update.Rumor.Name, update.Rumor.Msg)
                        } else if (update.Peer !== null) {
                            // This is a peer
                            addPeer(update.Peer.IP + ":" + update.Peer.Port)
                        } else if (update.PrivateMessage !== null) {
                            // This is a private message
                            if (update.PrivateMessage.Origin === document.getElementById("my_name").innerHTML) {
                                appendMessage(update.PrivateMessage.Destination, update.PrivateMessage.Origin, update.PrivateMessage.Msg)                                
                            } else {
                                appendMessage(update.PrivateMessage.Origin, update.PrivateMessage.Origin, update.PrivateMessage.Msg)
                            }
                        } else if (update.PrivateContact !== null) {
                            // This is a private contact
                            addContact(update.PrivateContact.Name)
                        } else if (update.IndexedFile !== null) {
                            // This is a new indexed file
                            addIndexedFile(update.IndexedFile.Filename, update.IndexedFile.Metahash)
                            removeConstructingFile(update.IndexedFile.Metahash)
                        } else if (update.ConstructingFile !== null) {
                            // This is a new file in construction
                            addConstructingFile(update.IndexedFile.Filename, update.IndexedFile.Metahash, update.IndexedFile.Origin)
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
    setTimeout(refresh, 500);
}
