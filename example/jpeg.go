package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video0", "v4l2 device")

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
	if caps.DeviceCaps&v4l2.V4L2_CAP_VIDEO_M2M == 0 {
		log.Fatal("The device not support mutli-planar m2m")
	}

	// set source format
	var format v4l2.V4L2_Format
	var pf_out v4l2.V4L2_Pix_Format
	pf_out.Width = 800
	pf_out.Height = 600
	pf_out.PixelFormat = v4l2.GetFourCCByName("YUYV")
	pf_out.BytesPerLine = 0
	pf_out.SizeImage = 960000
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	format.Fmt = &pf_out
	err = v4l2.IoctlSetFmt(fd, &format)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pf_out)

	// set destination format
	var pf_in v4l2.V4L2_Pix_Format
	pf_in.Width = 800
	pf_in.Height = 600
	pf_in.PixelFormat = v4l2.GetFourCCByName("JPEG")
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	format.Fmt = &pf_in
	err = v4l2.IoctlSetFmt(fd, &format)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pf_in)

	var dst_reqbufs v4l2.V4L2_Requestbuffers
	dst_reqbufs.Count = 1
	dst_reqbufs.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	dst_reqbufs.Memory = v4l2.V4L2_MEMORY_MMAP
	err = v4l2.IoctlRequestBuffers(fd, &dst_reqbufs)
	if err != nil {
		log.Fatal(err)
	}
	if dst_reqbufs.Count == 0 {
		log.Fatal("Out of memory")
	}
	fmt.Println(dst_reqbufs)

	dst_data := make([][]byte, 0, dst_reqbufs.Count)
	for i := 0; i < int(dst_reqbufs.Count); i++ {
		vb := v4l2.V4L2_Buffer{
			Index:  uint32(i),
			Type:   dst_reqbufs.Type,
			Memory: dst_reqbufs.Memory,
		}
		if err := v4l2.IoctlQueryBuf(fd, &vb); err != nil {
			log.Fatal(err)
		}
		var offset uint32
		v4l2.GetValueFromUnion(vb.M, &offset)
		fmt.Println("dst offset: ", offset)

		buf, err := syscall.Mmap(fd, int64(offset), int(vb.Length),
			syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
		if err != nil {
			log.Fatal(err)
		}
		dst_data = append(dst_data, buf)
	}

	var src_reqbufs v4l2.V4L2_Requestbuffers
	src_reqbufs.Count = 1
	src_reqbufs.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	src_reqbufs.Memory = v4l2.V4L2_MEMORY_MMAP
	err = v4l2.IoctlRequestBuffers(fd, &src_reqbufs)
	if err != nil {
		log.Fatal(err)
	}
	if src_reqbufs.Count == 0 {
		log.Fatal("Out of memory")
	}
	fmt.Println(src_reqbufs)

	src_data := make([][]byte, 0, src_reqbufs.Count)
	for i := 0; i < int(src_reqbufs.Count); i++ {
		vb := v4l2.V4L2_Buffer{
			Index:  uint32(i),
			Type:   src_reqbufs.Type,
			Memory: src_reqbufs.Memory,
		}
		if err := v4l2.IoctlQueryBuf(fd, &vb); err != nil {
			log.Fatal(err)
		}
		var offset uint32
		v4l2.GetValueFromUnion(vb.M, &offset)
		fmt.Println("src offset: ", offset)

		buf, err := syscall.Mmap(fd, int64(offset), int(vb.Length),
			syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
		if err != nil {
			log.Fatal(err)
		}
		src_data = append(src_data, buf)
	}

	vb := v4l2.V4L2_Buffer{
		Type:   v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT,
		Memory: v4l2.V4L2_MEMORY_MMAP,
		Index:  0,
	}
	err = v4l2.IoctlQueryBuf(fd, &vb)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(vb)

	err = v4l2.IoctlQBuf(fd, &vb)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(vb)

	vb.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlQueryBuf(fd, &vb)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(vb)
	err = v4l2.IoctlQBuf(fd, &vb)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(vb)

	var stream int = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlStreamOn(fd, &stream)
	if err != nil {
		log.Fatal(err)
	}

	stream = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	err = v4l2.IoctlStreamOn(fd, &stream)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("111111111111111111")

	/*
		err = v4l2.IoctlDQBuf(fd, &vb)
		if err != nil {
			log.Fatal(err)
		}
	*/
	stream = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	err = v4l2.IoctlStreamOff(fd, &stream)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("222222222222222")
	for _, v := range src_data {
		err := syscall.Munmap(v)
		if err != nil {
			log.Fatal(err)
		}
	}

	stream = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlStreamOff(fd, &stream)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range dst_data {
		err := syscall.Munmap(v)
		if err != nil {
			log.Fatal(err)
		}
	}
}
