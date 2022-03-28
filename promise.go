package promise

import (
	"sync"
)

type Promise[T any] struct {
	result  chan T
	fail    chan error
	then    []func(T)
	catch   []func(error)
	finally []func()
}

func New[T any](handler func(resolve func(T), reject func(error))) *Promise[T] {
	prom := &Promise[T]{
		result:  make(chan T, 1),
		fail:    make(chan error, 1),
		then:    make([]func(T), 0, 1),
		catch:   make([]func(error), 0, 1),
		finally: make([]func(), 0, 1),
	}
	go handler(
		func(t T) {
			prom.result <- t
			close(prom.fail)
		},
		func(e error) {
			close(prom.result)
			prom.fail <- e
		},
	)
	return prom
}

func Resolve[T any](data T) *Promise[T] {
	prom := &Promise[T]{
		result:  make(chan T, 1),
		fail:    make(chan error, 0),
		then:    make([]func(T), 0, 1),
		catch:   make([]func(error), 0, 0),
		finally: make([]func(), 0, 1),
	}
	prom.result <- data
	close(prom.fail)
	return prom
}

func Reject[T any](data error) *Promise[T] {
	prom := &Promise[T]{
		result:  make(chan T, 0),
		fail:    make(chan error, 1),
		then:    make([]func(T), 0, 0),
		catch:   make([]func(error), 0, 1),
		finally: make([]func(), 0, 1),
	}
	close(prom.result)
	prom.fail <- data
	return prom
}

func (prom *Promise[T]) Await() (T, error) {
	var isCatched = false
	var resultT T
	var resultE error = nil
	select {
	case res, isResOk := <-prom.result:
		if isResOk {
			resultT = res
		} else {
			resultE = <-prom.fail
			isCatched = true
		}
	case err, isErrOk := <-prom.fail:
		if isErrOk {
			resultE = err
			isCatched = true
		} else {
			resultT = <-prom.result
		}
	}
	if !isCatched {
		// handler then
		for _, then := range prom.then {
			then(resultT)
		}
		close(prom.result)
	} else {
		// handler catch
		for _, catch := range prom.catch {
			catch(resultE)
		}

		close(prom.fail)
	}
	for _, finally := range prom.finally {
		finally()
	}
	return resultT, resultE
}

func (prom *Promise[T]) Then(handler func(T)) *Promise[T] {
	prom.then = append(prom.then, handler)
	return prom
}

func (prom *Promise[T]) Catch(handler func(error)) *Promise[T] {
	prom.catch = append(prom.catch, handler)
	return prom
}

func (prom *Promise[T]) Finally(handler func()) *Promise[T] {
	prom.finally = append(prom.finally, handler)
	return prom
}

func All[T any](promises ...*Promise[T]) *Promise[[]T] {
	prom := &Promise[[]T]{
		result:  make(chan []T, 0),
		fail:    make(chan error, 1),
		then:    make([]func([]T), 0, 0),
		catch:   make([]func(error), 0, 0),
		finally: make([]func(), 0, 0),
	}
	data := make([]T, 0, len(promises))
	mtx := new(sync.Mutex)

	for _, v := range promises {
		go func(each *Promise[T]) {
			// if failed promise, below code must be paniced
			defer recover()
			t, err := each.Await()
			if err != nil {
				// it cause panic, cause below comment
				close(prom.result)
				prom.fail <- err
				return
			}
			mtx.Lock()
			data = append(data, t)
			if len(data) == len(promises) {
				// panic if there is result closed by above
				prom.result <- data
				close(prom.fail)
			}
			mtx.Unlock()
		}(v)
	}
	return prom

}
