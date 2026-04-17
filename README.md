# Zep-Fast-Ingest 

A high-performance, concurrent Go tool for deduplicated historical data ingestion into Zep AI.

## The Problem
Enterprise customers often have gigabytes of historical chat data. Ingesting this sequentially into a Knowledge Graph is:
1. **Expensive:** LLM-based extraction on redundant data burns API credits.
2. **Slow:** Sequential network requests create a massive bottleneck.

## The Solution
This tool uses a **client-side MinHash + LSH (Locality Sensitive Hashing)** pipeline to deduplicate data *before* it hits the network. 

### Key Features:
- **Streaming I/O:** Processes million-line JSONL files with < 50MB RAM using `bufio.Scanner`.
- **MinHash/LSH Engine:** Native Go implementation of Jaccard Similarity approximation to identify near-duplicates (e.g., messages differing only by a typo or punctuation).
- **Worker Pool Concurrency:** Fan-out/Fan-in pattern utilizing 10+ concurrent workers for CPU-bound hashing tasks.
- **Atomic Deduplication:** Thread-safe LSH banding using `sync.RWMutex` to ensure O(1) duplicate lookups across workers.
- **Batch Ingestion:** Integrates with `zep-go` SDK to ship only unique payloads in optimal batch sizes.

## Benchmarks (100k Records)
- **Total Records:** 100,000
- **Time:** ~2.3 seconds
- **Throughput:** ~43,000 docs/sec
- **Reduction:** 98.7% (Filtered 98,725 near-duplicates)