package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/initialize"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	fetchBatchSize  = 1000 // Documents to fetch per batch from MongoDB
	jsonWorkerCount = 4    // Parallel workers for JSON conversion per batch
	globalWorkers   = 8    // Global worker pool size for processing batches
)

type collectionResult struct {
	name     string
	count    int64
	duration time.Duration
	err      error
}

type progressUpdate struct {
	collName     string
	current      int64
	total        int64
	bytesWritten int64
}

// batchJob represents a single batch of work to fetch and convert
type batchJob struct {
	collName   string
	batchNum   int
	skip       int64
	limit      int64
	totalDocs  int64
	collection *mongo.Collection
	resultChan chan<- batchResult
}

// batchResult contains the processed batch data ready for writing
type batchResult struct {
	batchNum  int
	jsonDocs  [][]byte
	fetchTime time.Duration
	jsonTime  time.Duration
	err       error
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Parse output directory from args or use default
	outputDir := getOutputDir()

	config := hz_config.NewConfigFromPath("config/dev.json")
	if err := config.Read(); err != nil {
		panic(err)
	}

	client, err := initialize.MongoClient(
		config.ValueOrPanic("mongo.connection.host"),
		config.ValueOrPanic("mongo.connection.username"),
		config.ValueOrPanic("mongo.connection.password"),
	)
	if err != nil {
		panic(err)
	}
	defer initialize.MongoCleanup(ctx, client)

	dbName := config.ValueOrPanic("mongo.database.name")
	db := client.Database(dbName)

	fmt.Println("=== MongoDB Collection Dump ===")
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("Output Directory: %s\n", outputDir)
	fmt.Printf("Workers: %d\n\n", runtime.NumCPU())

	if err := dumpAllCollections(ctx, db, outputDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func getOutputDir() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	// Default to ~/Documents/hazelmere_dmp_<date>
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	dateStr := time.Now().Format("2006-01-02")
	return filepath.Join(homeDir, "Documents", fmt.Sprintf("hazelmere_dmp_%s", dateStr))
}

func dumpAllCollections(ctx context.Context, db *mongo.Database, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// List all collections
	fmt.Println("Fetching collection list...")
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if len(collections) == 0 {
		fmt.Println("No collections found in database")
		return nil
	}

	fmt.Printf("Found %d collections: %v\n\n", len(collections), collections)

	// Get document counts for all collections first
	fmt.Println("Counting documents in each collection...")
	collectionCounts := make(map[string]int64)
	var totalDocsExpected int64
	var totalBatches int
	for _, collName := range collections {
		count, err := db.Collection(collName).CountDocuments(ctx, bson.M{})
		if err != nil {
			fmt.Printf("  WARNING: Could not count %s: %v\n", collName, err)
			collectionCounts[collName] = 0
		} else {
			collectionCounts[collName] = count
			totalDocsExpected += count
			numBatches := (count + int64(fetchBatchSize) - 1) / int64(fetchBatchSize)
			totalBatches += int(numBatches)
			fmt.Printf("  %s: %d documents (%d batches)\n", collName, count, numBatches)
		}
	}
	fmt.Printf("\nTotal documents to dump: %d across %d batches\n", totalDocsExpected, totalBatches)
	fmt.Printf("Global worker pool: %d workers, batch size: %d, JSON workers per batch: %d\n\n", globalWorkers, fetchBatchSize, jsonWorkerCount)

	startTime := time.Now()

	// Create global job queue for all batches across all collections
	jobQueue := make(chan batchJob, totalBatches)

	// Track active workers
	var activeWorkers atomic.Int32
	activeWorkers.Store(int32(globalWorkers))

	// Start global worker pool
	var workerWg sync.WaitGroup
	for w := 0; w < globalWorkers; w++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			defer activeWorkers.Add(-1)
			for job := range jobQueue {
				processBatchJob(ctx, workerID, job)
			}
		}(w)
	}

	// Start a writer goroutine for each collection
	var writerWg sync.WaitGroup
	resultChans := make(map[string]chan batchResult)
	collResults := make(chan collectionResult, len(collections))

	for _, collName := range collections {
		count := collectionCounts[collName]
		numBatches := int((count + int64(fetchBatchSize) - 1) / int64(fetchBatchSize))
		if count == 0 {
			numBatches = 0
		}

		resultChan := make(chan batchResult, numBatches+1)
		resultChans[collName] = resultChan

		writerWg.Add(1)
		go func(collName string, expectedCount int64, numBatches int, resultChan <-chan batchResult) {
			defer writerWg.Done()
			result := writeCollectionBatches(ctx, collName, outputDir, expectedCount, numBatches, resultChan, &activeWorkers)
			collResults <- result
		}(collName, count, numBatches, resultChan)
	}

	// Enqueue all batch jobs
	for _, collName := range collections {
		count := collectionCounts[collName]
		if count == 0 {
			// Signal empty collection
			close(resultChans[collName])
			continue
		}

		collection := db.Collection(collName)
		resultChan := resultChans[collName]
		numBatches := int((count + int64(fetchBatchSize) - 1) / int64(fetchBatchSize))

		for b := 0; b < numBatches; b++ {
			jobQueue <- batchJob{
				collName:   collName,
				batchNum:   b,
				skip:       int64(b) * int64(fetchBatchSize),
				limit:      int64(fetchBatchSize),
				totalDocs:  count,
				collection: collection,
				resultChan: resultChan,
			}
		}
	}
	close(jobQueue)

	// Wait for all workers to finish processing
	workerWg.Wait()

	// Close all result channels to signal writers to finish
	for _, collName := range collections {
		if collectionCounts[collName] > 0 {
			close(resultChans[collName])
		}
	}

	// Wait for all writers to finish
	writerWg.Wait()
	close(collResults)

	// Collect results
	var totalDocs int64
	var totalBytes int64
	var results []collectionResult
	for result := range collResults {
		results = append(results, result)
		if result.err != nil {
			fmt.Printf("  ERROR   [%s]: %v\n", result.name, result.err)
		} else {
			fmt.Printf("  DONE    [%s]: %d documents in %v\n", result.name, result.count, result.duration.Round(time.Millisecond))
			totalDocs += result.count
		}
	}

	totalDuration := time.Since(startTime)

	// Calculate total bytes written
	for _, collName := range collections {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.json", collName))
		if info, err := os.Stat(outputPath); err == nil {
			totalBytes += info.Size()
		}
	}

	fmt.Printf("\n=====================================\n")
	fmt.Printf("           DUMP COMPLETE             \n")
	fmt.Printf("=====================================\n")
	fmt.Printf("Collections dumped: %d\n", len(collections))
	fmt.Printf("Total documents:    %d\n", totalDocs)
	fmt.Printf("Total size:         %s\n", formatBytes(totalBytes))
	fmt.Printf("Total time:         %v\n", totalDuration.Round(time.Millisecond))
	fmt.Printf("Throughput:         %.0f docs/sec\n", float64(totalDocs)/totalDuration.Seconds())
	fmt.Printf("Output directory:   %s\n", outputDir)
	fmt.Printf("=====================================\n")

	// Print file summary
	fmt.Printf("\nOutput files:\n")
	for _, collName := range collections {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.json", collName))
		if info, err := os.Stat(outputPath); err == nil {
			fmt.Printf("  %s.json: %s\n", collName, formatBytes(info.Size()))
		}
	}

	// Print error summary if any
	var errors []collectionResult
	for _, r := range results {
		if r.err != nil {
			errors = append(errors, r)
		}
	}
	if len(errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %s: %v\n", e.name, e.err)
		}
	}

	return nil
}

