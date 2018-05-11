// Package v4l2, a facade to the Video4Linux2 video capture interface
// Copyright (C) 2016 Zoltán Korándi <korandi.z@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

#include <stdio.h>
#include <stdlib.h>
#include <stddef.h>
#include <linux/videodev2.h>

int main() {
	printf("// +build linux\n");
	printf("// +build ");
	fflush(stdout);
	system("go env GOHOSTARCH");
	printf("\n");

	printf("/////////////////////////////////////////////////////\n");
	printf("//                                                 //\n");
	printf("//  !!! THIS IS A GENERATED FILE, DO NOT EDIT !!!  //\n");
	printf("//                                                 //\n");
	printf("/////////////////////////////////////////////////////\n");
	printf("\n");

	printf("package main\n\n");

    printf("const (\n");
    printf("\t__SIZEOF_POINTER__    =  %llu\n", (long long unsigned)__SIZEOF_POINTER__);
    printf(")\n\n");

	printf("const (\n");
	printf("\toffset_format_type                = %llu\n", (long long unsigned) offsetof(struct v4l2_format, type));
	printf("\toffset_streamparm_type            = %llu\n", (long long unsigned) offsetof(struct v4l2_streamparm, type));
	printf("\toffset_requestbuffers_type        = %llu\n", (long long unsigned) offsetof(struct v4l2_requestbuffers, type));
	printf("\toffset_buffer_type                = %llu\n", (long long unsigned) offsetof(struct v4l2_buffer, type));
	printf("\toffset_cropcap_type               = %llu\n", (long long unsigned) offsetof(struct v4l2_cropcap, type));
	printf("\toffset_crop_type                  = %llu\n", (long long unsigned) offsetof(struct v4l2_crop, type));
	printf("\toffset_fmtdesc_type               = %llu\n", (long long unsigned) offsetof(struct v4l2_fmtdesc, type));
	printf("\toffset_frmsizeenum_type           = %llu\n", (long long unsigned) offsetof(struct v4l2_frmsizeenum, type));
	printf("\toffset_frmivalenum_type           = %llu\n", (long long unsigned) offsetof(struct v4l2_frmivalenum, type));
	printf("\toffset_queryctrl_type             = %llu\n", (long long unsigned) offsetof(struct v4l2_queryctrl, type));
	printf("\toffset_event_subscription_type    = %llu\n", (long long unsigned) offsetof(struct v4l2_event_subscription, type));
	printf("\toffset_event_type                 = %llu\n", (long long unsigned) offsetof(struct v4l2_event, type));
	printf("\toffset_querymenu_union            = %llu\n", (long long unsigned) offsetof(struct v4l2_querymenu, name));
	printf("\toffset_input_type                 = %llu\n", (long long unsigned) offsetof(struct v4l2_input, type));
	printf("\toffset_output_type                = %llu\n", (long long unsigned) offsetof(struct v4l2_output, type));
	printf("\toffset_selection_type             = %llu\n", (long long unsigned) offsetof(struct v4l2_selection, type));
    printf("\toffset_timecode_type              = %llu\n", (long long unsigned) offsetof(struct v4l2_timecode, type));
    printf("\toffset_pix_format_encoding        = %llu\n", (long long unsigned) offsetof(struct v4l2_pix_format, ycbcr_enc));
    printf("\toffset_pix_format_mplane_encoding = %llu\n", (long long unsigned) offsetof(struct v4l2_pix_format_mplane, ycbcr_enc));
	printf(")\n\n");

	return 0;
}
