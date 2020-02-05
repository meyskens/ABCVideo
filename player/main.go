package main

import (
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/zserge/lorca"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"regexp"
)

var player lorca.UI
var playerController = PlayerController{
	endVideo: make(chan string),
}

var playingVideo = false

// PlayerController helps with controling the player
type PlayerController struct {
	endVideo chan string
}

type EmptyRequest struct {
	Emptyness string // needed for RPC to be happy
}

type EmptyResponse struct {
	Emptyness string // needed for RPC to be happy
}

type FileRequest struct {
	File string
}

type FileResponse struct {
	File string
}

// SignalEndPlay gets called when a video ended by JS
func (p *PlayerController) SignalEndPlay(file string) {
	re := regexp.MustCompile(`^/videos/`)
	file = re.ReplaceAllString(file, `$1`)
	p.endVideo <- file
}

func (p *PlayerController) Play(req FileRequest, res *EmptyResponse) error {
	player.Eval(fmt.Sprintf("loadVideo('/videos/%s')", req.File))
	player.Eval("startVideo()")

	return nil
}

func (p *PlayerController) Cancel(req FileRequest, res *EmptyResponse) error {
	player.Eval("clearVideo()")

	return nil
}

func (p *PlayerController) Pause(req FileRequest, res *EmptyResponse) error {
	player.Eval("pauseVideo()")

	return nil
}

func (p *PlayerController) Resume(req FileRequest, res *EmptyResponse) error {
	player.Eval("startVideo()")

	return nil
}

func (p *PlayerController) WaitForEnd(req EmptyRequest, res *FileResponse) error {
	file := <-p.endVideo
	res.File = file

	return nil
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	playerBox := packr.New("Player", "../player-frontend")
	http.Handle("/player/", http.StripPrefix("/player/", http.FileServer(playerBox)))
	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir("./videos"))))

	go func() {
		fmt.Println("listening on", ln.Addr().String())
		log.Fatal(http.Serve(ln, nil))
	}()
	serveRPC()

	player = getPlayer(ln.Addr())
	defer player.Close()

	// Quit after all windows are closed
	<-player.Done()
}

func getPlayer(serverAddr net.Addr) lorca.UI {
	var err error
	ui, err := lorca.New("", "", 480, 320)
	if err != nil {
		log.Fatal(err)
	}

	ui.Load(fmt.Sprintf("http://%s/player/", serverAddr))

	log.Println("DOM bind")
	ui.Bind("signalEndPlay", playerController.SignalEndPlay)

	return ui
}

func serveRPC() {
	rpc.Register(&playerController)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer listener.Close()
		rpc.Accept(listener)
	}()
}
