package randomvars

import (
	"fmt"
	"math"
	"strings"

	"gonum.org/v1/gonum/stat/distuv"
)

// Law — поддерживаемый закон распределения непрерывной случайной величины.
type Law string

const (
	LawUniform     Law = "uniform"
	LawExponential Law = "exponential"
)

// lawDistribution — методы распределения gonum, которые нужны сервису.
type lawDistribution interface {
	CDF(float64) float64
	Quantile(float64) float64
	Mean() float64
	Variance() float64
}

// distribution — параметризация одного из законов из методички.
//
// Для равномерного закона используются A и B, для экспоненциального — Lambda.
// Неиспользуемые параметры остаются нулевыми.
type distribution struct {
	law  Law
	dist lawDistribution
}

func newDistribution(law string, a, b, lambda float64) (distribution, error) {
	d := distribution{law: Law(strings.ToLower(strings.TrimSpace(law)))}

	switch d.law {
	case LawUniform:
		if !finite(a) || !finite(b) {
			return distribution{}, fmt.Errorf("%w: a и b должны быть конечными", ErrParameter)
		}
		if a >= b {
			return distribution{}, fmt.Errorf("%w: для равномерного закона нужно a < b", ErrParameter)
		}
		d.dist = distuv.Uniform{Min: a, Max: b}
	case LawExponential:
		if !finite(lambda) || lambda <= 0 {
			return distribution{}, fmt.Errorf("%w: lambda должна быть положительной", ErrParameter)
		}
		d.dist = distuv.Exponential{Rate: lambda}
	default:
		return distribution{}, fmt.Errorf("%w: %q", ErrLaw, law)
	}

	if !finite(d.mean()) || !finite(d.variance()) {
		return distribution{}, fmt.Errorf("%w: параметры дают бесконечные моменты", ErrParameter)
	}

	return d, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func (d distribution) name() string {
	switch d.law {
	case LawUniform:
		return "Равномерный закон"
	case LawExponential:
		return "Экспоненциальный закон"
	default:
		return string(d.law)
	}
}

func (d distribution) formula() string {
	switch d.law {
	case LawUniform:
		return "y = a + (b − a)·x"
	case LawExponential:
		return "y = −ln(1 − x) / λ"
	default:
		return ""
	}
}

// inverse — обратная функция распределения G⁻¹(x), x ∈ (0, 1).
func (d distribution) inverse(x float64) float64 {
	return d.dist.Quantile(x)
}

// cdf — функция распределения G(y) = P{Y < y}.
func (d distribution) cdf(y float64) float64 {
	return d.dist.CDF(y)
}

// support возвращает нижнюю и верхнюю границы носителя. Бесконечность
// означает, что соответствующий хвост распределения не ограничен.
func (d distribution) support() (float64, float64) {
	switch dist := d.dist.(type) {
	case distuv.Uniform:
		return dist.Min, dist.Max
	case distuv.Exponential:
		return 0, math.Inf(1)
	default:
		return math.Inf(-1), math.Inf(1)
	}
}

func (d distribution) mean() float64 {
	return d.dist.Mean()
}

func (d distribution) variance() float64 {
	return d.dist.Variance()
}
