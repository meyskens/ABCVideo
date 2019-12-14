ABCVideo
========

ABCVideo is a video player written in Go and React for use in theater productions at ABCTheater. It is based on the sounboard app [ABCBoard](https://github.com/meyskens/abcboard).
Due lack of multimedia support in all current Go powered GUI frameworks this program makes use of a webview and the HTML5 video API.

## Building
I only tested this on Linux but it should work on Mac and Windows too. For build instructions I suggest looking at [webview](https://github.com/zserge/webview) and the `build.sh` script.