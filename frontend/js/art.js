/* ------- Adding artists ------- */

function addArtist(name, signature) {

    // Create new artist tab
    let newArtist = document.createElement("div");
    newArtist.id = "artist_" + name;
    newArtist.className = "private_wrap";
    newArtist.innerHTML = name

    // Create new subscribe button
    let subscribeButton = document.createElement("div");
    subscribeButton.className = "button subscribe";
    subscribeButton.onclick = onSubscribe(signature);
    subscribeButton.innerHTML = 'Subscribe';
    subscribeButton.id = signature
    newArtist.appendChild(subscribeButton);

    // Spawn artist
    document.getElementById('artists').appendChild(newArtist);
    // contactAttachListeners(newArtist);

    // Create new chat history
    let newChat = document.createElement("div");
    newChat.className = "conversation";
    newChat.id = "chat_artist_" + name;
    newChat.innerHTML =
        '<div class="begin_conv"><span>This is ' + name + '\'s  showroom</span></div>';
    document.getElementById('chat_scrollable_wrap').appendChild(newChat);
}

function onSubscribe(signature) {
    return function() {
        // Change button style
        this.style.backgroundColor = 'rgb(' + 114 + ',' + 255 + ',' + 109 + ')';
        this.style.borderColor = 'rgb(' + 114 + ',' + 255 + ',' + 109 + ')';
        this.innerHTML = "Subscribed!";
        contactAttachListeners(this.parentElement)

        // POST data
        let xhr = new XMLHttpRequest();
        xhr.open("POST", "/subscribe", true);
        xhr.setRequestHeader("Content-Type", "application/json");
        let data = JSON.stringify({"signature": signature});
        xhr.send(data);
    };
}

/* ------- Adding artworks ------- */

function downloadArtwork(artistName, metahash, name, description) {

    var xhr = getXMLHttpRequest();
    xhr.open("GET", "/download", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.overrideMimeType('text/plain; charset=x-user-defined');
    let data = JSON.stringify({"filename": filename});
    
    xhr.onreadystatechange= function() {
        console.log("status: " + xhr.status)
        if(xhr.readyState==4 && xhr.status==200) {
            console.log("length: " + xhr.responseText.length)
            
            str = "data:image/jpg;base64," + encode64(xhr.responseText);

            let newArtwork = document.createElement("div");
            newArtwork.id = metahash;
            newArtwork.className = "artwork_wrap";
            newArtwork.innerHTML = '\
                <div class="artwork_title">' + name + '</div>\n\
                <div class="artwork_desc">' + description + '</div>\n\
                <div class="artwork"><img src="' + str + '"></div>\
                ';

            document.getElementById("chat_artist_" + artistName).appendChild(newArtwork)
        }
    };
    xhr.send(data);

}

/* ------- File Downloading ------- */

function getXMLHttpRequest() {
	var xhr = null;
	
	if (window.XMLHttpRequest || window.ActiveXObject) {
		if (window.ActiveXObject) {
			try {
				xhr = new ActiveXObject("Msxml2.XMLHTTP");
			} catch(e) {
				xhr = new ActiveXObject("Microsoft.XMLHTTP");
			}
		} else {
			xhr = new XMLHttpRequest(); 
		}
	} else {
		alert("Your browser does not support XMLHTTP");
		return null;
	}
	
	return xhr;
}

function encode64(inputStr) 
{
    var b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";
    var outputStr = "";
    var i = 0;
    
    while (i<inputStr.length) {
        //all three "& 0xff" added below are there to fix a known bug 
        //with bytes returned by xhr.responseText
        var byte1 = inputStr.charCodeAt(i++) & 0xff;
        var byte2 = inputStr.charCodeAt(i++) & 0xff;
        var byte3 = inputStr.charCodeAt(i++) & 0xff;

        var enc1 = byte1 >> 2;
        var enc2 = ((byte1 & 3) << 4) | (byte2 >> 4);
        
        var enc3, enc4;
        if (isNaN(byte2)) {
            enc3 = enc4 = 64;
        } else {
            enc3 = ((byte2 & 15) << 2) | (byte3 >> 6);
            if (isNaN(byte3)) {
                enc4 = 64;
            } else {
                enc4 = byte3 & 63;
            }
        }
        outputStr +=  b64.charAt(enc1) + b64.charAt(enc2) + b64.charAt(enc3) + b64.charAt(enc4);
    } 
    
    return outputStr;
}
