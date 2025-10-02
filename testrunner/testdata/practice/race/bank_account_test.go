package account

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcDeposit(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skip("Multiple CPU cores required for concurrency tests.")
	}
	previousMaxProcs := runtime.GOMAXPROCS(0)

	if previousMaxProcs < 2 {
		runtime.GOMAXPROCS(2)
		defer runtime.GOMAXPROCS(previousMaxProcs)
	}

	a := Open(0)
	if a == nil {
		t.Fatal("Open(0) = nil, want non-nil *Account.")
	}
	const amt = 10
	const c = 1000
	var negBal int32
	var start, g sync.WaitGroup
	start.Add(1)
	g.Add(3 * c)
	for i := 0; i < c; i++ {
		go func() { // deposit
			start.Wait()
			a.Deposit(amt) // ignore return values
			g.Done()
		}()
		go func() { // withdraw
			start.Wait()
			for {
				if _, ok := a.Deposit(-amt); ok {
					break
				}
				time.Sleep(time.Microsecond) // retry
			}
			g.Done()
		}()
		go func() { // watch that balance stays >= 0
			start.Wait()
			if p, _ := a.Balance(); p < 0 {
				atomic.StoreInt32(&negBal, 1)
			}
			g.Done()
		}()
	}
	start.Done()
	g.Wait()
	if negBal == 1 {
		t.Fatal("Balance went negative with concurrent deposits and " +
			"withdrawals.  Want balance always >= 0.")
	}
	if p, ok := a.Balance(); !ok || p != 0 {
		t.Fatalf("After equal concurrent deposits and withdrawals, "+
			"a.Balance = %d, %t.  Want 0, true", p, ok)
	}
}
