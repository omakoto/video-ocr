package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/omakoto/go-common/src/common"
	"github.com/otiai10/gosseract/v2"
	"github.com/pborman/getopt/v2"
	"gocv.io/x/gocv"
)

var (
	ocrRegions = make([]image.Rectangle, 0)
)

type ocrRegionArg struct {
	x, y, w, h int
}

var _ getopt.Value = (*ocrRegionArg)(nil)

// Removed invalid type conversion
func (o *ocrRegionArg) String() string {
	return fmt.Sprintf("%d,%d,%d,%d", o.x, o.y, o.w, o.h)
}

func (o *ocrRegionArg) Set(value string, opt getopt.Option) error {
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return fmt.Errorf("Invalid region format: %v", value)
	}
	atoi := func(s string) (int, error) {
		x, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("Invalid region format: %v: %w", value, err)
		}
		return x, nil
	}

	x, err := atoi(parts[0])
	if err != nil {
		return err
	}
	y, err := atoi(parts[1])
	if err != nil {
		return err
	}
	w, err := atoi(parts[2])
	if err != nil {
		return err
	}
	h, err := atoi(parts[3])
	if err != nil {
		return err
	}
	o.x, o.y, o.w, o.h = x, y, w, h

	// Add to the global list
	ocrRegions = append(ocrRegions, image.Rect(x, y, x+w, y+h))

	return nil
}

var (
	// sourceFile  = flag.String("f", "/dev/video0", "Input device file")
	// ocrInterval = flag.Int("i", 8, "Min interval for performing OCR")
	// sleep       = flag.Int("s", 1, "Sleep millis between frames")
	// languages   = flag.String("l", "eng", "Comma-separated list of languages")
	// width       = flag.Int("w", 1920, "Width of the video capture")
	// height      = flag.Int("h", 1080, "Height of the video capture")
	// verbose     = flag.Bool("v", false, "Make verbose")
	// fps         = flag.Int("fps", 30, "Capture FPS")
	// ocrScale    = flag.Float64("q", 1, "Image scale for feeding OCR [0.1-1]")

	sourceFile  = getopt.StringLong("source", 's', "/dev/video0", "Input device file")
	ocrInterval = getopt.IntLong("interval", 'i', 8, "Min interval for performing OCR")
	sleep       = getopt.IntLong("Wait", 't', 1, "Sleep millis between frames")
	languages   = getopt.StringLong("lang", 'l', "eng", "Comma-separated list of languages")
	width       = getopt.IntLong("width", 'w', 1920, "Width of the video capture")
	height      = getopt.IntLong("height", 'h', 1080, "Height of the video capture")
	verbose     = getopt.BoolLong("verbose", 'v', "Make verbose")
	fps         = getopt.IntLong("fps", 'f', 30, "Capture FPS")

	ocrScale float64 = 1
	_                = getopt.FlagLong(&ocrScale, "ocr-scale", 'q', "Image scale for feeding OCR [0.1 - 1]")

	r = ocrRegionArg{}
	_ = getopt.FlagLong((*ocrRegionArg)(&r), "region", 'r', "Region to OCR in the form of x,y,w,h")
	// ocrRegions = make([]image.Rectangle, 0)
)

// func init() {
// 	flag.Func("r", "Region to OCR in the form of x,y,w,h", func(value string) error {
// 		e := fmt.Errorf("Invalid region format: %v", value)
// 		parts := strings.Split(value, ",")
// 		if len(parts) != 4 {
// 			return e
// 		}
// 		x, err := strconv.Atoi(parts[0])
// 		if err != nil {
// 			return e
// 		}
// 		y, err := strconv.Atoi(parts[1])
// 		if err != nil {
// 			return e
// 		}
// 		w, err := strconv.Atoi(parts[2])
// 		if err != nil {
// 			return e
// 		}
// 		h, err := strconv.Atoi(parts[3])
// 		if err != nil {
// 			return e
// 		}
// 		ocrRegions = append(ocrRegions, image.Rect(x, y, x+w, y+h))

// 		return nil
// 	})
// }

func toOutput(text string) string {
	return strings.ReplaceAll(text, "\n", " ")
}

var logMutex sync.Mutex

func logf(format string, a ...any) {
	logMutex.Lock()
	fmt.Fprintf(os.Stdout, format, a...)
	logMutex.Unlock()
}

