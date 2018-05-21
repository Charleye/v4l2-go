package v4l2

import (
	"log"
	"syscall"
)

type Camera struct {
	Device
	Port
	Width             uint32
	Height            uint32
	PixelFormat       uint32
	PixFmtDescription string
}

func (c *Camera) VerifyCaps() {
	var caps V4L2_Capability
	err := IoctlQueryCap(c.FD, &caps)
	if err != nil {
		log.Fatal("Failed to query capability")
	}
	if caps.Capabilities&V4L2_CAP_VIDEO_CAPTURE == 0 {
		log.Fatal("The device not support video capture")
	}
}

func (c *Camera) SetFormat() {
	if c.Width == 0 || c.Height == 0 {
		log.Fatal("Not configure width or height in pixel")
	}
	if c.PixelFormat == 0 && c.PixFmtDescription == "" {
		log.Fatal("Not assign pixel format")
	}
	if c.PixelFormat > 0 && c.PixFmtDescription != "" {
		if GetFourCCByName(c.PixFmtDescription) != c.PixelFormat {
			log.Fatal("Inconsistent in pixel format")
		}
	}

	var format V4L2_Format
	var pixfmt V4L2_Pix_Format
	pixfmt.Width = c.Width
	pixfmt.Height = c.Height
	if c.PixelFormat > 0 {
		pixfmt.PixelFormat = c.PixelFormat
	} else if c.PixFmtDescription != "" {
		pixfmt.PixelFormat = GetFourCCByName(c.PixFmtDescription)
	}
	pixfmt.Priv = 0
	format.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	format.Fmt = &pixfmt
	err := IoctlSetFmt(c.FD, &format)
	if err != nil {
		log.Fatal("Failed to set format")
	}
}

func (c *Camera) AllocBuffers(count uint32) {
	var reqbufs V4L2_Requestbuffers
	reqbufs.Count = count
	reqbufs.Memory = V4L2_MEMORY_MMAP
	reqbufs.Type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err := IoctlRequestBuffers(c.FD, &reqbufs)
	if err != nil {
		log.Fatal("Failed to request buffers: ", err)
	}
	if reqbufs.Count == 0 {
		log.Fatal("Out of memory")
	}

	var bufs Buffers
	bufs.Count = reqbufs.Count
	bufs.NPlanes = 1
	c.Type = reqbufs.Memory
	c.Bufs = &bufs

	data := make([][]byte, 0, bufs.Count)
	for i := 0; i < int(bufs.Count); i++ {
		vb := V4L2_Buffer{
			Index:  uint32(i),
			Type:   V4L2_BUF_TYPE_VIDEO_CAPTURE,
			Memory: c.Type,
		}
		if err := IoctlQueryBuf(c.FD, &vb); err != nil {
			log.Fatal("Failed to query buffers: ", err)
		}
		var offset uint32
		GetValueFromUnion(vb.M, &offset)
		buf, err := syscall.Mmap(c.FD, int64(offset), int(vb.Length),
			syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
		if err != nil {
			log.Fatal("Failed to mmap: ", err)
		}
		data = append(data, buf)
		if err := IoctlQBuf(c.FD, &vb); err != nil {
			log.Fatal("Failed ro enqueue buffer: ", err)
		}
	}
	bufs.Data = data
}

func (c *Camera) TurnOn() {
	var stream int = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err := IoctlStreamOn(c.FD, &stream)
	if err != nil {
		log.Fatal("Failed to stream on: ", err)
	}
}

func (c *Camera) TurnOff() {
	var stream int = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err := IoctlStreamOff(c.FD, &stream)
	if err != nil {
		log.Fatal("Failed to stream off: ", err)
	}

	for _, v := range c.Bufs.Data {
		err := syscall.Munmap(v)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (c *Camera) Capture() []byte {
	vb := V4L2_Buffer{
		Type:   V4L2_BUF_TYPE_VIDEO_CAPTURE,
		Memory: c.Type,
	}
	if err := IoctlDQBuf(c.FD, &vb); err != nil {
		log.Fatal("Failed to dequeue buffer: ", err)
	}
	c.Bufs.Data[vb.Index] = c.Bufs.Data[vb.Index][:vb.Length]
	data := c.Bufs.Data[vb.Index][:vb.BytesUsed]
	if err := IoctlQBuf(c.FD, &vb); err != nil {
		log.Fatal("Failed to enqueue buffer: ", err)
	}
	return data
}
