package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video0", "camera device")

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
	cam.PixFmtDescription = "MJPG"
	cam.SetFormat()
	cam.AllocBuffers(4)
	cam.TurnOn()
	defer cam.TurnOff()

	file, err := os.OpenFile("test.mjpg", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			return
		default:
			data := cam.Capture()
			n, _ := file.Write(data)
			fmt.Println("Write: ", n)
		}
	}
}
