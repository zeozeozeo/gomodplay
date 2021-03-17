package speaker

// Ripped out of https://github.com/faiface/beep

import (
	"errors"
	"sync"

	"github.com/hajimehoshi/oto"
)

// Streamer provides the interface to stream samples
type Streamer interface {
	Stream(samples [][2]float32) (n int, ok bool)
	Err() error
}

var (
	mu       sync.Mutex
	samples  [][2]float32
	buf      []byte
	context  *oto.Context
	player   *oto.Player
	done     chan struct{}
	streamer *Streamer
	callback *func()
)

// Init initializes audio playback through speaker. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the speaker's buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func Init(sampleRate uint32, bufferSize uint32) error {
	mu.Lock()
	defer mu.Unlock()

	Close()

	numBytes := int(bufferSize * 4)
	samples = make([][2]float32, bufferSize)
	buf = make([]byte, numBytes)

	var err error
	context, err = oto.NewContext(int(sampleRate), 2, 2, numBytes)
	if err != nil {
		return errors.New(("Could not initialise speaker"))
	}
	player = context.NewPlayer()

	done = make(chan struct{})

	go func() {
		for {
			select {
			default:
				update()
			case <-done:
				return
			}
		}
	}()

	return nil
}

// Close down everything
func Close() {
	if player != nil {
		if done != nil {
			done <- struct{}{}
			done = nil
		}
		player.Close()
		context.Close()
		player = nil
	}
}

// Play some music
func Play(s Streamer, callback func()) {
	mu.Lock()
	streamer = &s
	mu.Unlock()
}

func update() {
	mu.Lock()
	numSamples, ok := (*streamer).Stream(samples)
	mu.Unlock()

	if !ok {
		Close()
		(*callback)()
		return
	}

	for i := 0; i < numSamples; i++ {
		for c := range samples[i] {
			val := samples[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			buf[i*4+c*2+0] = low
			buf[i*4+c*2+1] = high
		}
	}

	player.Write(buf)
}
