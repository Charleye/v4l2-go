package v4l2

/*
#include <linux/videodev2.h>
*/
import "C"

import (
	"fmt"
	"log"
	"syscall"
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
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_fmtdesc_type))
	*tmp = C.__u32(f.Type)
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
	Fmt  interface{}
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
	Encoding     uint32
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
	Encoding     uint8
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

	// v4l2_pix_format_mplane in videodev2.h is a little difference between linux-4.4.0 and linux-4.14.0
	tmp := (*C.__u8)(unsafe.Pointer(
		uintptr(ptr) + offset_pix_format_mplane_encoding))
	*tmp = C.__u8(f.Encoding)

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

	// v4l2_pix_format_mplane in videodev2.h is a little difference between linux-4.4.0 and linux-4.14.0
	tmp := (*C.__u8)(unsafe.Pointer(
		uintptr(ptr) + offset_pix_format_mplane_encoding))
	f.Encoding = uint8(*tmp)

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

	// v4l2_pix_format in videodev2.h is a little difference between linux-4.4.0 and linux-4.14.0
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_pix_format_encoding))
	*tmp = C.__u32(f.Encoding)

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

	// v4l2_pix_format in videodev2.h is a little difference between linux-4.4.0 and linux-4.14.0
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_pix_format_encoding))
	f.Encoding = uint32(*tmp)

	f.Quantization = uint32(p.quantization)
	f.XferFunc = uint32(p.xfer_func)
}

func (f *V4L2_Format) set(ptr unsafe.Pointer) {
	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_format_type))
	*tmp = C.__u32(f.Type)
}

func (f *V4L2_Format) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_format)(ptr)

	switch pf := f.Fmt.(type) {
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
	switch pf := argp.Fmt.(type) {
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
	switch pf := argp.Fmt.(type) {
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

type V4L2_Queryctrl struct {
	ID           uint32
	Type         uint32
	Name         string
	Minimum      int32
	Maximum      int32
	Step         int32
	DefaultValue int32
	Flags        uint32
}

func (c *V4L2_Queryctrl) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_queryctrl)(ptr)
	p.id = C.__u32(c.ID)
}

func (c *V4L2_Queryctrl) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_queryctrl)(ptr)
	c.ID = uint32(p.id)

	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_queryctrl_type))
	c.Type = uint32(*tmp)

	c.Name = C.GoString((*C.char)(unsafe.Pointer(&p.name[0])))
	c.Minimum = int32(p.minimum)
	c.Maximum = int32(p.maximum)
	c.Step = int32(p.step)
	c.DefaultValue = int32(p.default_value)
	c.Flags = uint32(p.flags)
}

func IoctlQueryCtrl(fd int, argp *V4L2_Queryctrl) error {
	var qc C.struct_v4l2_queryctrl
	p := unsafe.Pointer(&qc)
	argp.set(p)
	err := ioctl(fd, VIDIOC_QUERYCTRL, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Querymenu struct {
	ID    uint32
	Index uint32
	Union []byte
}

func (m *V4L2_Querymenu) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_querymenu)(ptr)
	p.id = C.__u32(m.ID)
	p.index = C.__u32(m.Index)
}

func (m *V4L2_Querymenu) get(ptr unsafe.Pointer) {
	// due to anonymous union, cannot get it's field pointer
	p := unsafe.Pointer(uintptr(ptr) + offset_querymenu_union)
	m.Union = C.GoBytes(p, 32)
}

func IoctlQueryMenu(fd int, argp *V4L2_Querymenu) error {
	var vm C.struct_v4l2_querymenu
	p := unsafe.Pointer(&vm)
	argp.set(p)
	err := ioctl(fd, VIDIOC_QUERYMENU, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Query_Ext_Ctrl struct {
	ID           uint32
	Type         uint32
	Name         string
	Minimum      int64
	Maximum      int64
	Step         uint64
	DefaultValue int64
	Flags        uint32
	ElemSize     uint32
	Elems        uint32
	NrOfDims     uint32
	Dims         [V4L2_CTRL_MAX_DIMS]uint32
}

func (c *V4L2_Query_Ext_Ctrl) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_query_ext_ctrl)(ptr)
	p.id = C.__u32(c.ID)
}

