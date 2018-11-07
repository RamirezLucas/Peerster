function checkNewPeer(e) {
    
    var code = (e.keyCode ? e.keyCode : e.which);

    if (code == 13) {
        // Don't create a newline
        e.preventDefault();

        if (e.ctrlKey) {

            // Send new peer
            var newPeer = document.getElementById("new_peer").value;
            sendPeer(newPeer);

            // Reset textarea
            document.getElementById("new_peer").value = "";
        }
    }
}

function sendPeer(addr) {
    
    // POST data
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/node", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    var data = JSON.stringify({"peer": addr});
    xhr.send(data);

}

function addPeer(address) {

    // Create new contact tab
    var new_peer = document.createElement("div");
    new_peer.className = "peer_wrap";
    new_peer.innerHTML = '<span>' + address + '</span>'
    document.getElementById('peers_scrollable_wrap').appendChild(new_peer);

}