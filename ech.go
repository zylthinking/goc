package goc


type EventChan struct {
	ch   chan int
	stat int32
}

func NewEventChan() *EventCh {
	return &EventChan{
		ch: make(chan int, 1),
	}
}

func (this *EventChan) Chan() <-chan int {
	return this.ch
}

func (this *EventChan) Notify() {
	if atomic.CompareAndSwapInt32(&this.stat, 0, 1) {
		this.ch <- 1
	}
}

func (this *EventChan) Reset() {
	atomic.StoreInt32(&this.stat, 0)
}