func (c *V4L2_Query_Ext_Ctrl) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_query_ext_ctrl)(ptr)

	tmp := (*C.__u32)(unsafe.Pointer(uintptr(ptr) + offset_query_ext_ctrl_type))
	c.Type = uint32(*tmp)

	c.Name = C.GoString((*C.char)(&p.name[0]))
	c.Minimum = int64(p.minimum)
	c.Maximum = int64(p.maximum)
	c.Step = uint64(p.step)
	c.DefaultValue = int64(p.default_value)
	c.Flags = uint32(p.flags)
	c.ElemSize = uint32(p.elem_size)
	c.Elems = uint32(p.elems)
	c.NrOfDims = uint32(p.nr_of_dims)
	data := (*[V4L2_CTRL_MAX_DIMS]uint32)(unsafe.Pointer(&p.dims[0]))
	c.Dims = *data
}

func IoctlQueryExtCtrl(fd int, argp *V4L2_Query_Ext_Ctrl) error {
	var ctrl C.struct_v4l2_query_ext_ctrl
	p := unsafe.Pointer(&ctrl)
	argp.set(p)
	err := ioctl(fd, VIDIOC_QUERY_EXT_CTRL, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Crop struct {
	Type uint32
	C    V4L2_Rect
}

type V4L2_Rect struct {
	Left   int32
	Top    int32
	Width  uint32
	Height uint32
}

func (r *V4L2_Rect) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_rect)(ptr)
	p.left = C.__s32(r.Left)
	p.top = C.__s32(r.Top)
	p.width = C.__u32(r.Width)
	p.height = C.__u32(r.Height)
}

func (r *V4L2_Rect) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_rect)(ptr)
	r.Left = int32(p.left)
	r.Top = int32(p.top)
	r.Width = uint32(p.width)
	r.Height = uint32(p.height)
}

func (c *V4L2_Crop) set(ptr unsafe.Pointer) {
	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_crop_type))
	*tmp = C.__u32(c.Type)
}

func IoctlGetCrop(fd int, argp *V4L2_Crop) error {
	var vc C.struct_v4l2_crop
	p := unsafe.Pointer(&vc)
	argp.set(p)
	err := ioctl(fd, VIDIOC_G_CROP, p)
	if err != nil {
		return err
	}
	argp.C.get(unsafe.Pointer(&vc.c))
	return nil
}

func IoctlSetCrop(fd int, argp *V4L2_Crop) error {
	var vc C.struct_v4l2_crop
	p := unsafe.Pointer(&vc)
	argp.set(p)
	argp.C.set(unsafe.Pointer(&vc.c))
	err := ioctl(fd, VIDIOC_S_CROP, p)
	if err != nil {
		return err
	}
	return nil
}

type V4L2_Cropcap struct {
	Type        uint32
	Bounds      V4L2_Rect
	Defrect     V4L2_Rect
	PixelAspect V4L2_Fract
}

type V4L2_Fract struct {
	Numerator   uint32
	Denominator uint32
}

func (f *V4L2_Fract) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_fract)(ptr)
	p.numerator = C.__u32(f.Numerator)
	p.denominator = C.__u32(f.Denominator)
}

func (f *V4L2_Fract) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_fract)(ptr)
	f.Numerator = uint32(p.numerator)
	f.Denominator = uint32(p.denominator)
}

func (c *V4L2_Cropcap) set(ptr unsafe.Pointer) {
	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_crop_type))
	*tmp = C.__u32(c.Type)
}

func IoctlCropCap(fd int, argp *V4L2_Cropcap) error {
	var cc C.struct_v4l2_cropcap
	p := unsafe.Pointer(&cc)
	argp.set(p)
	err := ioctl(fd, VIDIOC_CROPCAP, p)
	if err != nil {
		return err
	}
	argp.Bounds.get(unsafe.Pointer(&cc.bounds))
	argp.Defrect.get(unsafe.Pointer(&cc.defrect))
	argp.PixelAspect.get(unsafe.Pointer(&cc.pixelaspect))
	return nil
}

type V4L2_Buffer struct {
	Index     uint32
	Type      uint32
	BytesUsed uint32
	Flags     uint32
	Field     uint32
	TimeStamp syscall.Timeval
	TimeCode  V4L2_Timecode
	Sequence  uint32
	Memory    uint32
	M         []byte
	Length    uint32
}

type V4L2_Timecode struct {
	Type     uint32
	Flags    uint32
	Frames   uint8
	Seconds  uint8
	Minutes  uint8
	Hours    uint8
	UserBits [4]uint8
}

