package downloader

import nodeconfig "github.com/harmony-one/harmony/internal/configs/node"

const (
	numBlocksByNumPerRequest int = 10 // number of blocks for each request
	blocksPerInsert          int = 50 // number of blocks for each insert batch

	numBlockHashesPerRequest  int = 20 // number of get block hashes for short range sync
	numBlocksByHashesUpperCap int = 10 // number of get blocks by hashes upper cap
	numBlocksByHashesLowerCap int = 3  // number of get blocks by hashes lower cap

	lastMileThres uint64 = 10

	// soft cap of size in resultQueue. When the queue size is larger than this limit,
	// no more request will be assigned to workers to wait for InsertChain to finish.
	softQueueCap int = 100

	defaultConcurrency = 16
)

// Config is the downloader config
type Config struct {
	Network     nodeconfig.NetworkType
	Concurrency int // Number of concurrent sync requests
	MinStreams  int // Minimum number of streams to do sync
	InitStreams int // Number of streams requirement for initial bootstrap
}

func (c *Config) fixValues() {
	if c.Concurrency == 0 {
		c.Concurrency = defaultConcurrency
	}
	if c.Concurrency > c.MinStreams {
		c.MinStreams = c.Concurrency
	}
	if c.MinStreams > c.InitStreams {
		c.InitStreams = c.MinStreams
	}
}
