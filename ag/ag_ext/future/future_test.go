package future

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func TestFuture(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42
	close(ch)

	val, ok := <-ch
	fmt.Println(val, ok) // 输出: 42 true ✅

	val2, ok2 := <-ch
	fmt.Println(val2, ok2) // 输出: 0 false
}

// func init() {
// 	// It releases the default pool from ants.
// 	ants.Release()
// }

func TestFutureGPanic(t *testing.T) {
	// ants.Reboot()

	gpanic1()
	gpanic2()
	gpanic3()
	time.Sleep(time.Second)
	fmt.Println("主goroutine未被panic中断")
}

func gpanic1() {
	fut := NewFutureFunc(func() (interface{}, error) {
		panic("gpanic1")
	})
	_, err := fut()
	if err != nil {
		// fmt.Printf("gpanic1: %v\n", err)
		slog.Error("gpanic1", "err", err)
	}
}

func gpanic2() {
	FutureCall(func() (interface{}, error) {
		panic("gpanic2")
	}, func(res interface{}, err error) {
		if err != nil {
			slog.Error("gpanic2", "err", err)
		}
	})
}

func gpanic3() {
	fut := NewFuture(func() (interface{}, error) {
		panic("gpanic3")
	})
	_, err := fut.Await(context.Background())
	if err != nil {
		// fmt.Printf("gpanic3: %v\n", err)
		slog.Error("gpanic3", "err", err)
	}
}

// =============================================================================
// V2 (NewFuture / Await) — 有 recover 保护
// =============================================================================

func TestNewFuture_PanicString(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		panic("boom")
	})

	val, err := fut.Await(context.Background())
	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
	if err.Error() != "panic: boom" {
		t.Fatalf("expected error 'panic: boom', got '%v'", err)
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %v", val)
	}
}

func TestNewFuture_PanicError(t *testing.T) {
	fut := NewFuture(func() (string, error) {
		panic(errors.New("something went wrong"))
	})

	val, err := fut.Await(context.Background())
	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
	if err.Error() != "panic: something went wrong" {
		t.Fatalf("expected error 'panic: something went wrong', got '%v'", err)
	}
	if val != "" {
		t.Fatalf("expected zero value, got '%v'", val)
	}
}

func TestNewFuture_PanicInt(t *testing.T) {
	fut := NewFuture(func() (float64, error) {
		panic(42)
	})

	val, err := fut.Await(context.Background())
	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
	if err.Error() != "panic: 42" {
		t.Fatalf("expected error 'panic: 42', got '%v'", err)
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %v", val)
	}
}

func TestNewFuture_PanicStruct(t *testing.T) {
	type myPanic struct {
		msg string
	}
	fut := NewFuture(func() (bool, error) {
		panic(myPanic{msg: "custom"})
	})

	val, err := fut.Await(context.Background())
	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
	if val != false {
		t.Fatalf("expected zero value, got %v", val)
	}
}

func TestNewFuture_NoPanic(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		return 42, nil
	})

	val, err := fut.Await(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestNewFuture_NoPanicWithError(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		return 0, errors.New("task failed")
	})

	val, err := fut.Await(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "task failed" {
		t.Fatalf("expected 'task failed', got '%v'", err)
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %v", val)
	}
}

func TestNewFuture_PanicWithContextCancellation(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		panic("delayed boom")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	val, err := fut.Await(ctx)
	if err == nil {
		t.Fatal("expected error (context deadline or panic), got nil")
	}
	if fmt.Sprintf("%v", err) != "context deadline exceeded" {
		t.Fatalf("expected context deadline exceeded, got '%v'", err)
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %v", val)
	}
}

func TestNewFuture_MultipleAwaitAfterPanic(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		panic("multi-boom")
	})

	// 第一次 Await: 成功消费，拿到 panic error
	val1, err1 := fut.Await(context.Background())
	// 第二次 Await: 被 CAS 拒绝，返回 "future already consumed"
	val2, err2 := fut.Await(context.Background())

	if err1 == nil {
		t.Fatal("first await should return panic error")
	}
	if err1.Error() != "panic: multi-boom" {
		t.Fatalf("first await: expected 'panic: multi-boom', got '%v'", err1)
	}
	if val1 != 0 {
		t.Fatalf("first await: expected zero value, got %v", val1)
	}
	if err2 == nil {
		t.Fatal("second await should be rejected")
	}
	if err2.Error() != "future already consumed" {
		t.Fatalf("second await: expected 'future already consumed', got '%v'", err2)
	}
	if val2 != 0 {
		t.Fatalf("second await: expected zero value, got %v", val2)
	}
}

func TestNewFuture_DuplicateAwaitRejected(t *testing.T) {
	fut := NewFuture(func() (string, error) {
		return "result", nil
	})

	val1, err1 := fut.Await(context.Background())
	if err1 != nil {
		t.Fatalf("first await: unexpected error: %v", err1)
	}
	if val1 != "result" {
		t.Fatalf("first await: expected 'result', got '%v'", val1)
	}

	val2, err2 := fut.Await(context.Background())
	if err2 == nil {
		t.Fatal("second await should be rejected")
	}
	if err2.Error() != "future already consumed" {
		t.Fatalf("second await: expected 'future already consumed', got '%v'", err2)
	}
	if val2 != "" {
		t.Fatalf("second await: expected zero value, got '%v'", val2)
	}
}

