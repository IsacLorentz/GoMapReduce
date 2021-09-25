package main

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
	"sync"
	"time"
)

const DataFile = "loremipsum.txt"

// Return the word frequencies of the text argument.
//
// Split load optimally across processor cores.
func WordCount(text string) map[string]int {

	numCPU := runtime.NumCPU()
	var splittedStrings []string
	first :=0
	last := findSpaceInString(text, len(text)/numCPU)

	for i := 0; i < numCPU; i++ {
		if i+1 == numCPU {
			last = len(text)
		}

		splittedStrings = append(splittedStrings, text[first:last])
		first, last = last, findSpaceInString(text, last + len(text)/numCPU)
	}
	results := make(chan map[string]int)
	finalVal := make(chan map[string]int)

	var wg sync.WaitGroup

	wg.Add(len(splittedStrings))

	for _, subStr := range splittedStrings {
		go func(aSubString string) {
			defer wg.Done()
			results <- Map(aSubString)
		}(subStr)
	}

	go reduce(results, finalVal)

	wg.Wait()
	close(results)

	freqs := <-finalVal

	return freqs
}

func findSpaceInString(str string, index int) int  {
	for ; index < len(str); index++{
		if str[index] == ' ' {
			return index
		}
	}
	return index
}

func Map(aSubString string) map[string]int {
	aSubString = strings.ToLower(aSubString)
	aSubString = strings.ReplaceAll(aSubString, ".", "")
	aSubString = strings.ReplaceAll(aSubString, ",", "")
	aSubList := strings.Fields(aSubString)
	freqs := make(map[string]int)
	for _, val := range aSubList {
		freqs[val]++
	}
	return freqs
}

func reduce(mapList chan map[string]int, sendFinalVal chan map[string]int) {
	final := map[string]int{}
	for aMap := range mapList {
		for key, count := range aMap {
			final[key] += count
		}
	}
	sendFinalVal <- final
}

// Benchmark how long it takes to count word frequencies in text numRuns times.
//
// Return the total time elapsed.
func benchmark(text string, numRuns int) int64 {
	start := time.Now()
	for i := 0; i < numRuns; i++ {
		WordCount(text)
	}
	runtimeMillis := time.Since(start).Nanoseconds() / 1e6

	return runtimeMillis
}

// Print the results of a benchmark
func printResults(runtimeMillis int64, numRuns int) {
	fmt.Printf("amount of runs: %d\n", numRuns)
	fmt.Printf("total time: %d ms\n", runtimeMillis)
	average := float64(runtimeMillis) / float64(numRuns)
	fmt.Printf("average time/run: %.2f ms\n", average)
}

func main() {
	// read in DataFile as a string called data
	data, _ := ioutil.ReadFile(DataFile)

	fmt.Printf("%#v", WordCount(string(data)))

	numRuns := 100
	runtimeMillis := benchmark(string(data), numRuns)
	printResults(runtimeMillis, numRuns)
}
