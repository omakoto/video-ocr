# video-ocr

Real-time OCR on a live video feed. Opens a capture device (webcam), displays the video in a window with highlighted OCR regions, and prints recognized text to stdout.

## Features

- Runs OCR asynchronously so frame capture is never blocked
- Supports multiple independent OCR regions in a single run
- Interactive: pause/resume OCR, toggle stats, and define regions by mouse drag
- Multi-language support via Tesseract (English, Japanese, and more)

## Usage

```
video-ocr [OPTIONS]

Options:
  -s, --source FILE       Video capture device (default: /dev/video0)
  -l, --lang LANG         OCR language(s), comma-separated (default: eng)
                            Examples: jpn, eng+jpn
  -h, --help              Show help and exit
  -W, --width N           Capture width in pixels (default: 1920)
  -H, --height N          Capture height in pixels (default: 1080)
  -f, --fps N             Target capture frame rate (default: 30)
  -r, --region x,y,w,h   OCR region; repeat for multiple regions
                            Default: full frame
  -i, --interval N        Run OCR once every N frames (default: 8)
  -q, --ocr-scale F       Scale image before OCR, 0.1–1.0 (default: 1.0)
                            Smaller values are faster but less accurate
  -n, --no-stats          Suppress FPS and OCR timing stats
  -t, --Wait N            Milliseconds to sleep between frame captures (default: 1)
  -v, --verbose           Enable verbose/debug output
```

### Examples

OCR with the default webcam, English only:
```
video-ocr
```

Use a second camera, Japanese text:
```
video-ocr -s /dev/video2 -l jpn
```

OCR only a specific region of the frame (e.g. a subtitle bar at the bottom):
```
video-ocr -r 0,900,1920,120
```

OCR two separate regions simultaneously:
```
video-ocr -r 100,50,400,80 -r 100,200,400,80
```

Mixed English/Japanese on a 720p camera, scaled down for faster OCR:
```
video-ocr -s /dev/video1 -l eng+jpn -w 1280 -h 720 -q 0.5
```

## Interactive controls

| Key / Action          | Effect                                      |
|-----------------------|---------------------------------------------|
| `ESC`                 | Quit                                        |
| `P`                   | Pause / resume OCR                          |
| `S`                   | Toggle FPS/OCR stats display                |
| Left-click and drag   | Print the region coordinates (`x,y,w,h`) to stdout so you can copy them into a `-r` flag |

## Output format

All output lines start with `#`. Lines without `#` are recognized text.

```
# Input: Frame Width: 1920
# Input: Frame Height: 1080
# Input: FPS: 30
# Languages: [eng]
# Info: [ESC] key to close app
# Info: [P] key to toggle OCR
# Info: [S] key to toggle stats
# Text 0: Hello, world
# Stats: FPS: 29    Last capture ms: 2    Last OCR ms: 145
```

To extract only recognized text, filter out comment lines:
```
video-ocr | grep -v '^#'
```

## Installation

```
go install github.com/omakoto/video-ocr/cmd/...@latest
```

Or, after `git clone`:
```
./00install.bash
```

## Dependencies

### OpenCV

Install the runtime libraries and development headers:
```
sudo apt install libopencv-dev
```

This provides the `opencv4.pc` pkg-config file that GoCV needs to locate headers and libraries at build time. The runtime libraries alone (e.g. `libopencv-core410`) are not sufficient.

### GoCV (Go OpenCV bindings)

- Source: https://github.com/hybridgroup/gocv
- Follow the install instructions in the GoCV README.

Known fixes for **Ubuntu 24**:
- `libtbb2` does not exist — remove it from `DEBS`.
- Rename `libdc1394-22-dev` → `libdc1394-dev`.

```diff
-DEBS=... libtbb2 libtbb-dev ... libdc1394-22-dev ...
+DEBS=... libtbb-dev ... libdc1394-dev ...
```

### gosseract (Tesseract OCR)

```
sudo apt install tesseract-ocr libtesseract-dev
sudo apt install tesseract-ocr-eng tesseract-ocr-jpn   # add languages as needed
```

Full language list: https://github.com/tesseract-ocr/tessdoc/blob/main/Installation.md
