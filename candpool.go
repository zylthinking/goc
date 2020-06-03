package goc

import (
	"math/rand"
	"sync"
	"unsafe"
)

type candidate struct {
	weight uint32
	any    unsafe.Pointer
}

type CandPool struct {
	sync.Mutex
	raw    []*candidate
	nr     int
	bg, fg []int
	items  []unsafe.Pointer
}

func (this *CandPool) standardization(raw []*candidate) ([]*candidate, int) {
	nr := len(raw)
	if nr == 0 {
		return nil, 0
	}

	sum := uint64(0)
	for i := 0; i < nr; i++ {
		if raw[i].weight == 0 {
			raw[i].weight = 1
		}
		sum += uint64(raw[i].weight)
	}

	var std []*candidate
	n := uint32(0)
	for i := 0; i < nr; i++ {
		weight := uint32(uint64((raw[i].weight * 100)) / sum)
		std = append(std, &candidate{
			weight: weight,
			any:    raw[i].any,
		})
		n += weight
	}
	return std, int(n)
}

func NewCandPool() *CandPool {
	return &CandPool{}
}

func (this *CandPool) AddCand(weight uint32, any unsafe.Pointer) {
	this.Lock()
	this.raw = append(this.raw, &candidate{
		weight: weight,
		any:    any,
	})
	this.Unlock()
}

func (this *CandPool) DelCand(fptr func(unsafe.Pointer) bool) {
	this.Lock()
	nr := len(this.raw)
	for i := 0; i < nr; {
		if fptr(this.raw[i].any) {
			this.raw[i] = this.raw[nr-1]
			nr--
		} else {
			i++
		}
	}
	this.raw = this.raw[:nr]
	this.Unlock()
}

func (this *CandPool) Shuffle() {
	this.Lock()
	raw := this.raw
	this.Unlock()

	std, sum := this.standardization(raw)
	this.Lock()
	defer this.Unlock()

	this.items = nil
	this.bg = nil
	this.fg = nil
	this.nr = 0
	if std == nil {
		return
	}

	this.items = make([]unsafe.Pointer, sum)
	for i := 0; i < len(std); i++ {
		for j := 0; j < int(std[i].weight); j++ {
			this.items[i] = std[i].any
		}
	}

	this.bg = make([]int, sum)
	this.fg = make([]int, sum)
	for i := 0; i < sum; i++ {
		this.fg[i] = i
	}
	this.nr = sum
}

func (this *CandPool) Pick() unsafe.Pointer {
	rn := rand.Int()
	this.Lock()
	defer this.Unlock()

	nr := len(this.fg)
	if nr == 0 {
		return nil
	}

	idx := rn % this.nr
	n := this.fg[idx]
	this.bg[nr-this.nr] = n
	this.nr--

	if this.nr == 0 {
		this.fg, this.bg, this.nr = this.bg, this.fg, nr
	} else {
		this.fg[idx] = this.fg[this.nr]
	}
	return this.items[n]
}
