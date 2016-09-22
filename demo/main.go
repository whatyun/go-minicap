package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"log"
	"net/http"

	"github.com/BigWavelet/go-minicap"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nfnt/resize"
)

var imC <-chan image.Image
var upgrader = websocket.Upgrader{}

type WebConfig struct {
	User    string
	Version string
}

func test() {
	m, err := minicap.NewService(minicap.Options{Serial: "EP7333W7XB"})
	if err != nil {
		log.Fatal(err)
	}
	err = m.Install()
	if err != nil {
		log.Fatal(err)
	}
	imC, err = m.Capture()
	if err != nil {
		log.Fatal(err)
	}
}

func hIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("t").ParseFiles("index.html"))
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func hImageWs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade err:", err)
		return
	}
	done := make(chan bool, 1)
	go func() {
		buf := new(bytes.Buffer)
		buf.Reset()
		log.Println("Prepare websocket send", imC)
		for im := range imC {
			log.Println("encode image")
			select {
			case <-done:
				log.Println("finished")
				return
			default:
			}
			size := im.Bounds().Size()
			newIm := resize.Resize(uint(size.X)/2, 0, im, resize.Lanczos3)
			wr, err := c.NextWriter(websocket.BinaryMessage)
			if err != nil {
				log.Println(err)
				break
			}

			if err := jpeg.Encode(wr, newIm, nil); err != nil {
				break
			}
			wr.Close()
		}
	}()
	for {
		mt, p, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			done <- true
			break
		}
		log.Println(mt, p, err)
	}
}

func startWebServer(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/", hIndex)
	//r.HandleFunc("/ws/screen", wsPerf)
	r.HandleFunc("/ws", hImageWs)
	http.Handle("/", r)
	log.Println("start webserver here...")
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func main() {
	test()
	startWebServer(5678)

}