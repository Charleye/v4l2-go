package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video0", "camera device")
var image = flag.Bool("i", false, "store frame into image file")
var fourcc = flag.String("f", "", "set pixel format")
var videoname = flag.String("v", "", "video name")

var file *os.File

func main() {
	flag.Parse()

	d, err := v4l2.Open(*device)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	var cam v4l2.Camera
	cam.Device = *d

	cam.VerifyCaps()

	cam.Width = 800
	cam.Height = 600
	cam.PixelFormat = v4l2.GetFourCCByName(*fourcc)
	cam.SetFormat()
	cam.AllocBuffers(4)
	cam.TurnOn()
	defer cam.TurnOff()

	if !*image {
		file, _ = os.OpenFile(*videoname, os.O_RDWR|os.O_CREATE, 0644)
		defer file.Close()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	var counter int
	for {
		select {
		case <-c:
			return
		default:
			if *image {
				imagename := "in" + strconv.Itoa(counter) + "_" + *fourcc + "_800_600.raw"
				file, _ = os.OpenFile(imagename, os.O_RDWR|os.O_CREATE, 0644)
			}
			data := cam.Capture()
			n, _ := file.Write(data)
			fmt.Println("Write: ", n)
			if *image {
				counter++
				file.Close()
			}
		}
	}
}