// processBatchJob fetches and converts a batch to JSON, sending result to the collection's writer
func processBatchJob(ctx context.Context, workerID int, job batchJob) {
	fetchStart := time.Now()

	// Fetch batch
	opts := options.Find().
		SetBatchSize(int32(job.limit)).
		SetLimit(job.limit).
		SetSkip(job.skip)

	cursor, err := job.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		job.resultChan <- batchResult{batchNum: job.batchNum, err: fmt.Errorf("fetch failed: %w", err)}
		return
	}

	var batchDocs []bson.Raw
	for cursor.Next(ctx) {
		var raw bson.Raw
		if err := cursor.Decode(&raw); err != nil {
			cursor.Close(ctx)
			job.resultChan <- batchResult{batchNum: job.batchNum, err: fmt.Errorf("decode failed: %w", err)}
			return
		}
		rawCopy := make(bson.Raw, len(raw))
		copy(rawCopy, raw)
		batchDocs = append(batchDocs, rawCopy)
	}
	cursor.Close(ctx)

	if err := cursor.Err(); err != nil {
		job.resultChan <- batchResult{batchNum: job.batchNum, err: fmt.Errorf("cursor error: %w", err)}
		return
	}

	fetchTime := time.Since(fetchStart)

	// Convert to JSON in parallel
	jsonStart := time.Now()
	jsonDocs := make([][]byte, len(batchDocs))
	jsonErrors := make([]error, len(batchDocs))

	var wg sync.WaitGroup
	docChan := make(chan int, len(batchDocs))

	numJsonWorkers := jsonWorkerCount
	if len(batchDocs) < numJsonWorkers {
		numJsonWorkers = len(batchDocs)
	}

	for w := 0; w < numJsonWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range docChan {
				extJSON, err := bson.MarshalExtJSON(batchDocs[idx], true, false)
				if err != nil {
					jsonErrors[idx] = err
				} else {
					jsonDocs[idx] = extJSON
				}
			}
		}()
	}

	for i := range batchDocs {
		docChan <- i
	}
	close(docChan)
	wg.Wait()

	jsonTime := time.Since(jsonStart)

	// Check for errors
	for i, err := range jsonErrors {
		if err != nil {
			job.resultChan <- batchResult{batchNum: job.batchNum, err: fmt.Errorf("json marshal doc %d: %w", i, err)}
			return
		}
	}

	pct := float64(job.skip+int64(len(batchDocs))) / float64(job.totalDocs) * 100
	fmt.Printf("  WORKER%d [%s]: Batch %d done - %d docs (fetch: %v, json: %v) - %.1f%%\n",
		workerID, job.collName, job.batchNum, len(batchDocs),
		fetchTime.Round(time.Millisecond), jsonTime.Round(time.Millisecond), pct)

	job.resultChan <- batchResult{
		batchNum:  job.batchNum,
		jsonDocs:  jsonDocs,
		fetchTime: fetchTime,
		jsonTime:  jsonTime,
	}
}

