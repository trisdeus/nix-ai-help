// Package cache provides response streaming optimization for real-time response delivery
package cache

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// StreamingOptimizer handles response streaming optimization
type StreamingOptimizer struct {
	responseStreams map[string]*ResponseStream
	config          *config.UserConfig
	logger          *logger.Logger
	mu              sync.RWMutex
	bufferSize      int
	flushInterval   time.Duration
}

// ResponseStream represents an optimized response stream
type ResponseStream struct {
	ID              string                 `json:"id"`
	Query           string                 `json:"query"`
	Context         QueryContext           `json:"context"`
	Writer          io.Writer              `json:"-"`
	Buffer          *StreamBuffer          `json:"-"`
	StartTime       time.Time              `json:"start_time"`
	LastActivity    time.Time              `json:"last_activity"`
	TotalBytes      int64                  `json:"total_bytes"`
	ChunksWritten   int                    `json:"chunks_written"`
	CompressionRate float64                `json:"compression_rate"`
	Latency         time.Duration          `json:"latency"`
	Active          bool                   `json:"active"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// StreamBuffer provides intelligent buffering for responses
type StreamBuffer struct {
	data           []byte
	size           int
	maxSize        int
	flushThreshold int
	lastFlush      time.Time
	writer         io.Writer
	mu             sync.Mutex
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	ID        string                 `json:"id"`
	StreamID  string                 `json:"stream_id"`
	Sequence  int                    `json:"sequence"`
	Data      []byte                 `json:"data"`
	Size      int                    `json:"size"`
	Timestamp time.Time              `json:"timestamp"`
	Type      StreamChunkType        `json:"type"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// StreamChunkType represents different types of stream chunks
type StreamChunkType string

const (
	ChunkTypeHeader   StreamChunkType = "header"
	ChunkTypeContent  StreamChunkType = "content"
	ChunkTypeMetadata StreamChunkType = "metadata"
	ChunkTypeEnd      StreamChunkType = "end"
	ChunkTypeError    StreamChunkType = "error"
)

// StreamStats tracks streaming performance statistics
type StreamStats struct {
	ActiveStreams     int           `json:"active_streams"`
	TotalStreams      int           `json:"total_streams"`
	TotalBytes        int64         `json:"total_bytes"`
	AverageLatency    time.Duration `json:"average_latency"`
	AverageChunkSize  int           `json:"average_chunk_size"`
	CompressionRatio  float64       `json:"compression_ratio"`
	BufferUtilization float64       `json:"buffer_utilization"`
	FlushRate         float64       `json:"flush_rate"`
	ErrorRate         float64       `json:"error_rate"`
	LastUpdate        time.Time     `json:"last_update"`
}

// StreamingConfig configures streaming behavior
type StreamingConfig struct {
	BufferSize      int           `json:"buffer_size"`      // Buffer size in bytes
	FlushInterval   time.Duration `json:"flush_interval"`   // Auto-flush interval
	ChunkSize       int           `json:"chunk_size"`       // Preferred chunk size
	CompressionMode string        `json:"compression_mode"` // "none", "gzip", "adaptive"
	MaxStreams      int           `json:"max_streams"`      // Maximum concurrent streams
	TimeoutDuration time.Duration `json:"timeout_duration"` // Stream timeout
	EnableBatching  bool          `json:"enable_batching"`  // Enable chunk batching
}

// NewStreamingOptimizer creates a new streaming optimizer
func NewStreamingOptimizer(cfg *config.UserConfig) *StreamingOptimizer {
	return &StreamingOptimizer{
		responseStreams: make(map[string]*ResponseStream),
		config:          cfg,
		logger:          logger.NewLogger(),
		bufferSize:      8192,  // 8KB default buffer
		flushInterval:   50 * time.Millisecond, // 50ms flush interval
	}
}

// CreateStream creates a new optimized response stream
func (so *StreamingOptimizer) CreateStream(ctx context.Context, query string, context QueryContext, writer io.Writer) (*ResponseStream, error) {
	so.mu.Lock()
	defer so.mu.Unlock()

	streamID := so.generateStreamID(query, context)
	
	buffer := &StreamBuffer{
		data:           make([]byte, 0, so.bufferSize),
		maxSize:        so.bufferSize,
		flushThreshold: so.bufferSize / 4, // Flush at 25% capacity
		lastFlush:      time.Now(),
		writer:         writer,
	}

	stream := &ResponseStream{
		ID:          streamID,
		Query:       query,
		Context:     context,
		Writer:      writer,
		Buffer:      buffer,
		StartTime:   time.Now(),
		LastActivity: time.Now(),
		Active:      true,
		Metadata:    make(map[string]interface{}),
	}

	so.responseStreams[streamID] = stream
	so.logger.Info(fmt.Sprintf("Created optimized stream %s for query: %s", streamID, query))

	return stream, nil
}

// WriteToStream writes data to a stream with optimization
func (so *StreamingOptimizer) WriteToStream(streamID string, data []byte) error {
	so.mu.RLock()
	stream, exists := so.responseStreams[streamID]
	so.mu.RUnlock()

	if !exists {
		return fmt.Errorf("stream %s not found", streamID)
	}

	if !stream.Active {
		return fmt.Errorf("stream %s is not active", streamID)
	}

	// Write to buffer with optimization
	return stream.Buffer.Write(data)
}

// WriteStreamChunk writes a structured chunk to the stream
func (so *StreamingOptimizer) WriteStreamChunk(streamID string, chunk StreamChunk) error {
	chunk.StreamID = streamID
	chunk.Timestamp = time.Now()
	
	// Serialize chunk (in real implementation, this might use efficient serialization)
	data := so.serializeChunk(chunk)
	
	return so.WriteToStream(streamID, data)
}

// FlushStream forces a flush of the stream buffer
func (so *StreamingOptimizer) FlushStream(streamID string) error {
	so.mu.RLock()
	stream, exists := so.responseStreams[streamID]
	so.mu.RUnlock()

	if !exists {
		return fmt.Errorf("stream %s not found", streamID)
	}

	return stream.Buffer.Flush()
}

// CloseStream closes and cleans up a stream
func (so *StreamingOptimizer) CloseStream(streamID string) error {
	so.mu.Lock()
	defer so.mu.Unlock()

	stream, exists := so.responseStreams[streamID]
	if !exists {
		return fmt.Errorf("stream %s not found", streamID)
	}

	// Final flush
	stream.Buffer.Flush()
	
	// Mark as inactive
	stream.Active = false
	stream.LastActivity = time.Now()

	// Clean up after delay (in real implementation)
	go func() {
		time.Sleep(5 * time.Minute)
		so.mu.Lock()
		delete(so.responseStreams, streamID)
		so.mu.Unlock()
		so.logger.Info(fmt.Sprintf("Cleaned up stream %s", streamID))
	}()

	so.logger.Info(fmt.Sprintf("Closed stream %s", streamID))
	return nil
}

// OptimizeStreamForQuery optimizes streaming based on query characteristics
func (so *StreamingOptimizer) OptimizeStreamForQuery(stream *ResponseStream, query string) {
	// Analyze query type to optimize streaming
	queryType := so.analyzeQueryType(query)
	
	switch queryType {
	case "code_generation":
		// Optimize for incremental code delivery
		stream.Buffer.flushThreshold = 512 // Smaller chunks for real-time feedback
		stream.Metadata["optimization"] = "incremental_code"
	case "configuration":
		// Optimize for structured config delivery
		stream.Buffer.flushThreshold = 1024 // Medium chunks for configuration
		stream.Metadata["optimization"] = "structured_config"
	case "troubleshooting":
		// Optimize for diagnostic information
		stream.Buffer.flushThreshold = 256 // Small chunks for step-by-step guidance
		stream.Metadata["optimization"] = "diagnostic_steps"
	case "explanation":
		// Optimize for educational content
		stream.Buffer.flushThreshold = 2048 // Larger chunks for coherent explanations
		stream.Metadata["optimization"] = "educational_content"
	default:
		// Default optimization
		stream.Buffer.flushThreshold = so.bufferSize / 4
		stream.Metadata["optimization"] = "default"
	}

	so.logger.Info(fmt.Sprintf("Optimized stream %s for query type: %s", stream.ID, queryType))
}

// GetStreamStats returns statistics for all active streams
func (so *StreamingOptimizer) GetStreamStats() StreamStats {
	so.mu.RLock()
	defer so.mu.RUnlock()

	stats := StreamStats{
		ActiveStreams: 0,
		TotalStreams:  len(so.responseStreams),
		LastUpdate:   time.Now(),
	}

	var totalLatency time.Duration
	var totalBytes int64
	var totalChunks int

	for _, stream := range so.responseStreams {
		if stream.Active {
			stats.ActiveStreams++
		}
		
		totalLatency += stream.Latency
		totalBytes += stream.TotalBytes
		totalChunks += stream.ChunksWritten
	}

	if stats.TotalStreams > 0 {
		stats.AverageLatency = totalLatency / time.Duration(stats.TotalStreams)
	}

	stats.TotalBytes = totalBytes
	
	if totalChunks > 0 {
		stats.AverageChunkSize = int(totalBytes) / totalChunks
	}

	return stats
}

// StreamBuffer methods

// Write adds data to the buffer and handles intelligent flushing
func (sb *StreamBuffer) Write(data []byte) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Check if we need to flush before adding new data
	if len(sb.data)+len(data) > sb.maxSize {
		if err := sb.flush(); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
	}

	// Add data to buffer
	sb.data = append(sb.data, data...)
	sb.size = len(sb.data)

	// Check if we should flush based on threshold or time
	shouldFlush := sb.size >= sb.flushThreshold || 
		time.Since(sb.lastFlush) > 100*time.Millisecond

	if shouldFlush {
		return sb.flush()
	}

	return nil
}

