function contactAttachListeners(contact) {
    contact.addEventListener("click", switchContact);
    contact.addEventListener("mouseenter", contactMouseEnter);
    contact.addEventListener("mouseleave", contactMouseLeave);
}

let curr_contact = null
window.onload = function(){
    // Attach event listeners to all existing contacts
    let contacts = document.getElementsByClassName("private_contact_wrap");
    for (let i = 0 ; i < contacts.length ; i++) {
        contactAttachListeners(contacts[i]);
    }

    // Get my own name and IP:PORT address
    whoAmI()

    // Initial call to refresh the page
    refresh()    

};