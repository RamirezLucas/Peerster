function contact_switch() {

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
    // TODO: change name in send message box

    // Update current contact
    curr_contact = this

}

function contact_mouse_enter() {
    if (curr_contact === null || curr_contact.innerHTML != this.innerHTML) {
        this.style.backgroundColor = 'rgb(' + 54 + ',' + 57 + ',' + 63 + ')';
        this.style.color = 'rgb(' + 255 + ',' + 255 + ',' + 255 + ')';
    }
}

function contact_mouse_leave() {
    if (curr_contact === null || curr_contact.innerHTML != this.innerHTML) {
        this.style.backgroundColor = 'rgb(' + 47 + ',' + 49 + ',' + 54 + ')';
        this.style.color = 'rgb(' + 105 + ',' + 106 + ',' + 110 + ')';
    }
}

function contact_attach_listeners(contact) {
    contact.addEventListener("click", contact_switch);
    contact.addEventListener("mouseenter", contact_mouse_enter);
    contact.addEventListener("mouseleave", contact_mouse_leave);
}

var curr_contact = null
window.onload = function(){
    
    // Attach event listeners to all existing contacts
    var contacts = document.getElementsByClassName("private_contact_wrap");
    for (var i = 0 ; i < contacts.length ; i++) {
        contact_attach_listeners(contacts[i]);
    }
    
};