type V4L2_Plane struct {
	BytesUsed  uint32
	Length     uint32
	Union      []byte
	DataOffset uint32
}

func (v *V4L2_Plane) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_plane)(ptr)
	v.BytesUsed = uint32(p.bytesused)
	v.Length = uint32(p.length)
	v.Union = C.GoBytes(unsafe.Pointer(&p.m), __SIZEOF_POINTER__)
	v.DataOffset = uint32(p.data_offset)
}

func (t *V4L2_Timecode) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_timecode)(ptr)

	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_timecode_type))
	t.Type = uint32(*tmp)

	t.Flags = uint32(p.flags)
	t.Frames = uint8(p.frames)
	t.Seconds = uint8(p.seconds)
	t.Minutes = uint8(p.minutes)
	t.Hours = uint8(p.hours)
	data := (*[4]uint8)(unsafe.Pointer(&p.userbits[0]))
	t.UserBits = *data
}

func (b *V4L2_Buffer) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_buffer)(ptr)
	p.index = C.__u32(b.Index)

	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_buffer_type))
	*tmp = C.__u32(b.Type)

	p.bytesused = C.__u32(b.BytesUsed)
	p.flags = C.__u32(b.Flags)
	p.field = C.__u32(b.Field)
	p.memory = C.__u32(b.Memory)
	p.length = C.__u32(b.Length)
}

func (b *V4L2_Buffer) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_buffer)(ptr)
	b.Index = uint32(p.index)

	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_buffer_type))
	b.Type = uint32(*tmp)

	b.BytesUsed = uint32(p.bytesused)
	b.Flags = uint32(p.flags)
	b.Field = uint32(p.field)

	t := (*syscall.Timeval)(unsafe.Pointer(&p.timestamp))
	b.TimeStamp = *t
	b.TimeCode.get(unsafe.Pointer(&p.timecode))
	b.Sequence = uint32(p.sequence)
	b.Memory = uint32(p.memory)
	b.M = C.GoBytes(unsafe.Pointer(&p.m), __SIZEOF_POINTER__)
	b.Length = uint32(p.length)
}

func IoctlQueryBuf(fd int, argp *V4L2_Buffer) error {
	var vb C.struct_v4l2_buffer
	var planes [VIDEO_MAX_PLANES]C.struct_v4l2_plane

	if argp.Type == V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE ||
		argp.Type == V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE {
		b := UintptrToBytes(uintptr(unsafe.Pointer(&planes[0])))
		for i := 0; i < __SIZEOF_POINTER__; i++ {
			vb.m[i] = b[i]
		}
	}
	p := unsafe.Pointer(&vb)
	argp.set(p)
	err := ioctl(fd, VIDIOC_QUERYBUF, p)
	if err != nil {
		return err
	}

	if argp.Type == V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE ||
		argp.Type == V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE {
		tmp := unsafe.Pointer(BytesToUintptr(argp.M))
		p := (*[VIDEO_MAX_PLANES]V4L2_Plane)(tmp)
		for i := 0; i < int(vb.length); i++ {
			(*p)[i].get(unsafe.Pointer(&planes[i]))
		}
	}
	argp.get(p)
	return nil
}

func IoctlQBuf(fd int, argp *V4L2_Buffer) error {
	var vb C.struct_v4l2_buffer
	p := unsafe.Pointer(&vb)
	argp.set(p)
	err := ioctl(fd, VIDIOC_QBUF, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

func IoctlDQBuf(fd int, argp *V4L2_Buffer) error {
	var vb C.struct_v4l2_buffer
	p := unsafe.Pointer(&vb)
	argp.set(p)
	err := ioctl(fd, VIDIOC_DQBUF, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Requestbuffers struct {
	Count  uint32
	Type   uint32
	Memory uint32
}

func (b *V4L2_Requestbuffers) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_requestbuffers)(ptr)
	p.count = C.__u32(b.Count)
	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_requestbuffers_type))
	*tmp = C.__u32(b.Type)
	p.memory = C.__u32(b.Memory)
}

func (b *V4L2_Requestbuffers) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_requestbuffers)(ptr)
	b.Count = uint32(p.count)
}