func parseArgs() {
	getopt.Parse()

	if *verbose {
		common.DebugEnabled = true
		common.VerboseEnabled = true
	}

	// Normalize arguments
	if ocrScale < 0.1 {
		ocrScale = 0.1
	} else if ocrScale > 1 {
		ocrScale = 1
	}

	if len(ocrRegions) == 0 {
		ocrRegions = append(ocrRegions, image.Rect(0, 0, *width, *height))
	}
	if common.VerboseEnabled {
		fmt.Printf("# OCR Regions: %v\n", ocrRegions)
	}
}

func showVideoCaptureProps(webcam *gocv.VideoCapture) {
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
}

func mustInitCapture(file string, width, height, fps int) *gocv.VideoCapture {
	capture, err := gocv.OpenVideoCapture(*sourceFile)
	common.Check(err, "Error opening video capture device")

	capture.Set(gocv.VideoCaptureFrameWidth, float64(width))
	capture.Set(gocv.VideoCaptureFrameHeight, float64(height))
	for f := fps; f > 0; f-- {
		// fmt.Printf("Trying FPS: %v\n", f)
		capture.Set(gocv.VideoCaptureFPS, float64(f))
		if int(capture.Get(gocv.VideoCaptureFPS)) == f {
			break
		}
	}
	return capture
}

func realMain() int {

	// Open the video source and initialize it
	webcam := mustInitCapture(*sourceFile, *width, *height, *fps)

	defer webcam.Close()
	showVideoCaptureProps(webcam)

	// Initialize Tesseract client
	client := gosseract.NewClient()
	defer client.Close()

	langs := strings.Split(*languages, ",")
	logf("# Languages: %v\n", langs)
	client.SetLanguage(langs...)

	// Create a window for display (if enabled)
	window := gocv.NewWindow("Video with OCR")
	defer window.Close()

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

	reader := func() {
		if *verbose {
			logf("# Reader started\n")
		}
		for img := range send {
			defer img.Close()
			readStart := time.Now()

			gray := gocv.NewMat()
			defer gray.Close()
			gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

			for i, ocrRect := range ocrRegions {
				var rect = gray.Region(ocrRect)
				defer rect.Close()

				gocv.Resize(rect, &rect, image.Point{}, ocrScale, ocrScale, gocv.InterpolationLinear)

				// window.IMShow(rect)

				// Get image bytes
				imgBytes, err := gocv.IMEncode(gocv.PNGFileExt, rect)
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

				if len(text) > 0 {
					logf("# Text %d: %s\n", i, toOutput(text))
				}
			}
			readTime := time.Now().Sub(readStart)
			readMillis.Store(int32(readTime.Milliseconds()))
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
			send <- img.Clone()
			ready.Add(1)
			frames = 0
		}

		// Draw the ORC rectangles on the window
		for _, ocrRect := range ocrRegions {
			gocv.Rectangle(&img, ocrRect, color.RGBA{0, 255, 0, 0}, 3)
		}

		window.IMShow(img)

		key := window.WaitKey(1) // Handle events.

		if key == 27 { // ESC
			break
		}

		readFps++
		now := time.Now().UnixNano()
		if now >= nextTick {
			logf("# FPS: %d (last capture ms: %d, read ms: %d)\n", readFps, captureTime.Milliseconds(), readMillis.Load())
			nextTick = now + 1e9
			readFps = 0
		}

	}
	return 0
}
func main() {
	common.RunAndExit(func() int {
		parseArgs()
		//args := getopt.Args()
		// if *help || len(args) == 0 {
		// 	getopt.Usage()
		// 	os.Exit(0)
		// }

		// options := gaze.Options{}
		// options.Input = os.Stdin
		// options.Output = os.Stdout
		// options.ForcedTerminalWidth = *width
		// options.ForcedTerminalHeight = *height
		// options.CommandLine = args
		// options.SetInterval(time.Duration(float64(time.Second) * interval))
		// options.Precise = *precise
		// options.NoTitle = *noTitle
		// options.UseExec = *exec

		// common.Dump("Options: ", options)
		// common.Debugf("Display command: %s\n", options.GetDisplayCommand())
		// common.Dump("Exec command: ", options.GetExecCommand())

		// gazer := gaze.NewGazer(options)
		// defer gazer.Finish()

		// gazer.RunLoop(*times)

		return 0
	})
}
