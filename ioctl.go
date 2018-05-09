package main

/*
#include <linux/videodev2.h>
*/
import "C"

import (
	"log"
	"unsafe"
)

type V4L2_Capability struct {
	Driver       string
	Card         string
	BusInfo      string
	Version      uint32
	Capabilities uint32
	DeviceCaps   uint32
}

func (c *V4L2_Capability) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_capability)(ptr)
	c.Driver = C.GoStringN((*C.char)(unsafe.Pointer(&p.driver[0])), 16)
	c.Card = C.GoStringN((*C.char)(unsafe.Pointer(&p.card[0])), 32)
	c.BusInfo = C.GoStringN((*C.char)(unsafe.Pointer(&p.bus_info[0])), 32)
	c.Version = uint32(p.version)
	c.Capabilities = uint32(p.capabilities)
	c.DeviceCaps = uint32(p.device_caps)
}

func IoctlQueryCap(fd int, argp *V4L2_Capability) error {
	var caps C.struct_v4l2_capability
	p := unsafe.Pointer(&caps)
	err := ioctl(fd, VIDIOC_QUERYCAP, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Fmtdesc struct {
	Index       uint32
	Type        uint32
	Flags       uint32
	Description string
	PixelFormat uint32
}

func (f *V4L2_Fmtdesc) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_fmtdesc)(ptr)
	p.index = C.__u32(f.Index)

	// due to type field, it is keyword in golang
	tmp := (*uint32)(unsafe.Pointer(
		uintptr(ptr) + offset_fmtdesc_type))
	*tmp = f.Type
}

func (f *V4L2_Fmtdesc) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_fmtdesc)(ptr)
	f.Flags = uint32(p.flags)
	f.Description = C.GoString((*C.char)(unsafe.Pointer(&p.description[0])))
	f.PixelFormat = uint32(p.pixelformat)
}

func IoctlEnumFmt(fd int, argp *V4L2_Fmtdesc) error {
	var f C.struct_v4l2_fmtdesc
	p := unsafe.Pointer(&f)
	argp.set(p)
	err := ioctl(fd, VIDIOC_ENUM_FMT, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Format struct {
	Type uint32
	fmt  interface{}
}

type V4L2_Pix_Format struct {
	Width        uint32
	Height       uint32
	PixelFormat  uint32
	Field        uint32
	BytesPerLine uint32
	SizeImage    uint32
	ColorSpace   uint32
	Priv         uint32
	Flags        uint32
	YcbcrEnc     uint32
	Quantization uint32
	XferFunc     uint32
}

type V4L2_Pix_Format_Mplane struct {
	Width        uint32
	Height       uint32
	PixelFormat  uint32
	Field        uint32
	ColorSpace   uint32
	PlaneFmt     [VIDEO_MAX_PLANES]V4L2_Plane_Pix_Format
	NumPlanes    uint8
	Flags        uint8
	YcbcrEnc     uint8
	Quantization uint8
	XferFunc     uint8
}

type V4L2_Plane_Pix_Format struct {
	SizeImage    uint32
	BytesPerLine uint32
}

func (f *V4L2_Plane_Pix_Format) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_plane_pix_format)(ptr)
	p.sizeimage = C.__u32(f.SizeImage)
	p.bytesperline = C.__u32(f.BytesPerLine)
}

func (f *V4L2_Plane_Pix_Format) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_plane_pix_format)(ptr)
	f.SizeImage = uint32(p.sizeimage)
	f.BytesPerLine = uint32(p.bytesperline)
}

func (f *V4L2_Pix_Format_Mplane) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_pix_format_mplane)(ptr)
	p.width = C.__u32(f.Width)
	p.height = C.__u32(f.Height)
	p.pixelformat = C.__u32(f.PixelFormat)
	p.field = C.__u32(f.Field)
	p.colorspace = C.__u32(f.ColorSpace)
	p.num_planes = C.__u8(f.NumPlanes)
	for i := 0; i < int(f.NumPlanes); i++ {
		f.PlaneFmt[i].set(unsafe.Pointer(&p.plane_fmt[i]))
	}
	p.flags = C.__u8(f.Flags)
	p.ycbcr_enc = C.__u8(f.YcbcrEnc)
	p.quantization = C.__u8(f.Quantization)
	p.xfer_func = C.__u8(f.XferFunc)
}

func (f *V4L2_Pix_Format_Mplane) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_pix_format_mplane)(ptr)
	f.Width = uint32(p.width)
	f.Height = uint32(p.height)
	f.PixelFormat = uint32(p.pixelformat)
	f.Field = uint32(p.field)
	f.ColorSpace = uint32(p.colorspace)
	f.NumPlanes = uint8(p.num_planes)
	for i := 0; i < int(f.NumPlanes); i++ {
		f.PlaneFmt[i].get(unsafe.Pointer(&p.plane_fmt[i]))
	}
	f.Flags = uint8(p.flags)
	f.YcbcrEnc = uint8(p.ycbcr_enc)
	f.Quantization = uint8(p.quantization)
	f.XferFunc = uint8(p.xfer_func)
}

