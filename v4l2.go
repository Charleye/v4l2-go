package main

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
	VIDIOC_QUERYCAP = C.VIDIOC_QUERYCAP // Query device capabilities
	VIDIOC_ENUM_FMT = C.VIDIOC_ENUM_FMT // Enumerate image formats
	VIDIOC_G_FMT    = C.VIDIOC_G_FMT    // Get or set the data format, try a format
	VIDIOC_S_FMT    = C.VIDIOC_S_FMT
	VIDIOC_TRY_FMT  = C.VIDIOC_TRY_FMT
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
