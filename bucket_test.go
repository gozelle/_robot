package _robot

import (
	"testing"
	"time"
)

func TestNewBucket(t *testing.T) {
	bucket := NewBucket(5 * time.Second)
	go func() {
		for {
			bucket.Push(time.Now().String())
			time.Sleep(500 * time.Millisecond)
		}
	}()
	
	bucket.PopTimely(func(messages []interface{}) {
		t.Log(len(messages), messages)
	})
}

func TestNewBucket2(t *testing.T) {
	bucket := NewBucket(5 * time.Second)
	go func() {
		for {
			bucket.Push(time.Now().String())
			time.Sleep(500 * time.Millisecond)
		}
	}()
	
	bucket.PopLazily(func(messages []interface{}) {
		t.Log(len(messages), messages)
	})
}

//func TestNewBucket3(t *testing.T) {
//	bucket := NewBucket(1 * time.Second)
//
//	go func() {
//		i := 0
//		for {
//			i++
//			bucket.PushWait(time.Now().String(), time.Duration(i)*time.Second)
//			//time.Sleep(500 * time.Millisecond)
//			time.Sleep(1 * time.Second)
//		}
//	}()
//
//	bucket.PopLazily(func(messages []interface{}) {
//		t.Log(len(messages), messages)
//	})
//}
