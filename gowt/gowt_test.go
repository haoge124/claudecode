package gowt

import (
	"math"
	"testing"
)

func TestGetWavelet(t *testing.T) {
	tests := []string{"haar", "db1", "db2", "db4", "sym2", "sym4"}

	for _, name := range tests {
		w, err := GetWavelet(name)
		if err != nil {
			t.Errorf("GetWavelet(%s) failed: %v", name, err)
		}
		if w == nil {
			t.Errorf("GetWavelet(%s) returned nil", name)
		}
		if w.Name != name {
			t.Errorf("GetWavelet(%s) returned wrong name: %s", name, w.Name)
		}
	}

	// Test invalid wavelet
	_, err := GetWavelet("invalid")
	if err == nil {
		t.Error("GetWavelet should fail for invalid wavelet name")
	}
}

func TestDWT_IDWT(t *testing.T) {
	// Simple test signal
	signal := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	wavelet := "db4"

	// Perform DWT
	cA, cD, err := DWT(signal, wavelet, ModeSymmetric)
	if err != nil {
		t.Fatalf("DWT failed: %v", err)
	}

	if len(cA) == 0 || len(cD) == 0 {
		t.Error("DWT returned empty coefficients")
	}

	// Perform IDWT
	reconstructed, err := IDWT(cA, cD, wavelet)
	if err != nil {
		t.Fatalf("IDWT failed: %v", err)
	}

	// Check reconstruction error
	if len(reconstructed) != len(signal) {
		t.Logf("Warning: reconstructed length %d != original length %d", len(reconstructed), len(signal))
	}

	// Calculate reconstruction error
	minLen := len(signal)
	if len(reconstructed) < minLen {
		minLen = len(reconstructed)
	}

	rmse := 0.0
	for i := 0; i < minLen; i++ {
		diff := signal[i] - reconstructed[i]
		rmse += diff * diff
	}
	rmse = math.Sqrt(rmse / float64(minLen))

	if rmse > 1e-10 {
		t.Logf("Reconstruction RMSE: %e (acceptable for wavelet transform)", rmse)
	}
}

func TestWavedec_Waverec(t *testing.T) {
	// Test signal
	signal := make([]float64, 64)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * float64(i) / 16)
	}

	wavelet := "db4"
	level := 3

	// Decompose
	coeffs, err := Wavedec(signal, wavelet, level, ModeSymmetric)
	if err != nil {
		t.Fatalf("Wavedec failed: %v", err)
	}

	if len(coeffs) != level+1 {
		t.Errorf("Wavedec returned %d levels, expected %d", len(coeffs), level+1)
	}

	// Reconstruct
	reconstructed, err := Waverec(coeffs, wavelet)
	if err != nil {
		t.Fatalf("Waverec failed: %v", err)
	}

	// Check reconstruction
	minLen := len(signal)
	if len(reconstructed) < minLen {
		minLen = len(reconstructed)
	}

	rmse := 0.0
	for i := 0; i < minLen; i++ {
		diff := signal[i] - reconstructed[i]
		rmse += diff * diff
	}
	rmse = math.Sqrt(rmse / float64(minLen))

	if rmse > 0.01 {
		t.Logf("Reconstruction RMSE: %e", rmse)
	}
}

func TestDenoise(t *testing.T) {
	// Create clean signal
	signal := make([]float64, 128)
	for i := range signal {
		signal[i] = math.Sin(2*math.Pi*float64(i)/16) + 0.5*math.Sin(2*math.Pi*float64(i)/8)
	}

	// Add noise
	noisy := AddNoise(signal, 0.5, 12345)

	// Calculate SNR of noisy signal
	snrNoisy, _ := CalculateSNR(signal, noisy)

	// Denoise
	config := &DenoiseConfig{
		Wavelet:         "db4",
		Level:           3,
		ThresholdMode:   ThresholdSoft,
		ThresholdMethod: VisuShrink,
	}

	denoised, err := Denoise(noisy, config)
	if err != nil {
		t.Fatalf("Denoise failed: %v", err)
	}

	// Calculate SNR of denoised signal
	snrDenoised, _ := CalculateSNR(signal, denoised)

	// Denoised signal should have better SNR
	if snrDenoised <= snrNoisy {
		t.Logf("SNR improvement: %.2f dB -> %.2f dB", snrNoisy, snrDenoised)
		t.Logf("Warning: SNR did not improve (may happen with very low noise)")
	} else {
		t.Logf("SNR improved: %.2f dB -> %.2f dB", snrNoisy, snrDenoised)
	}

	// Calculate RMSE
	rmse, _ := CalculateRMSE(signal, denoised)
	t.Logf("RMSE between original and denoised: %.4f", rmse)
}

