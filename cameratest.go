// GoCV test.
// Open a window and start WebCam recording.
// Run with: go run ./cameratest.go

package main

import (
	"flag"
	"fmt"
	"os"

	"gocv.io/x/gocv"
)

var (
	sourceIndex = flag.Int("d", 0, "Index of capture device to use")
	sourceFile  = flag.String("input", "", "Input device file")
	width       = flag.Int("w", 1920, "Width of the video capture")
	height      = flag.Int("h", 1080, "Height of the video capture")
	fps         = flag.Int("fps", 1, "Capture FPS")
)

func main() {
	flag.Parse()

	var input any
	if *sourceFile != "" {
		input = *sourceFile
	} else {
		input = *sourceIndex
	}

	webcam, err := gocv.OpenVideoCapture(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening video capture device: %v\n", err)
		os.Exit(1)
	}

	webcam.Set(gocv.VideoCaptureFrameWidth, float64(*width))
	webcam.Set(gocv.VideoCaptureFrameHeight, float64(*height))

	for f := *fps; f > 0; f-- {
		// fmt.Printf("Trying FPS: %v\n", f)
		webcam.Set(gocv.VideoCaptureFPS, float64(f))
		if int(webcam.Get(gocv.VideoCaptureFPS)) == f {
			break
		}
	}

	fmt.Printf("Frame Width: %v\n", webcam.Get(gocv.VideoCaptureFrameWidth))
	fmt.Printf("Frame Height: %v\n", webcam.Get(gocv.VideoCaptureFrameHeight))
	fmt.Printf("FPS: %v\n", webcam.Get(gocv.VideoCaptureFPS))
	fmt.Printf("Buffer Size: %v\n", webcam.Get(gocv.VideoCaptureBufferSize))
	fmt.Printf("Brightness: %v\n", webcam.Get(gocv.VideoCaptureBrightness))
	fmt.Printf("Contrast: %v\n", webcam.Get(gocv.VideoCaptureContrast))
	fmt.Printf("Saturation: %v\n", webcam.Get(gocv.VideoCaptureSaturation))
	fmt.Printf("Hue: %v\n", webcam.Get(gocv.VideoCaptureHue))
	fmt.Printf("Gain: %v\n", webcam.Get(gocv.VideoCaptureGain))
	fmt.Printf("Exposure: %v\n", webcam.Get(gocv.VideoCaptureExposure))
	fmt.Printf("Auto Exposure: %v\n", webcam.Get(gocv.VideoCaptureAutoExposure))
	fmt.Printf("Gamma: %v\n", webcam.Get(gocv.VideoCaptureGamma))
	fmt.Printf("Sharpness: %v\n", webcam.Get(gocv.VideoCaptureSharpness))
	fmt.Printf("Backlight Compensation: %v\n", webcam.Get(gocv.VideoCaptureBacklight))
	fmt.Printf("Focus: %v\n", webcam.Get(gocv.VideoCaptureFocus))
	fmt.Printf("Zoom: %v\n", webcam.Get(gocv.VideoCaptureZoom))
	fmt.Printf("ISO Speed: %v\n", webcam.Get(gocv.VideoCaptureISOSpeed))
	fmt.Printf("Temperature: %v\n", webcam.Get(gocv.VideoCaptureTemperature))
	fmt.Printf("HW Acceleration: %v\n", webcam.Get(gocv.VideoCaptureHWAcceleration))
	fmt.Printf("HW Device: %v\n", webcam.Get(gocv.VideoCaptureHWDevice))

	window := gocv.NewWindow("Hello")
	img := gocv.NewMat()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Fprintf(os.Stderr, "ERROR: Unable to read frame.\n")
			continue
		}
		window.IMShow(img)
		window.WaitKey(1)
	}
}
