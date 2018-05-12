package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video0", "video device")

func main() {
	flag.Parse()
	fd, err := syscall.Open(*device, syscall.O_RDWR, 0)
	if err != nil {
		log.Fatal(err)
	}

	var vc v4l2.V4L2_Capability
	err = v4l2.IoctlQueryCap(fd, &vc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Driver: %s\n", vc.Driver)
	fmt.Printf("Card: %s\n", vc.Card)
	fmt.Printf("BusInfo: %s\n", vc.BusInfo)
	fmt.Printf("Version: %#x\n", vc.Version)
	fmt.Printf("Capabilities: %#x\n", vc.Capabilities)
	fmt.Println("")

	for i := 0; ; i++ {
		vf := v4l2.V4L2_Fmtdesc{
			Index: uint32(i),
			Type:  v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE,
		}
		err := v4l2.IoctlEnumFmt(fd, &vf)
		if err != nil {
			if err == syscall.EINVAL {
				break
			}
			log.Fatal(err)
		}
		fmt.Println("Description: ", vf.Description)
		fmt.Printf("PixelFormat: %#x\n", vf.PixelFormat)
		fmt.Println("Flags: ", vf.Flags)
	}
	fmt.Println("")

	var vfmt v4l2.V4L2_Format
	var pf v4l2.V4L2_Pix_Format
	vfmt.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	vfmt.Fmt = &pf
	err = v4l2.IoctlGetFmt(fd, &vfmt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Width: ", pf.Width)
	fmt.Println("Height: ", pf.Height)
	fmt.Printf("PixelFormat: %#x\n", pf.PixelFormat)
	fmt.Println("SizeImage: ", pf.SizeImage)
	fmt.Println("Priv: ", pf.Priv)
	fmt.Println("BytesPerLine: ", pf.BytesPerLine)
	fmt.Println("")

	vpf := v4l2.V4L2_Pix_Format{
		Width:       1024,
		Height:      768,
		PixelFormat: v4l2.GetFourCCByName("YUYV"),
	}
	vfmt.Fmt = &vpf
	err = v4l2.IoctlSetFmt(fd, &vfmt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Width: ", vpf.Width)
	fmt.Println("Height: ", vpf.Height)
	fmt.Printf("PixelFormat: %#x\n", vpf.PixelFormat)
	fmt.Println("SizeImage: ", vpf.SizeImage)
	fmt.Println("Priv: ", vpf.Priv)
	fmt.Println("BytesPerLine: ", vpf.BytesPerLine)
}