// Flush writes all buffered data to the writer
func (sb *StreamBuffer) Flush() error {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.flush()
}

// flush internal method (caller must hold lock)
func (sb *StreamBuffer) flush() error {
	if len(sb.data) == 0 {
		return nil
	}

	_, err := sb.writer.Write(sb.data)
	if err != nil {
		return err
	}

	// Clear buffer
	sb.data = sb.data[:0]
	sb.size = 0
	sb.lastFlush = time.Now()

	return nil
}

// Helper methods

func (so *StreamingOptimizer) generateStreamID(query string, context QueryContext) string {
	// Generate a unique stream ID
	return fmt.Sprintf("stream_%d_%s", time.Now().Unix(), context.WorkingDirectory)
}

func (so *StreamingOptimizer) analyzeQueryType(query string) string {
	query = strings.ToLower(query)
	
	if strings.Contains(query, "generate") || strings.Contains(query, "create") || strings.Contains(query, "write") {
		return "code_generation"
	}
	if strings.Contains(query, "config") || strings.Contains(query, "setup") || strings.Contains(query, "install") {
		return "configuration"
	}
	if strings.Contains(query, "error") || strings.Contains(query, "fix") || strings.Contains(query, "debug") || strings.Contains(query, "problem") {
		return "troubleshooting"
	}
	if strings.Contains(query, "explain") || strings.Contains(query, "how") || strings.Contains(query, "what") || strings.Contains(query, "why") {
		return "explanation"
	}
	
	return "general"
}

