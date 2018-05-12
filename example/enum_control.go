package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"syscall"
)

var device = flag.String("d", "/dev/video11", "video device")

func main() {
	flag.Parse()
	fd, err := syscall.Open(*device, syscall.O_RDWR, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(fd)
	EnumerateAllCtrl(fd)
}

func EnumerateAllCtrl(fd int) {
	var vc V4L2_Queryctrl
	vc.ID = V4L2_CTRL_FLAG_NEXT_CTRL
	for {
		err := IoctlQueryCtrl(fd, &vc)
		if err != nil {
			if err != syscall.EINVAL {
				log.Fatal(err)
			}
			break
		}
		if vc.Flags&V4L2_CTRL_FLAG_DISABLED != 0 {
			continue
		}
		fmt.Println(vc)

		switch vc.Type {
		case V4L2_CTRL_TYPE_MENU:
			EnumerateMenu(fd, vc.ID, vc.Minimum, vc.Maximum)
		case V4L2_CTRL_TYPE_INTEGER_MENU:
			EnumerateIntergerMenu(fd, vc.ID, vc.Minimum, vc.Maximum)
		}

		vc.ID |= V4L2_CTRL_FLAG_NEXT_CTRL
	}
}

func EnumerateMenu(fd int, id uint32, min, max int32) {
	var vm V4L2_Querymenu
	vm.ID = id
	for i := min; i <= max; i++ {
		vm.Index = uint32(i)
		err := IoctlQueryMenu(fd, &vm)
		if err == nil {
			fmt.Println("\t", string(vm.union))
		}
	}
}

func EnumerateIntergerMenu(fd int, id uint32, min, max int32) {
	var vm V4L2_Querymenu
	vm.ID = id
	for i := min; i <= max; i++ {
		vm.Index = uint32(i)
		err := IoctlQueryMenu(fd, &vm)
		if err == nil {
			var value int64
			b := vm.union[:8]
			buf := bytes.NewReader(b)
			err := binary.Read(buf, binary.LittleEndian, &value)
			if err != nil {
				fmt.Println("binary.Read failed: ", err)
			}
			fmt.Println("\t", value)
		}
	}
}
