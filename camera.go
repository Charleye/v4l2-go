package v4l2

import "log"

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
		log.Fatal("Failed to request buffers")
	}
	if reqbufs.Count == 0 {
		log.Fatal("Out of memory")
	}

	var bufs Buffers
	bufs.Count = reqbufs.Count
	bufs.NPlanes = 1
	c.Type = reqbufs.Memory
	c.Data = &bufs

}
