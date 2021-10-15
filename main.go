package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/heptiolabs/healthcheck"
)

func main() {

	health := healthcheck.NewHandler()
	// Our app is not happy if we've got more than 100 goroutines running.
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))

	// Our app is not ready if we can't resolve our upstream dependency in DNS.
	health.AddReadinessCheck(
		"upstream-dep-dns",
		healthcheck.DNSResolveCheck("upstream.example.com", 50*time.Millisecond))

	health.AddLivenessCheck(
		"memory-allocation",
		memoryCheck())

	go http.ListenAndServe("0.0.0.0:8085", health)

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/memory", PrintMemUsage)
	router.GET("/blow", blowMemory)

	router.Run("localhost:8080")
}

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

var overall [][]int

func dummyCheck() healthcheck.Check {
	return nil
}

func memoryCheck() healthcheck.Check {
	// var m runtime.MemStats
	// runtime.ReadMemStats(&m)

	return func() error {
		var count runtime.MemStats
		runtime.ReadMemStats(&count)
		if bToMb(count.Alloc) > 200 {
			return fmt.Errorf("too large memory allocation (%d)", bToMb(count.Alloc))
		}
		return nil
	}

}

func PrintMemUsage(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	c.IndentedJSON(http.StatusOK, gin.H{
		"Alloc":      fmt.Sprintf("%v MiB", bToMb(m.Alloc)),
		"TotalAlloc": fmt.Sprintf("%v MiB", bToMb(m.TotalAlloc)),
		"Sys":        fmt.Sprintf("%v MiB", bToMb(m.Sys)),
		"NumGC":      fmt.Sprintf("%v", m.NumGC),
	})
	// fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	// fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	// fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	// fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func blowMemory(c *gin.Context) {

	for i := 0; i < 10; i++ {

		// Allocate memory using make() and append to overall (so it doesn't get
		// garbage collected). This is to create an ever increasing memory usage
		// which we can track. We're just using []int as an example.
		a := make([]int, 0, 999999)
		overall = append(overall, a)

		// Print our memory usage at each interval
		// time.Sleep(time.Second)
	}
	c.JSON(http.StatusOK, gin.H{})
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}
