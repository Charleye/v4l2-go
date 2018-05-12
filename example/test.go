package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"
)

var device = flag.String("d", "/dev/video0", "video device")

func main() {
	flag.Parse()
	fd, err := syscall.Open(*device, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(fd)

	var vc V4L2_Capability
	err = IoctlQueryCap(fd, &vc)
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
		vf := V4L2_Fmtdesc{
			Index: uint32(i),
			Type:  V4L2_BUF_TYPE_VIDEO_CAPTURE,
		}
		err := IoctlEnumFmt(fd, &vf)
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

	var vfmt V4L2_Format
	var pf V4L2_Pix_Format
	vfmt.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	vfmt.fmt = &pf
	err = IoctlGetFmt(fd, &vfmt)
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

	vpf := V4L2_Pix_Format{
		Width:       1024,
		Height:      768,
		PixelFormat: GetFourCCByName("YUYV"),
	}
	vfmt.fmt = &vpf
	err = IoctlSetFmt(fd, &vfmt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Width: ", vpf.Width)
	fmt.Println("Height: ", vpf.Height)
	fmt.Printf("PixelFormat: %#x\n", vpf.PixelFormat)
	fmt.Println("SizeImage: ", vpf.SizeImage)
	fmt.Println("Priv: ", vpf.Priv)
	fmt.Println("BytesPerLine: ", vpf.BytesPerLine)
	fmt.Println("")

	var ctrl V4L2_Control
	ctrl.ID = V4L2_CID_BRIGHTNESS
	ctrl.Value = 20
	err = IoctlSetCtrl(fd, &ctrl)
	if err != nil {
		log.Fatal(err)
	}

	ctrl.Value = 0
	err = IoctlGetCtrl(fd, &ctrl)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID: %#x\n", ctrl.ID)
	fmt.Println("Value: ", ctrl.Value)
	fmt.Println("")

	var cc V4L2_Cropcap
	cc.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = IoctlCropCap(fd, &cc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cc)

	/*
		var crop V4L2_Crop
		crop.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
		crop.C = cc.Defrect
		err = IoctlSetCrop(fd, &crop)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(crop)
	*/

	var rb V4L2_Requestbuffers
	rb.Count = 1
	rb.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	rb.Memory = V4L2_MEMORY_MMAP
	err = IoctlRequestBuffers(fd, &rb)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("V4L2_Requestbuffers: ", rb)

	var vb V4L2_Buffer
	vb.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	vb.Index = 0
	err = IoctlQueryBuf(fd, &vb)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("V4L2_Buffer:")
	fmt.Println(vb)

	var sp V4L2_Streamparm
	sp.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = IoctlGetParm(fd, &sp)
	if err != nil {
		log.Fatal(err)
	}
	cp := sp.Parm.(*V4L2_Captureparm)
	fmt.Println("V4L2_Captureparm: ", cp)

	var se V4L2_Event_Subscription
	se.Type = V4L2_EVENT_CTRL
	se.ID = V4L2_CID_BRIGHTNESS
	err = IoctlSubscribeEvent(fd, &se)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(se)

	var ve V4L2_Event
	err = IoctlDQEvent(fd, &ve)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ve)
}
