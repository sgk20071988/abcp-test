package main

import (
	"fmt"
	"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Task represents a meaninglessness of our life
type Task struct {
	id         int
	createTime time.Time // время создания
	finishTime time.Time // время выполнения
	taskResult string
}

const mainSleepTime = time.Second * 3
const workerSleepTime = time.Millisecond * 150

func main() {
	id := 1
	taskСreator := func(a chan Task) {
		go func() {
			for {
				currentTime := time.Now()
				nanosecond := currentTime.Nanosecond()
				if nanosecond/1000%2 > 0 { // вот такое условие появления ошибочных тасков
					currentTime = currentTime.Add(-30 * time.Second)
				}
				a <- Task{taskResult: "", createTime: currentTime, id: id} // передаем таск на выполнение
				id++
			}
		}()
	}

	superChan := make(chan Task, 10)

	go taskСreator(superChan)

	taskWorker := func(task Task) (Task, bool) {
		var succes bool
		if task.createTime.After(time.Now().Add(-20 * time.Second)) {
			task.taskResult = "task has been successed"
			succes = true
		} else {
			task.taskResult = "something went wrong"
			succes = false
		}
		task.finishTime = time.Now()

		time.Sleep(workerSleepTime)

		return task, succes
	}

	var result = struct {
		sync.RWMutex
		tasks []Task
	}{tasks: []Task{}}

	var err = struct {
		sync.RWMutex
		tasks []Task
	}{tasks: []Task{}}

	doneTasks := make(chan Task, 10)
	failedTasks := make(chan Task, 10)

	taskSorter := func(task Task, succes bool) {
		if succes {
			doneTasks <- task
		} else {
			failedTasks <- task
		}
	}

	go func() {
		// получение тасков
		for t := range superChan {
			t, succes := taskWorker(t)
			go taskSorter(t, succes)
		}
		close(superChan)
	}()

	stop := make(chan bool)

	go func() {
		for {
			select {
			case succesTask := <-doneTasks:
				result.tasks = append(result.tasks, succesTask)
			case failedTask := <-failedTasks:
				err.tasks = append(err.tasks, failedTask)
			case <-stop:
				return
			}
		}
	}()

	time.Sleep(mainSleepTime)

	stop <- true

	println("Errors:")
	for _, task := range err.tasks {
		fmt.Printf("id: %v | create time: %v | result: %s \n", task.id, task.createTime.Format(time.RFC3339), task.taskResult)
	}

	println("Done tasks:")
	for _, task := range result.tasks {
		fmt.Printf("id: %v | finish time: %v | result: %s \n", task.id, task.finishTime.Format(time.RFC3339), task.taskResult)
	}
}
