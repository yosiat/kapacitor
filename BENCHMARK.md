# Benchmarking


## What benchmark
The benchmarks are using "packets_benchmark.srpl" which has 4848 points with the measurement of "packets".
I ingested it directly to http handler using "/write" request.

The benchmark splitted to two types:
1. Good case scenario - measurement names *do* matches in the from('..')
1. Bad case scenario - measurement names *dont* matches in the from('..')

All of the benchmark ran with the same task:
```javascript
stream.from('MEASUREMENT_NAME')
```
And I ran each benchmark multiple times:
- With bad/good case scenario
- Number of tasks - 1000, 100, 10

That makes total: 6 (2 types * 3 groups of tasks count) benchmarks


## Results

I ran benchmark sessions using: 
`go test -v -bench=. ./integrations/write_benchmark_test.go ./integrations/helpers_test.go  | tee perf.0`

After that checked the cpu profile (ran another session using "-cpuprofile=prof.cpu") and saw that:

```
(pprof) list CollectPoint
Total: 4.76mins
ROUTINE ======================== github.com/influxdata/kapacitor.(*Edge).CollectPoint in /Users/yosi/code/go/src/github.com/influxdata/kapacitor/edge.go
     5.20s   1.54mins (flat, cum) 32.36% of Total
         .          .    183:		}
         .          .    184:	}
         .          .    185:	return
         .          .    186:}
         .          .    187:
     200ms     29.38s    188:func (e *Edge) CollectPoint(p models.Point) error {
     2.80s        16s    189:	e.statMap.Add(statCollected, 1)
     150ms     15.65s    190:	e.incCollected(&p)
     100ms     27.53s    191:	select {
     390ms      1.26s    192:	case <-e.aborted:
         .          .    193:		return ErrAborted
     1.46s      2.49s    194:	case e.stream <- p:
     100ms      100ms    195:		return nil
         .          .    196:	}
         .          .    197:}
         .          .    198:
         .          .    199:func (e *Edge) CollectBatch(b models.Batch) error {
         .          .    200:	e.statMap.Add(statCollected, 1)
```

So I commented out:
```go
e.statMap.Add(statCollected, 1)
e.incCollected(&p)
```

And ran another benchmark session:
`go test -v -bench=. ./integrations/write_benchmark_test.go ./integrations/helpers_test.go  | tee perf.1`


Compared the results using `benchcmp perf.0 perf.1`, and got the next results:

```
benchmark                                            old ns/op       new ns/op       delta
Benchmark_Write_MeasurementNameNotMatches_1000-4     15057077777     9273169621      -38.41%
Benchmark_Write_MeasurementNameMatches_1000-4        26866797210     20475112351     -23.79%
Benchmark_Write_MeasurementNameNotMatches_100-4      17269207804     9242848814      -46.48%
Benchmark_Write_MeasurementNameMatches_100-4         17185451808     9263158453      -46.10%
Benchmark_Write_MeasurementNameNotMatches_10-4       15581726506     7605390150      -51.19%
Benchmark_Write_MeasurementNameMatches_10-4          16172117992     8545672907      -47.16%

benchmark                                            old allocs     new allocs     delta
Benchmark_Write_MeasurementNameNotMatches_1000-4     15825754       8124723        -48.66%
Benchmark_Write_MeasurementNameMatches_1000-4        24209180       14424353       -40.42%
Benchmark_Write_MeasurementNameNotMatches_100-4      23860781       14146063       -40.71%
Benchmark_Write_MeasurementNameMatches_100-4         24373527       14449352       -40.72%
Benchmark_Write_MeasurementNameNotMatches_10-4       24283917       14292482       -41.14%
Benchmark_Write_MeasurementNameMatches_10-4          24286921       14327471       -41.01%

benchmark                                            old bytes      new bytes      delta
Benchmark_Write_MeasurementNameNotMatches_1000-4     2257844232     1272840424     -43.63%
Benchmark_Write_MeasurementNameMatches_1000-4        3257172728     2006710960     -38.39%
Benchmark_Write_MeasurementNameNotMatches_100-4      3000318344     1758955048     -41.37%
Benchmark_Write_MeasurementNameMatches_100-4         3058094440     1791036808     -41.43%
Benchmark_Write_MeasurementNameNotMatches_10-4       3026568552     1751300984     -42.14%
Benchmark_Write_MeasurementNameMatches_10-4          3026105272     1754809144     -42.01%
```
