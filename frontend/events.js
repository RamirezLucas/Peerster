function switch_conv(contact) {

    // If the contact was already selected, do nothing
    if (curr_contact !== null && contact.innerHTML === curr_contact.innerHTML) {
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
    contact.style.backgroundColor = 'rgb(' + 66 + ',' + 70 + ',' + 77 + ')';
    contact.style.color = 'rgb(' + 255 + ',' + 255 + ',' + 255 + ')';
    var newConv = document.getElementById(contact.innerHTML)
    newConv.style.display = "block"

    // Update current contact
    curr_contact = contact

}

var curr_contact = null
window.onload = function(){
    var tmp = document.getElementById("default_conv")
    switch_conv(tmp)
};