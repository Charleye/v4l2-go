package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	v4l2 "github.com/Charleye/v4l2-go"
)

var input_file = flag.String("f", "", "input file")
var mfc_node = flag.String("v", "", "MFC device node")
var width = flag.Uint("w", 0, "width  in pixel")
var height = flag.Uint("h", 0, "height in pixel")
var fourcc = flag.String("r", "NM12", "pixel format for input interface")

/* global varibales */
var data_input_file []byte
var input_file_size int64
var num_src_planes, num_dst_planes int
var data_src_buf, data_dst_buf [][][]byte
var src_frame_size, dst_frame_size uint32
var num_src_bufs, num_dst_bufs uint32

func main() {
	flag.Parse()

	file_fd := InitInputFile()
	defer syscall.Close(file_fd)

	video_fd := InitMFCVideoNode()
	defer syscall.Close(video_fd)

	SetInputFormat(video_fd)
	AllocInputBuffers(video_fd)
	defer FreeInputBuffers()

	SetOutputFormat(video_fd)
	AllocOutputBuffers(video_fd)
	defer FreeOutputBuffers()
}

func InitInputFile() int {
	if *input_file == "" {
		log.Fatal("Failed to specify input file")
	}

	var st syscall.Stat_t
	var fd int
	fd, err := syscall.Open(*input_file, syscall.O_RDONLY, 0)
	if err != nil {
		log.Fatalf("Failed to open input file: %s", *input_file)
	}
	syscall.Fstat(fd, &st)
	input_file_size = st.Size
	fmt.Printf("input file size: %v\n", input_file_size)

	data_input_file, err = syscall.Mmap(fd, 0, int(input_file_size),
		syscall.PROT_READ, syscall.MAP_SHARED)

	fmt.Println("len cap: ", len(data_input_file), cap(data_input_file))
	return fd
}

func process(video_fd int) {
	var src_planes [v4l2.VIDEO_MAX_PLANES]v4l2.V4L2_Plane
	var dst_planes [v4l2.VIDEO_MAX_PLANES]v4l2.V4L2_Plane
	var src_buf, dst_buf v4l2.V4L2_Buffer

	src_buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	src_buf.Memory = v4l2.V4L2_MEMORY_MMAP
	src_buf.M = v4l2.PointerToBytes(&src_planes[0])
	src_buf.Length = uint32(num_src_planes)

	for i := 0; i < int(num_dst_bufs); i++ {
		copy(data_src_buf[i][0], data_input_file)
		copy(data_src_buf[i][1], data_input_file)
		src_buf.Index = uint32(i)
		err := v4l2.IoctlQBuf(video_fd, &src_buf)
		if err != nil {
			fmt.Printf("Failed to enqueue buffer %d/%d to %d\n", i, num_dst_bufs, video_fd)
		}
	}

	dst_buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	dst_buf.Memory = v4l2.V4L2_MEMORY_MMAP
	dst_buf.Index = 0
	dst_buf.M = v4l2.PointerToBytes(&dst_planes[0])
	dst_buf.Length = uint32(num_dst_planes)

}

func InitMFCVideoNode() int {
	vid_fd, err := syscall.Open(*mfc_node, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		log.Fatalf("Failed to open MFC node: %s", *mfc_node)
	}

	var caps v4l2.V4L2_Capability
	err = v4l2.IoctlQueryCap(vid_fd, &caps)
	if err != nil {
		log.Fatal("Failed to query capabilities")
	}
	if caps.Capabilities&v4l2.V4L2_CAP_DEVICE_CAPS != 0 {
		if caps.DeviceCaps&v4l2.V4L2_CAP_VIDEO_M2M_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes mem-to-mem (%#x)\n",
				*mfc_node, caps.DeviceCaps)
		}
	} else {
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_CAPTURE_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes capture (%#x)\n",
				*mfc_node, caps.Capabilities)
		}
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_OUTPUT_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes output (%#x)\n",
				*mfc_node, caps.Capabilities)
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
	pixmp.NumPlanes = 2
	pixmp.PlaneFmt[0].BytesPerLine = 1280
	pixmp.PlaneFmt[0].SizeImage = 0xE1000
	pixmp.PlaneFmt[1].BytesPerLine = 1280
	pixmp.PlaneFmt[1].SizeImage = 0x70800
	format.Fmt = &pixmp
	err := v4l2.IoctlSetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to set input format")
	}
	fmt.Println(planes)
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
	pixmp.PixelFormat = v4l2.GetFourCCByName("H264")
	pixmp.PlaneFmt[0].SizeImage = 2 * 1024 * 1024
	pixmp.NumPlanes = 1
	format.Fmt = &pixmp
	err := v4l2.IoctlSetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to set output format")
	}
	num_dst_planes = int(pixmp.NumPlanes)
	fmt.Printf("out_width: %v, out_height: %v\n", pixmp.Width, pixmp.Height)
}

func AllocInputBuffers(video_fd int) {
	/* request input buffer */
	var reqbuf v4l2.V4L2_Requestbuffers
	reqbuf.Count = 16
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
	reqbuf.Count = 4
	err := v4l2.IoctlRequestBuffers(video_fd, &reqbuf)
	if err != nil {
		log.Fatalf("Failed to request output buffer: %v", err)
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
		buf.Length = v4l2.VIDEO_MAX_PLANES
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

			// mmap output buffer
			buf, err := syscall.Mmap(video_fd, int64(offset), int(planes[i].Length),
				syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
			if err != nil {
				log.Fatalf("mmap error: %v\n", err)
			}
			data_dst_planes = append(data_dst_planes, buf)
		}
		data_dst_buf = append(data_dst_buf, data_dst_planes)

		err = v4l2.IoctlQBuf(video_fd, &buf)
		if err != nil {
			log.Fatalf("Failed to enqueue output buffer: %v", err)
		} else {
			fmt.Printf("Enqueued buffer %d/%d to %d\n", index, num_dst_bufs, video_fd)
		}
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
