package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video0", "camera device node")
var fimc_node = flag.String("v", "", "FIMC device node")
var width = flag.Uint("w", 0, "width  in pixel")
var height = flag.Uint("h", 0, "height in pixel")
var fourcc = flag.String("r", "YUYV", "pixel format for input interface")

/* global varibales */
var num_src_planes, num_dst_planes int
var data_src_buf, data_dst_buf [][][]byte
var src_frame_size, dst_frame_size uint32
var num_src_bufs, num_dst_bufs uint32

func main() {
	flag.Parse()

	dev, err := v4l2.Open(*device)
	if err != nil {
		log.Fatalf("Failed to open device: %s, err: %v", *device, err)
	}
	defer dev.Close()

	var cam v4l2.Camera
	cam.Device = *dev

	cam.VerifyCaps()
	cam.Width = uint32(*width)
	cam.Height = uint32(*height)
	cam.PixFmtDescription = *fourcc
	cam.SetFormat()
	cam.AllocBuffers(4)
	cam.TurnOn()
	defer cam.TurnOff()

	video_fd := InitCSCVideoNode()
	defer syscall.Close(video_fd)

	SetInputFormat(video_fd)
	AllocInputBuffers(video_fd)
	defer FreeInputBuffers()

	SetOutputFormat(video_fd)
	AllocOutputBuffers(video_fd)
	defer FreeOutputBuffers()

	process(&cam, video_fd)
	streamoff(video_fd)
}

func process(cam *v4l2.Camera, video_fd int) {
	var src_planes [v4l2.VIDEO_MAX_PLANES]v4l2.V4L2_Plane
	var dst_planes [v4l2.VIDEO_MAX_PLANES]v4l2.V4L2_Plane
	var src_buf, dst_buf v4l2.V4L2_Buffer

	src_buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	src_buf.Memory = v4l2.V4L2_MEMORY_MMAP
	src_buf.Index = 0
	src_buf.M = v4l2.PointerToBytes(&src_planes[0])
	src_buf.Length = uint32(num_src_planes)

	dst_buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	dst_buf.Memory = v4l2.V4L2_MEMORY_MMAP
	dst_buf.Index = 0
	dst_buf.M = v4l2.PointerToBytes(&dst_planes[0])
	dst_buf.Length = uint32(num_dst_planes)

	copy(data_src_buf[0][0], cam.Capture())

	var num_frames int
	for ; num_frames < 1; num_frames++ {
		err := v4l2.IoctlQBuf(video_fd, &src_buf)
		if err != nil {
			log.Fatalf("Failed to enqueue input buffer: %v", err)
		}
		err = v4l2.IoctlQBuf(video_fd, &dst_buf)
		if err != nil {
			log.Fatalf("Failed to enqueue output buffer: %v", err)
		}

		if num_frames == 0 {
			Type := v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
			err := v4l2.IoctlStreamOn(video_fd, &Type)
			if err != nil {
				log.Fatalf("Failed to stream on capture interface: %v", err)
			}
			Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
			err = v4l2.IoctlStreamOn(video_fd, &Type)
			if err != nil {
				log.Fatalf("Failed to stream on output interface: %v", err)
			}
		}

		/* dequeue buffer */
		var read_fds syscall.FdSet
		var write_fds syscall.FdSet
		read_fds.Bits[video_fd/64] = 1 << uint32(video_fd%64)
		write_fds.Bits[video_fd/64] = 1 << uint32(video_fd%64)
		r, err := syscall.Select(video_fd+1, &read_fds, &write_fds, nil, nil)
		if r < 0 {
			log.Fatalf("select errors: %v\n", err)
		}

		err = v4l2.IoctlDQBuf(video_fd, &dst_buf)
		if err != nil {
			log.Fatalf("Failed to dequeue capture interface buffer: %v", err)
		}
		err = v4l2.IoctlDQBuf(video_fd, &src_buf)
		if err != nil {
			log.Fatalf("Failed to dequeue output interface buffer: %v", err)
		}
	}
}

