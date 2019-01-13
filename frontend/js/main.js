function contactAttachListeners(contact) {
    contact.addEventListener("click", switchContact);
    contact.addEventListener("mouseenter", contactMouseEnter);
    contact.addEventListener("mouseleave", contactMouseLeave);
}

let curr_contact = null
window.onload = function(){
    
    // Create the global channel
    addGroup("Global")

    // Get my own name and IP:PORT address
    whoAmI()

    // Initial call to refresh the page
    refresh()  

};