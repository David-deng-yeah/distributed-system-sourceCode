package mr

import (
	"sync"
	"log"
	"net"
	"os"
	"net/rpc"
	"net/http"
	"math"
)

// define a Task structure, you know struct can be use everywhere in one packge(with capitalized)
type Task struct {
	Type string
	Index int
	MapInputFile string // filename
}

// coordinator should maintain those status information
type Coordinator struct {
	// Your definitions here.
	lock sync.Mutex
	// 1.basic configure information: num of total Map tasks, num of total reduce task
	nMap int
	nReduce int
	// 2.schedule needed information: 
	// 	2.1 current status: Map or Reduce ?
	stage string
	//	2.2 all {runing Task, belonging Worker and Deadline(optional)}
	tasks map[string]Task
	//	2.3 a task-pool contains all unassigned tasks, for responding worker's application and Failover, implemented y golang channel
	taskPool chan Task
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false
	// Your code here.


	return ret
}

// generate taskID with Type and Index
func genTaskID(Type string, Id int) string {
	return fmt.Sprintf("%s-%v", Type, Id)
}

// ApplyForTask as the entrance of RPC
// parameter: reply is what we return to worker
func ApplyForTask(args *ApplyForTaskArgs, reply *ApplyForTaskReply) error {

	// 1. in the current phase, all tasks are completed, enter next phase

	// 2. get a availble task from coordinator.taskPool to tmp "task"
}


//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	//c := Coordinator{}

	// Your code here.
	// 1. generate specific Map task based on files[] and put into taskPool
	// 1.1 make coordinator and generate Map task
	c := Coordinator{
		nMap : len(files),
		nReduce : nReduce,
		stage : Map,
		tasks : make(map[string]Task),
		taskPool : make(chan Task, int(math.Max(float64(len(files)), float64(nReduce)))),
	}

	for i, file := range files {
		task := Task {
			Type : Map,
			Index : i,
			MapInputFile : file,
		}
		c.tasks[genTaskID(task.Type, task.Index)] := task
		c.taskPool <- task
	}
	// 1.2 turn on coordinator and begin to response worker's request
	log.Printf("coordinator start!\n")
	c.server()

	// 2. handle application RPC from workder, assigned a task to worker from taskPool
	// in this phase we design a function to handle it
	// 3. handle notification(task doned) RPC from workder, done task and commmit data
	// 4. after all Map task being done, turn into reduce phase and adding reduce task to taskPool
	// 5. after all Reduce task being done, marked the completion of MR, exit
	// 6. examine the running task periodic, finding those last for over 10s task, assigned them to another worker


	//c.server()



	return &c
}
