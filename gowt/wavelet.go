// Package gowt provides wavelet transform and denoising functionality similar to Python's pywt library
package gowt

import (
	"fmt"
	"math"
)

// Wavelet represents a wavelet with its filter coefficients
type Wavelet struct {
	Name          string
	DecLowFilter  []float64 // Decomposition low-pass filter
	DecHighFilter []float64 // Decomposition high-pass filter
	RecLowFilter  []float64 // Reconstruction low-pass filter
	RecHighFilter []float64 // Reconstruction high-pass filter
}

// WaveletFamily holds predefined wavelets
var wavelets = make(map[string]*Wavelet)

func init() {
	// Register common wavelets
	registerHaar()
	registerDb()
	registerSym()
}

// GetWavelet returns a wavelet by name
func GetWavelet(name string) (*Wavelet, error) {
	w, ok := wavelets[name]
	if !ok {
		return nil, fmt.Errorf("wavelet '%s' not found", name)
	}
	return w, nil
}

// registerHaar registers the Haar wavelet
func registerHaar() {
	sqrt2 := math.Sqrt(2)
	dec_lo := []float64{1 / sqrt2, 1 / sqrt2}
	dec_hi := []float64{-1 / sqrt2, 1 / sqrt2}
	rec_lo := []float64{1 / sqrt2, 1 / sqrt2}
	rec_hi := []float64{1 / sqrt2, -1 / sqrt2}

	wavelets["haar"] = &Wavelet{
		Name:          "haar",
		DecLowFilter:  dec_lo,
		DecHighFilter: dec_hi,
		RecLowFilter:  rec_lo,
		RecHighFilter: rec_hi,
	}
}

// registerDb registers Daubechies wavelets
func registerDb() {
	// db1 is the same as Haar (but with different name)
	sqrt2 := math.Sqrt(2)
	db1_dec_lo := []float64{1 / sqrt2, 1 / sqrt2}
	wavelets["db1"] = createOrthogonalWavelet("db1", db1_dec_lo)

	// db2 coefficients
	sqrt3 := math.Sqrt(3)
	dec_lo := []float64{
		(1 + sqrt3) / (4 * sqrt2),
		(3 + sqrt3) / (4 * sqrt2),
		(3 - sqrt3) / (4 * sqrt2),
		(1 - sqrt3) / (4 * sqrt2),
	}

	wavelets["db2"] = createOrthogonalWavelet("db2", dec_lo)

	// db4 coefficients
	db4_dec_lo := []float64{
		0.23037781330885523,
		0.7148465705525415,
		0.6308807679295904,
		-0.02798376941698385,
		-0.18703481171888114,
		0.030841381835986965,
		0.032883011666982945,
		-0.010597401784997278,
	}
	wavelets["db4"] = createOrthogonalWavelet("db4", db4_dec_lo)
}

// registerSym registers Symlet wavelets
func registerSym() {
	// sym2 is the same as db2 (but with different name)
	sqrt3 := math.Sqrt(3)
	sqrt2 := math.Sqrt(2)
	sym2_dec_lo := []float64{
		(1 + sqrt3) / (4 * sqrt2),
		(3 + sqrt3) / (4 * sqrt2),
		(3 - sqrt3) / (4 * sqrt2),
		(1 - sqrt3) / (4 * sqrt2),
	}
	wavelets["sym2"] = createOrthogonalWavelet("sym2", sym2_dec_lo)

	// sym4 coefficients
	sym4_dec_lo := []float64{
		-0.07576571478927333,
		-0.02963552764599851,
		0.49761866763201545,
		0.8037387518059161,
		0.29785779560527736,
		-0.09921954357684722,
		-0.012603967262037833,
		0.0322231006040427,
	}
	wavelets["sym4"] = createOrthogonalWavelet("sym4", sym4_dec_lo)
}

// createOrthogonalWavelet creates an orthogonal wavelet from decomposition low-pass filter
func createOrthogonalWavelet(name string, dec_lo []float64) *Wavelet {
	n := len(dec_lo)

	// Create high-pass decomposition filter (quadrature mirror filter)
	dec_hi := make([]float64, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			dec_hi[i] = dec_lo[n-1-i]
		} else {
			dec_hi[i] = -dec_lo[n-1-i]
		}
	}

	// For orthogonal wavelets, reconstruction filters are time-reversed decomposition filters
	rec_lo := make([]float64, n)
	rec_hi := make([]float64, n)
	for i := 0; i < n; i++ {
		rec_lo[i] = dec_lo[n-1-i]
		rec_hi[i] = dec_hi[n-1-i]
	}

	return &Wavelet{
		Name:          name,
		DecLowFilter:  dec_lo,
		DecHighFilter: dec_hi,
		RecLowFilter:  rec_lo,
		RecHighFilter: rec_hi,
	}
}

// ListWavelets returns a list of available wavelet names
func ListWavelets() []string {
	var names []string
	for name := range wavelets {
		names = append(names, name)
	}
	return names
}
