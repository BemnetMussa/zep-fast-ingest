# Zep-Fast-Ingest ⚡️

A high-performance, edge-deduplication ingestion engine built specifically for the **Zep Cloud (Graphiti Temporal Memory) API**.

## The Mission
As AI agents move toward long-term memory, the cost of ingesting historical enterprise data (Zendesk, Slack, CRM) becomes a bottleneck. Often, data is massively redundant. 

**Zep-Fast-Ingest** moves the semantic deduplication intelligence to the client side. By utilizing Locality Sensitive Hashing (LSH), it filters up to 98% of redundant data before it ever hits Zep's network—drastically reducing API transit costs and preventing your underlying Temporal Knowledge Graph from becoming saturated with noise.

## Technical Architecture
- **Language:** 100% Go (Native performance, zero-dependency parser).
- **Concurrency:** Worker Pool pattern with `10` concurrent hashing workers and built-in channel backpressure.
- **Algorithms:** 
  - **MinHash:** Approximates Jaccard Similarity using 100 bitwise-permuted hash functions.
  - **LSH Banding:** Enables $O(1)$ near-duplicate lookups.
- **Memory Safety & Scale:** Implements **Lock Striping / Map Sharding**. Our in-memory index is split across 256 independent `sync.RWMutex` shards, effectively erasing thread lock-contention and allowing true parallel deduplication.
- **Integration:** Verified directly against **Zep Cloud V2/V3** architecture. Strictly adheres to Zep's entity graph constraints by creating `Users`, initializing `Threads`, and batch-uploading `Messages` strictly within API rate limits (30 messages per max payload).

## Benchmarks & Results
We built a dual-mode testing framework to prove this system works.

**1. The Integration Validation (Real Zep Cloud API)**
- Fully tested and integrated with the live `https://api.getzep.com` environment.
- Correctly constructs Users, generates Threads, assigns required `user_id`s, and successfully pushes authenticated payloads using `ZEP_API_KEY`. 

**2. The Raw Throughput Benchmark (Local Mock Server)**
- Tested locally against a 100,000 document simulated dataset injected with 98% semantic duplication.
- **Execution Time:** ~0.43 seconds (431ms)
- **Scale Impact:** Lock striping dropped execution time from an initial 2.38 seconds down to <500ms.
- **Data Reduction:** Filtered over 99,000 near-duplicates dynamically.

## Setup & Run

### 1. Configure Credentials
Create a `.env` file in the root directory:
```env
ZEP_API_URL="https://api.getzep.com"
ZEP_API_KEY="your_zep_cloud_key_here"
```

### 2. Build and Execute (Real Cloud)
Ensure your dataset is prepared (e.g. `data.jsonl`) and run the pipeline:
```bash
go build -o zep-ingest ./cmd/zep-fast-ingest
./zep-ingest --file data.jsonl
```

### 3. Local Benchmarking (No Network Cost)
If you want to test throughput locally without burning Zep Cloud API limits:
```bash
go build -o mock-server ./cmd/mock-server
./mock-server &
export ZEP_API_URL="http://localhost:8000"
./zep-ingest --file data.jsonl
```