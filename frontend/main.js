function contact_attach_listeners(contact) {
    contact.addEventListener("click", switchContact);
    contact.addEventListener("mouseenter", contactMouseEnter);
    contact.addEventListener("mouseleave", contactMouseLeave);
}

var curr_contact = null
window.onload = function(){
    // Attach event listeners to all existing contacts
    let contacts = document.getElementsByClassName("private_contact_wrap");
    for (let i = 0 ; i < contacts.length ; i++) {
        contact_attach_listeners(contacts[i]);
    }

    // Get my own name and IP:PORT address
    whoAmI()

    // Initial call to refresh the page
    refresh()    

};