function checkNewPeer(e) {
    
    let code = (e.keyCode ? e.keyCode : e.which);

    if (code == 13) {
        // Don't create a newline
        e.preventDefault();

        // Send new peer
        let newPeer = document.getElementById("new_peer").value;
        sendPeer(newPeer);

        // Reset textarea
        document.getElementById("new_peer").value = "";
    }
}

function sendPeer(addr) {
    
    // POST data
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/node", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    let data = JSON.stringify({"peer": addr});
    xhr.send(data);

}

function addPeer(address) {

    // Create new contact tab
    let new_peer = document.createElement("div");
    new_peer.className = "peer_wrap";
    new_peer.innerHTML = '<span>' + address + '</span>'
    document.getElementById('peers_scrollable_wrap').appendChild(new_peer);

}