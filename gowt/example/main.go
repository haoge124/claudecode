package main

import (
	"fmt"
	"math"

	"github.com/haoge124/gowt"
)

func main() {
	fmt.Println("=== GoWT - Go Wavelet Transform Library ===")
	fmt.Println()

	// Example 1: Basic Wavelet Transform
	fmt.Println("Example 1: Basic Wavelet Transform")
	basicWaveletTransform()
	fmt.Println()

	// Example 2: Wavelet Denoising
	fmt.Println("Example 2: Wavelet Denoising")
	waveletDenoising()
	fmt.Println()

	// Example 3: Multi-level Decomposition
	fmt.Println("Example 3: Multi-level Decomposition")
	multiLevelDecomposition()
	fmt.Println()

	// Example 4: Comparing Different Wavelets
	fmt.Println("Example 4: Comparing Different Wavelets")
	compareWavelets()
	fmt.Println()

	// Example 5: Threshold Methods Comparison
	fmt.Println("Example 5: Threshold Methods Comparison")
	compareThresholdMethods()
}

func basicWaveletTransform() {
	// Create a simple signal
	signal := make([]float64, 64)
	for i := range signal {
		signal[i] = math.Sin(2*math.Pi*float64(i)/16) + 0.5*math.Cos(2*math.Pi*float64(i)/8)
	}

	fmt.Printf("Original signal length: %d\n", len(signal))

	// Perform DWT
	cA, cD, err := gowt.DWT(signal, "db4", gowt.ModeSymmetric)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Approximation coefficients length: %d\n", len(cA))
	fmt.Printf("Detail coefficients length: %d\n", len(cD))

	// Reconstruct
	reconstructed, err := gowt.IDWT(cA, cD, "db4")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Calculate reconstruction error
	rmse, _ := gowt.CalculateRMSE(signal, reconstructed)
	fmt.Printf("Reconstruction RMSE: %.6f\n", rmse)
}

func waveletDenoising() {
	// Create a clean signal (sine wave)
	signal := make([]float64, 256)
	for i := range signal {
		signal[i] = math.Sin(2*math.Pi*float64(i)/32) + 0.5*math.Sin(2*math.Pi*float64(i)/16)
	}

	fmt.Printf("Clean signal length: %d\n", len(signal))

	// Add Gaussian noise
	noiseLevel := 0.5
	noisy := gowt.AddNoise(signal, noiseLevel, 12345)

	// Calculate SNR of noisy signal
	snrNoisy, _ := gowt.CalculateSNR(signal, noisy)
	rmseNoisy, _ := gowt.CalculateRMSE(signal, noisy)
	fmt.Printf("Noisy signal - SNR: %.2f dB, RMSE: %.4f\n", snrNoisy, rmseNoisy)

	// Denoise using wavelet transform
	config := &gowt.DenoiseConfig{
		Wavelet:         "db4",
		Level:           4,
		ThresholdMode:   gowt.ThresholdSoft,
		ThresholdMethod: gowt.VisuShrink,
	}

	denoised, err := gowt.Denoise(noisy, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Calculate metrics for denoised signal
	snrDenoised, _ := gowt.CalculateSNR(signal, denoised)
	rmseDenoised, _ := gowt.CalculateRMSE(signal, denoised)
	fmt.Printf("Denoised signal - SNR: %.2f dB, RMSE: %.4f\n", snrDenoised, rmseDenoised)
	fmt.Printf("SNR improvement: %.2f dB\n", snrDenoised-snrNoisy)
	fmt.Printf("RMSE reduction: %.2f%%\n", (rmseNoisy-rmseDenoised)/rmseNoisy*100)
}

func multiLevelDecomposition() {
	// Create a signal with multiple frequency components
	signal := make([]float64, 512)
	for i := range signal {
		t := float64(i)
		signal[i] = math.Sin(2*math.Pi*t/64) + // Low frequency
			0.5*math.Sin(2*math.Pi*t/16) + // Medium frequency
			0.25*math.Sin(2*math.Pi*t/4) // High frequency
	}

	// Decompose using multiple levels
	wc, err := gowt.Decompose(signal, "db4", 5)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Decomposition level: %d\n", wc.Level)
	fmt.Printf("Approximation coefficients: %d\n", len(wc.Approximation))
	for i, detail := range wc.Details {
		energy := gowt.GetEnergy(detail)
		fmt.Printf("Detail level %d: length=%d, energy=%.2f\n", i+1, len(detail), energy)
	}

	// Reconstruct
	reconstructed, err := wc.Reconstruct()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	rmse, _ := gowt.CalculateRMSE(signal, reconstructed)
	fmt.Printf("Reconstruction RMSE: %.6f\n", rmse)
}

func compareWavelets() {
	// Create a noisy signal
	signal := make([]float64, 256)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * float64(i) / 32)
	}
	noisy := gowt.AddNoise(signal, 0.5, 54321)

	wavelets := []string{"haar", "db2", "db4", "sym2", "sym4"}

	fmt.Println("Wavelet\t\tSNR (dB)\tRMSE")
	fmt.Println("----------------------------------------")

	for _, wavelet := range wavelets {
		denoised, err := gowt.DenoiseSimple(noisy, wavelet, 3)
		if err != nil {
			fmt.Printf("%s\t\tError: %v\n", wavelet, err)
			continue
		}

		snr, _ := gowt.CalculateSNR(signal, denoised)
		rmse, _ := gowt.CalculateRMSE(signal, denoised)
		fmt.Printf("%s\t\t%.2f\t\t%.4f\n", wavelet, snr, rmse)
	}
}

func compareThresholdMethods() {
	// Create a noisy signal
	signal := make([]float64, 256)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * float64(i) / 32)
	}
	noisy := gowt.AddNoise(signal, 0.5, 99999)

	methods := []struct {
		name   string
		method gowt.ThresholdMethod
	}{
		{"VisuShrink", gowt.VisuShrink},
		{"BayesShrink", gowt.BayesShrink},
	}

	fmt.Println("Method\t\tMode\t\tSNR (dB)\tRMSE")
	fmt.Println("----------------------------------------------------")

	for _, method := range methods {
		for _, mode := range []gowt.ThresholdMode{gowt.ThresholdSoft, gowt.ThresholdHard} {
			config := &gowt.DenoiseConfig{
				Wavelet:         "db4",
				Level:           3,
				ThresholdMode:   mode,
				ThresholdMethod: method.method,
			}

			denoised, err := gowt.Denoise(noisy, config)
			if err != nil {
				fmt.Printf("%s\t%s\tError: %v\n", method.name, mode, err)
				continue
			}

			snr, _ := gowt.CalculateSNR(signal, denoised)
			rmse, _ := gowt.CalculateRMSE(signal, denoised)
			fmt.Printf("%s\t%s\t%.2f\t\t%.4f\n", method.name, mode, snr, rmse)
		}
	}
}
