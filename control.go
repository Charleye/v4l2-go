package v4l2

/*
#include <linux/v4l2-controls.h>
*/
import "C"

const (
	/* User-class control IDs */
	V4L2_CID_BRIGHTNESS           = C.V4L2_CID_BRIGHTNESS
	V4L2_CID_HUE                  = C.V4L2_CID_HUE
	V4L2_CID_EXPOSURE             = C.V4L2_CID_EXPOSURE
	V4L2_CID_AUTOBRIGHTNESS       = C.V4L2_CID_AUTOBRIGHTNESS
	V4L2_CID_POWER_LINE_FREQUENCY = C.V4L2_CID_POWER_LINE_FREQUENCY
)
