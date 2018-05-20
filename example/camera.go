package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video0", "camera device")

func main() {
	flag.Parse()

	fd, err := syscall.Open(*device, syscall.O_RDWR, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(fd)

	var stat syscall.Stat_t
	if err := syscall.Fstat(fd, &stat); err != nil {
		log.Fatal(err)
	}
	if stat.Mode&syscall.S_IFCHR == 0 || stat.Rdev>>8 != 81 {
		log.Fatal("Wrong V4L2 Device")
	}

	var caps v4l2.V4L2_Capability
	err = v4l2.IoctlQueryCap(fd, &caps)
	if err != nil {
		log.Fatal(err)
	}

	if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_CAPTURE == 0 {
		log.Fatal("The device not support video capture")
	}

	var format v4l2.V4L2_Format
	var pixfmt v4l2.V4L2_Pix_Format
	pixfmt.Width = 800
	pixfmt.Height = 600
	pixfmt.PixelFormat = v4l2.GetFourCCByName("YUYV")
	pixfmt.Priv = 0
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	format.Fmt = &pixfmt
	err = v4l2.IoctlSetFmt(fd, &format)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(v4l2.GetNameByFourCC(pixfmt.PixelFormat))

	var reqbufs v4l2.V4L2_Requestbuffers
	reqbufs.Count = 4
	reqbufs.Memory = v4l2.V4L2_MEMORY_MMAP
	reqbufs.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlRequestBuffers(fd, &reqbufs)
	if err != nil {
		log.Fatal(err)
	}
	if reqbufs.Count == 0 {
		log.Fatal("Out of Memory")
	}

	data := make([][]byte, 0, reqbufs.Count)
	for i := 0; i < int(reqbufs.Count); i++ {
		vb := v4l2.V4L2_Buffer{
			Index:  uint32(i),
			Type:   v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE,
			Memory: v4l2.V4L2_MEMORY_MMAP,
		}
		if err := v4l2.IoctlQueryBuf(fd, &vb); err != nil {
			log.Fatal(err)
		}
		var offset uint32
		v4l2.GetValueFromUnion(vb.M, &offset)
		fmt.Println("offset: ", offset)

		buf, err := syscall.Mmap(fd, int64(offset), int(vb.Length),
			syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
		if err != nil {
			log.Fatal(err)
		}
		data = append(data, buf)
		if err := v4l2.IoctlQBuf(fd, &vb); err != nil {
			log.Fatal(err)
		}
	}

	var stream int = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlStreamOn(fd, &stream)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile("video.yuv", os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()

	for i := 0; i < 1; i++ {
		vb := v4l2.V4L2_Buffer{
			Type:   v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE,
			Memory: v4l2.V4L2_MEMORY_MMAP,
		}
		if err := v4l2.IoctlDQBuf(fd, &vb); err != nil {
			log.Fatal(err)
		}

		data[vb.Index] = data[vb.Index][:vb.Length]
		//	frame := make([]byte, vb.Length)
		//	copy(frame, data[vb.Index])
		//	fmt.Println(len(frame), cap(frame))

		//	n, err := file.Write(frame)
		n, err := file.Write(data[vb.Index][:vb.BytesUsed])
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("count: ", n)

		if err := v4l2.IoctlQBuf(fd, &vb); err != nil {
			log.Fatal(err)
		}
	}

	err = v4l2.IoctlStreamOff(fd, &stream)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range data {
		err := syscall.Munmap(v)
		if err != nil {
			log.Fatal(err)
		}
	}
}
