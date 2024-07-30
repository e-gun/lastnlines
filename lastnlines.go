package lastnlines

import (
	"fmt"
	"github.com/e-gun/safestack"
	"github.com/nxadm/tail"
	"strings"
	"time"
)

type LastNLines struct {
	stack  *safestack.SafeStack[string]
	tailer *tail.Tail
	tcfg   tail.Config
	stop   chan struct{}
	live   bool
}

func (ll *LastNLines) Start() {
	ll.live = true
	go func() {
		// this function will panic at "case line := <-ll.tailer.Lines" in the for loop if the tail file disappears
		// hence the defer recover()
		defer func() {
			if r := recover(); r != nil {
				// ll.SetDepth(1)
				// ll.stack.Push("(the watched file was deleted)")
				// tail itself will send a message: "Stopping tail as file no longer exists"
				ll.SetDepth(0)
				ll.live = false
			}
		}()

		for {
			select {
			case line := <-ll.tailer.Lines:
				ll.stack.Push(line.Text)
			case <-ll.stop:
				_ = ll.tailer.Stop()
				return
			}
		}
	}()

}

func (ll *LastNLines) Stop() {
	ll.live = false
	ll.stop <- struct{}{}
}

func (ll *LastNLines) SetDepth(d int) {
	ll.stack.NewMax(d)
}

func (ll *LastNLines) LastItem() (string, error) {
	return ll.stack.Peek()
}

func (ll *LastNLines) Get() []string {
	return ll.GetFILO()
}

func (ll *LastNLines) GetFILO() []string {
	return ll.stack.PeekAtSlice()
}

func (ll *LastNLines) GetLIFO() []string {
	return ll.stack.PeekAll()
}

func (ll *LastNLines) IsAlive() bool {
	return ll.live
}

func NewLNL(f string) *LastNLines {
	lnl := lnlfactory(f)
	lnl.stack = safestack.NewSafeStack([]string{})
	return lnl
}

func lnlfactory(f string) *LastNLines {
	tc := tail.Config{
		Location:      nil,
		ReOpen:        false,
		MustExist:     false,
		Poll:          false,
		Pipe:          false,
		Follow:        true,
		MaxLineSize:   0,
		CompleteLines: false,
		RateLimiter:   nil,
		Logger:        nil,
	}

	t, err := tail.TailFile(f, tc)
	if err != nil {
		panic(err)
	}

	return &LastNLines{
		stack:  nil,
		tailer: t,
		tcfg:   tc,
		stop:   make(chan struct{}),
	}
}

func main() {
	ln := NewLNL("./test.txt")
	ln.SetDepth(10)
	ln.Start()
	iter := 0
	stop := false
	for {
		iter += 1
		fmt.Println(iter)
		for _, line := range ln.Get() {
			fmt.Println(line)
			if strings.Contains(line, "stop") {
				ln.Stop()
				stop = true
			}
		}
		if stop {
			break
		}
		time.Sleep(3 * time.Second)
	}
}
