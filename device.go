package v4l2

import (
	"syscall"
)

func Open(name string) (d *Device, err error) {
	fd, err := syscall.Open(name, syscall.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var st syscall.Stat_t
	err = syscall.Fstat(fd, &st)
	if err != nil {
		return nil, err
	}
	if st.Mode&syscall.S_IFCHR == 0 || st.Rdev>>8 != 81 {
		err = ErrorWrongDevice
		return nil, err
	}
	d = &Device{
		FD:   fd,
		Path: name,
	}
	return d, nil
}

func (d *Device) Open() (err error) {
	if d.Path == "" {
		err = ErrorNotSpecified
		return
	}
	tmp, err := Open(d.Path)
	if err != nil {
		return err
	}
	d.FD = tmp.FD
	return nil
}

func (d *Device) Close() {
	syscall.Close(d.FD)
	d.Path = ""
	d.FD = -1
}

type Device struct {
	Path string
	FD   int
}

type Buffers struct {
	Count     uint32
	NPlanes   uint32
	Data      [][]byte
	BytesUsed []uint
}

type Port struct {
	Type    uint32 // stream type
	State   uint32
	Counter uint32 /* total number of dequeued buffers */
	NBufs   uint32 // number of buffers in queue
	Bufs    *Buffers
}

type DeviceOps interface {
	Read() (n int, err error)
	Write() (n int, err error)
	RequestBuffers()
	DequeueBuufer()
	EnqueueBuffer()
	DequeueEvent()
	Destroy()
}
