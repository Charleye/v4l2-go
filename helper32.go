// +build arm i386

package v4l2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

// get value from v4l2_buffer union field
func GetValueFromUnion(union []byte, value interface{}) {
	tmp := bytes.NewReader(union)
	switch x := value.(type) {
	case *uint32: // offset
		err := binary.Read(tmp, binary.LittleEndian, x)
		if err != nil {
			goto BinaryError
		}
	case *int: // fd
		var m uint32
		err := binary.Read(tmp, binary.LittleEndian, &m)
		if err != nil {
			goto BinaryError
		}
		*x = int(m)
	case *uintptr:
		var m uint32
		err := binary.Read(tmp, binary.LittleEndian, &m)
		if err != nil {
			goto BinaryError
		}
		*x = uintptr(m)
	}
	return
BinaryError:
	fmt.Printf("Read for package binary failed\n")
}

func UintptrToBytes(n uintptr) []byte {
	tmp := uint32(n)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.LittleEndian, tmp)
	return buffer.Bytes()
}

func BytesToUintptr(b []byte) uintptr {
	buffer := bytes.NewBuffer(b)
	var tmp uint32
	binary.Read(buffer, binary.LittleEndian, &tmp)
	return uintptr(tmp)
}

func PointerToBytes(p interface{}) []byte {
	switch x := p.(type) {
	case *V4L2_Plane:
		tmp := unsafe.Pointer(x)
		return UintptrToBytes(uintptr(tmp))
	}
	return nil
}
