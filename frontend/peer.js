function checkNewPeer(e) {
    
    var code = (e.keyCode ? e.keyCode : e.which);

    if (code == 13) {
        // Don't create a newline
        e.preventDefault();

        if (e.ctrlKey) {

            // Send new peer
            var newPeer = document.getElementById("new_peer").value;
            sendToServer("peer", newPeer, "/node")

            // Reset textarea
            document.getElementById("new_peer").value = "";
        }
    }
}

function addPeer(address) {

    // Create new contact tab
    var new_peer = document.createElement("div");
    new_peer.className = "peer_wrap";
    new_peer.innerHTML = '<span>' + address + '</span>'
    document.getElementById('peers_scrollable_wrap').appendChild(new_peer);

}