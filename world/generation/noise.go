package generation

import "math"

func fade(t float64) float64 {
	return t * t * (3 - 2*t)
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Hash2D — детерминированный хеш по координатам и seed (экспорт для world).
func Hash2D(x, y, seed int) int {
	h := x*374761393 + y*668265263 + seed*69069
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	if h < 0 {
		h = -h
	}
	return h
}

func randomValue2D(x, y, seed int) float64 {
	return float64(Hash2D(x, y, seed)%10000) / 10000.0
}

func valueNoise2D(x, y float64, seed int) float64 {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1
	sx := x - float64(x0)
	sy := y - float64(y0)
	n00 := randomValue2D(x0, y0, seed)
	n10 := randomValue2D(x1, y0, seed)
	n01 := randomValue2D(x0, y1, seed)
	n11 := randomValue2D(x1, y1, seed)
	ux := fade(sx)
	uy := fade(sy)
	ix0 := lerp(n00, n10, ux)
	ix1 := lerp(n01, n11, ux)
	return lerp(ix0, ix1, uy)
}

// FractalNoise2D — фрактальный шум (экспорт для generation.go).
func FractalNoise2D(x, y float64, seed, octaves int, persistence, lacunarity float64) float64 {
	total := 0.0
	amplitude := 1.0
	frequency := 1.0
	maxAmplitude := 0.0
	for i := 0; i < octaves; i++ {
		n := valueNoise2D(x*frequency, y*frequency, seed+i*101)
		total += n * amplitude
		maxAmplitude += amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}
	if maxAmplitude == 0 {
		return 0
	}
	return total / maxAmplitude
}
