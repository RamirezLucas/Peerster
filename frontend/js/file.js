function indexNewFile() {
    document.getElementById('file-input').click();
}

function getFilename() {
    // Get filename
    let input = document.getElementById("file-input").value.replace(/\\/g, "/");
    let filename = input.substring(input.lastIndexOf("/") + 1);
    
    // Send file index request
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/fileIndex", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    let data = JSON.stringify({"filename": filename});
    xhr.send(data); 
}

function remoteFileRequest() {
    let filename = document.getElementById("new_filename").value;
    let metahash = document.getElementById("metahash").value;
    if (filename !== "" && metahash !== "" && curr_contact.innerHTML !== "Global Channel") {
        
        // Send remote file request
        let xhr = new XMLHttpRequest();
        xhr.open("POST", "/fileRequest", true);
        xhr.setRequestHeader("Content-Type", "application/json");
        let data = JSON.stringify({ "filename": filename,
                                    "metahash": metahash,
                                    "destination": curr_contact.innerHTML});
        xhr.send(data);

        // Clear fields
        document.getElementById("new_filename").value = ""
        document.getElementById("metahash").value = ""
    }
}

function addIndexedFile(filename, metahash) {

    // Create new indexed file
    let newFile = document.createElement("div");
    newFile.className = "file_wrap";
    newFile.innerHTML = '<div class="filename">' + filename + '</div>\
                        <div class="metahash">' + metahash + '</div>'

    document.getElementById('indexed_files').appendChild(newFile);

}

function addConstructingFile(filename, metahash, origin) {

    // Create new indexed file
    let newFile = document.createElement("div");
    newFile.className = "file_wrap";
    newFile.innerHTML = '<div class="filename">' + filename + ' <em>from ' + origin + '</em></div>\
                        <div class="metahash">' + metahash + '</div>'

    document.getElementById('reconstructing_files').appendChild(newFile);

}

function addAvailableFile(filename, metahash, origin) {

    // Create new indexed file
    let newFile = document.createElement("div");
    newFile.className = "file_wrap";
    newFile.innerHTML = '<div class="filename">' + filename + '</div>\
                        <div class="metahash">' + metahash + '</div>'

    document.getElementById('available_files').appendChild(newFile);

}

function removeFile(metahash, category) {
    
    let constructingFiles = document.getElementById(category);
    let children = constructingFiles.children;
    for (let i = 0; i < children.length; i++) {
        if (children[i].innerHTML.includes(metahash)) {
            constructingFiles.removeChild(children[i]);
            return
        }
    }
}

function checkSearchRequest(e) {
    
    let code = (e.keyCode ? e.keyCode : e.which);

    if (code == 13) {
        // Don't create a newline
        e.preventDefault();

        // Send new peer
        let request = document.getElementById("search_request").value;
        sendSearchRequest(request);

        // Reset textarea
        document.getElementById("search_request").value = "";
    }
}

function sendSearchRequest(request) {
    
    // Split the search string
    let splitted = request.split(":")
    if (splitted.length != 2) {
        alert("bitch lasagna")
        return
    } 

    // POST data
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/fileSearch", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    let data = JSON.stringify({"budget": splitted[0], "keywords": splitted[1]});
    xhr.send(data);

}
