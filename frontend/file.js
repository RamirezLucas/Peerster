function indexNewFile() {
    document.getElementById('file-input').click();
}

function getFilename() {
    // Get filename
    let input = document.getElementById("file-input").value.replace(/\\/g, "/");
    let filename = input.substring(input.lastIndexOf("/") + 1);
    
    // Send file index request
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/file_index", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    let data = JSON.stringify({"filename": filename});
    xhr.send(data); 
}

function addIndexedFile(filename, metahash) {

    // Create new indexed file
    let newFile = document.createElement("div");
    newFile.className = "file_wrap";
    newFile.innerHTML = '<div class="filename">' + filename + '</div>\
                        <div class="metahash">' + metahash + '</div>'

    document.getElementById('files_indexed_scrollable_wrap').appendChild(newFile);

}

function addConstructingFile(filename, metahash, origin) {

    // Create new indexed file
    let newFile = document.createElement("div");
    newFile.className = "file_wrap";
    newFile.innerHTML = '<div class="filename">' + filename + ' <em>from ' + origin + '</em></div>\
                        <div class="metahash">' + metahash + '</div>'

    document.getElementById('files_construct_scrollable_wrap').appendChild(newFile);

}

function removeConstructingFile(metahash) {
    
    let constructingFiles = document.getElementById("files_construct_scrollable_wrap");
    let children = constructingFiles.children;
    for (let i = 0; i < children.length; i++) {
        if (children[i].innerHTML.includes(metahash)) {
            constructingFiles.removeChild(children[i]);
            return
        }
    }
}