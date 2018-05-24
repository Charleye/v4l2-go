/* Tested on odroid-xu4 board */
package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	v4l2 "github.com/Charleye/v4l2-go"
)

var in = flag.String("f", "", "input file")
var video_node = flag.String("v", "", "video device")
var width = flag.Uint("w", 0, "width  in pixel")
var height = flag.Uint("h", 0, "height in pixel")
var fourcc = flag.String("r", "", "pixel format")

var data_input_file []byte
var input_file_sz int64
var num_planes int
var num_src_bufs uint32
var data_src_buf [][][]byte
var num_dst_bufs uint32

func InitInputFile() int {
	if *in == "" {
		log.Fatal("Failed to specify input file")
	}

	var st syscall.Stat_t
	var fd int
	fd, err := syscall.Open(*in, syscall.O_RDONLY, 0)
	if err != nil {
		log.Fatalf("Failed to open input file: %s", *in)
	}
	syscall.Fstat(fd, &st)
	input_file_sz = st.Size
	fmt.Printf("input file size: %v\n", input_file_sz)

	data_input_file, err = syscall.Mmap(fd, 0, int(input_file_sz),
		syscall.PROT_READ, syscall.MAP_SHARED)

	return fd
}

func InitVideoNode() int {
	vid_fd, err := syscall.Open(*video_node, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		log.Fatalf("Failed to open video node: %s", *video_node)
	}

	var caps v4l2.V4L2_Capability
	err = v4l2.IoctlQueryCap(vid_fd, &caps)
	if err != nil {
		log.Fatal("Failed to query capabilities")
	}
	if caps.Capabilities&v4l2.V4L2_CAP_DEVICE_CAPS != 0 {
		if caps.DeviceCaps&v4l2.V4L2_CAP_VIDEO_M2M_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes mem-to-mem (%#x)\n",
				*video_node, caps.DeviceCaps)
		}
	} else {
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_CAPTURE_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes capture (%#x)\n",
				*video_node, caps.Capabilities)
		}
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_OUTPUT_MPLANE == 0 {
			log.Fatalf("Device %s does not support multi-planes output (%#x)\n",
				*video_node, caps.Capabilities)
		}
	}
	return vid_fd
}

func main() {
	flag.Parse()

	input_fd := InitInputFile()
	defer syscall.Close(input_fd)
	defer syscall.Munmap(data_input_file)

	video_fd := InitVideoNode()
	defer syscall.Close(video_fd)

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
	num_planes = int(pixmp.NumPlanes)
	var frame_size uint32
	for i := 0; i < num_planes; i++ {
		frame_size += pixmp.PlaneFmt[i].SizeImage
		fmt.Printf("plane[%d]: bytesperline: %d, sizeimage: %d\n",
			i, pixmp.PlaneFmt[i].BytesPerLine, pixmp.PlaneFmt[i].SizeImage)
	}
	fmt.Printf("SRC frame_size: %v\n", frame_size)

	err = v4l2.IoctlGetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to get input format")
	}
	*width = uint(pixmp.Width)
	*height = uint(pixmp.Height)
	fmt.Printf("width: %v, height: %v\n", *width, *height)

	/* request input buffer */
	var reqbuf v4l2.V4L2_Requestbuffers
	reqbuf.Count = 1
	reqbuf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
	reqbuf.Memory = v4l2.V4L2_MEMORY_MMAP
	err = v4l2.IoctlRequestBuffers(video_fd, &reqbuf)
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
		buf.Length = uint32(num_planes)
		buf.Memory = v4l2.V4L2_MEMORY_MMAP
		buf.Index = uint32(index)

		err := v4l2.IoctlQueryBuf(video_fd, &buf)
		if err != nil {
			log.Fatal(err)
		}

		data_src_planes := make([][]byte, 0, num_planes)
		for i := 0; i < num_planes; i++ {
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
	defer func() {
		for _, v := range data_src_buf {
			for _, d := range v {
				syscall.Munmap(d)
			}
		}
	}()

}
