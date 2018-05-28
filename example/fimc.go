/* Tested on odroid-xu4 board */
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
var num_src_planes int
var num_src_bufs uint32
var src_frame_size uint32
var data_src_buf [][][]byte

var num_dst_planes int
var num_dst_bufs uint32
var data_dst_buf [][][]byte
var dst_frame_size uint32

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
	defer func() {
		for _, v := range data_src_buf {
			for _, d := range v {
				syscall.Munmap(d)
			}
		}
	}()

	/* copy file data into the buffer */
	copy(data_src_buf[0][0], data_input_file)

	/* set output format */
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	pixmp = v4l2.V4L2_Pix_Format_Mplane{}
	pixmp.Width = uint32(*width)
	pixmp.Height = uint32(*height)
	pixmp.Field = v4l2.V4L2_FIELD_ANY
	pixmp.PlaneFmt[0].BytesPerLine = uint32(*width)
	pixmp.PixelFormat = v4l2.GetFourCCByName("UYVY")
	format.Fmt = &pixmp
	err = v4l2.IoctlSetFmt(video_fd, &format)
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

	/* request output buffers */
	reqbuf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	reqbuf.Count = 1
	err = v4l2.IoctlRequestBuffers(video_fd, &reqbuf)
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
	defer func() {
		for _, v := range data_dst_buf {
			for _, d := range v {
				syscall.Munmap(d)
			}
		}
	}()

	/* process frames */
	process(video_fd)
	defer streamoff(video_fd)
}

func process(video_fd int) {
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
	out_file, err := os.OpenFile("in422_uyvy_800_600.raw", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Failed to open output file")
	}
	fmt.Println("Generating output file...")
	n, _ := out_file.Write(data_dst_buf[0][0][:dst_planes[0].BytesUsed])
	fmt.Println(n)
	out_file.Close()
	fmt.Printf("Output file: %s, size: %v\n", "test.rgb", dst_planes[0].BytesUsed)
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