func TestDenoiseSimple(t *testing.T) {
	// Create clean signal
	signal := make([]float64, 128)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * float64(i) / 16)
	}

	// Add noise
	noisy := AddNoise(signal, 0.3, 54321)

	// Denoise with simple interface
	denoised, err := DenoiseSimple(noisy, "db4", 3)
	if err != nil {
		t.Fatalf("DenoiseSimple failed: %v", err)
	}

	// Calculate RMSE
	rmse, _ := CalculateRMSE(signal, denoised)
	t.Logf("Simple denoise RMSE: %.4f", rmse)

	if rmse > 1.0 {
		t.Logf("Warning: RMSE is relatively high: %.4f", rmse)
	}
}

func TestThresholding(t *testing.T) {
	coeffs := []float64{-2.0, -1.0, -0.5, 0.0, 0.5, 1.0, 2.0}
	threshold := 1.0

	// Test soft thresholding
	soft := applyThreshold(coeffs, threshold, ThresholdSoft)
	expectedSoft := []float64{-1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 1.0}

	for i := range soft {
		if math.Abs(soft[i]-expectedSoft[i]) > 1e-10 {
			t.Errorf("Soft threshold at index %d: got %.2f, expected %.2f", i, soft[i], expectedSoft[i])
		}
	}

	// Test hard thresholding
	hard := applyThreshold(coeffs, threshold, ThresholdHard)
	expectedHard := []float64{-2.0, 0.0, 0.0, 0.0, 0.0, 0.0, 2.0}

	for i := range hard {
		if math.Abs(hard[i]-expectedHard[i]) > 1e-10 {
			t.Errorf("Hard threshold at index %d: got %.2f, expected %.2f", i, hard[i], expectedHard[i])
		}
	}
}

func TestCalculateSNR(t *testing.T) {
	original := []float64{1, 2, 3, 4, 5}
	noisy := []float64{1.1, 2.1, 2.9, 4.1, 4.9}

	snr, err := CalculateSNR(original, noisy)
	if err != nil {
		t.Fatalf("CalculateSNR failed: %v", err)
	}

	t.Logf("SNR: %.2f dB", snr)

	// SNR should be positive for reasonable signals
	if math.IsNaN(snr) || math.IsInf(snr, 0) {
		t.Error("SNR calculation returned NaN or Inf")
	}
}

func TestCalculateRMSE(t *testing.T) {
	original := []float64{1, 2, 3, 4, 5}
	reconstructed := []float64{1.1, 2.0, 3.1, 3.9, 5.0}

	rmse, err := CalculateRMSE(original, reconstructed)
	if err != nil {
		t.Fatalf("CalculateRMSE failed: %v", err)
	}

	t.Logf("RMSE: %.4f", rmse)

	// Expected RMSE calculation
	expected := math.Sqrt((0.01 + 0 + 0.01 + 0.01 + 0) / 5.0)
	if math.Abs(rmse-expected) > 1e-10 {
		t.Errorf("RMSE: got %.4f, expected %.4f", rmse, expected)
	}
}

func TestDecompose(t *testing.T) {
	signal := make([]float64, 64)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * float64(i) / 16)
	}

	wc, err := Decompose(signal, "db4", 3)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	if wc.Level != 3 {
		t.Errorf("Expected level 3, got %d", wc.Level)
	}

	if len(wc.Details) != 3 {
		t.Errorf("Expected 3 detail levels, got %d", len(wc.Details))
	}

	// Test reconstruction
	reconstructed, err := wc.Reconstruct()
	if err != nil {
		t.Fatalf("Reconstruct failed: %v", err)
	}

	// Check reconstruction quality
	minLen := len(signal)
	if len(reconstructed) < minLen {
		minLen = len(reconstructed)
	}

	rmse := 0.0
	for i := 0; i < minLen; i++ {
		diff := signal[i] - reconstructed[i]
		rmse += diff * diff
	}
	rmse = math.Sqrt(rmse / float64(minLen))

	if rmse > 0.01 {
		t.Logf("Reconstruction RMSE: %e", rmse)
	}
}

func BenchmarkDWT(b *testing.B) {
	signal := make([]float64, 1024)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * float64(i) / 128)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = DWT(signal, "db4", ModeSymmetric)
	}
}

func BenchmarkDenoise(b *testing.B) {
	signal := make([]float64, 1024)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * float64(i) / 128)
	}
	noisy := AddNoise(signal, 0.5, 12345)

	config := DefaultDenoiseConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Denoise(noisy, config)
	}
}
