/* ------- EVENTS ------- */

function switchContact() {

    // If the contact was already selected, do nothing
    if (curr_contact !== null && this.innerHTML === curr_contact.innerHTML) {
        return
    }

    if (curr_contact !== null) {
        // Unselect current contact and hide current conversation
        curr_contact.style.backgroundColor = 'rgb(' + 47 + ',' + 49 + ',' + 54 + ')';
        curr_contact.style.color = 'rgb(' + 105 + ',' + 106 + ',' + 110 + ')';
        var prevConv = document.getElementById(curr_contact.innerHTML)
        prevConv.style.display = "none"
    }
  
    // Select new contact and show new conversation
    this.style.backgroundColor = 'rgb(' + 66 + ',' + 70 + ',' + 77 + ')';
    this.style.color = 'rgb(' + 255 + ',' + 255 + ',' + 255 + ')';
    var newConv = document.getElementById(this.innerHTML)
    newConv.style.display = "block"
    
    // Change name in textbox
    changeMessageBoxText(this.innerHTML)

    // Update current contact
    curr_contact = this
}

function contactMouseEnter() {
    if (curr_contact === null || curr_contact.innerHTML != this.innerHTML) {
        this.style.backgroundColor = 'rgb(' + 54 + ',' + 57 + ',' + 63 + ')';
        this.style.color = 'rgb(' + 255 + ',' + 255 + ',' + 255 + ')';
    }
}

function contactMouseLeave() {
    if (curr_contact === null || curr_contact.innerHTML != this.innerHTML) {
        this.style.backgroundColor = 'rgb(' + 47 + ',' + 49 + ',' + 54 + ')';
        this.style.color = 'rgb(' + 105 + ',' + 106 + ',' + 110 + ')';
    }
}

/* ------- API ------- */

function addContact(name) {

    // Create new contact tab
    let newContact = document.createElement("div");
    newContact.className = "private_contact_wrap";
    newContact.innerHTML = name;
    document.getElementById('contact_scrollable_wrap').appendChild(newContact);
    contact_attach_listeners(newContact);

    // Create new chat history
    let newChat = document.createElement("div");
    newChat.className = "conversation";
    newChat.id = name;
    newChat.innerHTML =
        '<div class="begin_conv"><span>This is the beginning of \
        your conversation with ' + name + '</span></div>';
    document.getElementById('chat_scrollable_wrap').appendChild(newChat);
}

function whoAmI() {
     
    // POST data
    let xhr = new XMLHttpRequest();
    let url = "/id";
    xhr.open("GET", url, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (xhr.responseText !== "") {
                let json = JSON.parse(xhr.responseText); // Parse JSON
                if(json.hasOwnProperty("name") && json.hasOwnProperty("addr")){ // Check that the keys exist
                    document.getElementById("my_name").innerHTML = json.name
                    document.getElementById("my_address").innerHTML = json.addr                
                }
            }
        }
    };
    
    xhr.send();
}