func InitCSCVideoNode() int {
	vid_fd, err := syscall.Open(*fimc_node, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		log.Fatalf("Failed to open CSC node: %s", *fimc_node)
	}

	var caps v4l2.V4L2_Capability
	err = v4l2.IoctlQueryCap(vid_fd, &caps)
	if err != nil {
		log.Fatal("Failed to query capabilities")
	}
	if caps.Capabilities&v4l2.V4L2_CAP_DEVICE_CAPS != 0 {
		if caps.DeviceCaps&v4l2.V4L2_CAP_VIDEO_M2M_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes mem-to-mem (%#x)\n",
				*fimc_node, caps.DeviceCaps)
		}
	} else {
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_CAPTURE_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes capture (%#x)\n",
				*fimc_node, caps.Capabilities)
		}
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_OUTPUT_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes output (%#x)\n",
				*fimc_node, caps.Capabilities)
		}
	}
	return vid_fd
}

func SetInputFormat(video_fd int) {
	/* set input format */
	var format v4l2.V4L2_Format
	var pixmp v4l2.V4L2_Pix_Format_Mplane
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	pixmp.Width = uint32(*width)
	pixmp.Height = uint32(*height)
	pixmp.Field = v4l2.V4L2_FIELD_ANY
	pixmp.PixelFormat = v4l2.GetFourCCByName(*fourcc)
	format.Fmt = &pixmp
	err := v4l2.IoctlSetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to set input format")
	}
	num_src_planes = int(pixmp.NumPlanes)
	for i := 0; i < num_src_planes; i++ {
		src_frame_size += pixmp.PlaneFmt[i].SizeImage
		fmt.Printf("plane[%d]: bytesperline: %d, sizeimage: %d\n",
			i, pixmp.PlaneFmt[i].BytesPerLine, pixmp.PlaneFmt[i].SizeImage)
	}
	fmt.Printf("SRC frame_size: %v\n", src_frame_size)

	err = v4l2.IoctlGetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to get input format")
	}
	*width = uint(pixmp.Width)
	*height = uint(pixmp.Height)
	fmt.Printf("width: %v, height: %v\n", *width, *height)
}

func SetOutputFormat(video_fd int) {
	/* set output format */
	var format v4l2.V4L2_Format
	var pixmp v4l2.V4L2_Pix_Format_Mplane
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	pixmp = v4l2.V4L2_Pix_Format_Mplane{}
	pixmp.Width = uint32(*width)
	pixmp.Height = uint32(*height)
	pixmp.Field = v4l2.V4L2_FIELD_ANY
	pixmp.PlaneFmt[0].BytesPerLine = uint32(*width)
	pixmp.PixelFormat = v4l2.GetFourCCByName("UYVY")
	format.Fmt = &pixmp
	err := v4l2.IoctlSetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to set output format")
	}
	fmt.Println("output format: ", pixmp)
	num_dst_planes = int(pixmp.NumPlanes)
	for i := 0; i < num_dst_planes; i++ {
		dst_frame_size += pixmp.PlaneFmt[i].SizeImage
		fmt.Printf("plane[%d]: bytesperline: %v, sizeimage: %v\n",
			i, pixmp.PlaneFmt[i].BytesPerLine, pixmp.PlaneFmt[i].SizeImage)
	}
	fmt.Printf("out_width: %v, out_height: %v\n", pixmp.Width, pixmp.Height)
	fmt.Printf("DST framesize: %v\n", dst_frame_size)
}

