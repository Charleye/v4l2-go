package v4l2

/*
#include <linux/videodev2.h>
*/
import "C"

import (
	"log"
	"syscall"
	"unsafe"
)

const (
	VIDIOC_QUERYCAP  = C.VIDIOC_QUERYCAP // Query device capabilities
	VIDIOC_ENUM_FMT  = C.VIDIOC_ENUM_FMT // Enumerate image formats
	VIDIOC_G_FMT     = C.VIDIOC_G_FMT    // Get or set the data format, try a format
	VIDIOC_S_FMT     = C.VIDIOC_S_FMT
	VIDIOC_TRY_FMT   = C.VIDIOC_TRY_FMT
	VIDIOC_G_CTRL    = C.VIDIOC_G_CTRL
	VIDIOC_S_CTRL    = C.VIDIOC_S_CTRL
	VIDIOC_QUERYCTRL = C.VIDIOC_QUERYCTRL
	VIDIOC_QUERYMENU = C.VIDIOC_QUERYMENU //  Enumerate controls and menu control items
	VIDIOC_G_CROP    = C.VIDIOC_G_CROP    // Get or set the current cropping rectangle
	VIDIOC_S_CROP    = C.VIDIOC_S_CROP
	VIDIOC_CROPCAP   = C.VIDIOC_CROPCAP  // Information about the video cropping and scaling abilities
	VIDIOC_QUERYBUF  = C.VIDIOC_QUERYBUF // Query the status of a buffer
	VIDIOC_REQBUFS   = C.VIDIOC_REQBUFS  //  Initiate Memory Mapping, User Pointer I/O or DMA buffer I/O
	VIDIOC_QBUF      = C.VIDIOC_QBUF     // Exchange a buffer with the driver
	VIDIOC_DQBUF     = C.VIDIOC_DQBUF
	VIDIOC_G_PARM    = C.VIDIOC_G_PARM // Get or set streaming parameters
	VIDIOC_S_PARM    = C.VIDIOC_S_PARM

	// Subscribe or unsubscribe event
	VIDIOC_SUBSCRIBE_EVENT   = C.VIDIOC_SUBSCRIBE_EVENT
	VIDIOC_UNSUBSCRIBE_EVENT = C.VIDIOC_UNSUBSCRIBE_EVENT
	VIDIOC_DQEVENT           = C.VIDIOC_DQEVENT // Dequeue event

	// Get or set the value of several controls, try control values
	VIDIOC_G_EXT_CTRLS   = C.VIDIOC_G_EXT_CTRLS
	VIDIOC_S_EXT_CTRLS   = C.VIDIOC_S_EXT_CTRLS
	VIDIOC_TRY_EXT_CTRLS = C.VIDIOC_TRY_EXT_CTRLS
<<<<<<< HEAD

	// Start or stop streaming I/O
	VIDIOC_STREAMON  = C.VIDIOC_STREAMON
	VIDIOC_STREAMOFF = C.VIDIOC_STREAMOFF
=======
>>>>>>> 1ffb17fe995de6201193b915469e6f6be7c5dff0
)

// v4l2 buffer type
const (
	V4L2_BUF_TYPE_VIDEO_CAPTURE        = C.V4L2_BUF_TYPE_VIDEO_CAPTURE
	V4L2_BUF_TYPE_VIDEO_OUTPUT         = C.V4L2_BUF_TYPE_VIDEO_OUTPUT
	V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE = C.V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE
	V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE  = C.V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE
)

const (
	VIDEO_MAX_PLANES = C.VIDEO_MAX_PLANES
)

// Pixel format FOURCC depth Description
const (
	/* Luminance+Chrominance formats */
	V4L2_PIX_FMT_YVU410  = C.V4L2_PIX_FMT_YVU410
	V4L2_PIX_FMT_YVU420  = C.V4L2_PIX_FMT_YVU420
	V4L2_PIX_FMT_YUYV    = C.V4L2_PIX_FMT_YUYV
	V4L2_PIX_FMT_YUV422P = C.V4L2_PIX_FMT_YUV422P

	/* two non contiguous planes - one Y, one Cr + Cb interleaved */
	V4L2_PIX_FMT_NV12M = C.V4L2_PIX_FMT_NV12M
	V4L2_PIX_FMT_NV21M = C.V4L2_PIX_FMT_NV21M
)

const (
	/* Query flags, to be ORed with the control ID */
	V4L2_CTRL_FLAG_NEXT_CTRL = C.V4L2_CTRL_FLAG_NEXT_CTRL

	/*  Control flags  */
	V4L2_CTRL_FLAG_DISABLED = C.V4L2_CTRL_FLAG_DISABLED
)

// control type
const (
	V4L2_CTRL_TYPE_MENU         = C.V4L2_CTRL_TYPE_MENU
	V4L2_CTRL_TYPE_INTEGER_MENU = C.V4L2_CTRL_TYPE_INTEGER_MENU
)

// memory type
const (
	V4L2_MEMORY_MMAP    = C.V4L2_MEMORY_MMAP
	V4L2_MEMORY_USERPTR = C.V4L2_MEMORY_USERPTR
	V4L2_MEMORY_OVERLAY = C.V4L2_MEMORY_OVERLAY
	V4L2_MEMORY_DMABUF  = C.V4L2_MEMORY_DMABUF
)

// Event types
const (
	V4L2_EVENT_ALL   = C.V4L2_EVENT_ALL
	V4L2_EVENT_VSYNC = C.V4L2_EVENT_VSYNC
	V4L2_EVENT_EOS   = C.V4L2_EVENT_EOS
	V4L2_EVENT_CTRL  = C.V4L2_EVENT_CTRL
)

func GetNameByFourCC(fourcc uint32) string {
	const mask = 0xFF
	b := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b[i] = byte((fourcc) >> uint32(i*8) & mask)
	}

	return string(b)
}

func GetFourCCByName(name string) uint32 {
	switch name {
	case "YUV9", "YVU 4:1:0":
		return V4L2_PIX_FMT_YVU410
	case "YV12", "YVU 4:2:0":
		return V4L2_PIX_FMT_YVU420
	case "YUYV", "YUV 4:2:2":
		return V4L2_PIX_FMT_YUYV
	case "422P", "YVU422 planar", "YVU422P":
		return V4L2_PIX_FMT_YUV422P
	case "NM12", "Y/CbCr 4:2:0":
		return V4L2_PIX_FMT_NV12M
	case "NM21", "Y/CrCb 4:2:0":
		return V4L2_PIX_FMT_NV21M

	default:
		log.Fatal("Unexpected FourCC name")
	}
	return 0
}

func ioctl(fd int, request uint, argp unsafe.Pointer) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), uintptr(request), uintptr(argp))
	if err != 0 {
		return err
	}
	return nil
}
