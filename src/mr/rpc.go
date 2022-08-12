package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// define a function for RPC communication between coordinator and worker
// parameter: Worker ID, types and index of previous Task
// return: type and inded of new task, if null worker can exit
//	and other information for Worker conduction
//		1. if is Map task: corresponding input file name, num of all reduce task
//		2. if is Reduce task: need total num of Map Task to generate filename of corresponding intermediate file


func ApplyForTask(WorkID string, TaskType string) bool {
	return true
}


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
