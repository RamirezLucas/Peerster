function changeMessageBoxText(gossipingTo) {
    // Enable textarea
    document.getElementById("send_message_wrap").style.display = "block";
    document.getElementById("request_file_wrap").style.display = "block";
    // Change text    
    document.getElementById("send_message").setAttribute("placeholder", 
        "> Message " + gossipingTo + " (Ctrl + Enter to send)");
    document.getElementById("request_file_btn").innerHTML = 
        "Request file from " + gossipingTo;
}   

function sendMessage(e) {
    
    let code = (e.keyCode ? e.keyCode : e.which);
    if (e.ctrlKey && code == 13) {
        // Don't create a newline
        e.preventDefault();

        // Send message
        let newMsg = document.getElementById("send_message").value;
        if (curr_contact.innerHTML === "Global Channel") {
            let xhr = new XMLHttpRequest();
            xhr.open("POST", "/rumor", true);
            xhr.setRequestHeader("Content-Type", "application/json");
            let data = JSON.stringify({"message": newMsg});
            xhr.send(data);      
        } else {
            let xhr = new XMLHttpRequest();
            xhr.open("POST", "/private", true);
            xhr.setRequestHeader("Content-Type", "application/json");
            let data = JSON.stringify({ "destination": curr_contact.innerHTML,
                                        "message": newMsg});
            xhr.send(data);       
        }

        // Reset textarea
        document.getElementById("send_message").value = "";
    }

}

function suppressEnter(e) {
    
    let code = (e.keyCode ? e.keyCode : e.which);
    if (code == 13) {
        // Don't create a newline
        e.preventDefault();
    }
}

function appendMessage(channel, sender, msg_content) {

    /* The function assumes that the channel already exists */

    let idChat = "chat_private_" + channel;
    if (channel === "Global") {
        idChat = "chat_group_" + channel;
    }

    // Check who was the last to talk on the channel
    let childs_conv = document.getElementById(idChat).children;
    let last_monologue = null;
    let new_monologue = null;

    // No one has talked on this channel yet
    if (childs_conv.length === 1) {
        new_monologue = create_monologue(sender);
    } else {
        last_monologue = childs_conv[childs_conv.length - 1];
        let last_author = last_monologue.children[0].innerHTML;
        // Check the last person who talked
        if (last_author !== sender) {
            new_monologue = create_monologue(sender);
        }
    }

    // Append the message to the last monologue
    let new_msg = '<div class="message">' + msg_content + '</div>';

    // Append a new monologue if necessary
    if (new_monologue !== null) {
        new_monologue.innerHTML += new_msg;
        document.getElementById(idChat).appendChild(new_monologue);
    } else {
        last_monologue.innerHTML += new_msg;
    }

}

function displayArtwork(filename, name, description) {

    

}

function create_monologue(author) {
    
    // Create new contact tab
    let new_contact = document.createElement("div");
    new_contact.className = "monologue";
    new_contact.innerHTML = '<div class="author">' + author + '</div>';
    return new_contact;
    
}