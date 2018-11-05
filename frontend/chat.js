function changeMessageBoxTest(gossipingTo) {
    // Enable textarea
    document.getElementById("send_message_wrap").style.display = "block";
    // Chnage text    
    document.getElementById("send_message").setAttribute("placeholder", 
        "> Message " + gossipingTo + " (Ctrl + Enter to send)");
}   

function checkSend(e) {
    
    var code = (e.keyCode ? e.keyCode : e.which);
    if (e.ctrlKey && code == 13) {
        // Don't create a newline
        e.preventDefault();

        // Send message
        var newMsg = document.getElementById("send_message").value;
        if (curr_contact.innerHTML === "Global Channel") {
            sendToServer("message", newMsg, "/rumor")        
        } else {
            sendToServer("message", newMsg, "/private")        
        }

        // Reset textarea
        document.getElementById("send_message").value = "";
    }

}

function appendMessage(channel, sender, msg_content) {

    /* The function assumes that the channel already exists */
    
    // Check who was the last to talk on the channel
    var conversation = document.getElementById(channel);
    var childs_conv = conversation.children;
    var last_monologue = null;
    var new_monologue = null;

    // No one has talked on this channel yet
    if (childs_conv.length === 1) {
        new_monologue = create_monologue(sender);
    } else {
        last_monologue = childs_conv[childs_conv.length - 1];
        var last_author = last_monologue.children[0].innerHTML;
        // Check the last person who talked
        if (last_author !== sender) {
            new_monologue = create_monologue(sender);
        }
    }

    // Append the message to the last monologue
    var new_msg = '<div class="message">' + msg_content + '</div>';

    // Append a new monologue if necessary
    if (new_monologue !== null) {
        new_monologue.innerHTML += new_msg;
        document.getElementById(channel).appendChild(new_monologue);
    } else {
        last_monologue.innerHTML += new_msg;
    }

}

function create_monologue(author) {
    
    // Create new contact tab
    var new_contact = document.createElement("div");
    new_contact.className = "monologue";
    new_contact.innerHTML = '<div class="author">' + author + '</div>';
    return new_contact;
    
}