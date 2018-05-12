package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video11", "video device")

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
			Type:  v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE,
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

	// first VIDIOC_G_FMT
	var vfmt v4l2.V4L2_Format
	var pfm v4l2.V4L2_Pix_Format_Mplane
	vfmt.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	vfmt.Fmt = &pfm
	err = v4l2.IoctlGetFmt(fd, &vfmt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("get mplane format Before VIDIOC_S_FMT")
	fmt.Println(pfm)
	fmt.Println("Width: ", pfm.Width)
	fmt.Println("Height: ", pfm.Height)
	fmt.Printf("PixelFormat: %#x\n", pfm.PixelFormat)
	fmt.Println("NumPlanes: ", pfm.NumPlanes)
	fmt.Println("Field: ", pfm.Field)
	fmt.Println("")

	// first VIDIOC_S_FMT
	vpfm := v4l2.V4L2_Pix_Format_Mplane{
		Width:       1024,
		Height:      768,
		PixelFormat: v4l2.GetFourCCByName("NM12"),
	}
	vpfm.NumPlanes = 2
	vpfm.PlaneFmt[0].SizeImage = 0xC0000
	vpfm.PlaneFmt[0].BytesPerLine = 1024
	vpfm.PlaneFmt[1].SizeImage = 0x60000
	vpfm.PlaneFmt[1].BytesPerLine = 1024
	vfmt.Fmt = &vpfm
	err = v4l2.IoctlSetFmt(fd, &vfmt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Width: ", vpfm.Width)
	fmt.Println("Height: ", vpfm.Height)
	fmt.Printf("PixelFormat: %#x\n", vpfm.PixelFormat)
	fmt.Println("NumPlanes: ", vpfm.NumPlanes)
	fmt.Println("Field: ", vpfm.Field)
	fmt.Println("")

	// second VIDIOC_G_FMT
	pfm = v4l2.V4L2_Pix_Format_Mplane{}
	vfmt.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	vfmt.Fmt = &pfm
	err = v4l2.IoctlGetFmt(fd, &vfmt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("get mplane format After VIDIOC_S_FMT")
	fmt.Println(pfm)
	fmt.Println("Width: ", pfm.Width)
	fmt.Println("Height: ", pfm.Height)
	fmt.Printf("PixelFormat: %#x\n", pfm.PixelFormat)
	fmt.Println("NumPlanes: ", pfm.NumPlanes)
	fmt.Println("Field: ", pfm.Field)
}