func AllocInputBuffers(video_fd int) {
	/* request input buffer */
	var reqbuf v4l2.V4L2_Requestbuffers
	reqbuf.Count = 1
	reqbuf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	reqbuf.Memory = v4l2.V4L2_MEMORY_MMAP
	err := v4l2.IoctlRequestBuffers(video_fd, &reqbuf)
	if err != nil {
		log.Fatal("Failed to request input buffers")
	}
	if reqbuf.Count < 1 {
		log.Fatal("Out of memory")
	}
	num_src_bufs = reqbuf.Count

	data_src_buf = make([][][]byte, 0, num_src_bufs)
	for index := 0; index < int(num_src_bufs); index++ {
		/* get buffer parameters */
		var buf v4l2.V4L2_Buffer
		var planes [v4l2.VIDEO_MAX_PLANES]v4l2.V4L2_Plane

		buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
		buf.M = v4l2.PointerToBytes(&planes[0])
		buf.Length = uint32(num_src_planes)
		buf.Memory = v4l2.V4L2_MEMORY_MMAP
		buf.Index = uint32(index)

		err := v4l2.IoctlQueryBuf(video_fd, &buf)
		if err != nil {
			log.Fatal(err)
		}

		data_src_planes := make([][]byte, 0, num_src_planes)
		for i := 0; i < num_src_planes; i++ {
			var offset uint32
			v4l2.GetValueFromUnion(planes[i].Union, &offset)
			fmt.Printf("QUERYBUF: plane [%d]: Length: %v, bytesused: %v, offset: %v\n",
				i, planes[i].Length, planes[i].BytesUsed, offset)

			/* mmap input buffer */
			buf, err := syscall.Mmap(video_fd, int64(offset), int(planes[i].Length),
				syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
			if err != nil {
				log.Fatalf("mmap error: %v\n", err)
			}
			data_src_planes = append(data_src_planes, buf)
		}
		data_src_buf = append(data_src_buf, data_src_planes)
	}
}

func AllocOutputBuffers(video_fd int) {
	/* request output buffers */
	var reqbuf v4l2.V4L2_Requestbuffers
	reqbuf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	reqbuf.Memory = v4l2.V4L2_MEMORY_MMAP
	reqbuf.Count = 1
	err := v4l2.IoctlRequestBuffers(video_fd, &reqbuf)
	if err != nil {
		log.Fatal("Failed to request output buffer")
	}
	if reqbuf.Count < 1 {
		log.Fatal("request output buffer: Out of memory")
	}
	num_dst_bufs = reqbuf.Count

	data_dst_buf = make([][][]byte, 0, num_dst_bufs)
	for index := 0; index < int(num_dst_bufs); index++ {
		/* get buffer parameters */
		var buf v4l2.V4L2_Buffer
		var planes [v4l2.VIDEO_MAX_PLANES]v4l2.V4L2_Plane

		buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
		buf.M = v4l2.PointerToBytes(&planes[0])
		buf.Length = uint32(num_dst_planes)
		buf.Memory = v4l2.V4L2_MEMORY_MMAP
		buf.Index = uint32(index)

		err := v4l2.IoctlQueryBuf(video_fd, &buf)
		if err != nil {
			log.Fatal(err)
		}

		data_dst_planes := make([][]byte, 0, num_dst_planes)
		for i := 0; i < num_dst_planes; i++ {
			var offset uint32
			v4l2.GetValueFromUnion(planes[i].Union, &offset)
			fmt.Printf("QUERYBUF: plane [%d]: Length: %v, bytesused: %v, offset: %v\n",
				i, planes[i].Length, planes[i].BytesUsed, offset)

			/* mmap output buffer */
			buf, err := syscall.Mmap(video_fd, int64(offset), int(planes[i].Length),
				syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
			if err != nil {
				log.Fatalf("mmap error: %v\n", err)
			}
			data_dst_planes = append(data_dst_planes, buf)
		}
		data_dst_buf = append(data_dst_buf, data_dst_planes)
	}
}

func FreeInputBuffers() {
	for _, v := range data_src_buf {
		for _, d := range v {
			if err := syscall.Munmap(d); err != nil {
				fmt.Printf("SRC: munmap error: %v\n", err)
			}
		}
	}
}

func FreeOutputBuffers() {
	for _, v := range data_dst_buf {
		for _, d := range v {
			if err := syscall.Munmap(d); err != nil {
				fmt.Printf("DST: munmap error: %v\n", err)
			}
		}
	}
}

func streamoff(video_fd int) {
	Type := v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	err := v4l2.IoctlStreamOff(video_fd, &Type)
	if err != nil {
		log.Fatalf("Failed to stream off output interface: %v", err)
	}
	Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	err = v4l2.IoctlStreamOff(video_fd, &Type)
	if err != nil {
		log.Fatalf("Failed to stream off capture interface: %v", err)
	}
}
