package main

import (
	"fmt"
	"image"
	"image/color"
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

const (
	// From https://docs.opencv.org/4.x/d4/dd5/highgui_8hpp.html
	CV_EVENT_MOUSEMOVE     = 0
	CV_EVENT_LBUTTONDOWN   = 1
	CV_EVENT_RBUTTONDOWN   = 2
	CV_EVENT_MBUTTONDOWN   = 3
	CV_EVENT_LBUTTONUP     = 4
	CV_EVENT_RBUTTONUP     = 5
	CV_EVENT_MBUTTONUP     = 6
	CV_EVENT_LBUTTONDBLCLK = 7
	CV_EVENT_RBUTTONDBLCLK = 8
	CV_EVENT_MBUTTONDBLCLK = 9
	CV_EVENT_MOUSEWHEEL    = 10
	CV_EVENT_MOUSEHWHEEL   = 11
)

type ocrRegionArg struct {
}

var _ getopt.Value = (*ocrRegionArg)(nil)

// Removed invalid type conversion
func (o *ocrRegionArg) String() string {
	return fmt.Sprintf("%v", ocrRegions)
}

func (o *ocrRegionArg) Set(value string, opt getopt.Option) error {
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return fmt.Errorf("invalid region format: %v", value)
	}
	atoi := func(s string) (int, error) {
		x, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid region format: %v: %w", value, err)
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

	// Add to the global list
	ocrRegions = append(ocrRegions, image.Rect(x, y, x+w, y+h))

	return nil
}

var (
	sourceFile  = getopt.StringLong("source", 's', "/dev/video0", "Input device file")
	ocrInterval = getopt.IntLong("interval", 'i', 8, "Min interval for performing OCR")
	wait        = getopt.IntLong("Wait", 't', 1, "Sleep millis between frames")
	languages   = getopt.StringLong("lang", 'l', "eng", "Comma-separated list of languages")
	width       = getopt.IntLong("width", 'w', 1920, "Width of the video capture")
	height      = getopt.IntLong("height", 'h', 1080, "Height of the video capture")
	verbose     = getopt.BoolLong("verbose", 'v', "Make verbose")
	fps         = getopt.IntLong("fps", 'f', 30, "Capture FPS")
	noStats     = getopt.BoolLong("no-stats", 'n', "Disable stats")

	ocrScale float64 = 1
	_                = getopt.FlagLong(&ocrScale, "ocr-scale", 'q', "Image scale for feeding OCR [0.1 - 1]")

	r = ocrRegionArg{}
	_ = getopt.FlagLong((*ocrRegionArg)(&r), "region", 'r', "Region to OCR in the form of x,y,w,h")
)

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

	common.Verbosef("# OCR Regions: %v\n", ocrRegions)
}

func showVideoCaptureProps(webcam *gocv.VideoCapture) {
	fmt.Printf("# Input: Frame Width: %v\n", webcam.Get(gocv.VideoCaptureFrameWidth))
	fmt.Printf("# Input: Frame Height: %v\n", webcam.Get(gocv.VideoCaptureFrameHeight))
	fmt.Printf("# Input: FPS: %v\n", webcam.Get(gocv.VideoCaptureFPS))
	fmt.Printf("# Input: Buffer Size: %v\n", webcam.Get(gocv.VideoCaptureBufferSize))
	fmt.Printf("# Input: Brightness: %v\n", webcam.Get(gocv.VideoCaptureBrightness))
	fmt.Printf("# Input: Contrast: %v\n", webcam.Get(gocv.VideoCaptureContrast))
	fmt.Printf("# Input: Saturation: %v\n", webcam.Get(gocv.VideoCaptureSaturation))
	fmt.Printf("# Input: Hue: %v\n", webcam.Get(gocv.VideoCaptureHue))
	fmt.Printf("# Input: Gain: %v\n", webcam.Get(gocv.VideoCaptureGain))
	fmt.Printf("# Input: Exposure: %v\n", webcam.Get(gocv.VideoCaptureExposure))
	fmt.Printf("# Input: Auto Exposure: %v\n", webcam.Get(gocv.VideoCaptureAutoExposure))
	fmt.Printf("# Input: Gamma: %v\n", webcam.Get(gocv.VideoCaptureGamma))
	fmt.Printf("# Input: Sharpness: %v\n", webcam.Get(gocv.VideoCaptureSharpness))
	fmt.Printf("# Input: Backlight Compensation: %v\n", webcam.Get(gocv.VideoCaptureBacklight))
	fmt.Printf("# Input: Focus: %v\n", webcam.Get(gocv.VideoCaptureFocus))
	fmt.Printf("# Input: Zoom: %v\n", webcam.Get(gocv.VideoCaptureZoom))
	fmt.Printf("# Input: ISO Speed: %v\n", webcam.Get(gocv.VideoCaptureISOSpeed))
	fmt.Printf("# Input: Temperature: %v\n", webcam.Get(gocv.VideoCaptureTemperature))
	fmt.Printf("# Input: HW Acceleration: %v\n", webcam.Get(gocv.VideoCaptureHWAcceleration))
	fmt.Printf("# Input: HW Device: %v\n", webcam.Get(gocv.VideoCaptureHWDevice))
}