// writeCollectionBatches receives batch results and writes them to file in order
func writeCollectionBatches(ctx context.Context, collName string, outputDir string, expectedCount int64, numBatches int, resultChan <-chan batchResult, activeWorkers *atomic.Int32) collectionResult {
	start := time.Now()
	result := collectionResult{name: collName}

	if expectedCount == 0 {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.json", collName))
		if err := os.WriteFile(outputPath, []byte("[]"), 0644); err != nil {
			result.err = fmt.Errorf("failed to write empty file: %w", err)
			return result
		}
		result.count = 0
		result.duration = time.Since(start)
		fmt.Printf("  START   [%s]: Empty collection, created empty file\n", collName)
		return result
	}

	fmt.Printf("  START   [%s]: Expecting %d documents in %d batches\n", collName, expectedCount, numBatches)

	// Create output file
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.json", collName))
	file, err := os.Create(outputPath)
	if err != nil {
		result.err = fmt.Errorf("failed to create output file: %w", err)
		return result
	}
	defer file.Close()

	bufWriter := newBufferedFileWriter(file, 8*1024*1024)
	defer bufWriter.Flush()

	if _, err := bufWriter.WriteString("[\n"); err != nil {
		result.err = fmt.Errorf("failed to write opening bracket: %w", err)
		return result
	}

	// Buffer to hold out-of-order batches
	pendingBatches := make(map[int]batchResult)
	nextBatchToWrite := 0
	var totalDocCount int64
	var bytesWritten int64
	first := true

	// Receive and write batches in order
	for batchResult := range resultChan {
		if batchResult.err != nil {
			result.err = batchResult.err
			return result
		}

		// Store batch (might be out of order)
		pendingBatches[batchResult.batchNum] = batchResult

		// Write any consecutive batches we can
		for {
			batch, ok := pendingBatches[nextBatchToWrite]
			if !ok {
				break
			}
			delete(pendingBatches, nextBatchToWrite)

			// Write this batch
			writeStart := time.Now()
			for _, extJSON := range batch.jsonDocs {
				if !first {
					n, err := bufWriter.WriteString(",\n")
					if err != nil {
						result.err = fmt.Errorf("failed to write separator: %w", err)
						return result
					}
					bytesWritten += int64(n)
				}
				first = false

				n, err := bufWriter.Write(extJSON)
				if err != nil {
					result.err = fmt.Errorf("failed to write document: %w", err)
					return result
				}
				bytesWritten += int64(n)
				totalDocCount++
			}
			writeDuration := time.Since(writeStart)

			pct := float64(totalDocCount) / float64(expectedCount) * 100
			fmt.Printf("  WRITE   [%s]: Batch %d written - %d docs (write: %v) - %d/%d (%.1f%%) - %d workers active\n",
				collName, nextBatchToWrite, len(batch.jsonDocs),
				writeDuration.Round(time.Millisecond),
				totalDocCount, expectedCount, pct,
				activeWorkers.Load())

			nextBatchToWrite++
		}
	}

	// Write closing bracket
	if _, err := bufWriter.WriteString("\n]"); err != nil {
		result.err = fmt.Errorf("failed to write closing bracket: %w", err)
		return result
	}

	if err := bufWriter.Flush(); err != nil {
		result.err = fmt.Errorf("failed to flush buffer: %w", err)
		return result
	}

	result.count = totalDocCount
	result.duration = time.Since(start)
	return result
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// bufferedFileWriter wraps a file with a buffer for better write performance
type bufferedFileWriter struct {
	file   *os.File
	buffer []byte
	pos    int
}

func newBufferedFileWriter(file *os.File, bufferSize int) *bufferedFileWriter {
	return &bufferedFileWriter{
		file:   file,
		buffer: make([]byte, bufferSize),
		pos:    0,
	}
}

func (w *bufferedFileWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		space := len(w.buffer) - w.pos
		if space == 0 {
			if err := w.Flush(); err != nil {
				return written, err
			}
			space = len(w.buffer)
		}
		n := copy(w.buffer[w.pos:], p)
		w.pos += n
		written += n
		p = p[n:]
	}
	return written, nil
}

func (w *bufferedFileWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *bufferedFileWriter) Flush() error {
	if w.pos == 0 {
		return nil
	}
	_, err := w.file.Write(w.buffer[:w.pos])
	w.pos = 0
	return err
}
