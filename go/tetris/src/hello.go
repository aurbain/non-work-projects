package main

import (
	"fmt"
	"math"
)

func main() {
	const limit = 1000000

	// Initialize a slice of booleans; true means potentially prime
	isPrime := make([]bool, limit+1)
	for i := 2; i <= limit; i++ {
		isPrime[i] = true
	}

	// Sieve of Eratosthenes
	sqrtLimit := int(math.Sqrt(float64(limit)))
	for p := 2; p <= sqrtLimit; p++ {
		if isPrime[p] {
			// Mark all multiples of p starting from p*p as composite
			for i := p * p; i <= limit; i += p {
				isPrime[i] = false
			}
		}
	}

	// Output all primes
	for i := 2; i <= limit; i++ {
		if isPrime[i] {
			fmt.Printf("%d is prime\n", i)
		}
	}
}
