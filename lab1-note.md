# build and run
 the first thing that we should do is to build the plugin derived from /mrapps/wc.go, with command "go build -race -buildmode=plugin ../mrapps/wc.go", and it will generate a .so file "wc.so" in /main

 after that we can conduct programm "mrsequential.go" through command "go run -race mesequential.go wc.so pg*.txt", after that a result "mr-out-0" will be produced, 

 ## take a look at ~/mrapps/wc.go

 this program is word-count application "plugin" for mapreduce

 this program define two functions: Map and Reduce
 * function Map take two string filename and contents, return an array of mr.KeyValue
 * function Reduce the number of occurances of the key


## take a look at ~/main/mrsequential.go

this program is a simple sequential MapReduce

in main function, it loadPlugin of Map and Reduce, then read each inputfile into function map, accumulate the intermediate Map output, and finally sort it.

at last, call reduce on each distinct key in intermediatep[] 

## my job

our job is to implementa a distributed system consists of "rpc.go", "coordinator.go", "worker.go".

here's how to run our code on the word-count MapReduce application.

first we require to run the mrcoordinator process
```bash
go build -race -buildmode=plugin ../mrapps/wc.go
go run -race mrcoordinator.go pg_*.txt
```

then we can run mrworker processes in one or more terminal
```bash
go run -race mrworker.go wc.so
```

after mrcoordinator and mrworder complete their jobs, we will obtain result files the same as what mrconsequential.go produces.
```bash
$ cat mr-out-* | sort | more
A 509
ABOUT 2
ACT 8
...
```

and the way we test whether our program correct or not is running shell test-mr.sh
```bash
$ bash test-mr.sh
*** Starting wc test.
--- wc test: PASS
*** Starting indexer test.
--- indexer test: PASS
*** Starting map parallelism test.
--- map parallelism test: PASS
*** Starting reduce parallelism test.
--- reduce parallelism test: PASS
*** Starting crash test.
--- crash test: PASS
*** PASSED ALL TESTS
$
```

## get start first

let's get start by getting familiar with the code
the mrcoordinator.go and mrworker.go is framework of conduction, who will call coordinator.go and worker.go respectively.

### coordinator.go

first define coordinator's structure
then implement multiple interface of coordinator:
* server: start a thread that listens for RPCs from a worker.go
* Done: coordinator calls Done() periodically to find out if the entire job has finished

then implement function MakeCoordinator
* create a coordinator
* mrcoordinator calls this function, and nReduce is the number of reduce tasks to use





