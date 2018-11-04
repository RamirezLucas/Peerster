function change_message_box_text(gossipingTo) {
    // Enable textarea
    document.getElementById("send_message_wrap").style.display = "block";
    // Chnage text    
    document.getElementById("send_message").setAttribute("placeholder", 
        "> Message " + gossipingTo + " (Ctrl + Enter to send)");
}   

function check_send(e) {
    
    var code = (e.keyCode ? e.keyCode : e.which);
    if (e.ctrlKey && code == 13) {
        // Don't create a newline
        e.preventDefault();

        // TODO: Send message
        var textarea = document.getElementById("send_message");
        

        // Reset textarea
        document.getElementById("send_message").value = "";
    }

}