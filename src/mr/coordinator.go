package mr

import (
	"sync"
	"log"
	"net"
	"os"
	"net/rpc"
	"net/http"
	"math"
	"fmt"
	"time"
)

/* note:
// handle application RPC from workder, assigned a task to worker from taskPool
// in this phase we design a function to handle it
// handle notification(task doned) RPC from workder, done task and commmit data
// after all Map task being done, turn into reduce phase and adding reduce task to taskPool
// after all Reduce task being done, marked the completion of MR, exit
// examine the running task periodic, finding those last for over 10s task, assigned them to another worker
*/

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

// start a thread that listens for RPCs from worker.go
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

/*
main/mrcoordinator.go calls Done() periodically to find out
if the entire job has finished.
*/
func (c *Coordinator) Done() bool {
	c.lock.Lock()
	ret := (c.stage == DONE)
	defer c.lock.Unlock()
	return ret
}



/* Your code here -- RPC handlers for the worker to call.*/

//generate taskID with Type and Index
func genTaskID(Type string, Id int) string {
	return fmt.Sprintf("%s-%d", Type, Id)
}

// change phase
func (c *Coordinator)transit() {
	if c.stage == MAP {
		log.Printf("all map tasks have been done\n")
		c.stage = REDUCE
		for i:=0; i<c.nReduce; i++ {
			task := Task{
				Type : REDUCE,
				Index : i,
			}
			c.tasks[genTaskID(task.Type, task.Index)] = task
			c.taskPool <- task
		}
	}else if c.stage == REDUCE {
		log.Printf("all reduce tasks have been done, and MR finished\n")
		close(c.taskPool)
		c.stage = DONE
	}
}

/*
an RPC handler
the RPC argument and reply types are defined in rpc.go
call by worker
*/
func (c *Coordinator)ApplyForTask(args *ApplyForTaskArgs, reply *ApplyForTaskReply) error {
	// if reply.LastTaskType is not nil
	// in the current phase, all tasks are completed, enter next phase
	log.Printf("args.LastTaskType: %s", args.LastTaskType)
	if args.LastTaskType != "" {
		c.lock.Lock()
		lastTaskID := genTaskID(args.LastTaskType, args.LastTaskIndex)
		lastTask, ok:= c.tasks[lastTaskID]
		// if this task still belongs to worker
		if ok && lastTask.WorkerID == args.WorkerID {
			// cause this task was finished, we will rename th tmpOuput as finalOutput
			if args.LastTaskType == MAP {
				// rename
				for ri:=0; ri<c.nReduce; ri++ {
					tmpName := tmpMapOutFile(args.WorkerID, args.LastTaskIndex, ri)
					finalName := finalMapOutFile(args.LastTaskIndex, ri)
					err := os.Rename(tmpName, finalName)
					if err != nil {
						log.Fatalf("map failed to rename %s to %s by error: %s\n",tmpName, finalName, err)
					}
				}
			}else if args.LastTaskType == REDUCE {
				// rename
				tmpName := tmpReduceOutFile(args.WorkerID, args.LastTaskIndex)
				finalName := finalReduceOutFile(args.LastTaskIndex)
				err := os.Rename(tmpName, finalName)
				if err != nil {
					log.Fatalf("reduce failed to rename %s to %s by error: %s\n",tmpName, finalName, err)
				}
			}
			// after that, since this task was finished, we delete it from tasks map
			delete(c.tasks, lastTaskID)
			// if all tasks were done, then turn into next phase
			if len(c.tasks) == 0{
				c.transit()
			}
		}
		c.lock.Unlock()
		// if not, just ignore it
	}
	// if reply.LastTaskTYpe is nil, then we assign a task in worker
	// get a availble task from coordinator.taskPool to tmp "task"
	task, err := <-c.taskPool
	if !err {
		return nil// channel turn off, there are no task to do, MR done
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	log.Printf("coordinator is assigning task to worker")
	// 
	task.WorkerID = args.WorkerID
	task.Deadline = time.Now().Add(10*time.Second)
	// task information
	reply.TaskType = task.Type
	reply.TaskIndex = task.Index
	reply.MapInputFile = task.MapInputFile
	// coordinator's information
	reply.nMap = c.nMap
	reply.nReduce = c.nReduce

	return nil
}


/*
create a Coordinator.
main/mrcoordinator.go calls this function.
nReduce is the number of reduce tasks to use.
*/
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	// part 1
	// generate specific Map task based on files[] and put into taskPool
	// make coordinator and generate Map task
	c := Coordinator{
		nMap : len(files),
		nReduce : nReduce,
		stage : MAP,
		tasks : make(map[string]Task),
		taskPool : make(chan Task, int(math.Max(float64(len(files)), float64(nReduce)))),
	}

	// produce tasks base files, then marking task in c.tasks and putting into channel
	for i, file := range files {
		task := Task {
			Type : MAP,
			Index : i,
			MapInputFile : file,
		}
		c.tasks[genTaskID(task.Type, task.Index)] = task
		c.taskPool <- task
	}
	// turn on coordinator and begin to response worker's request
	log.Printf("coordinator start!\n")
	c.server()

	// part 2
	// coordinator periodically examine tasks, to reassign those failed tasks
	go func() {
		time.Sleep(500*time.Millisecond)
		c.lock.Lock()
		for _, task := range c.tasks {// string : Task
			if task.WorkerID != -1 && time.Now().After(task.Deadline) {
				log.Printf("find time-out %s task %d previously running on %d worker", task.Type, task.Index, task.WorkerID)
				task.WorkerID = -1
				c.taskPool <- task
			}
		}
		c.lock.Unlock()
	}()

	return &c
}








/**************example*****************************/

/*
an example RPC handler.
the RPC argument and reply types are defined in rpc.go.
*/
// func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
// 	reply.Y = args.X + 1
// 	return nil
// }


// func MakeCoordinator(files []string, nReduce int) *Coordinator {
// 	c := Coordinator{}
// 	c.server()
// 	return &c
// }