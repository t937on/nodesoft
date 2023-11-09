package main

import (
	"fmt"
	"time"
)

const (
	cnt = 10
)

type ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskResult []byte
}

func (a *ttype) taskWorker() {
	tt, _ := time.Parse(time.RFC3339, a.cT)
	if tt.After(time.Now().Add(-20 * time.Second)) {
		a.taskResult = []byte("task has been successed")
	} else {
		a.taskResult = []byte("something went wrong")
	}
	a.fT = time.Now().Format(time.RFC3339Nano)

	time.Sleep(time.Millisecond * 150)
}

func createTask(a chan ttype) {
	defer close(a)
	for {
		ft := time.Now().Format(time.RFC3339)
		if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
			ft = "Some error occured"
		}
		a <- ttype{cT: ft, id: int(time.Now().Unix())} // передаем таск на выполнение
	}
}

func createDoneTask(t ttype, ch chan ttype) {
	ch <- t
}

func createUndoneTask(t ttype, ch chan error) {
	ch <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, t.taskResult)
}

func main() {

	superChan := make(chan ttype, cnt)
	doneTasks := make(chan ttype)
	undoneTasks := make(chan error)

	result := map[int]ttype{}
	err := []error{}

	go createTask(superChan)

	go func() {
		for t := range superChan {
			t.taskWorker()

			if string(t.taskResult[14:]) == "successed" {
				go createDoneTask(t, doneTasks)
			} else {
				go createUndoneTask(t, undoneTasks)
			}

			select {
			case v, ok := <-doneTasks:
				if !ok {
					doneTasks = nil
					continue
				}
				result[v.id] = v
			case v, ok := <-undoneTasks:
				if !ok {
					undoneTasks = nil
					continue
				}
				err = append(err, v)
			}
		}
		close(doneTasks)
		close(undoneTasks)
	}()

	time.Sleep(time.Second * 3)

	fmt.Println("Errors:")
	for i, v := range err {
		fmt.Println(i, v)
	}

	fmt.Println("Done tasks:")
	for k, v := range result {
		fmt.Printf("%d=\"%s\"\n", k, v)
	}

}
