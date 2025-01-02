# video-ocr

## Description

Using OpenCV, open the first capture device and run OCR on the input.

## Installation

```
$ ./00install.bash
```

## Installing dependencies

### GoCV (Golang OpenCV binding)

- Get from https://github.com/hybridgroup/gocv
- Follow the install instructions in the README
- As of 2024-12-30, I had to make the following change to `Makefile` on Ubuntu 24:
  - `libtbb2` doesn't exist, so just remove it.
  - Change `libdc1394-22-dev` to `libdc1394-dev`

```
diff --git a/Makefile b/Makefile
index ccc035d..bcf4bbb 100644
--- a/Makefile
+++ b/Makefile
@@ -18,7 +18,7 @@ BUILD_SHARED_LIBS?=ON
 
 # Package list for each well-known Linux distribution
 RPMS=cmake curl wget git gtk2-devel libpng-devel libjpeg-devel libtiff-devel tbb tbb-devel libdc1394-devel unzip gcc-c++
-DEBS=unzip wget build-essential cmake curl git libgtk2.0-dev pkg-config libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev libjpeg-dev libpng-dev libtiff-dev libdc1394-22-dev libharfbuzz-dev libfreetype6-dev
+DEBS=unzip wget build-essential cmake curl git libgtk2.0-dev pkg-config libavcodec-dev libavformat-dev libswscale-dev libtbb-dev libjpeg-dev libpng-dev libtiff-dev libdc1394-dev libharfbuzz-dev libfreetype6-dev
 DEBS_BOOKWORM=unzip wget build-essential cmake curl git libgtk2.0-dev pkg-config libavcodec-dev libavformat-dev libswscale-dev libtbbmalloc2 libtbb-dev libjpeg-dev libpng-dev libtiff-dev libharfbuzz-dev libfreetype6-dev
 DEBS_UBUNTU_JAMMY=unzip wget build-essential cmake curl git libgtk2.0-dev pkg-config libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev libjpeg-dev libpng-dev libtiff-dev libdc1394-dev libharfbuzz-dev libfreetype6-dev
 JETSON=build-essential cmake git unzip pkg-config libjpeg-dev libpng-dev libtiff-dev libavcodec-dev libavformat-dev libswscale-dev libgtk2.0-dev libcanberra-gtk* libxvidcore-dev libx264-dev libgtk-3-dev libtbb2 libtbb-dev libdc1394-22-dev libv4l-dev v4l-utils libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev libavresample-dev libvorbis-dev libxine2-dev libfaac-dev libmp3lame-dev libtheora-dev libopencore-amrnb-dev libopencore-amrwb-dev libopenblas-dev libatlas-base-dev libblas-dev liblapack-dev libeigen3-dev gfortran libhdf5-dev protobuf-compiler libprotobuf-dev libgoogle-glog-dev libgflags-dev
```

### gosseract (OCR using Tesseract)

- Install `tesseract`: https://github.com/tesseract-ocr/tessdoc/blob/main/Installation.md
  - `sudo apt install tesseract-ocr libtesseract-dev`
  - `sudo apt install -y tesseract-ocr-eng tesseract-ocr-jpn`