func (f *V4L2_Pix_Format) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_pix_format)(ptr)
	p.width = C.__u32(f.Width)
	p.height = C.__u32(f.Height)
	p.pixelformat = C.__u32(f.PixelFormat)
	p.field = C.__u32(f.Field)
	p.bytesperline = C.__u32(f.BytesPerLine)
	p.sizeimage = C.__u32(f.SizeImage)
	p.colorspace = C.__u32(f.ColorSpace)
	p.priv = C.__u32(f.Priv)
	p.flags = C.__u32(f.Flags)
	p.ycbcr_enc = C.__u32(f.YcbcrEnc)
	p.quantization = C.__u32(f.Quantization)
	p.xfer_func = C.__u32(f.XferFunc)
}

func (f *V4L2_Pix_Format) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_pix_format)(ptr)
	f.Width = uint32(p.width)
	f.Height = uint32(p.height)
	f.PixelFormat = uint32(p.pixelformat)
	f.Field = uint32(p.field)
	f.BytesPerLine = uint32(p.bytesperline)
	f.SizeImage = uint32(p.sizeimage)
	f.ColorSpace = uint32(p.colorspace)
	f.Priv = uint32(p.priv)
	f.Flags = uint32(p.flags)
	f.YcbcrEnc = uint32(p.ycbcr_enc)
	f.Quantization = uint32(p.quantization)
	f.XferFunc = uint32(p.xfer_func)
}

func (f *V4L2_Format) set(ptr unsafe.Pointer) {
	// due to type field, it is keyword in golang
	tmp := (*uint32)(unsafe.Pointer(
		uintptr(ptr) + offset_format_type))
	*tmp = f.Type
}

func (f *V4L2_Format) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_format)(ptr)

	switch pf := f.fmt.(type) {
	case *V4L2_Pix_Format:
		pf.get(unsafe.Pointer(&p.fmt))
	case *V4L2_Pix_Format_Mplane:
		pf.get(unsafe.Pointer(&p.fmt))
	default:
		log.Fatalf("Unexpected type %T\n", pf)
	}
}

func IoctlGetFmt(fd int, argp *V4L2_Format) error {
	var vf C.struct_v4l2_format
	p := unsafe.Pointer(&vf)
	argp.set(p)
	err := ioctl(fd, VIDIOC_G_FMT, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

func IoctlSetFmt(fd int, argp *V4L2_Format) error {
	var vf C.struct_v4l2_format
	p := unsafe.Pointer(&vf)
	argp.set(p)
	switch pf := argp.fmt.(type) {
	case *V4L2_Pix_Format:
		pf.set(unsafe.Pointer(&vf.fmt))
	case *V4L2_Pix_Format_Mplane:
		pf.set(unsafe.Pointer(&vf.fmt))
	default:
		log.Fatalf("Unexpected type %T\n", pf)
	}
	err := ioctl(fd, VIDIOC_S_FMT, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

func IoctlTryFmt(fd int, argp *V4L2_Format) error {
	var vf C.struct_v4l2_format
	p := unsafe.Pointer(&vf)
	argp.set(p)
	switch pf := argp.fmt.(type) {
	case *V4L2_Pix_Format:
		pf.set(unsafe.Pointer(&vf.fmt))
	case *V4L2_Pix_Format_Mplane:
		pf.set(unsafe.Pointer(&vf.fmt))
	default:
		log.Fatalf("Unexpected type %T", pf)
	}
	err := ioctl(fd, VIDIOC_TRY_FMT, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Control struct {
	ID    uint32
	Value int32
}

func (c *V4L2_Control) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_control)(ptr)
	p.id = C.__u32(c.ID)
	p.value = C.__s32(c.Value)
}

func (c *V4L2_Control) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_control)(ptr)
	c.ID = uint32(p.id)
	c.Value = int32(p.value)
}

func IoctlGetCtrl(fd int, argp *V4L2_Control) error {
	var vc C.struct_v4l2_control
	p := unsafe.Pointer(&vc)
	argp.set(p)
	err := ioctl(fd, VIDIOC_G_CTRL, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

func IoctlSetCtrl(fd int, argp *V4L2_Control) error {
	var vc C.struct_v4l2_control
	p := unsafe.Pointer(&vc)
	argp.set(p)
	err := ioctl(fd, VIDIOC_S_CTRL, p)
	if err != nil {
		return err
	}
	return nil
}
