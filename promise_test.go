package promise_test

import (
	"errors"
	"reflect"
	"testing"

	promise "github.com/egoavara/blog-go-promise-test"
)

type Passed struct{ isPassed bool }

func (pass *Passed) Pass() { pass.isPassed = true }
func (pass *Passed) Defer(t *testing.T) {
	if !pass.isPassed {
		t.Errorf("not passed ")
	}
}

func same(t *testing.T, expected any, got any) {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, but got %v", expected, got)
	}
}
func TestNew(t *testing.T) {
	// resolve test
	resolvePromise := promise.New(func(resolve func(int), reject func(error)) {
		resolve(1)
	})
	resolved, rejected := resolvePromise.Await()
	same(t, 1, resolved)
	same(t, nil, rejected)
	// reject test
	var err = errors.New("Hello, error!")
	rejectPromise := promise.New(func(resolve func(int), reject func(error)) {
		reject(err)
	})
	resolved, rejected = rejectPromise.Await()
	same(t, 0, resolved)
	same(t, err, rejected)
}

func TestResolve(t *testing.T) {
	ok, err := promise.Resolve("Hello, World").Await()
	same(t, "Hello, World", ok)
	same(t, nil, err)
}

func TestReject(t *testing.T) {
	var e = errors.New("Hello, Error!")
	ok, err := promise.Reject[any](e).Await()
	same(t, nil, ok)
	same(t, e, err)
}

func TestThen(t *testing.T) {
	var pass = new(Passed)
	defer pass.Defer(t)
	ok, err := promise.Resolve("Hello, World").
		Then(func(s string) {
			same(t, "Hello, World", s)
			pass.Pass()
		}).
		Await()
	same(t, "Hello, World", ok)
	same(t, nil, err)
}

func TestCatch(t *testing.T) {
	var pass = new(Passed)
	defer pass.Defer(t)
	var e = errors.New("Hello, Error!")
	ok, err := promise.Reject[any](e).
		Catch(func(err error) {
			same(t, e, err)
			pass.Pass()
		}).
		Await()
	same(t, nil, ok)
	same(t, e, err)
}

func TestFanally(t *testing.T) {
	var pass = new(Passed)
	defer pass.Defer(t)
	ok, err := promise.Resolve("Hello, World").
		Finally(func() {
			pass.Pass()
		}).
		Await()
	same(t, "Hello, World", ok)
	same(t, nil, err)
}

func TestAll(t *testing.T) {
	// all passed
	ok, fail := promise.All(
		promise.Resolve(1),
		promise.Resolve(2),
		promise.Resolve(3),
		promise.Resolve(4),
	).Await()
	resultSum := 0
	for _, v := range ok {
		resultSum += v
	}
	same(t, 10, resultSum)
	same(t, nil, fail)
	// some passed
	var e = errors.New("Hello, Error!")
	ok, fail = promise.All(
		promise.Resolve(1),
		promise.Resolve(2),
		promise.Resolve(3),
		promise.Resolve(4),
		promise.Reject[int](e),
	).Await()
	same(t, 0, len(ok))
	same(t, e, fail)
}
