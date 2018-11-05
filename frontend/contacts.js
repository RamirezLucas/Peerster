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
    changeMessageBoxTest(this.innerHTML)

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
    var newContact = document.createElement("div");
    newContact.className = "private_contact_wrap";
    newContact.innerHTML = name;
    document.getElementById('contact_scrollable_wrap').appendChild(newContact);
    contact_attach_listeners(newContact);

    // Create new chat history
    var newChat = document.createElement("div");
    newChat.className = "conversation";
    newChat.id = name;
    newChat.innerHTML =
        '<div class="begin_conv"><span>This is the beginning of \
        your conversation with ' + name + '</span></div>';
    document.getElementById('chat_scrollable_wrap').appendChild(newChat);
}

function whoAmI() {
     
    // POST data
    var xhr = new XMLHttpRequest();
    var url = "/id";
    xhr.open("GET", url, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (xhr.responseText !== "") {
                var json = JSON.parse(xhr.responseText); // Parse JSON
                if(json.hasOwnProperty("name")){ // Check that the key exists
                    document.getElementById("myID").innerText = 'MESSAGES (gossiping as ' + json.name + ')'
                }
            }
        }
    };
    
    xhr.send();
}

/* ------- ONLOAD ------- */

function contact_attach_listeners(contact) {
    contact.addEventListener("click", switchContact);
    contact.addEventListener("mouseenter", contactMouseEnter);
    contact.addEventListener("mouseleave", contactMouseLeave);
}

var curr_contact = null
window.onload = function(){
    // Attach event listeners to all existing contacts
    var contacts = document.getElementsByClassName("private_contact_wrap");
    for (var i = 0 ; i < contacts.length ; i++) {
        contact_attach_listeners(contacts[i]);
    }

    // Get my own name and IP:PORT address
    whoAmI()

    // Initial call to refresh the page
    refresh()    

};