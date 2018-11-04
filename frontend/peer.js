function check_new_peer(e) {
    
    var code = (e.keyCode ? e.keyCode : e.which);

    if (code == 13) {
        // Don't create a newline
        e.preventDefault();

        if (e.ctrlKey) {
            // TODO: Send new peer
            var textarea = document.getElementById("new_peer");
                    
            // Reset textarea
            document.getElementById("new_peer").value = "";
        }
    }
}

function peer_add(address) {

    // Create new contact tab
    var new_peer = document.createElement("div");
    new_peer.className = "peer_wrap";
    new_peer.innerHTML = '<span>' + address + '</span>'
    document.getElementById('peers_scrollable_wrap').appendChild(new_peer);

}