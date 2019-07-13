// An implementation of Conway's Game of Life.
package main

import (
	"bytes"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"math/rand"
	"time"
)

var s1 = rand.NewSource(time.Now().UnixNano())
var r1 = rand.New(s1)

// Field represents a two-dimensional field of cells.
type Field struct {
	s    [][]bool
	w, h int
}

// NewField returns an empty field of the specified width and height.
func NewField(w, h int) *Field {
	s := make([][]bool, h)
	for i := range s {
		s[i] = make([]bool, w)
	}
	return &Field{s: s, w: w, h: h}
}

// Set sets the state of the specified cell to the given value.
func (f *Field) Set(x, y int, b bool) {
	f.s[y][x] = b
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are wrapped
// toroidally. For instance, an x value of -1 is treated as width-1.
func (f *Field) Alive(x, y int) bool {
	x += f.w
	x %= f.w
	y += f.h
	y %= f.h
	return f.s[y][x]
}

// Next returns the state of the specified cell at the next time step.
func (f *Field) Next(x, y int) bool {
	// Count the adjacent cells that are alive.
	alive := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && f.Alive(x+i, y+j) {
				alive++
			}
		}
	}
	// Return next state according to the game rules:
	//   exactly 3 neighbors: on,
	//   exactly 2 neighbors: maintain current state,
	//   otherwise: off.
	return alive == 3 || alive == 2 && f.Alive(x, y)
}

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b *Field
	w, h int
}

// NewLife returns a new Life game state with a random initial state.
func NewLife(w, h int) *Life {
	a := NewField(w, h)
	for i := 0; i < (w * h ); i++ {
		a.Set(r1.Intn(w), r1.Intn(h), true)
	}
	return &Life {
		a: a, b: NewField(w, h),
		w: w, h: h,
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			l.b.Set(x, y, l.a.Next(x, y))
		}
	}
	if l.a == l.b  {
		for i := 0; i < (l.w * l.h / 4); i++ {
			l.b.Set(r1.Intn(l.w), r1.Intn(l.h), true)
		}
	}
	// Swap fields a and b.
	l.a, l.b = l.b, l.a

}

// String returns the game board as a string.
func (l *Life) String() string {
	var buf bytes.Buffer
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			b := byte(' ')
			if l.a.Alive(x, y) {
				b = '*'
			}
			buf.WriteByte(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (l *Life) Image() *image.Gray {
	m := image.NewGray(image.Rect(0, 0, l.w, l.h))

	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			pixel := m.At(x, y).(color.Gray)

			if l.a.Alive(x, y) {
				pixel.Y = 0
			} else {
				pixel.Y = 255
			}
			m.Set(x, y, pixel)

		}
	}
	return m
}

func main() {
	imgChan := make(chan gocv.Mat)
	window := gocv.NewWindow("Hello")

	wd := 100
	ht := 100
	go func() {
		for{
			l := NewLife(wd, ht)

			for i := 0; i < 1500; i++{
				l.Step()

				img, err := gocv.ImageGrayToMatGray(l.Image())
				if err != nil {
					panic("SOmething wrong")
				}
				//fmt.Println("Sending to the channel")
				imgChan <- img
				//time.Sleep(time.Second * 2)
			}
			fmt.Println("Next Wave")
		}
	}()

	for {
		select {
		case comp :=<- imgChan:
			//fmt.Println("Got to the goroutine")
			window.IMShow(comp)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}
}
