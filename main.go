package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"

	"github.com/gobuffalo/packr/v2"
	"github.com/zserge/lorca"
)

var panel lorca.UI
var player lorca.UI
var controller = PanelController{Panels: []Panel{}, playCancelers: map[string]context.CancelFunc{}, playPausers: map[string]*sync.Mutex{}}
var playerController = PlayerController{}

var playingVideo = false

// Panel is one panel to be shown
type Panel struct {
	Name     string `json:"name"`
	Shortcut string `json:"shortcut"`
	File     string `json:"file"`
}

// PlayerController helps with controling the player
type PlayerController struct{}

// SignalEndPlay gets sent when a video ended
func (p *PlayerController) SignalEndPlay(file string) {
	re := regexp.MustCompile(`^/videos/`)
	file = re.ReplaceAllString(file, `$1`)
	panel.Eval("window.eventEmitter.emit('endPlay','" + file + "')")
	playingVideo = false
}

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

	player.Eval(fmt.Sprintf("loadVideo('/videos/%s')", file))
	player.Eval("startVideo()")
	playingVideo = true
}

// Cancel stops playing a specific file
func (p *PanelController) Cancel(file string) {
	player.Eval("clearVideo()")
	playingVideo = false
}

// Pause pauses playing a specific file
func (p *PanelController) Pause(file string) {
	player.Eval("pauseVideo()")
}

// Resume pauses playing a specific file
func (p *PanelController) Resume(file string) {
	player.Eval("startVideo()")
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
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	playerBox := packr.New("Player", "./player-frontend")
	panelBox := packr.New("Panel", "./panel-frontend/build")

	http.Handle("/api/panels", http.HandlerFunc(handleAPIPanels))
	// load in bindata
	http.Handle("/panel/", http.StripPrefix("/panel/", http.FileServer(panelBox)))
	http.Handle("/player/", http.StripPrefix("/player/", http.FileServer(playerBox)))
	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir("./videos"))))

	go func() {
		fmt.Println("listening on", ln.Addr().String())
		log.Fatal(http.Serve(ln, nil))
	}()

	panel = getPanel(ln.Addr())
	defer panel.Close()

	player = getPlayer(ln.Addr())
	defer player.Close()

	// Quit after all windows are closed
	<-panel.Done()
	<-player.Done()
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

	return ui
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
