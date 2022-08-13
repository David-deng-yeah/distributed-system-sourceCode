package mr

import (
	"fmt"
	"log"
	"net/rpc"
	"hash/fnv"
	"os"
	"io/ioutil"
	"strings"
	"sort"
)


/*
Map functions return a slice of KeyValue.
*/
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }


/*
use ihash(key) % NReduce to choose the reduce
task number for each KeyValue emitted by Map.
*/
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}


/*
main/mrworker.go calls this function.
*/
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	// Your worker implementation here.
	// worker id is pid
	//id := strconv.Itoa(os.Getpid())
	id := os.Getpid()
	var lastTaskType string
	var lastTaskIndex int
	// loop, build arg and reply, call applyfortask
	for {
		args := ApplyForTaskArgs {
			WorkerID : id,
			LastTaskType : lastTaskType,
			LastTaskIndex : lastTaskIndex,
		}
		reply := ApplyForTaskReply{}
		call("Coordinator.ApplyForTask", &args, &reply)

		// 	1. if reply.tasktype = nil, then break(coordinator has no task, mr is done)
		if reply.TaskType == "" {
			break
		}
		//	2. if reply.tasktype = map or reduce
		if reply.TaskType == MAP {
			// handle map
			// 1.write input data
			file, err := os.Open(reply.MapInputFile)
			if err != nil {
				log.Fatalf("failed to open file: %s by error: %e",reply.MapInputFile, err)
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("failed to read file: %s by error: %e", reply.MapInputFile, err)
			}

			// 2.pass to map function
			kva := mapf(reply.MapInputFile, string(content))// key:value map
			// 3.divide into bucket
			hashedKva := make(map[int][] KeyValue)

			for _, kv := range kva {
				reduceID := ihash(kv.Key) % reply.nReduce
				hashedKva[reduceID] = append(hashedKva[reduceID], kv)			
			}	
			// 4.write into intermediate file
			for i:=0; i<reply.nReduce; i++ {
				ofile, _ := os.Create(tmpMapOutFile(id, reply.TaskIndex, i))
				for _, kv := range hashedKva[i] {
					fmt.Fprintf(ofile, "%v\t%v\n", kv.Key, kv.Value)
				}
				ofile.Close()
			}
		}else if reply.TaskType == "Reduce" {
			// handle reduce
			var lines []string
			for mi:=0; mi<reply.nMap; mi++ {
				inputfile := finalMapOutFile(mi, reply.TaskIndex,)
				file, err := os.Open(inputfile)
				if err != nil {
					log.Fatalf("failed to open map output file %s: %e", inputfile, err)
				}
				content, err := ioutil.ReadAll(file)
				if err != nil {
					log.Fatalf("failed to read all map output file %s: %e", inputfile, err)
				}
				lines = append(lines, strings.Split(string(content), "\n")...)
			}

			var kva []KeyValue
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				parts := strings.Split(line, "\t")
				kva = append(kva, KeyValue{
					Key : parts[0],
					Value : parts[1],
				})
			}

			// sort data
			sort.Sort(ByKey(kva))
			ofile, _ := os.Create(tmpReduceOutFile(id, reply.TaskIndex))
			i := 0
			for i<len(kva) {
				j := i+1
				for j <len(kva) && kva[j].Key == kva[i].Key {
					j++
				}
				values := []string{}
				for k:=i; k<j; k++ {
					values = append(values, kva[k].Value)
				}
				output := reducef(kva[i].Key, values)

				fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output)

				i=j
			}
			ofile.Close()
		}

		// record the information of doned task
		lastTaskType = reply.TaskType
		lastTaskIndex = reply.TaskIndex
		log.Printf("Finished %s task %d\n", reply.TaskType, reply.TaskIndex)
	}
	log.Printf("worker %d is exited\n", id)
}


/*
send an RPC request to the coordinator, wait for the response.
usually returns true.
returns false if something goes wrong.
*/
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	// DialHTTP connects to an HTTP RPC server at the specified network address 
	// listening on the default HTTP RPC path.
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}


/**************example*****************************/



// func Worker(mapf func(string, string) []KeyValue,
// 	reducef func(string, []string) string) {

// 	// Your worker implementation here.

// 	// uncomment to send the Example RPC to the coordinator.
// 	CallExample()

// }


/*
example function to show how to make an RPC call to the coordinator.
the RPC argument and reply types are defined in rpc.go.
*/
// func CallExample() {

// 	// declare an argument structure.
// 	args := ExampleArgs{}
// 	// args := ApplyForTaskArgs{}

// 	// fill in the argument(s).
// 	args.X = 99

// 	// declare a reply structure.
// 	reply := ExampleReply{}
// 	// reply := ApplyForTaskReply{}

// 	// send the RPC request, wait for the reply.
// 	// the "Coordinator.Example" tells the
// 	// receiving server that we'd like to call
// 	// the Example() method of struct Coordinator.
// 	ok := call("Coordinator.Example", &args, &reply)
// 	if ok {
// 		// reply.Y should be 100.
// 		fmt.Printf("reply.Y %v\n", reply.Y)
// 	} else {
// 		fmt.Printf("call failed!\n")
// 	}
// }