func IoctlRequestBuffers(fd int, argp *V4L2_Requestbuffers) error {
	var rb C.struct_v4l2_requestbuffers
	p := unsafe.Pointer(&rb)
	argp.set(p)
	err := ioctl(fd, VIDIOC_REQBUFS, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Streamparm struct {
	Type uint32
	Parm interface{}
}

type V4L2_Captureparm struct {
	Capability   uint32
	CaptureMode  uint32
	TimePerFrame V4L2_Fract
	ExtendedMode uint32
	ReadBuffers  uint32
}

type V4L2_Outputparm struct {
	Capability   uint32
	OutputMode   uint32
	TimePerFrame V4L2_Fract
	ExtendedMode uint32
	WriteBuffers uint32
}

func (o *V4L2_Outputparm) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_outputparm)(ptr)
	p.outputmode = C.__u32(o.OutputMode)
	o.TimePerFrame.set(unsafe.Pointer(&p.timeperframe))
	p.writebuffers = C.__u32(o.WriteBuffers)
}

func (o *V4L2_Outputparm) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_outputparm)(ptr)
	o.Capability = uint32(p.capability)
	o.OutputMode = uint32(p.outputmode)
	o.TimePerFrame.get(unsafe.Pointer(&p.timeperframe))
	o.ExtendedMode = uint32(p.extendedmode)
	o.WriteBuffers = uint32(p.writebuffers)
}

func (c *V4L2_Captureparm) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_captureparm)(ptr)
	p.capturemode = C.__u32(c.CaptureMode)
	c.TimePerFrame.set(unsafe.Pointer(&p.timeperframe))
	p.readbuffers = C.__u32(c.ReadBuffers)
}

func (c *V4L2_Captureparm) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_captureparm)(ptr)
	c.Capability = uint32(p.capability)
	c.CaptureMode = uint32(p.capturemode)
	c.TimePerFrame.get(unsafe.Pointer(&p.timeperframe))
	c.ExtendedMode = uint32(p.extendedmode)
	c.ReadBuffers = uint32(p.readbuffers)
}

func (s *V4L2_Streamparm) set(ptr unsafe.Pointer) {
	// due to type field, it is keyword in golang
	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_streamparm_type))
	*tmp = C.__u32(s.Type)
}

func (s *V4L2_Streamparm) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_streamparm)(ptr)

	switch s.Type {
	case V4L2_BUF_TYPE_VIDEO_CAPTURE,
		V4L2_BUF_TYPE_VIDEO_CAPTURE_MPLANE:
		cp := V4L2_Captureparm{}
		cp.get(unsafe.Pointer(&p.parm))
		s.Parm = &cp
	case V4L2_BUF_TYPE_VIDEO_OUTPUT,
		V4L2_BUF_TYPE_VIDEO_OUTPUT_MPLANE:
		op := V4L2_Outputparm{}
		op.get(unsafe.Pointer(&p.parm))
		s.Parm = &op
	default:
		log.Fatalf("Unexpected buffer type %v\n", s.Type)
	}
}

func IoctlGetParm(fd int, argp *V4L2_Streamparm) error {
	var sp C.struct_v4l2_streamparm
	p := unsafe.Pointer(&sp)
	argp.set(p)
	err := ioctl(fd, VIDIOC_G_PARM, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

func IoctlSetParm(fd int, argp *V4L2_Streamparm) error {
	var sp C.struct_v4l2_streamparm
	p := unsafe.Pointer(&sp)
	argp.set(p)
	switch parm := argp.Parm.(type) {
	case *V4L2_Captureparm:
		parm.set(unsafe.Pointer(&sp.parm))
	case *V4L2_Outputparm:
		parm.set(unsafe.Pointer(&sp.parm))
	default:
		panic(fmt.Sprintf("Unexpected type %T\n", parm))
	}
	err := ioctl(fd, VIDIOC_S_PARM, p)
	if err != nil {
		return err
	}
	return nil
}

type V4L2_Event_Subscription struct {
	Type  uint32
	ID    uint32
	Flags uint32
}

func (e *V4L2_Event_Subscription) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_event_subscription)(ptr)

	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_event_subscription_type))
	*tmp = C.__u32(e.Type)

	p.id = C.__u32(e.ID)
	p.flags = C.__u32(e.Flags)
}

func IoctlSubscribeEvent(fd int, argp *V4L2_Event_Subscription) error {
	var se C.struct_v4l2_event_subscription
	p := unsafe.Pointer(&se)
	argp.set(p)
	err := ioctl(fd, VIDIOC_SUBSCRIBE_EVENT, p)
	if err != nil {
		return err
	}
	return nil
}

