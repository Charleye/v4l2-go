package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/Charleye/v4l2-go"
)

var device = flag.String("d", "/dev/video0", "video device")
var cid = flag.Int("c", 10029671, "control ID")

func main() {
	flag.Parse()
	fd, err := syscall.Open(*device, syscall.O_RDWR, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(fd)

	var control v4l2.V4L2_Control
	control.ID = uint32(*cid)
	err = v4l2.IoctlGetCtrl(fd, &control)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Before:")
	fmt.Println("ID: ", control.ID)
	fmt.Println("Value: ", control.Value)

	var ctrls v4l2.V4L2_Ext_Controls
	var ctrl v4l2.V4L2_Ext_Control
	ctrl.ID = uint32(*cid)
	ctrl.Union = int32(1)
	ctrls.Controls = append(ctrls.Controls, ctrl)

	ctrls.ClassWhich = v4l2.V4L2_CTRL_CLASS_MPEG
	ctrls.Count = 1

	err = v4l2.IoctlSetExtCtrls(fd, &ctrls)
	if err != nil {
		log.Fatal(err)
	}

	control.Value = 0
	err = v4l2.IoctlGetCtrl(fd, &control)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("After:")
	fmt.Println("ID: ", control.ID)
	fmt.Println("Value: ", control.Value)
}
