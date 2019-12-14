var currentFile = ""

function loadVideo(file) {
    document.getElementById("mp4_src").src = file
    document.getElementById("video").load()
    currentFile = file
}

function startVideo() {
    document.getElementById("video").play()
}

function pauseVideo() {
    document.getElementById("video").pause()
}

function clearVideo() {
    document.getElementById("video").pause()
    document.getElementById("video").currentTime = 0
    document.getElementById("mp4_src").src = ""
    document.getElementById("video").load()
    currentFile = ""
}

document.addEventListener('DOMContentLoaded', function(){ 
    document.getElementById("video").addEventListener("ended", function() {
        console.log("play ended")
        window.signalEndPlay(currentFile)
        clearVideo()
    })
}, false);