func IoctlUnsubscribeEvent(fd int, argp *V4L2_Event_Subscription) error {
	var ue C.struct_v4l2_event_subscription
	p := unsafe.Pointer(&ue)
	argp.set(p)
	err := ioctl(fd, VIDIOC_UNSUBSCRIBE_EVENT, p)
	if err != nil {
		return err
	}
	return nil
}

type V4L2_Event struct {
	Type      uint32
	Union     interface{}
	Pending   uint32
	Sequence  uint32
	TimeStamp syscall.Timespec
	ID        uint32
}

func (e *V4L2_Event) get(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_event)(ptr)

	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_event_type))
	e.Type = uint32(*tmp)

	e.Pending = uint32(p.pending)
	e.Sequence = uint32(p.sequence)

	t := (*syscall.Timespec)(unsafe.Pointer(&p.timestamp))
	e.TimeStamp = *t

	e.ID = uint32(p.id)
}

func IoctlDQEvent(fd int, argp *V4L2_Event) error {
	var ve C.struct_v4l2_event
	p := unsafe.Pointer(&ve)
	err := ioctl(fd, VIDIOC_DQEVENT, p)
	if err != nil {
		return err
	}
	argp.get(p)
	return nil
}

type V4L2_Ext_Control struct {
	ID    uint32
	Size  uint32
	Union interface{}
}

type V4L2_Ext_Controls struct {
	ClassWhich uint32
	Count      uint32
	ErrorIdx   uint32
	Controls   []V4L2_Ext_Control
}

func (c *V4L2_Ext_Control) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_ext_control)(ptr)
	p.id = C.__u32(c.ID)
	p.size = C.__u32(c.Size)

	tmp := unsafe.Pointer(uintptr(ptr) + offset_ext_control_union)
	switch v := c.Union.(type) {
	case int32:
		*(*C.__s32)(tmp) = C.__s32(v)
	case int64:
		*(*C.__s64)(tmp) = C.__s64(v)
	case *string:
		*(**C.char)(tmp) = (*C.char)(unsafe.Pointer(v))
	case *uint8:
		*(**C.__u8)(tmp) = (*C.__u8)(unsafe.Pointer(v))
	case *uint16:
		*(**C.__u16)(tmp) = (*C.__u16)(unsafe.Pointer(v))
	case *uint32:
		*(**C.__u32)(tmp) = (*C.__u32)(unsafe.Pointer(v))
	case unsafe.Pointer:
		*(**C.void)(tmp) = (*C.void)(v)
	default:
		panic(fmt.Sprintf("Unexpected type %T", v))
	}
}

func (c *V4L2_Ext_Control) get(ptr unsafe.Pointer) {
}

func (c *V4L2_Ext_Controls) set(ptr unsafe.Pointer) {
	p := (*C.struct_v4l2_ext_controls)(ptr)

	tmp := (*C.__u32)(unsafe.Pointer(
		uintptr(ptr) + offset_ext_controls_ctrl_class))
	*tmp = C.__u32(c.ClassWhich)

	p.count = C.__u32(c.Count)
}

func (c *V4L2_Ext_Controls) get(ptr unsafe.Pointer) {
}

func IoctlSetExtCtrls(fd int, argp *V4L2_Ext_Controls) error {
	var ctrls C.struct_v4l2_ext_controls
	p := unsafe.Pointer(&ctrls)
	var ctrl []C.struct_v4l2_ext_control
	for i := 0; i < int(argp.Count); i++ {
		var c C.struct_v4l2_ext_control
		argp.Controls[i].set(unsafe.Pointer(&c))
		ctrl = append(ctrl, c)
	}
	argp.set(p)
	ctrls.controls = (*C.struct_v4l2_ext_control)(unsafe.Pointer(&ctrl[0]))
	err := ioctl(fd, VIDIOC_S_EXT_CTRLS, p)
	if err != nil {
		return err
	}
	return nil
}

func IoctlStreamOn(fd int, argp *int) error {
	var i C.int
	i = C.int(*argp)
	p := unsafe.Pointer(&i)
	err := ioctl(fd, VIDIOC_STREAMON, p)
	if err != nil {
		return err
	}
	return nil
}

func IoctlStreamOff(fd int, argp *int) error {
	var i C.int
	i = C.int(*argp)
	p := unsafe.Pointer(&i)
	err := ioctl(fd, VIDIOC_STREAMOFF, p)
	if err != nil {
		return err
	}
	return nil
}
