# GoWT - Go Wavelet Transform Library

GoWT 是一个 Go 语言实现的小波变换库，提供类似于 Python pywt 库的功能，特别是小波降噪功能。

## 特性

- **离散小波变换 (DWT)** - 单层和多层小波分解
- **逆小波变换 (IDWT)** - 信号重构
- **小波降噪** - 支持多种阈值方法和模式
- **多种小波基** - Haar, Daubechies (db1-db4), Symlets (sym2, sym4)
- **阈值方法** - VisuShrink, BayesShrink
- **阈值模式** - 软阈值、硬阈值
- **信号质量评估** - SNR, RMSE 计算

## 安装

```bash
go get github.com/haoge124/gowt
```

## 快速开始

### 基本小波变换

```go
package main

import (
    "fmt"
    "github.com/haoge124/gowt"
)

func main() {
    // 创建信号
    signal := []float64{1, 2, 3, 4, 5, 6, 7, 8}

    // 执行小波变换
    cA, cD, err := gowt.DWT(signal, "db4", gowt.ModeSymmetric)
    if err != nil {
        panic(err)
    }

    // 逆变换重构信号
    reconstructed, err := gowt.IDWT(cA, cD, "db4")
    if err != nil {
        panic(err)
    }

    fmt.Println("重构信号:", reconstructed)
}
```

### 小波降噪

```go
package main

import (
    "fmt"
    "math"
    "github.com/haoge124/gowt"
)

func main() {
    // 创建干净信号
    signal := make([]float64, 256)
    for i := range signal {
        signal[i] = math.Sin(2 * math.Pi * float64(i) / 32)
    }

    // 添加噪声
    noisy := gowt.AddNoise(signal, 0.5, 12345)

    // 小波降噪
    config := &gowt.DenoiseConfig{
        Wavelet:         "db4",
        Level:           4,
        ThresholdMode:   gowt.ThresholdSoft,
        ThresholdMethod: gowt.VisuShrink,
    }

    denoised, err := gowt.Denoise(noisy, config)
    if err != nil {
        panic(err)
    }

    // 计算性能指标
    snr, _ := gowt.CalculateSNR(signal, denoised)
    rmse, _ := gowt.CalculateRMSE(signal, denoised)

    fmt.Printf("SNR: %.2f dB\n", snr)
    fmt.Printf("RMSE: %.4f\n", rmse)
}
```

### 简化降噪接口

```go
// 使用默认参数快速降噪
denoised, err := gowt.DenoiseSimple(noisySignal, "db4", 3)
```

## API 文档

### 小波变换函数

#### DWT
```go
func DWT(data []float64, waveletName string, mode Mode) ([]float64, []float64, error)
```
执行单层离散小波变换。

**参数:**
- `data`: 输入信号
- `waveletName`: 小波名称 ("haar", "db1", "db2", "db4", "sym2", "sym4")
- `mode`: 边界扩展模式 (ModeZero, ModeSymmetric, ModePeriodic)

**返回:**
- 近似系数 (cA)
- 细节系数 (cD)
- 错误信息

#### IDWT
```go
func IDWT(cA, cD []float64, waveletName string) ([]float64, error)
```
执行单层逆离散小波变换。

#### Wavedec
```go
func Wavedec(data []float64, waveletName string, level int, mode Mode) ([][]float64, error)
```
执行多层小波分解。

#### Waverec
```go
func Waverec(coeffs [][]float64, waveletName string) ([]float64, error)
```
执行多层小波重构。

### 降噪函数

#### Denoise
```go
func Denoise(signal []float64, config *DenoiseConfig) ([]float64, error)
```
使用小波变换对信号进行降噪。

**DenoiseConfig 结构:**
```go
type DenoiseConfig struct {
    Wavelet         string           // 小波名称
    Level           int              // 分解层数
    ThresholdMode   ThresholdMode    // 阈值模式
    ThresholdMethod ThresholdMethod  // 阈值计算方法
    ManualThreshold float64          // 手动阈值（仅用于 Manual 方法）
}
```

**阈值模式:**
- `ThresholdSoft`: 软阈值
- `ThresholdHard`: 硬阈值

**阈值方法:**
- `VisuShrink`: 通用阈值 σ√(2log(n))
- `BayesShrink`: 贝叶斯阈值估计
- `Manual`: 手动指定阈值

#### DenoiseSimple
```go
func DenoiseSimple(signal []float64, waveletName string, level int) ([]float64, error)
```
使用默认参数进行降噪（软阈值 + VisuShrink）。

### 辅助函数

#### CalculateSNR
```go
func CalculateSNR(original, noisy []float64) (float64, error)
```
计算信噪比 (dB)。

#### CalculateRMSE
```go
func CalculateRMSE(original, reconstructed []float64) (float64, error)
```
计算均方根误差。

#### AddNoise
```go
func AddNoise(signal []float64, noiseLevel float64, seed int64) []float64
```
向信号添加高斯噪声。

#### GetWavelet
```go
func GetWavelet(name string) (*Wavelet, error)
```
获取指定名称的小波。

#### ListWavelets
```go
func ListWavelets() []string
```
列出所有可用的小波名称。

## 支持的小波

- **Haar**: `"haar"`, `"db1"`
- **Daubechies**: `"db2"`, `"db4"`
- **Symlets**: `"sym2"`, `"sym4"`

## 示例程序

运行示例程序：

```bash
cd gowt/example
go run main.go
```

示例程序包含：
1. 基本小波变换
2. 小波降噪
3. 多层分解
4. 不同小波对比
5. 阈值方法对比

## 测试

运行测试：

```bash
cd gowt
go test -v
```

运行基准测试：

```bash
go test -bench=.
```

## 性能

在现代 CPU 上，处理 1024 点信号的性能：
- DWT: ~50-100 µs
- 降噪: ~500-1000 µs

## 与 Python pywt 的对比

| 功能 | Python pywt | GoWT |
|------|-------------|------|
| DWT/IDWT | ✓ | ✓ |
| 多层分解 | ✓ | ✓ |
| 降噪 | ✓ | ✓ |
| 小波基 | 50+ | 6 (常用) |
| 性能 | 中 | 高 |
| 类型安全 | 弱 | 强 |

## 应用场景

- 信号降噪
- 数据压缩
- 特征提取
- 时频分析
- 图像处理（一维信号）

## 算法原理

### 小波变换
小波变换将信号分解为不同频率的成分，类似于傅里叶变换，但提供了更好的时频局部化。

### 降噪流程
1. **分解**: 使用小波变换将信号分解为多个层次
2. **阈值处理**: 对细节系数应用阈值，去除小幅度（噪声）系数
3. **重构**: 从处理后的系数重构信号

### 阈值选择
- **VisuShrink**: 保守的通用阈值，适用于未知噪声水平
- **BayesShrink**: 基于贝叶斯估计的自适应阈值，通常效果更好

## 许可证

MIT License

## 作者

GoWT 是 Python pywt 库的 Go 语言实现版本。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 参考文献

- Mallat, S. (1989). A theory for multiresolution signal decomposition: the wavelet representation.
- Donoho, D. L., & Johnstone, I. M. (1994). Ideal spatial adaptation by wavelet shrinkage.
- PyWavelets Documentation: https://pywavelets.readthedocs.io/
