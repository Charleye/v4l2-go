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
	VIDIOC_QUERYCAP       = C.VIDIOC_QUERYCAP // Query device capabilities
	VIDIOC_ENUM_FMT       = C.VIDIOC_ENUM_FMT // Enumerate image formats
	VIDIOC_G_FMT          = C.VIDIOC_G_FMT    // Get or set the data format, try a format
	VIDIOC_S_FMT          = C.VIDIOC_S_FMT
	VIDIOC_TRY_FMT        = C.VIDIOC_TRY_FMT
	VIDIOC_G_CTRL         = C.VIDIOC_G_CTRL
	VIDIOC_S_CTRL         = C.VIDIOC_S_CTRL
	VIDIOC_QUERYCTRL      = C.VIDIOC_QUERYCTRL
	VIDIOC_QUERYMENU      = C.VIDIOC_QUERYMENU //  Enumerate controls and menu control items
	VIDIOC_QUERY_EXT_CTRL = C.VIDIOC_QUERY_EXT_CTRL
	VIDIOC_G_CROP         = C.VIDIOC_G_CROP // Get or set the current cropping rectangle
	VIDIOC_S_CROP         = C.VIDIOC_S_CROP
	VIDIOC_CROPCAP        = C.VIDIOC_CROPCAP  // Information about the video cropping and scaling abilities
	VIDIOC_QUERYBUF       = C.VIDIOC_QUERYBUF // Query the status of a buffer
	VIDIOC_REQBUFS        = C.VIDIOC_REQBUFS  //  Initiate Memory Mapping, User Pointer I/O or DMA buffer I/O
	VIDIOC_QBUF           = C.VIDIOC_QBUF     // Exchange a buffer with the driver
	VIDIOC_DQBUF          = C.VIDIOC_DQBUF
	VIDIOC_G_PARM         = C.VIDIOC_G_PARM // Get or set streaming parameters
	VIDIOC_S_PARM         = C.VIDIOC_S_PARM

	// Subscribe or unsubscribe event
	VIDIOC_SUBSCRIBE_EVENT   = C.VIDIOC_SUBSCRIBE_EVENT
	VIDIOC_UNSUBSCRIBE_EVENT = C.VIDIOC_UNSUBSCRIBE_EVENT
	VIDIOC_DQEVENT           = C.VIDIOC_DQEVENT // Dequeue event

	// Get or set the value of several controls, try control values
	VIDIOC_G_EXT_CTRLS   = C.VIDIOC_G_EXT_CTRLS
	VIDIOC_S_EXT_CTRLS   = C.VIDIOC_S_EXT_CTRLS
	VIDIOC_TRY_EXT_CTRLS = C.VIDIOC_TRY_EXT_CTRLS

	// Start or stop streaming I/O
	VIDIOC_STREAMON  = C.VIDIOC_STREAMON
	VIDIOC_STREAMOFF = C.VIDIOC_STREAMOFF
)

const (
	V4L2_CAP_VIDEO_CAPTURE        = C.V4L2_CAP_VIDEO_CAPTURE
	V4L2_CAP_VIDEO_CAPTURE_MPLANE = C.V4L2_CAP_VIDEO_CAPTURE_MPLANE
	V4L2_CAP_VIDEO_OUTPUT         = C.V4L2_CAP_VIDEO_OUTPUT
	V4L2_CAP_VIDEO_OUTPUT_MPLANE  = C.V4L2_CAP_VIDEO_OUTPUT_MPLANE
	V4L2_CAP_VIDEO_M2M            = C.V4L2_CAP_VIDEO_M2M
	V4L2_CAP_VIDEO_M2M_MPLANE     = C.V4L2_CAP_VIDEO_M2M_MPLANE
	V4L2_CAP_STREAMING            = C.V4L2_CAP_STREAMING
	V4L2_CAP_DEVICE_CAPS          = C.V4L2_CAP_DEVICE_CAPS
)

