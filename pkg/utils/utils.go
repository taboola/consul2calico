package utils

import (
	"github.com/cenkalti/backoff"
	"os"
	"time"
)
// GetBackOff will return a backoff for running Ops that require retry
func GetBackOff() backoff.BackOff{
	constBackOff := backoff.ConstantBackOff{Interval: 15 * time.Second}
	backOff := backoff.WithMaxRetries(&constBackOff,3)
	return backOff
}

// GetEnv get key environment variable if exist otherwise return defalutValue
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// CompareSlice
// This function will return 2 slices with give 2 slices a,b
// slice ra : elements that exist in a and NOT in b
// slice rb : elements that exist in b and NOT in a
func CompareSlice(a []string, b []string)  ([]string,[]string){
	var ra []string
	var rb []string
	var m = make(map[string]int)

	for _,valA := range a {
		m[valA] = 1
	}

	for _,valB := range b {
		//Check if key exist in map
		_, ok := m[valB]
		if ok {
			//valB is both in slice a and in slice b  - remove from returned slice .
			delete(m, valB)
		}else {
			//Add to returned slice
			rb = append(rb,valB)
		}
	}
	for key := range m {
		ra = append(ra, key)
	}
	return ra,rb
}