func (so *StreamingOptimizer) serializeChunk(chunk StreamChunk) []byte {
	// Simple serialization (in real implementation, might use protobuf or msgpack)
	serialized := fmt.Sprintf("[%s][%d][%s]%s\n", 
		chunk.Type, chunk.Sequence, chunk.Timestamp.Format(time.RFC3339), string(chunk.Data))
	return []byte(serialized)
}

// StreamingReader provides optimized reading of streamed responses
type StreamingReader struct {
	reader    io.Reader
	buffer    []byte
	bufferPos int
	bufferEnd int
	eof       bool
}

// NewStreamingReader creates a new streaming reader
func NewStreamingReader(reader io.Reader) *StreamingReader {
	return &StreamingReader{
		reader: reader,
		buffer: make([]byte, 8192),
	}
}

// ReadChunk reads the next chunk from the stream
func (sr *StreamingReader) ReadChunk() (*StreamChunk, error) {
	if sr.eof {
		return nil, io.EOF
	}

	// Read line-by-line for chunk boundaries
	scanner := bufio.NewScanner(sr.reader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		sr.eof = true
		return nil, io.EOF
	}

	line := scanner.Text()
	return sr.parseChunk(line)
}

func (sr *StreamingReader) parseChunk(line string) (*StreamChunk, error) {
	// Parse serialized chunk format: [type][sequence][timestamp]data
	// This is a simplified parser
	if len(line) < 10 {
		return nil, fmt.Errorf("invalid chunk format")
	}

	// Extract type
	typeEnd := strings.Index(line[1:], "]") + 1
	if typeEnd < 2 {
		return nil, fmt.Errorf("invalid chunk type")
	}
	chunkType := StreamChunkType(line[1:typeEnd])

	// For simplicity, just return basic chunk
	return &StreamChunk{
		Type:      chunkType,
		Data:      []byte(line[typeEnd+1:]),
		Timestamp: time.Now(),
	}, nil
}

