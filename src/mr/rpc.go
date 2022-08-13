package mr

/*
RPC definitions.
remember to capitalize all names.
*/

import (
	"time"
	"os"
	"strconv"
	"fmt"
)


const (
	MAP = "MAP"
	REDUCE = "REDUCE"
	DONE = "DONE"
)



/*
 define a Task structure, you know struct can be use everywhere 
 in one packge(with capitalized)
 */
 type Task struct {
	Type string
	Index int
	MapInputFile string // filename
	WorkerID int
	Deadline time.Time
}

/*
declare the arguments and reply for an RPC
*/
type ApplyForTaskArgs struct {
	WorkerID int
	LastTaskType string
	LastTaskIndex int
}

type ApplyForTaskReply struct {
	TaskType string
	TaskIndex int
	MapInputFile string
	nMap int
	nReduce int
}

// Add your RPC definitions here.

/*
define a function for RPC communication between coordinator and worker
parameter: Worker ID, types and index of previous Task
return: type and inded of new task, if null worker can exit
and other information for Worker conduction
	1. if is Map task: corresponding input file name, num of all reduce task
	2. if is Reduce task: need total num of Map Task to generate filename of corresponding intermediate file
*/

/*
Cook up a unique-ish UNIX-domain socket name
in /var/tmp, for the coordinator.
Can't use the current directory since
Athena AFS doesn't support UNIX-domain sockets.
*/
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}


/*some file naming function*/
func tmpMapOutFile(WokrderID int, MapID int, ReduceID int) string {
	return fmt.Sprintf("tmp-worker-%d-%d-%d", WokrderID, MapID, ReduceID)
}

func finalMapOutFile(MapID int, ReduceID int) string {
	return fmt.Sprintf("mr-%d-%d", MapID, ReduceID)
}

func tmpReduceOutFile(WokrderID int, ReduceID int) string {
	return fmt.Sprintf("tmp-worker-%d-out-%d", WokrderID, ReduceID)
}

func finalReduceOutFile(ReduceID int) string {
	return fmt.Sprintf("mr-out-%d", ReduceID)
}



/**************example*****************************/

/*
example to show how to declare the arguments
and reply for an RPC.
*/
// type ExampleArgs struct {
// 	X int
// }

// type ExampleReply struct {
// 	Y int
// }

