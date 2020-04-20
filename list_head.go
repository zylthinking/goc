package goc

type ListHead struct {
	Next, Prev *ListHead
	addr       interface{}
}

func InitListHead(head *ListHead, addr ...interface{}) {
	head.Next = head
	head.Prev = head
	if len(addr) != 0 {
		head.addr = addr[0]
	}
}

func ListEntry(head *ListHead) interface{} {
	return head.addr
}

func list_add(new, prev, next *ListHead) {
	next.Prev = new
	new.Next = next
	new.Prev = prev
	prev.Next = new
}

func ListAdd(entry, head *ListHead) {
	list_add(entry, head, head.Next)
}

func ListAddTail(entry, head *ListHead) {
	list_add(entry, head.Prev, head)
}

func list_del(prev, next *ListHead) {
	next.Prev = prev
	prev.Next = next
}

func ListDel(head *ListHead) {
	list_del(head.Prev, head.Next)
}

func ListDelInit(head *ListHead) {
	list_del(head.Prev, head.Next)
	InitListHead(head)
}

func ListEmpty(head *ListHead) bool {
	return head.Next == head
}

func ListJoin(one, two *ListHead) {
	one = one.Prev
	one.Next.Prev = two.Prev
	two.Prev.Next = one.Next
	one.Next = two
	two.Prev = one
}

func ListSplit(one, two *ListHead) {
	entry := two.Prev
	one.Prev.Next = two
	two.Prev = one.Prev
	one.Prev = entry
	entry.Next = one
}
