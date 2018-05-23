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

const (
	ENCODE = 0
	DECODE = 1
)

var mode = flag.Int("m", 0, "mode 0-encode 1-decode")
var in = flag.String("f", "", "input file")
var out = flag.String("o", "", "output file")
var video_node = flag.String("v", "/dev/video30", "video node")
var width = flag.Uint("w", 0, "width in pixel")
var height = flag.Uint("h", 0, "height in pixel")
var fourcc = flag.String("r", "", "pixel format string")

var def_outfile = "test.raw"
var input_file_sz int64
var data_input_file []byte
var num_src_bufs uint32
var num_dst_bufs uint32
var src_buf_size uint32
var data_src_buf []byte
var data_dst_buf []byte
var capture_buffer_sz uint32

func InitInputFile() int {
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

func main() {
	flag.Parse()

	input_fd := InitInputFile()
	defer syscall.Close(input_fd)

	video_fd, err := syscall.Open(*video_node, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		log.Fatalf("Failed to open video node: %s\n", *video_node)
	}

	var caps v4l2.V4L2_Capability
	err = v4l2.IoctlQueryCap(video_fd, &caps)
	if err != nil {
		log.Fatal("Failed to query capabilities of video node")
	}
	if caps.DeviceCaps|v4l2.V4L2_CAP_DEVICE_CAPS != 0 {
		if caps.DeviceCaps&v4l2.V4L2_CAP_VIDEO_M2M == 0 {
			log.Fatalf("Device %s does not support mem-to-mem (%#x)\n",
				*video_node, caps.DeviceCaps)
		}
	} else {
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_CAPTURE == 0 {
			log.Fatalf("Device %s does not support capture (%#x)\n",
				*video_node, caps.Capabilities)
		}
		if caps.Capabilities&v4l2.V4L2_CAP_VIDEO_OUTPUT == 0 {
			log.Fatalf("Device %s does not support output (%#x)\n",
				*video_node, caps.Capabilities)
		}
	}

	/* set input format */
	var format v4l2.V4L2_Format
	var pixfmt v4l2.V4L2_Pix_Format
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	pixfmt.Width = uint32(*width)
	pixfmt.Height = uint32(*height)
	pixfmt.SizeImage = uint32(input_file_sz)
	if *mode == ENCODE {
		pixfmt.PixelFormat = v4l2.GetFourCCByName(*fourcc)
	} else {
		pixfmt.PixelFormat = v4l2.V4L2_PIX_FMT_JPEG
	}
	pixfmt.Field = v4l2.V4L2_FIELD_ANY
	pixfmt.BytesPerLine = 0
	format.Fmt = &pixfmt
	fmt.Println(pixfmt)
	err = v4l2.IoctlSetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to set format")
	}

	/* request input buffer */
	var reqbuf v4l2.V4L2_Requestbuffers
	reqbuf.Count = 1
	reqbuf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	reqbuf.Memory = v4l2.V4L2_MEMORY_MMAP
	err = v4l2.IoctlRequestBuffers(video_fd, &reqbuf)
	if err != nil {
		log.Fatal("Failed to request buffers")
	}
	num_src_bufs = reqbuf.Count

	/* query buffer parameters */
	var buf v4l2.V4L2_Buffer
	buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	buf.Memory = v4l2.V4L2_MEMORY_MMAP
	buf.Index = 0
	err = v4l2.IoctlQueryBuf(video_fd, &buf)
	if err != nil {
		log.Fatal("Failed to query buffer")
	}
	src_buf_size = buf.Length

	/* mmap buffer */
	var offset uint32
	v4l2.GetValueFromUnion(buf.M, &offset)
	data_src_buf, err = syscall.Mmap(video_fd, int64(offset), int(buf.Length),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("mmap error: %v\n", err)
	}
	defer syscall.Munmap(data_src_buf)

	/* copy input file data to the bufffer */
	copy(data_src_buf, data_input_file)

	/* queue input buffer */
	buf = v4l2.V4L2_Buffer{}
	buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	buf.Memory = v4l2.V4L2_MEMORY_MMAP
	buf.Index = 0
	buf.BytesUsed = uint32(input_file_sz)
	err = v4l2.IoctlQBuf(video_fd, &buf)
	if err != nil {
		log.Fatal("Failed to queue buffer")
	}

	var stream_type int = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	err = v4l2.IoctlStreamOn(video_fd, &stream_type)
	if err != nil {
		log.Fatal("Failed to stream on")
	}

	if *mode == DECODE {
		format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
		err = v4l2.IoctlGetFmt(video_fd, &format)
		if err != nil {
			log.Fatal("Failed to get format")
		}
		*width = uint(pixfmt.Width)
		*height = uint(pixfmt.Height)
		fmt.Printf("input JPEG dimensions: %vx%v\n", *width, *height)

		/* apply scaling */
	}

	/* set output format */
	format.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	pixfmt.Width = uint32(*width)
	pixfmt.Height = uint32(*height)
	pixfmt.SizeImage = uint32(*width * (*height) * 4)
	if *mode == DECODE {
		pixfmt.PixelFormat = v4l2.GetFourCCByName(*fourcc)
	} else {
		pixfmt.PixelFormat = v4l2.V4L2_PIX_FMT_JPEG
	}
	pixfmt.Field = v4l2.V4L2_FIELD_ANY
	err = v4l2.IoctlSetFmt(video_fd, &format)
	if err != nil {
		log.Fatal("Failed to set output format")
	}
	fmt.Printf("output image dimensions: %vx%v\n", pixfmt.Width, pixfmt.Height)

	if *mode == ENCODE {
		/* set necoder ctrls */
	}

	/* request output buffer */
	reqbuf.Count = 1
	reqbuf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlRequestBuffers(video_fd, &reqbuf)
	if err != nil {
		log.Fatal("Failed to request output buffer")
	}
	num_dst_bufs = reqbuf.Count

	/* query buffer parameters */
	buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	buf.Memory = v4l2.V4L2_MEMORY_MMAP
	buf.Index = 0
	err = v4l2.IoctlQueryBuf(video_fd, &buf)
	if err != nil {
		log.Fatal("Failed to query output buffer")
	}

	/* mmap output buffer */
	v4l2.GetValueFromUnion(buf.M, &offset)
	data_dst_buf, err = syscall.Mmap(video_fd, int64(offset), int(buf.Length),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("Failed to mmap output buffer: %v", err)
	}

	buf = v4l2.V4L2_Buffer{} // zero
	buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	buf.Memory = v4l2.V4L2_MEMORY_MMAP
	buf.Index = 0
	err = v4l2.IoctlQBuf(video_fd, &buf)
	if err != nil {
		log.Fatal("Failed to enqueue output buffer")
	}

	stream_type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlStreamOn(video_fd, &stream_type)
	if err != nil {
		log.Fatalf("Failed to stream on output interface")
	}

	/* dequeue buffer */
	var read_fds syscall.FdSet
	read_fds.Bits[video_fd/64] = 1 << uint32(video_fd%64)
	r, err := syscall.Select(video_fd+1, &read_fds, nil, nil, nil)
	if r < 0 {
		log.Fatalf("select errors: %v\n", err)
	}

	var out_file *os.File
	if dq_frame(video_fd) > 0 {
		fmt.Println("dequeue frame failed")
		goto done
	}

	if *mode == DECODE {
		/* get input JPEG subsampling info */
	}

	/* generate output file */
	if *out == "" {
		*out = def_outfile
	}

	out_file, err = os.OpenFile(*out, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Failed to open output file")
	}
	fmt.Println("Generating output file...")
	out_file.Write(data_dst_buf[:capture_buffer_sz])
	out_file.Close()
	fmt.Printf("Output file: %s, size: %v\n", *out, capture_buffer_sz)

done:
	syscall.Close(video_fd)
	syscall.Munmap(data_src_buf)
	syscall.Munmap(data_dst_buf)
	syscall.Munmap(data_input_file)
}

func dq_frame(vid_fd int) int {
	var buf v4l2.V4L2_Buffer
	buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_OUTPUT
	buf.Memory = v4l2.V4L2_MEMORY_MMAP
	err := v4l2.IoctlDQBuf(vid_fd, &buf)
	fmt.Printf("Dequeued dst buffer, index: %d\n", buf.Index)
	if err != nil {
		switch err {
		case syscall.EAGAIN:
			fmt.Println("Got EAGAIN")
			return 0
		case syscall.EIO:
			fmt.Println("Got EIO")
			return 0
		default:
			fmt.Println("ioctl error")
			return 1
		}
	}

	buf.Type = v4l2.V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = v4l2.IoctlDQBuf(vid_fd, &buf)
	fmt.Printf("Dequeued dst buffer, index: %d\n", buf.Index)
	if err != nil {
		switch err {
		case syscall.EAGAIN:
			fmt.Println("Got EAGAIN")
			return 0
		case syscall.EIO:
			fmt.Println("Got EIO")
			return 0
		default:
			fmt.Println("ioctl error")
			return 1
		}
	}

	capture_buffer_sz = buf.BytesUsed
	return 0
}
