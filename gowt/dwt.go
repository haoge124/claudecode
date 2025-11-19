package gowt

import (
	"fmt"
	"math"
)

// Mode represents the signal extension mode
type Mode string

const (
	// ModeZero pads with zeros
	ModeZero Mode = "zero"
	// ModeSymmetric symmetric padding
	ModeSymmetric Mode = "symmetric"
	// ModePeriodic periodic extension
	ModePeriodic Mode = "periodic"
)

// DWT performs single-level discrete wavelet transform
func DWT(data []float64, waveletName string, mode Mode) ([]float64, []float64, error) {
	wavelet, err := GetWavelet(waveletName)
	if err != nil {
		return nil, nil, err
	}

	if len(data) < len(wavelet.DecLowFilter) {
		return nil, nil, fmt.Errorf("data length must be at least %d", len(wavelet.DecLowFilter))
	}

	// Extend signal
	extended := extendSignal(data, len(wavelet.DecLowFilter), mode)

	// Convolve and downsample
	cA := convolveDownsample(extended, wavelet.DecLowFilter)
	cD := convolveDownsample(extended, wavelet.DecHighFilter)

	return cA, cD, nil
}

// IDWT performs single-level inverse discrete wavelet transform
func IDWT(cA, cD []float64, waveletName string) ([]float64, error) {
	wavelet, err := GetWavelet(waveletName)
	if err != nil {
		return nil, err
	}

	if cA == nil && cD == nil {
		return nil, fmt.Errorf("both cA and cD cannot be nil")
	}

	var length int
	if cA != nil {
		length = len(cA)
	} else {
		length = len(cD)
	}

	// Upsample and convolve
	var rec1, rec2 []float64

	if cA != nil {
		rec1 = upsampleConvolve(cA, wavelet.RecLowFilter)
	} else {
		rec1 = make([]float64, length*2+len(wavelet.RecLowFilter)-1)
	}

	if cD != nil {
		rec2 = upsampleConvolve(cD, wavelet.RecHighFilter)
	} else {
		rec2 = make([]float64, length*2+len(wavelet.RecHighFilter)-1)
	}

	// Ensure both arrays have the same length
	maxLen := len(rec1)
	if len(rec2) > maxLen {
		maxLen = len(rec2)
	}

	// Pad shorter array
	if len(rec1) < maxLen {
		temp := make([]float64, maxLen)
		copy(temp, rec1)
		rec1 = temp
	}
	if len(rec2) < maxLen {
		temp := make([]float64, maxLen)
		copy(temp, rec2)
		rec2 = temp
	}

	// Add the two reconstructed signals
	result := make([]float64, maxLen)
	for i := 0; i < maxLen; i++ {
		result[i] = rec1[i] + rec2[i]
	}

	// Trim the result to proper length
	outputLen := length * 2
	start := (len(result) - outputLen) / 2
	if start < 0 {
		start = 0
	}
	end := start + outputLen
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// Wavedec performs multilevel discrete wavelet decomposition
func Wavedec(data []float64, waveletName string, level int, mode Mode) ([][]float64, error) {
	if level < 1 {
		return nil, fmt.Errorf("level must be at least 1")
	}

	coeffs := make([][]float64, level+1)
	current := data

	for i := 0; i < level; i++ {
		cA, cD, err := DWT(current, waveletName, mode)
		if err != nil {
			return nil, err
		}
		coeffs[level-i] = cD
		current = cA
	}
	coeffs[0] = current

	return coeffs, nil
}

// Waverec performs multilevel discrete wavelet reconstruction
func Waverec(coeffs [][]float64, waveletName string) ([]float64, error) {
	if len(coeffs) < 2 {
		return nil, fmt.Errorf("coeffs must have at least 2 levels")
	}

	current := coeffs[0]
	for i := 1; i < len(coeffs); i++ {
		rec, err := IDWT(current, coeffs[i], waveletName)
		if err != nil {
			return nil, err
		}
		current = rec
	}

	return current, nil
}

// extendSignal extends the signal according to the mode
func extendSignal(data []float64, filterLen int, mode Mode) []float64 {
	n := len(data)
	padLen := filterLen - 1

	switch mode {
	case ModeSymmetric:
		extended := make([]float64, n+2*padLen)
		// Left padding (symmetric)
		for i := 0; i < padLen; i++ {
			idx := padLen - i - 1
			if idx < n {
				extended[i] = data[idx]
			}
		}
		// Copy original data
		copy(extended[padLen:], data)
		// Right padding (symmetric)
		for i := 0; i < padLen; i++ {
			idx := n - 1 - i
			if idx >= 0 {
				extended[padLen+n+i] = data[idx]
			}
		}
		return extended

	case ModePeriodic:
		extended := make([]float64, n+2*padLen)
		// Left padding (periodic)
		for i := 0; i < padLen; i++ {
			extended[i] = data[n-padLen+i]
		}
		// Copy original data
		copy(extended[padLen:], data)
		// Right padding (periodic)
		for i := 0; i < padLen; i++ {
			extended[padLen+n+i] = data[i]
		}
		return extended

	default: // ModeZero
		extended := make([]float64, n+2*padLen)
		copy(extended[padLen:], data)
		return extended
	}
}

// convolveDownsample performs convolution and downsampling by 2
func convolveDownsample(signal, filter []float64) []float64 {
	n := len(signal)
	m := len(filter)
	convLen := n - m + 1

	result := make([]float64, 0, convLen/2+1)

	for i := 0; i < convLen; i += 2 {
		sum := 0.0
		for j := 0; j < m; j++ {
			sum += signal[i+j] * filter[m-1-j]
		}
		result = append(result, sum)
	}

	return result
}

// upsampleConvolve performs upsampling by 2 and convolution
func upsampleConvolve(signal, filter []float64) []float64 {
	n := len(signal)
	m := len(filter)

	// Upsample by inserting zeros
	upsampled := make([]float64, n*2)
	for i := 0; i < n; i++ {
		upsampled[i*2] = signal[i]
	}

	// Convolve
	resultLen := len(upsampled) + m - 1
	result := make([]float64, resultLen)

	for i := 0; i < resultLen; i++ {
		sum := 0.0
		for j := 0; j < m; j++ {
			idx := i - j
			if idx >= 0 && idx < len(upsampled) {
				sum += upsampled[idx] * filter[j]
			}
		}
		result[i] = sum
	}

	return result
}

// WaveletCoefficients represents the coefficients from wavelet decomposition
type WaveletCoefficients struct {
	Approximation []float64
	Details       [][]float64
	WaveletName   string
	Level         int
}

// Decompose is a convenience function that performs wavelet decomposition
func Decompose(data []float64, waveletName string, level int) (*WaveletCoefficients, error) {
	coeffs, err := Wavedec(data, waveletName, level, ModeSymmetric)
	if err != nil {
		return nil, err
	}

	wc := &WaveletCoefficients{
		Approximation: coeffs[0],
		Details:       coeffs[1:],
		WaveletName:   waveletName,
		Level:         level,
	}

	return wc, nil
}

// Reconstruct reconstructs the signal from wavelet coefficients
func (wc *WaveletCoefficients) Reconstruct() ([]float64, error) {
	coeffs := make([][]float64, wc.Level+1)
	coeffs[0] = wc.Approximation
	copy(coeffs[1:], wc.Details)

	return Waverec(coeffs, wc.WaveletName)
}

// GetEnergy calculates the energy of a signal
func GetEnergy(signal []float64) float64 {
	energy := 0.0
	for _, v := range signal {
		energy += v * v
	}
	return energy
}

// GetMAD calculates the Median Absolute Deviation
func GetMAD(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	// Calculate absolute values
	absData := make([]float64, len(data))
	for i, v := range data {
		absData[i] = math.Abs(v)
	}

	// Sort to find median
	sorted := make([]float64, len(absData))
	copy(sorted, absData)
	quickSort(sorted, 0, len(sorted)-1)

	median := sorted[len(sorted)/2]
	return median / 0.6745 // Normalized MAD
}

// quickSort is a helper function to sort float64 slices
func quickSort(arr []float64, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func partition(arr []float64, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
