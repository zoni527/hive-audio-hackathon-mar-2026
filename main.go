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

const PI float64 = math.Pi
const BUFFER_SIZE int32 = 1024
const WINDOW_HEIGHT int32 = 1000
const WINDOW_WIDTH int32 = 1800
const TARGET_FPS int32 = 500
const AMPLITUDE_GRAPH_AMPLITUDE int32 = 500
const FFT_GRAPH_AMPLITUDE int32 = 300
const FFT_GRAPH_AMPLITUDE_Y_OFFSET int32 = 500

type vis struct {
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
	go fft(a0, invert)
	go fft(a1, invert)

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

	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(buf), buf)
	chk(err)
	defer stream.Close()
	chk(stream.Start())

	var time, sampleMemorySize uint64
	sampleMemorySize = 20
	var max int32
	var fmax, fprev, textSize, boxSize float64
	points := make([]vis, WINDOW_WIDTH)
	sampleMemory := make([]float64, sampleMemorySize)
	for !rl.WindowShouldClose() {
		chk(stream.Read())

		// Look for maximum element & form complex numbers for fft
		max = 0
		for i, v := range buf {
			if i < largestPowerOfTwo {
				num := float64(v) / float64(math.MaxInt32)
				complexNumbers[i] = complex(num, 0)
			}
			if v > max {
				max = v
			}
		}
		fmax = float64(max) / float64(math.MaxInt32)
		// Threshold
		if fmax > 0.1 {
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
		textSize = averageSample
		// fmt.Printf("max:\t%v\n", fmax)
		// fmt.Println(complexNumbers)

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		speedy := 0.01
		phasey := float64(time) * speedy
		points[time%uint64(WINDOW_WIDTH)] = vis{w: 2, h: int32(float64(AMPLITUDE_GRAPH_AMPLITUDE) * boxSize)}
		for i := range points {
			rl.DrawRectangle(points[i].x, points[i].y, points[i].w, points[i].h, rl.Red)
			points[i].x += 1
		}
		for i := range len(complexNumbers) / 2 {
			offset := float64(i) * (float64(WINDOW_WIDTH) / float64(len(complexNumbers)/2))
			width := WINDOW_WIDTH/int32(len(complexNumbers)) + 3
			height := int32(cmplx.Abs(complexNumbers[i]) * float64(FFT_GRAPH_AMPLITUDE))
			rl.DrawRectangle(int32(offset), FFT_GRAPH_AMPLITUDE_Y_OFFSET, width, height, rl.Green)
		}
		rl.DrawText("Congrats! You created your first window!", 100, 500+int32(math.Sin(phasey)*100.0), 20+int32(textSize*300), rl.Black)

		rl.EndDrawing()
		fprev = boxSize

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
