package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {
	// Отримуємо налаштування з оточення з дефолтними значеннями
	cpuLoad, _ := strconv.Atoi(getEnv("CPU_LOAD_ITERATIONS", "0"))
	memLoad, _ := strconv.Atoi(getEnv("MEM_LOAD_MB", "0"))
	maxDelay, _ := strconv.Atoi(getEnv("MAX_DELAY_MS", "0"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. Емуляція випадкової затримки
		if maxDelay > 0 {
			delay := rand.Intn(maxDelay)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		// 2. Емуляція навантаження на CPU (паралельна обробка)
		if cpuLoad > 0 {
			numWorkers := runtime.NumCPU()
			iterationsPerWorker := cpuLoad / numWorkers

			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for w := 0; w < numWorkers; w++ {
				go func() {
					defer wg.Done()
					// Кожен воркер виконує свою частину ітерацій
					for i := 0; i < iterationsPerWorker; i++ {
						_ = math.Sqrt(float64(i))
					}
				}()
			}
			wg.Wait()
		}

		// 3. Емуляція навантаження на RAM
		if memLoad > 0 {
			dummy := make([]byte, memLoad*1024*1024)
			if len(dummy) > 0 {
				dummy[0] = 1
			}
		}

		fmt.Fprintf(w, "Node: %s | CPU Iterations: %d | RAM: %dMB\n", os.Getenv("HOSTNAME"), cpuLoad, memLoad)
	})

	// Використовуємо порт :80 згідно твого останнього запиту
	port := ":80"
	fmt.Printf("Stress server started on %s with %d cores available\n", port, runtime.NumCPU())
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}