package pipe

import (
	"math"
	"math/cmplx"
	"sync"

	"gonum.org/v1/gonum/dsp/fourier"
)

// https://numpy.org/doc/stable/reference/generated/numpy.fft.fft.html
// https://numpy.org/doc/stable/reference/routines.fft.html#module-numpy.fft

type fftTwiddlesCacheItem struct {
	mutex    sync.RWMutex
	twiddles *fourier.CmplxFFT
}

type fftTwiddlesCache struct {
	mutex       sync.RWMutex
	twiddlesMap map[int]*fftTwiddlesCacheItem
}

var cache = fftTwiddlesCache{
	twiddlesMap: make(map[int]*fftTwiddlesCacheItem),
}

func getTwiddleFactors(n int) *fourier.CmplxFFT {
	cache.mutex.RLock()
	item, ok := cache.twiddlesMap[n]
	cache.mutex.RUnlock()
	if !ok {
		cache.mutex.Lock()
		// Make sure the size still isn't present
		item, ok = cache.twiddlesMap[n]
		if !ok {
			// if it still isn't there then build it
			item = &fftTwiddlesCacheItem{}
			item.mutex.Lock()
			cache.twiddlesMap[n] = item
			cache.mutex.Unlock()
			item.twiddles = fourier.NewCmplxFFT(n)
			item.mutex.Unlock()
		} else {
			cache.mutex.Unlock()
		}
	}

	item.mutex.RLock()
	defer item.mutex.RUnlock()
	return item.twiddles
}

func realSliceToComplexSlice[T RealNumber](in []T) []complex128 {
	ret := make([]complex128, len(in))
	for i, x := range in {
		ret[i] = complex(float64(x), 0)
	}
	return ret
}

func NumberSliceToComplexSlice(inAny any) []complex128 {
	var in []complex128

	switch v := inAny.(type) {
	case []complex128:
		in = v
	case []complex64:
		in = make([]complex128, len(v))
		for i, x := range v {
			in[i] = complex(float64(real(x)), float64(imag(x)))
		}
	case []float64:
		in = realSliceToComplexSlice(v)
	case []float32:
		in = realSliceToComplexSlice(v)
	case []int:
		in = realSliceToComplexSlice(v)
	case []int8:
		in = realSliceToComplexSlice(v)
	case []int16:
		in = realSliceToComplexSlice(v)
	case []int32:
		in = realSliceToComplexSlice(v)
	case []int64:
		in = realSliceToComplexSlice(v)
	case []uint:
		in = realSliceToComplexSlice(v)
	case []uint8:
		in = realSliceToComplexSlice(v)
	case []uint16:
		in = realSliceToComplexSlice(v)
	case []uint32:
		in = realSliceToComplexSlice(v)
	case []uint64:
		in = realSliceToComplexSlice(v)
	}

	return in
}

// Coefficients computes the Fourier coefficients of a complex input sequence,
// converting the time series in seq into the frequency spectrum, placing
// the result in dst and returning it. This transform is unnormalized; a call
// to Coefficients followed by a call of Sequence will multiply the input
// sequence by the length of the sequence.
//
// If the length of seq is not t.Len(), Coefficients will panic.
// If dst is nil, a new slice is allocated and returned. If dst is not nil and
// the length of dst does not equal the length of seq, Coefficients will panic.
// It is safe to use the same slice for dst and seq.
func FFT(inAny any, dest []complex128) []complex128 {
	in := NumberSliceToComplexSlice(inAny)
	twiddles := getTwiddleFactors(len(in))
	return twiddles.Coefficients(dest, in)
}

// Sequence computes the complex periodic sequence from the Fourier coefficients,
// converting the frequency spectrum in coeff into a time series, placing the
// result in dst and returning it. This transform is unnormalized; a call to
// Coefficients followed by a call of Sequence will multiply the input sequence
// by the length of the sequence.
//
// If the length of coeff is not t.Len(), Sequence will panic.
// If dst is nil, a new slice is allocated and returned. If dst is not nil and
// the length of dst does not equal the length of coeff, Sequence will panic.
// It is safe to use the same slice for dst and coeff.
func IFFT(inAny any, dest []complex128) []complex128 {
	in := NumberSliceToComplexSlice(inAny)
	twiddles := getTwiddleFactors(len(in))
	return twiddles.Sequence(dest, in)
}

// Divide every coefficient by 1/n
func Norm(in []complex128) {
	factor := 1.0 / float64(len(in))
	for i, x := range in {
		in[i] = complex(real(x)*factor, imag(x)*factor)
	}
}

// Divide every coefficient by 1/Sqrt(n)
func NormOrtho(in []complex128) {
	factor := 1.0 / math.Sqrt(float64(len(in)))
	for i, x := range in {
		in[i] = complex(real(x)*factor, imag(x)*factor)
	}
}

func Freq(n int, samplingFreq float64) []float64 {
	factor := samplingFreq / float64(n)
	ret := make([]float64, n)
	for i := range ret {
		ret[i] = float64(i) * factor
	}

	return ret
}

func Mag(in []complex128, dest []float64) []float64 {
	if dest == nil {
		dest = make([]float64, len(in))
	}

	for i := range in {
		dest[i] = cmplx.Abs(in[i])
	}

	return dest
}
