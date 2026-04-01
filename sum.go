package example

// Calculator defines basic arithmetic operations.
type Calculator interface {
	Sum(a, b int) int
	Diff(a, b int) int
}

//go:generate go tool mockgen -source=sum.go -destination=sum_mock.go -package=example

// Summarize returns the sum of all values using the provided Calculator.
func Summarize(c Calculator, values ...int) int {
	result := 0
	for _, v := range values {
		result = c.Sum(result, v)
	}
	return result
}
