package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"os/signal"
	"syscall"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gordonklaus/portaudio"
)

const PI = math.Pi
const BUFFER_SIZE = 1024
const WINDOW_HEIGHT int32 = 1000
const WINDOW_WIDTH int32 = 1800
const TARGET_FPS = 500
const AMPLITUDE_GRAPH_AMPLITUDE = WINDOW_HEIGHT/2 - 200
const FFT_GRAPH_AMPLITUDE = 100
const FFT_GRAPH_AMPLITUDE_Y_OFFSET = WINDOW_HEIGHT
const SAMPLE_GRAPH_AMPLITUDE float64 = 500
const POLAR_BASE_RADIUS float64 = 200.0
const POLAR_AMPLITUDE float64 = 180.0
const POLAR_THICKNESS = 3
const GRAVITY = 0.95

type pixel struct {
	x, y int32
}

type rect struct {
	x, y, w, h int32
}

func fft(a []complex128, invert bool) {
	n := len(a)
	if n == 1 {
		return
	}

	a0 := make([]complex128, n/2)
	a1 := make([]complex128, n/2)
	for i := 0; 2*i < n; i++ {
		a0[i] = a[2*i]
		a1[i] = a[2*i+1]
	}

	fft(a0, invert)
	fft(a1, invert)
	ang := 2 * PI / float64(n)
	if invert {
		ang *= -1
	}

	w := complex(1, 0)
	wn := complex(math.Cos(ang), math.Sin(ang))
	for i := 0; 2*i < n; i++ {
		a[i] = a0[i] + w*a1[i]
		a[i+n/2] = a0[i] - w*a1[i]
		if invert {
			a[i] /= 2
			a[i+n/2] /= 2
		}
		w *= wn
	}
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowHighdpi)
	rl.InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "Sound visualizer")
	defer rl.CloseWindow()

	rl.SetTargetFPS(TARGET_FPS)

	fmt.Println("Recording.  Press Ctrl-C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	portaudio.Initialize()
	defer portaudio.Terminate()

	buf := make([]int32, BUFFER_SIZE)
	largestPowerOfTwo := 2
	for largestPowerOfTwo <= len(buf) {
		largestPowerOfTwo *= 2
	}
	largestPowerOfTwo /= 2

	complexNumbers := make([]complex128, largestPowerOfTwo)
	visualHeights := make([]float64, largestPowerOfTwo/2)

	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(buf), buf)
	chk(err)
	defer stream.Close()
	chk(stream.Start())

	var time, sampleMemorySize uint64
	sampleMemorySize = 20
	var max int32
	var fmax, fprev, boxSize float64
	rects := make([]rect, WINDOW_WIDTH)
	sampleMemory := make([]float64, sampleMemorySize)
	polarPoints := make([]pixel, len(buf))
	for !rl.WindowShouldClose() {
		chk(stream.Read())

		// Loop over samples in buffer
		max = 0
		for i, v := range buf {
			num := float64(v) / math.MaxInt32
			if i < largestPowerOfTwo {
				windowingFactor := (0.5 - 0.5*math.Cos(2*PI*float64(i)/float64(len(buf))))
				complexNumbers[i] = complex(num*windowingFactor, 0)
			}
			if v > max {
				max = v
			}

			// Calculate points for polar coordinate visualizer
			centerX := WINDOW_WIDTH / 2
			centerY := WINDOW_HEIGHT / 2
			angle := float64(i) * 2 * PI / float64(len(buf))
			sampleX := centerX + int32(math.Cos(angle)*(POLAR_BASE_RADIUS+num*POLAR_AMPLITUDE))
			sampleY := centerY + int32(math.Sin(angle)*(POLAR_BASE_RADIUS+num*POLAR_AMPLITUDE))
			polarPoints[i] = pixel{sampleX, sampleY}
		}

		fmax = float64(max) / math.MaxInt32
		if fmax > 0.1 { // Threshold
			sampleMemory[time%sampleMemorySize] = fmax
		} else {
			sampleMemory[time%sampleMemorySize] = 0
		}

		fft(complexNumbers, false)

		// Smoothing for text scaling
		var averageSample float64
		for _, v := range sampleMemory {
			averageSample += v
		}
		averageSample /= float64(len(sampleMemory))

		// Scaling drawable elements
		boxSize = 0.1*fprev + 0.9*fmax
		rects[time%uint64(WINDOW_WIDTH)] = rect{w: 2, h: int32(float64(AMPLITUDE_GRAPH_AMPLITUDE) * boxSize)}
		bgCol := rl.Color{R: 30, G: 30, B: 40, A: 255}
		rl.BeginDrawing()
		rl.ClearBackground(bgCol)

		// Sample visualizer
		for i, v := range buf {
			num := float64(v) / float64(math.MaxInt32)
			height := int32(num * SAMPLE_GRAPH_AMPLITUDE)
			leftMargin := int32((WINDOW_WIDTH - BUFFER_SIZE) / 2)
			rl.DrawRectangle(leftMargin+int32(i), WINDOW_HEIGHT/2-height/2, 1, height, rl.Purple)
		}

		// Loudness visualizer
		loudnessColor := rl.Red
		loudnessColor.A = 180
		for i := range rects {
			rl.DrawRectangle(rects[i].x, rects[i].y, rects[i].w, rects[i].h, loudnessColor)
			rects[i].x += 1
		}

		// FFT visualizer
		fftCol := rl.Green
		fftCol.A = 180
		for i := range len(complexNumbers) / 2 {
			offset := int32(2 * float64(i) * (float64(WINDOW_WIDTH) / float64(len(complexNumbers)/2)))
			width := WINDOW_WIDTH/int32(len(complexNumbers)) + 3
			magnitude := cmplx.Abs(complexNumbers[i])
			height := math.Log1p(magnitude) * float64(FFT_GRAPH_AMPLITUDE)
			if height > visualHeights[i] {
				visualHeights[i] = height
			} else {
				visualHeights[i] *= GRAVITY
			}
			rl.DrawRectangle(offset,
				FFT_GRAPH_AMPLITUDE_Y_OFFSET-int32(visualHeights[i]),
				width, int32(visualHeights[i]),
				fftCol)
		}

		// Centered text
		textString := "ENJOY THE SOUNDS!"
		textSize := float32(80)
		textSpacing := float32(2)
		textFont := rl.GetFontDefault()
		textDimensions := rl.MeasureTextEx(textFont, textString, textSize, textSpacing)
		textPosition := rl.Vector2{
			X: float32(WINDOW_WIDTH / 2),
			Y: float32(WINDOW_HEIGHT / 2),
		}
		textRotation := float32(0)
		textOrigin := rl.Vector2{
			X: textDimensions.X / 2,
			Y: textDimensions.Y / 2,
		}
		textCol := bgCol
		rl.DrawTextPro(
			textFont,
			textString,
			textPosition,
			textOrigin,
			textRotation,
			textSize, textSpacing,
			textCol,
		)

		// Polar coordinates sample visualizer
		polarCol := rl.Color{R: 200, G: 200, B: 200, A: 255}
		for _, v := range polarPoints {
			rl.DrawCircle(v.x, v.y, POLAR_THICKNESS, polarCol)
		}
		rl.EndDrawing()
		fprev = boxSize

		// Closing signal
		select {
		case <-sig:
			return
		default:
		}
		time++
	}
}

func chk(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
