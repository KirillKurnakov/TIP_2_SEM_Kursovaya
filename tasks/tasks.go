package tasks

import (
    "encoding/json"
    "os"
    "sync"
    "time"
	"log"
	"context"
)

type Task struct {
    ID       string `json:"id"`
    Status   string `json:"status"`   // pending, in_progress, done, error
    Filename string `json:"filename"`
	cancel     chan struct{}
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
}

var (
	tasks           = make(map[string]*Task)
	mutex           = &sync.Mutex{}
	activeTaskCount = 0
	maxActiveTasks  = 5
	activeTaskSem   = make(chan struct{}, maxActiveTasks)
)

func CreateTask() string {
	taskID := generateID()
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	task := &Task{
		ID:         taskID,
		Status:     "pending",
		cancel:     make(chan struct{}),
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	mutex.Lock()
	tasks[taskID] = task
	mutex.Unlock()

	log.Printf("Task created: ID=%s, Status=%s", taskID, task.Status)
	return taskID
}

func GetTask(taskID string) *Task {
	mutex.Lock()
	defer mutex.Unlock()

	task, exists := tasks[taskID]
	if exists {
		log.Printf("Task retrieved: ID=%s, Status=%s", task.ID, task.Status)
	} else {
		log.Printf("Task not found: ID=%s", taskID)
	}
	return task
}

func RunTask(taskID string) {
	if !acquireTaskSlot() {
		log.Printf("Task ID=%s cannot start. Max concurrent tasks reached.", taskID)
		return
	}
	defer releaseTaskSlot()

	task := GetTask(taskID)
	if task == nil {
		log.Printf("RunTask aborted: Task not found for ID=%s", taskID)
		return
	}

	log.Printf("Task started: ID=%s", task.ID)
	task.Status = "in_progress"

	filename := "export_" + taskID + ".json"
	task.Filename = filename

	for i := 0; i < 20; i++ {
		select {
		case <-task.cancelCtx.Done():
			task.Status = "cancelled"
			log.Printf("Task cancelled: ID=%s", task.ID)
			return
		default:
			// Эмуляция работы
			time.Sleep(1 * time.Second)
			log.Printf("Task in progress: ID=%s, Step=%d", task.ID, i+1)
		}
	}

	// Запись данных в файл
	file, err := os.Create(filename)
	if err != nil {
		task.Status = "error"
		log.Printf("Task failed: ID=%s, Error creating file: %v", task.ID, err)
		return
	}
	defer file.Close()

	data := map[string]string{"message": "Data exported successfully"}
	if err := json.NewEncoder(file).Encode(data); err != nil {
		task.Status = "error"
		log.Printf("Task failed: ID=%s, Error writing file: %v", task.ID, err)
		return
	}

	task.Status = "done"
	log.Printf("Task completed: ID=%s, Filename=%s", task.ID, task.Filename)
}

func CancelTask(taskID string) {
	task := GetTask(taskID)
	if task == nil {
		log.Printf("CancelTask aborted: Task not found for ID=%s", taskID)
		return
	}

	task.cancelFunc()
	close(task.cancel)
	log.Printf("Task cancellation requested: ID=%s", task.ID)
}

func generateID() string {
	return time.Now().Format("20060102150405")
}

func acquireTaskSlot() bool {
	select {
	case activeTaskSem <- struct{}{}:
		mutex.Lock()
		activeTaskCount++
		mutex.Unlock()
		log.Printf("Slot acquired, active tasks: %d", activeTaskCount)
		return true
	default:
		log.Printf("Max active tasks reached, cannot acquire slot.")
		return false
	}
}

func releaseTaskSlot() {
	mutex.Lock()
	activeTaskCount--
	mutex.Unlock()
	<-activeTaskSem
	log.Printf("Slot released, active tasks: %d", activeTaskCount)
}