package pipe

import (
	"fmt"
	"math"
	"math/cmplx"
)

type RealNumber interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type ComplexNumber interface {
	complex64 | complex128
}

type Number interface {
	RealNumber | ComplexNumber
}

func SrcComplexFrequency(c *Controller, frequencyHz float64, startPhaseRad float64) *Wire[complex128] {
	outWire := NewWire[complex128](c.channelDepth)

	go func() {
		sin, cos := math.Sincos(startPhaseRad)
		v := complex(cos, sin)

		sin, cos = math.Sincos(frequencyHz * 2 * math.Pi * c.timeStepSec)
		twiddle := complex(cos, sin)

		c.WaitForStart()

	MainLoop:
		for {
			select {
			case <-c.doneChan:
				break MainLoop
			default:
				// nothing
			}

			buf := make([]complex128, c.blockSize)
			for i := 0; i < c.blockSize; i++ {
				buf[i] = v
				v = v * twiddle
			}

			outWire.Publish(buf)

			// Normalize the vector back to a length of 1.
			mag := cmplx.Abs(v)
			v = complex(real(v)/mag, imag(v)/mag)
		}

		outWire.Stop()
	}()

	return outWire.GetWire()
}

func SrcSine(c *Controller, frequencyHz float64, startPhaseRad float64) *Wire[float64] {
	outWire := NewWire[float64](c.channelDepth)

	go func() {
		sin, cos := math.Sincos(startPhaseRad)
		v := complex(cos, sin)

		sin, cos = math.Sincos(frequencyHz * 2 * math.Pi * c.timeStepSec)
		twiddle := complex(cos, sin)

		c.WaitForStart()

	MainLoop:
		for {
			select {
			case <-c.doneChan:
				break MainLoop
			default:
				// nothing
			}

			buf := make([]float64, c.blockSize)
			for i := 0; i < c.blockSize; i++ {
				buf[i] = imag(v)
				v = v * twiddle
			}

			outWire.Publish(buf)

			// Normalize the vector back to a length of 1.
			mag := math.Sqrt(real(v)*real(v) + imag(v)*imag(v))
			v = complex(real(v)/mag, imag(v)/mag)
		}

		outWire.Stop()
	}()

	return outWire.GetWire()
}

func SrcCosine(c *Controller, frequencyHz float64, startPhaseRad float64) *Wire[float64] {
	return SrcSine(c, frequencyHz, startPhaseRad+math.Pi/2.0)
}

// y = x[0] + x[1] + x[2] + ...
func Add[T Number](c *Controller, inputs ...*Wire[T]) *Wire[T] {
	if len(inputs) < 2 {
		panic("Need 2 or more inputs")
	}

	operation := func(inputs [][]T) []T {
		// make the output block
		out := make([]T, c.blockSize)

		// Add the values
		for i := 0; i < c.blockSize; i++ {
			y := inputs[i][0]
			for _, x := range inputs[i][1:] {
				y += x
			}
			out[i] = y
		}
		return out
	}

	return GenericProcessingBlock_NIn_1Out(c, operation, inputs...)
}

// y = x[0] - x[1] - x[2] - x[3] - ...
func Subtract[T Number](c *Controller, inputs ...*Wire[T]) *Wire[T] {
	if len(inputs) < 2 {
		panic("Need 2 or more inputs")
	}

	operation := func(inputs [][]T) []T {
		// make the output block
		out := make([]T, c.blockSize)

		// Add the values
		for i := 0; i < c.blockSize; i++ {
			y := inputs[i][0]
			for _, x := range inputs[i][1:] {
				y -= x
			}
			out[i] = y
		}
		return out
	}

	return GenericProcessingBlock_NIn_1Out(c, operation, inputs...)
}

// y = x[0] * x[1] * x[2] * ...
func Multiply[T Number](c *Controller, inputs ...*Wire[T]) *Wire[T] {
	operation := func(inputs [][]T) []T {
		// make the output block
		out := make([]T, c.blockSize)

		// Add the values
		for i := 0; i < c.blockSize; i++ {
			y := inputs[i][0]
			for _, x := range inputs[i][1:] {
				y *= x
			}
			out[i] = y
		}
		return out
	}

	return GenericProcessingBlock_NIn_1Out(c, operation, inputs...)
}

// y = x[0] / x[1] / x[2] / ...
func Divide[T Number](c *Controller, inputs ...*Wire[T]) *Wire[T] {
	operation := func(inputs [][]T) []T {
		// make the output block
		out := make([]T, c.blockSize)

		// Add the values
		for i := 0; i < c.blockSize; i++ {
			y := inputs[i][0]
			for _, x := range inputs[i][1:] {
				y /= x
			}
			out[i] = y
		}
		return out
	}

	return GenericProcessingBlock_NIn_1Out(c, operation, inputs...)
}

func RealToFloat64[T RealNumber](c *Controller, inWire *Wire[T]) *Wire[float64] {
	operation := func(input []T) []float64 {
		// make the output block
		out := make([]float64, c.blockSize)

		// Add the values
		for i := 0; i < c.blockSize; i++ {
			out[i] = float64(input[i])
		}
		return out
	}

	return GenericProcessingBlock_1In_1Out(c, operation, inWire)
}

func RealToComplex128[T RealNumber](c *Controller, inWire *Wire[T]) *Wire[complex128] {
	operation := func(input []T) []complex128 {
		// make the output block
		out := make([]complex128, c.blockSize)

		// Add the values
		for i := 0; i < c.blockSize; i++ {
			out[i] = complex(float64(input[i]), 0)
		}
		return out
	}

	return GenericProcessingBlock_1In_1Out(c, operation, inWire)
}

func Complex128ToFloat64(c *Controller, inWire *Wire[complex128]) *Wire[float64] {
	operation := func(input []complex128) []float64 {
		// make the output block
		out := make([]float64, c.blockSize)

		// Add the values
		for i := 0; i < c.blockSize; i++ {
			x := real(input[i])
			fmt.Println(x)

			// out[i] = real(input[i])
		}
		return out
	}

	return GenericProcessingBlock_1In_1Out(c, operation, inWire)
}
