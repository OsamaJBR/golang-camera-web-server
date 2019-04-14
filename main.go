package main

import (
	"fmt"
	"html/template"
	"image"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

var (
	err      error
	webcam   *gocv.VideoCapture
	frame_id int
)

var buffer = make(map[int][]byte)
var frame []byte
var mutex = &sync.Mutex{}

func main() {

	host := "0.0.0.0:3000"

	// open webcam
	if len(os.Args) < 2 {
		fmt.Println(">> device /dev/video0 (default)")
		webcam, err = gocv.VideoCaptureDevice(0)
	} else {
		fmt.Println(">> file/url :: " + os.Args[1])
		webcam, err = gocv.VideoCaptureFile(os.Args[1])
	}

	if err != nil {
		fmt.Printf("Error opening capture device: \n")
		return
	}
	defer webcam.Close()

	// start capturing
	go getframes()

	fmt.Println("Capturing. Open http://" + host)

	// start http server
	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		data := ""
		for {
			/*			fmt.Println("Frame ID: ", frame_id)
			 */mutex.Lock()
			data = "--frame\r\n  Content-Type: image/jpeg\r\n\r\n" + string(frame) + "\r\n\r\n"
			mutex.Unlock()
			time.Sleep(33 * time.Millisecond)
			w.Write([]byte(data))
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, "index")
	})

	log.Fatal(http.ListenAndServe(host, nil))
}

func getframes() {
	img := gocv.NewMat()
	defer img.Close()
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed\n")
			return
		}
		if img.Empty() {
			continue
		}
		frame_id++
		gocv.Resize(img, &img, image.Point{}, float64(0.5), float64(0.5), 0)
		frame, _ = gocv.IMEncode(".jpg", img)

	}
}