/* field order */
const (
	V4L2_FIELD_ANY  = C.V4L2_FIELD_ANY
	V4L2_FIELD_NONE = C.V4L2_FIELD_NONE
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
	V4L2_PIX_FMT_YYUV    = C.V4L2_PIX_FMT_YYUV
	V4L2_PIX_FMT_YVYU    = C.V4L2_PIX_FMT_YVYU
	V4L2_PIX_FMT_UYVY    = C.V4L2_PIX_FMT_UYVY
	V4L2_PIX_FMT_VYUY    = C.V4L2_PIX_FMT_VYUY
	V4L2_PIX_FMT_YUV422P = C.V4L2_PIX_FMT_YUV422P
	V4L2_PIX_FMT_YUV411P = C.V4L2_PIX_FMT_YUV411P
	V4L2_PIX_FMT_Y41P    = C.V4L2_PIX_FMT_Y41P
	V4L2_PIX_FMT_YUV444  = C.V4L2_PIX_FMT_YUV444
	V4L2_PIX_FMT_YUV555  = C.V4L2_PIX_FMT_YUV555
	V4L2_PIX_FMT_YUV565  = C.V4L2_PIX_FMT_YUV565
	V4L2_PIX_FMT_YUV32   = C.V4L2_PIX_FMT_YUV32
	V4L2_PIX_FMT_YUV410  = C.V4L2_PIX_FMT_YUV410
	V4L2_PIX_FMT_YUV420  = C.V4L2_PIX_FMT_YUV420
	V4L2_PIX_FMT_HI240   = C.V4L2_PIX_FMT_HI240
	V4L2_PIX_FMT_HM12    = C.V4L2_PIX_FMT_HM12
	V4L2_PIX_FMT_M420    = C.V4L2_PIX_FMT_M420

	/* two non contiguous planes - one Y, one Cr + Cb interleaved */
	V4L2_PIX_FMT_NV12M = C.V4L2_PIX_FMT_NV12M
	V4L2_PIX_FMT_NV21M = C.V4L2_PIX_FMT_NV21M

	/* compressed formats */
	V4L2_PIX_FMT_MJPEG = C.V4L2_PIX_FMT_MJPEG
	V4L2_PIX_FMT_JPEG  = C.V4L2_PIX_FMT_JPEG
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

const (
	V4L2_CTRL_MAX_DIMS = C.V4L2_CTRL_MAX_DIMS
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
	case "YYUV":
		return V4L2_PIX_FMT_YYUV
	case "UYVY":
		return V4L2_PIX_FMT_UYVY
	case "VYUY":
		return V4L2_PIX_FMT_VYUY
	case "422P", "YVU422 planar", "YVU422P":
		return V4L2_PIX_FMT_YUV422P
	case "411P", "YVU411P":
		return V4L2_PIX_FMT_YUV411P
	case "Y41P", "YUV 4:1:1":
		return V4L2_PIX_FMT_Y41P
	case "Y444":
		return V4L2_PIX_FMT_YUV444
	case "YUVO":
		return V4L2_PIX_FMT_YUV555
	case "YUVP":
		return V4L2_PIX_FMT_YUV565
	case "YUV4":
		return V4L2_PIX_FMT_YUV32
	case "YUV9":
		return V4L2_PIX_FMT_YUV410
	case "YU12":
		return V4L2_PIX_FMT_YUV420
	case "HI24":
		return V4L2_PIX_FMT_HI240
	case "HM12":
		return V4L2_PIX_FMT_HM12
	case "M420":
		return V4L2_PIX_FMT_HM12

	case "NM12", "Y/CbCr 4:2:0":
		return V4L2_PIX_FMT_NV12M
	case "NM21", "Y/CrCb 4:2:0":
		return V4L2_PIX_FMT_NV21M
	case "MJPG", "Motion-JPEG":
		return V4L2_PIX_FMT_MJPEG
	case "JPEG", "JFIF JPEG":
		return V4L2_PIX_FMT_JPEG

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
