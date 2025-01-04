package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

var (
	sourceFile   = flag.String("f", "/dev/video0", "Input device file")
	ocrInterval  = flag.Int("i", 8, "Min interval for performing OCR")
	sleep        = flag.Int("s", 1, "Sleep millis between frames")
	languages    = flag.String("l", "eng", "Comma-separated list of languages")
	width        = flag.Int("w", 1920, "Width of the video capture")
	height       = flag.Int("h", 1080, "Height of the video capture")
	verbose      = flag.Bool("v", false, "Make verbose")
	fps          = flag.Int("fps", 30, "Capture FPS")
	ocrScale     = flag.Float64("q", 1, "Image scale for feeding OCR [0.1-1]")
	topMargin    = flag.Int("top-margin", 0, "Margin for feeding OCR")
	bottomMargin = flag.Int("bottom-margin", 0, "Margin for feeding OCR")
	leftMargin   = flag.Int("left-margin", 0, "Margin for feeding OCR")
	rightMargin  = flag.Int("right-margin", 0, "Margin for feeding OCR")
	margin       = flag.Int("margin", 0, "Margin for feeding OCR")
)

func toOutput(text string) string {
	return strings.ReplaceAll(text, "\n", " ")
}

var logMutex sync.Mutex

func logf(format string, a ...any) {
	logMutex.Lock()
	fmt.Fprintf(os.Stdout, format, a...)
	logMutex.Unlock()
}

func main() {
	// Define a custom usage function
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", "video-ocr")
		flag.PrintDefaults()
	}

	// Parse the flags
	flag.Parse()

	// Configuration
	showVideo := true // Display video window

	// Normalize arguments
	if *ocrScale < 0.1 {
		*ocrScale = 0.1
	} else if *ocrScale > 1 {
		*ocrScale = 1
	}

	if *margin > 0 {
		if *topMargin == 0 {
			*topMargin = *margin
		}
		if *bottomMargin == 0 {
			*bottomMargin = *margin
		}
		if *leftMargin == 0 {
			*leftMargin = *margin
		}
		if *rightMargin == 0 {
			*rightMargin = *margin
		}
	}

	// Open the video source
	webcam, err := gocv.OpenVideoCapture(*sourceFile)
	if err != nil {
		logf("Error opening video capture device: %v\n", err)
		return
	}
	defer webcam.Close()

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

	// Initialize Tesseract client
	client := gosseract.NewClient()
	defer client.Close()
	langs := strings.Split(*languages, ",")
	logf("# Languages: %v\n", langs)
	client.SetLanguage(langs...)

	// Create a window for display (if enabled)
	var window *gocv.Window
	if showVideo {
		window = gocv.NewWindow("Video with OCR")
		defer window.Close()
	}

	// Prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	frames := 0

	if *verbose {
		logf("# Started\n")
	}

	var ready atomic.Int32

	send := make(chan gocv.Mat)

	var readMillis atomic.Int32
	readMillis.Store(-1)

	ocrRect := image.Rect(*leftMargin, *topMargin, *width-*rightMargin, *height-*bottomMargin)

	reader := func() {
		if *verbose {
			logf("# Reader started\n")
		}
		for img := range send {
			readStart := time.Now()

			gray := gocv.NewMat()
			defer gray.Close()
			gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

			rect := image.Rect(*leftMargin, *topMargin, *width-*rightMargin, *height-*bottomMargin)
			gray = gray.Region(rect)
			gocv.Resize(gray, &gray, image.Point{}, *ocrScale, *ocrScale, gocv.InterpolationLinear)

			// Get image bytes
			imgBytes, err := gocv.IMEncode(gocv.PNGFileExt, gray) // Or use 'thresh' if you did preprocessing
			if err != nil {
				logf("# Error: encoding image: %v\n", err)
				return
			}

			if *verbose {
				logf("# Scanning the image...\n")
			}
			client.SetImageFromBytes(imgBytes.GetBytes())
			text, err := client.Text()
			if err != nil {
				logf("# Error: OCR failed: %v\n", err)
				return
			}

			readTime := time.Now().Sub(readStart)
			readMillis.Store(int32(readTime.Milliseconds()))

			if len(text) > 0 {
				logf("# Text: %s\n", toOutput(text))
			}
			ready.Add(-1)
		}
	}

	go reader()

	nextTick := time.Now().UnixNano() + 1e9
	readFps := 0

	for {
		if *verbose {
			logf(".")
		}
		time.Sleep(time.Duration(*sleep) * time.Millisecond)

		captureStart := time.Now()
		if ok := webcam.Read(&img); !ok {
			logf("# ERROR: Unable to read frame.\n")
			continue
		}
		captureTime := time.Now().Sub(captureStart)

		if img.Empty() {
			continue
		}

		frames++

		// // Perform OCR every ocrInterval frames
		if frames >= *ocrInterval && ready.Load() == 0 {
			send <- img
			ready.Add(1)
			frames = 0
		}

		if showVideo {
			// Draw a rectangle on the window
			// rect := image.Rect(*leftMargin, *topMargin, *width-*rightMargin, *height-*bottomMargin)
			gocv.Rectangle(&img, ocrRect, color.RGBA{0, 255, 0, 0}, 3)

			window.IMShow(img)

			window.WaitKey(1) // Handle events.
		}

		readFps++
		now := time.Now().UnixNano()
		if now >= nextTick {
			logf("# FPS: %d (last capture ms: %d, read ms: %d)\n", readFps, captureTime.Milliseconds(), readMillis.Load())
			nextTick = now + 1e9
			readFps = 0
		}

	}
}
