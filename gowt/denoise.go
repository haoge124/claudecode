package gowt

import (
	"fmt"
	"math"
)

// ThresholdMode represents the thresholding mode
type ThresholdMode string

const (
	// ThresholdSoft applies soft thresholding
	ThresholdSoft ThresholdMode = "soft"
	// ThresholdHard applies hard thresholding
	ThresholdHard ThresholdMode = "hard"
)

// ThresholdMethod represents the method for calculating threshold
type ThresholdMethod string

const (
	// VisuShrink uses universal threshold (sigma * sqrt(2*log(n)))
	VisuShrink ThresholdMethod = "visushrink"
	// BayesShrink uses Bayes threshold estimation
	BayesShrink ThresholdMethod = "bayeshrink"
	// Manual allows manual threshold specification
	Manual ThresholdMethod = "manual"
)

// DenoiseConfig holds configuration for wavelet denoising
type DenoiseConfig struct {
	Wavelet         string
	Level           int
	ThresholdMode   ThresholdMode
	ThresholdMethod ThresholdMethod
	ManualThreshold float64
}

// DefaultDenoiseConfig returns default denoising configuration
func DefaultDenoiseConfig() *DenoiseConfig {
	return &DenoiseConfig{
		Wavelet:         "db4",
		Level:           3,
		ThresholdMode:   ThresholdSoft,
		ThresholdMethod: VisuShrink,
	}
}

// Denoise performs wavelet denoising on the signal
func Denoise(signal []float64, config *DenoiseConfig) ([]float64, error) {
	if config == nil {
		config = DefaultDenoiseConfig()
	}

	// Decompose signal
	wc, err := Decompose(signal, config.Wavelet, config.Level)
	if err != nil {
		return nil, err
	}

	// Apply thresholding to detail coefficients
	for i := range wc.Details {
		var threshold float64

		switch config.ThresholdMethod {
		case VisuShrink:
			threshold = calculateVisuShrinkThreshold(wc.Details[i])
		case BayesShrink:
			threshold = calculateBayesShrinkThreshold(wc.Details[i])
		case Manual:
			threshold = config.ManualThreshold
		default:
			threshold = calculateVisuShrinkThreshold(wc.Details[i])
		}

		wc.Details[i] = applyThreshold(wc.Details[i], threshold, config.ThresholdMode)
	}

	// Reconstruct signal
	denoised, err := wc.Reconstruct()
	if err != nil {
		return nil, err
	}

	// Ensure output length matches input length
	if len(denoised) > len(signal) {
		denoised = denoised[:len(signal)]
	} else if len(denoised) < len(signal) {
		// Pad if necessary
		padded := make([]float64, len(signal))
		copy(padded, denoised)
		denoised = padded
	}

	return denoised, nil
}

// calculateVisuShrinkThreshold calculates universal threshold
func calculateVisuShrinkThreshold(coeffs []float64) float64 {
	n := float64(len(coeffs))
	sigma := estimateNoiseSigma(coeffs)
	threshold := sigma * math.Sqrt(2*math.Log(n))
	return threshold
}

// calculateBayesShrinkThreshold calculates Bayes threshold
func calculateBayesShrinkThreshold(coeffs []float64) float64 {
	sigma := estimateNoiseSigma(coeffs)
	sigmaY := calculateStdDev(coeffs)

	// Calculate signal variance
	sigmaX := math.Sqrt(math.Max(0, sigmaY*sigmaY-sigma*sigma))

	if sigmaX == 0 {
		return sigma * math.Sqrt(2*math.Log(float64(len(coeffs))))
	}

	threshold := (sigma * sigma) / sigmaX
	return threshold
}

// estimateNoiseSigma estimates noise standard deviation using MAD
func estimateNoiseSigma(coeffs []float64) float64 {
	return GetMAD(coeffs)
}

// calculateStdDev calculates standard deviation
func calculateStdDev(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	mean := 0.0
	for _, v := range data {
		mean += v
	}
	mean /= float64(len(data))

	variance := 0.0
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data))

	return math.Sqrt(variance)
}

// applyThreshold applies threshold to coefficients
func applyThreshold(coeffs []float64, threshold float64, mode ThresholdMode) []float64 {
	result := make([]float64, len(coeffs))

	for i, c := range coeffs {
		switch mode {
		case ThresholdSoft:
			result[i] = softThreshold(c, threshold)
		case ThresholdHard:
			result[i] = hardThreshold(c, threshold)
		default:
			result[i] = softThreshold(c, threshold)
		}
	}

	return result
}

// softThreshold applies soft thresholding
func softThreshold(value, threshold float64) float64 {
	absValue := math.Abs(value)
	if absValue <= threshold {
		return 0
	}
	if value > 0 {
		return value - threshold
	}
	return value + threshold
}

// hardThreshold applies hard thresholding
func hardThreshold(value, threshold float64) float64 {
	if math.Abs(value) <= threshold {
		return 0
	}
	return value
}

// DenoiseSimple performs simple wavelet denoising with default parameters
func DenoiseSimple(signal []float64, waveletName string, level int) ([]float64, error) {
	config := &DenoiseConfig{
		Wavelet:         waveletName,
		Level:           level,
		ThresholdMode:   ThresholdSoft,
		ThresholdMethod: VisuShrink,
	}
	return Denoise(signal, config)
}

// AddNoise adds Gaussian noise to a signal
func AddNoise(signal []float64, noiseLevel float64, seed int64) []float64 {
	noisy := make([]float64, len(signal))

	// Simple linear congruential generator for reproducible noise
	rng := seed
	for i := range signal {
		// Box-Muller transform for Gaussian noise
		rng = (rng*1103515245 + 12345) & 0x7fffffff
		u1 := float64(rng) / float64(0x7fffffff)

		rng = (rng*1103515245 + 12345) & 0x7fffffff
		u2 := float64(rng) / float64(0x7fffffff)

		noise := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2) * noiseLevel
		noisy[i] = signal[i] + noise
	}

	return noisy
}

// CalculateSNR calculates Signal-to-Noise Ratio in dB
func CalculateSNR(original, noisy []float64) (float64, error) {
	if len(original) != len(noisy) {
		return 0, fmt.Errorf("signals must have the same length")
	}

	signalPower := 0.0
	noisePower := 0.0

	for i := range original {
		signalPower += original[i] * original[i]
		noise := noisy[i] - original[i]
		noisePower += noise * noise
	}

	if noisePower == 0 {
		return math.Inf(1), nil
	}

	snr := 10 * math.Log10(signalPower/noisePower)
	return snr, nil
}

// CalculateRMSE calculates Root Mean Square Error
func CalculateRMSE(original, reconstructed []float64) (float64, error) {
	if len(original) != len(reconstructed) {
		return 0, fmt.Errorf("signals must have the same length")
	}

	mse := 0.0
	for i := range original {
		diff := original[i] - reconstructed[i]
		mse += diff * diff
	}
	mse /= float64(len(original))

	return math.Sqrt(mse), nil
}
