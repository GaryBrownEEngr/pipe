package pipe

import (
	"math"
	"time"
)

type Controller struct {
	blockSize    int
	channelDepth int
	timeStep     time.Duration
	done         chan bool
}

func NewController(blockSize, channelDepth int, timeStep time.Duration) *Controller {
	ret := &Controller{
		blockSize:    blockSize,
		channelDepth: channelDepth,
		timeStep:     timeStep,
		done:         make(chan bool),
	}

	return ret
}

func SrcSine(c *Controller, frequencyHz float64, startPhaseRad float64) chan []float64 {
	outChan := make(chan []float64, c.channelDepth)

	go func() {
		sin, cos := math.Sincos(startPhaseRad)
		v := complex(cos, sin)

		sin, cos = math.Sincos(frequencyHz * 2 * math.Pi * c.timeStep.Seconds())
		twiddle := complex(cos, sin)

	MainLoop:
		for {
			select {
			case <-c.done:
				break MainLoop
			default:
				// nothing
			}

			buf := make([]float64, c.blockSize)
			for i := 0; i < c.blockSize; i++ {
				buf[i] = imag(v)
				v = v * twiddle
			}

			outChan <- buf

			// Normalize the vector back to a length of 1.
			mag := math.Sqrt(real(v)*real(v) + imag(v)*imag(v))
			v = complex(real(v)/mag, imag(v)/mag)
		}

		close(outChan)

	}()

	return outChan
}

func SrcCosine(c *Controller, frequencyHz float64, startPhaseRad float64) chan []float64 {
	return SrcSine(c, frequencyHz, startPhaseRad+math.Pi/2.0)
}

func Add(c *Controller, a, b chan []float64) chan []float64 {
	outChan := make(chan []float64, c.channelDepth)

	go func() {
	MainLoop:
		for {
			select {
			case <-c.done:
				break MainLoop
			default:
				// nothing
			}

			aData := <-a
			bData := <-b

			if len(aData) != c.blockSize || len(bData) != c.blockSize {
				panic("Bad block size")
			}

			buf := make([]float64, c.blockSize)
			for i := 0; i < c.blockSize; i++ {
				buf[i] = aData[i] + bData[i]
			}

			outChan <- buf
		}

		go drainChannelTillClosed(a)
		drainChannelTillClosed(b)

		close(outChan)

	}()

	return outChan
}

func drainChannelTillClosed[T any](in chan []T) {
	for {
		_, ok := <-in
		if !ok {
			return
		}
	}

}

func Multiply(c *Controller, a, b chan []float64) chan []float64 {
	outChan := make(chan []float64, c.channelDepth)

	go func() {
	MainLoop:
		for {
			select {
			case <-c.done:
				break MainLoop
			default:
				// nothing
			}

			aData := <-a
			bData := <-b

			if len(aData) != c.blockSize || len(bData) != c.blockSize {
				panic("Bad block size")
			}

			buf := make([]float64, c.blockSize)
			for i := 0; i < c.blockSize; i++ {
				buf[i] = aData[i] * bData[i]
			}

			outChan <- buf
		}

		go drainChannelTillClosed(a)
		drainChannelTillClosed(b)

		close(outChan)

	}()

	return outChan
}
