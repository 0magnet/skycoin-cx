package benchmarks

func ackermann(m, n uint) uint {
	switch uint(0) {
	case m:
		return n + uint(1)
	case n:
		return ackermann(m-uint(1), uint(1))
	}
	return ackermann(m-uint(1), ackermann(m, n-uint(1)))
}
