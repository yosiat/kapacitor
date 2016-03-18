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
benchmark                                            old ns/op       new ns/op      delta
Benchmark_Write_MeasurementNameNotMatches_1000-4     13577781619     5284088591     -61.08%
Benchmark_Write_MeasurementNameMatches_1000-4        12410649978     5702849076     -54.05%
Benchmark_Write_MeasurementNameNotMatches_100-4      14668078017     6735436249     -54.08%
Benchmark_Write_MeasurementNameMatches_100-4         15342610617     6887722284     -55.11%
Benchmark_Write_MeasurementNameNotMatches_10-4       15439700628     7222307428     -53.22%
Benchmark_Write_MeasurementNameMatches_10-4          17612397820     7525348482     -57.27%

benchmark                                            old allocs     new allocs     delta
Benchmark_Write_MeasurementNameNotMatches_1000-4     15468417       7760752        -49.83%
Benchmark_Write_MeasurementNameMatches_1000-4        19327860       11610918       -39.93%
Benchmark_Write_MeasurementNameNotMatches_100-4      23168533       13770301       -40.56%
Benchmark_Write_MeasurementNameMatches_100-4         23837235       14111713       -40.80%
Benchmark_Write_MeasurementNameNotMatches_10-4       24214825       14250581       -41.15%
Benchmark_Write_MeasurementNameMatches_10-4          24088652       14296632       -40.65%

benchmark                                            old bytes      new bytes      delta
Benchmark_Write_MeasurementNameNotMatches_1000-4     1975011816     989465960      -49.90%
Benchmark_Write_MeasurementNameMatches_1000-4        2406223336     1420145576     -40.98%
Benchmark_Write_MeasurementNameNotMatches_100-4      2890837336     1689738088     -41.55%
Benchmark_Write_MeasurementNameMatches_100-4         2967947288     1726483144     -41.83%
Benchmark_Write_MeasurementNameNotMatches_10-4       3015660424     1743931480     -42.17%
Benchmark_Write_MeasurementNameMatches_10-4          2998693272     1748984872     -41.68%
```
