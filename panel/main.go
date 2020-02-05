package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"

	"github.com/gobuffalo/packr/v2"
	"github.com/zserge/lorca"
)

var panel lorca.UI
var controller = PanelController{Panels: []Panel{}, playCancelers: map[string]context.CancelFunc{}, playPausers: map[string]*sync.Mutex{}}
var playingVideo = false

var rpcClient *rpc.Client

// RPC structs
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

// Panel is one panel to be shown
type Panel struct {
	Name     string `json:"name"`
	Shortcut string `json:"shortcut"`
	File     string `json:"file"`
}

// PlayerController helps with controling the player
type PlayerController struct{}

// PanelController helps with controling the pannels
type PanelController struct {
	Panels        []Panel `json:"panels"`
	panelsForFile map[string]*Panel
	playCancelers map[string]context.CancelFunc
	playPausers   map[string]*sync.Mutex
}

// GetFromDisk loads the panels from the config file
func (p *PanelController) GetFromDisk() []Panel {
	f, err := os.Open("./panels.json")
	jsonParser := json.NewDecoder(f)
	jsonParser.Decode(&p.Panels)
	if err != nil {
		fmt.Println(err)
		return p.Panels
	}

	p.panelsForFile = map[string]*Panel{}
	for i := range p.Panels {
		p.panelsForFile[p.Panels[i].File] = &p.Panels[i]
	}
	return p.Panels
}

// Play starts playing a specific file
func (p *PanelController) Play(file string) {
	fmt.Println(file)
	rpcClient.Call("PlayerController.Play", FileRequest{File: file}, &EmptyResponse{})
	playingVideo = true
}

// Cancel stops playing a specific file
func (p *PanelController) Cancel(file string) {
	rpcClient.Call("PlayerController.Cancel", FileRequest{File: file}, &EmptyResponse{})
	playingVideo = false
}

// Pause pauses playing a specific file
func (p *PanelController) Pause(file string) {
	rpcClient.Call("PlayerController.Oause", FileRequest{File: file}, &EmptyResponse{})
}

// Resume pauses playing a specific file
func (p *PanelController) Resume(file string) {
	rpcClient.Call("PlayerController.Resume", FileRequest{File: file}, &EmptyResponse{})
}

// CanPlay tells if a new file can be started
func (p *PanelController) CanPlay() bool {
	return !playingVideo
}

func handleAPIPanels(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	out, _ := json.Marshal(controller.GetFromDisk())
	w.Write(out)
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("Need to specify IP of the player")
	}

	var err error
	rpcClient, err = rpc.Dial("tcp", os.Args[1]+":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	panelBox := packr.New("Panel", "../panel-frontend/build")

	http.Handle("/api/panels", http.HandlerFunc(handleAPIPanels))
	// load in bindata
	http.Handle("/panel/", http.StripPrefix("/panel/", http.FileServer(panelBox)))

	go func() {
		fmt.Println("listening on", ln.Addr().String())
		log.Fatal(http.Serve(ln, nil))
	}()

	panel = getPanel(ln.Addr())
	defer panel.Close()

	<-panel.Done()
}

func getPanel(serverAddr net.Addr) lorca.UI {
	var err error
	ui, err := lorca.New("", "", 480, 320)
	if err != nil {
		log.Fatal(err)
	}

	ui.Load(fmt.Sprintf("http://%s/panel/", serverAddr))

	log.Println("DOM bind")

	ui.Bind("play", controller.Play)
	ui.Bind("pause", controller.Pause)
	ui.Bind("cancel", controller.Cancel)
	ui.Bind("resume", controller.Resume)
	ui.Bind("canPlay", controller.CanPlay)
	ui.Eval(`
			window.panelController = {
				play,
				pause,
				cancel,
				resume,
				canPlay,
			}
		`)
	log.Println(err)

	go listenForEnd()

	return ui
}

func listenForEnd() {
	for {
		var res FileResponse
		err := rpcClient.Call("PlayerController.WaitForEnd", EmptyRequest{}, &res)
		if err != nil {
			log.Println(err)
			continue
		}
		panel.Eval("window.eventEmitter.emit('endPlay','" + res.File + "')")
	}
}