// StreamingAnalytics provides analytics for streaming performance
type StreamingAnalytics struct {
	metrics map[string]*StreamMetric
	mu      sync.RWMutex
}

// StreamMetric tracks metrics for a specific stream
type StreamMetric struct {
	StreamID        string        `json:"stream_id"`
	TotalBytes      int64         `json:"total_bytes"`
	ChunkCount      int           `json:"chunk_count"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	AverageChunkSize int          `json:"average_chunk_size"`
	Throughput      float64       `json:"throughput"` // bytes per second
}

// NewStreamingAnalytics creates new streaming analytics
func NewStreamingAnalytics() *StreamingAnalytics {
	return &StreamingAnalytics{
		metrics: make(map[string]*StreamMetric),
	}
}

// RecordStreamStart records the start of a stream
func (sa *StreamingAnalytics) RecordStreamStart(streamID string) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	sa.metrics[streamID] = &StreamMetric{
		StreamID:  streamID,
		StartTime: time.Now(),
	}
}

// RecordChunk records a chunk being written to a stream
func (sa *StreamingAnalytics) RecordChunk(streamID string, size int) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	metric, exists := sa.metrics[streamID]
	if !exists {
		return
	}

	metric.TotalBytes += int64(size)
	metric.ChunkCount++
	
	if metric.ChunkCount > 0 {
		metric.AverageChunkSize = int(metric.TotalBytes) / metric.ChunkCount
	}
}

// RecordStreamEnd records the end of a stream
func (sa *StreamingAnalytics) RecordStreamEnd(streamID string) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	metric, exists := sa.metrics[streamID]
	if !exists {
		return
	}

	metric.EndTime = time.Now()
	metric.Duration = metric.EndTime.Sub(metric.StartTime)
	
	if metric.Duration > 0 {
		metric.Throughput = float64(metric.TotalBytes) / metric.Duration.Seconds()
	}
}

// GetStreamMetrics returns metrics for a specific stream
func (sa *StreamingAnalytics) GetStreamMetrics(streamID string) *StreamMetric {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	metric, exists := sa.metrics[streamID]
	if !exists {
		return nil
	}

	// Return a copy
	metricCopy := *metric
	return &metricCopy
}