func mustInitCapture(file string, width, height, fps int) *gocv.VideoCapture {
	capture, err := gocv.OpenVideoCapture(file)
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

func mustInitOcr() *gosseract.Client {
	client := gosseract.NewClient()

	langs := strings.Split(*languages, ",")
	fmt.Printf("# Languages: %v\n", langs)
	client.SetLanguage(langs...)

	return client
}

type UiOptions struct {
	pauseOcr bool
	noStats  bool
}

type WindowManager struct {
	window *gocv.Window

	uiOptions *UiOptions

	mouseLDownX, mouseLDownY int
}

func NewWindowManager(window *gocv.Window, uiOptions *UiOptions) *WindowManager {
	ret := &WindowManager{
		window:    window,
		uiOptions: uiOptions,
	}
	window.SetMouseHandler(ret.mouseHandler, nil)
	return ret
}

func (w *WindowManager) Close() {
	w.window.Close()
}

func (w *WindowManager) ShowImage(img gocv.Mat) {
	w.window.IMShow(img)
}

func (w *WindowManager) mouseHandler(event int, x int, y int, flags int, data interface{}) {
	switch event {
	case CV_EVENT_LBUTTONDOWN:
		fmt.Printf("# Mouse: Left button down: %d, %d\n", x, y)
		w.mouseLDownX = x
		w.mouseLDownY = y
	case CV_EVENT_LBUTTONUP:
		fmt.Printf("# Mouse: Left button up: %d, %d\n", x, y)
		fmt.Printf("# Region: %d,%d,%d,%d\n", w.mouseLDownX, w.mouseLDownY, x-w.mouseLDownX, y-w.mouseLDownY)
	}
}

func (w *WindowManager) HandleEvents() bool {
	key := w.window.WaitKey(1) // Handle events.

	// if common.VerboseEnabled {
	// 	common.Verbosef("# Key: %v\n", key)
	// }

	switch key {
	case 27: // ESC
		return false
	case 'p':
		w.uiOptions.pauseOcr = !w.uiOptions.pauseOcr
		if w.uiOptions.pauseOcr {
			fmt.Printf("# OCR paused. Press [p] again to resume.\n")
		}
	case 's':
		w.uiOptions.noStats = !w.uiOptions.noStats
		if w.uiOptions.noStats {
			fmt.Printf("# Stats disabled. Press [s] again to enable.\n")
		}
	}
	return true
}

func ocrSingleFrame(client *gosseract.Client, img gocv.Mat) ([]string, time.Duration) {
	defer img.Close()
	readStart := time.Now()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

	ret := make([]string, 0)

	for _, ocrRect := range ocrRegions {
		var rect = gray.Region(ocrRect)
		defer rect.Close()

		gocv.Resize(rect, &rect, image.Point{}, ocrScale, ocrScale, gocv.InterpolationLinear)

		// Get image bytes
		imgBytes, err := gocv.IMEncode(gocv.PNGFileExt, rect)
		if err != nil {
			panic(fmt.Errorf("error: encoding image: %w", err))
		}

		if common.VerboseEnabled {
			fmt.Printf("# Scanning the image...\n")
		}
		client.SetImageFromBytes(imgBytes.GetBytes())
		text, err := client.Text()
		if err != nil {
			panic(fmt.Errorf("error: OCR failed: %w", err))
		}

		ret = append(ret, text)
	}
	readTime := time.Since(readStart)
	return ret, readTime
}

func realMain() int {

	uiOptions := UiOptions{}
	uiOptions.noStats = *noStats

	// Open the video source and initialize it
	webcam := mustInitCapture(*sourceFile, *width, *height, *fps)

	defer webcam.Close()
	showVideoCaptureProps(webcam)

	// Initialize Tesseract client
	client := mustInitOcr()
	defer client.Close()

	// Create a window for display (if enabled)
	window := NewWindowManager(gocv.NewWindow("Video with OCR"), &uiOptions)
	defer window.Close()

	// Prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// Various initialization...
	closing := atomic.Bool{}

	frameCounterForOcr := 0
	frameCounterForFps := 0

	// If it's 0, it means OCR is ready to process the next frame.
	// If it's 1, it means OCR is processing a frame.
	var ocrReady atomic.Int32

	var readMillis atomic.Int32
	readMillis.Store(-1)

	imageChannel := make(chan gocv.Mat)

	// Start OCR go routine that reads each frame from imageChannel and run OCR on it.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		common.Verbose("# Reader started\n")
		for img := range imageChannel {
			if img.Empty() {
				return
			}
			texts, duration := ocrSingleFrame(client, img)
			if closing.Load() {
				return
			}
			readMillis.Store(int32(duration.Milliseconds()))
			ocrReady.Add(-1)

			for i, text := range texts {
				if len(text) > 0 {
					fmt.Printf("# Text %d: %s\n", i, strings.ReplaceAll(text, "\n", " "))
				}
			}
		}
	}()

	lastTick := time.Now()
	nextTick := lastTick.Add(time.Second)

	fmt.Printf("# Info: [ESC] key to close app\n")
	fmt.Printf("# Info: [P] key to toggle OCR\n")
	fmt.Printf("# Info: [S] key to toggle stats\n")
	common.Verbose("# Started\n")

	// Main loop
	for {
		time.Sleep(time.Duration(*wait) * time.Millisecond)

		captureStart := time.Now()
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("# ERROR: Unable to read frame.\n")
			continue
		}
		captureTime := time.Since(captureStart)
		common.Verbose(".")

		if img.Empty() {
			continue
		}

		frameCounterForOcr++
		frameCounterForFps++

		// // Perform OCR every ocrInterval frames
		if !uiOptions.pauseOcr && frameCounterForOcr >= *ocrInterval && ocrReady.Load() == 0 {
			ocrReady.Add(1)
			imageChannel <- img.Clone()
			frameCounterForOcr = 0
		}

		// Draw the ORC rectangles on the image to show on the window.
		for _, ocrRect := range ocrRegions {
			gocv.Rectangle(&img, ocrRect, color.RGBA{0, 255, 0, 0}, 1)
		}

		// Update the window, and handle events.
		window.ShowImage(img)

		if !window.HandleEvents() {
			break
		}

		// Show FPS, etc
		now := time.Now()
		if now.Compare(nextTick) >= 0 {
			tickDuration := now.Sub(lastTick).Seconds()
			fps := float64(frameCounterForFps) / tickDuration
			if !uiOptions.pauseOcr && !uiOptions.noStats {
				fmt.Printf("# Stats: FPS: %d    Last capture ms: %d    Last OCR ms: %d\n", int(fps), captureTime.Milliseconds(), readMillis.Load())
			}
			lastTick = now
			nextTick = now.Add(time.Second)
			frameCounterForFps = 0
		}

	}
	fmt.Printf("# Exiting...\n")

	imageChannel <- gocv.NewMat()
	closing.Store(true)
	wg.Wait()
	return 0
}
func main() {
	common.RunAndExit(func() int {
		parseArgs()

		realMain()

		return 0
	})
}