func TestNewFuture_ConcurrentAwaitRace(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		time.Sleep(50 * time.Millisecond)
		return 99, nil
	})

	var successCount int32
	var rejectCount int32
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			_, err := fut.Await(context.Background())
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else if err.Error() == "future already consumed" {
				atomic.AddInt32(&rejectCount, 1)
			} else {
				t.Errorf("unexpected error: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	if successCount != 1 {
		t.Fatalf("exactly 1 goroutine should succeed, got %d", successCount)
	}
	if rejectCount != 4 {
		t.Fatalf("exactly 4 goroutines should be rejected, got %d", rejectCount)
	}
}

func TestNewFuture_ContextCancelLeavesFutureUnconsumable(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 42, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err1 := fut.Await(ctx)
	if err1 == nil {
		t.Fatal("first await (cancelled) should return error")
	}

	// context 取消后，consumed 已被 CAS 设为 1，后续 Await 也应被拒绝
	_, err2 := fut.Await(context.Background())
	if err2 == nil {
		t.Fatal("second await should be rejected after context cancel consumed it")
	}
	if err2.Error() != "future already consumed" {
		t.Fatalf("expected 'future already consumed', got '%v'", err2)
	}
}

func TestNewFuture_PanicInGoroutineWithSharedFuture(t *testing.T) {
	fut := NewFuture(func() (int, error) {
		time.Sleep(50 * time.Millisecond)
		panic("shared panic")
	})

	// 多个 goroutine 同时 Await
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			_, err := fut.Await(context.Background())
			if err == nil {
				t.Errorf("expected panic error")
			}
			done <- true
		}()
	}

	for i := 0; i < 3; i++ {
		<-done
	}
}

// =============================================================================
// V1 (NewFutureFunc) — 无 recover 保护，验证行为
// =============================================================================

func TestNewFutureFunc_Panic(t *testing.T) {
	getter := NewFutureFunc(func() (interface{}, error) {
		panic("v1 boom")
	})

	val, err := getter()
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
	if err.Error() != "panic: v1 boom" {
		t.Fatalf("expected 'panic: v1 boom', got '%v'", err)
	}
}

func TestNewFutureFunc_PanicError(t *testing.T) {
	getter := NewFutureFunc(func() (interface{}, error) {
		panic(errors.New("v1 error panic"))
	})

	val, err := getter()
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
	if err.Error() != "panic: v1 error panic" {
		t.Fatalf("expected 'panic: v1 error panic', got '%v'", err)
	}
}

func TestNewFutureFunc_NoPanicSuccess(t *testing.T) {
	getter := NewFutureFunc(func() (interface{}, error) {
		return "hello", nil
	})

	val, err := getter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hello" {
		t.Fatalf("expected 'hello', got %v", val)
	}
}

func TestNewFutureFunc_NoPanicError(t *testing.T) {
	getter := NewFutureFunc(func() (interface{}, error) {
		return nil, errors.New("normal error")
	})

	val, err := getter()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "normal error" {
		t.Fatalf("expected 'normal error', got '%v'", err)
	}
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
}

func TestNewFutureFunc_NoPanicInt(t *testing.T) {
	getter := NewFutureFunc(func() (interface{}, error) {
		return 100, nil
	})

	val, err := getter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 100 {
		t.Fatalf("expected 100, got %v", val)
	}
}

// =============================================================================
// V1 (FutureCall) — 无 recover 保护
// =============================================================================

func TestFutureCall_Panic(t *testing.T) {
	callbackCh := make(chan error, 1)

	FutureCall(func() (interface{}, error) {
		panic("callback panic")
	}, func(res interface{}, err error) {
		callbackCh <- err
	})

	select {
	case err := <-callbackCh:
		if err == nil {
			t.Fatal("expected error from panic, got nil")
		}
		if err.Error() != "panic: callback panic" {
			t.Fatalf("expected 'panic: callback panic', got '%v'", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for callback")
	}
}

func TestFutureCall_NoPanic(t *testing.T) {
	callbackCh := make(chan struct {
		res interface{}
		err error
	}, 1)

	FutureCall(func() (interface{}, error) {
		return "ok", nil
	}, func(res interface{}, err error) {
		callbackCh <- struct {
			res interface{}
			err error
		}{res, err}
	})

	select {
	case result := <-callbackCh:
		if result.res != "ok" {
			t.Fatalf("expected 'ok', got %v", result.res)
		}
		if result.err != nil {
			t.Fatalf("expected nil error, got %v", result.err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for callback")
	}
}

func TestFutureCall_Error(t *testing.T) {
	callbackCh := make(chan error, 1)

	FutureCall(func() (interface{}, error) {
		return nil, errors.New("task error")
	}, func(res interface{}, err error) {
		callbackCh <- err
	})

	select {
	case err := <-callbackCh:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "task error" {
			t.Fatalf("expected 'task error', got '%v'", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for callback")
	}
}

func TestFutureCall_PanicError(t *testing.T) {
	callbackCh := make(chan error, 1)

	FutureCall(func() (interface{}, error) {
		panic(errors.New("futurecall panic"))
	}, func(res interface{}, err error) {
		callbackCh <- err
	})

	select {
	case err := <-callbackCh:
		if err == nil {
			t.Fatal("expected error from panic, got nil")
		}
		if err.Error() != "panic: futurecall panic" {
			t.Fatalf("expected 'panic: futurecall panic', got '%v'", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for callback")
	}
}

