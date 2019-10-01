package rpc

import (
    "reflect"
    "unsafe"
)

type TrieUptr interface {
    Free(interface{});
}

type Trie struct {
    link ListHead;
    leaf *Trie;
    parent *Trie;
    child [256]*Trie;
    level int;
    branch int;
    Uptr interface{};
}

func QuickBytes(str string) []byte {
    var buf []byte;
    pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&buf));
    pstring := (*reflect.StringHeader)(unsafe.Pointer(&str))
    pbytes.Data = pstring.Data;
    pbytes.Len = pstring.Len;
    pbytes.Cap = pstring.Len;
    return buf;
}

func NewTrie() *Trie {
    trie := &Trie {};
    InitListHead(&trie.link, trie);
    return trie;
}

func (trie *Trie) Replace(key string, uptr interface{}) interface{} {
    cursor := trie;

    intp := QuickBytes(key);
    for i := 0; i < len(intp); i++ {
        key := intp[i];
        if (cursor.child[key] == nil) {
            cursor.branch += 1;
            cursor.child[key] = NewTrie();
            cursor.child[key].level = cursor.level + 1;
            cursor.child[key].leaf = cursor.leaf;
            cursor.child[key].parent = cursor;
            ListAddTail(&cursor.child[key].link, &trie.link);
        }
        cursor = cursor.child[key];
    }

    var ptr interface{};
    if (cursor != trie) {
        ptr = cursor.Uptr;
        cursor.leaf = cursor;
        cursor.Uptr = uptr;
        cursor.leaf_adjust(cursor);
    }
    return ptr;
}

func (trie *Trie) leaf_adjust(leaf *Trie) {
    for i := 0; i < 256; i++ {
        child := trie.child[i];
        if (child == nil || child.leaf == child || child.leaf == leaf) {
            continue;
        }
        child.leaf = leaf;
        child.leaf_adjust(leaf);
    }
}

func (trie *Trie) Find(key string) (*Trie, bool) {
    intp := QuickBytes(key);
    for i := 0; i < len(key); i++ {
        key := intp[i];
        if (trie.child[key] == nil) {
            break;
        }
        trie = trie.child[key];
    }

    leaf := trie.leaf;
    return leaf, (trie == leaf);
}

func (parent *Trie) unlink(idx uint8) {
    node := parent.child[idx];
    ListDelInit(&node.link);
    node.leaf = nil;
    node.parent = nil;
    parent.child[idx] = nil;
    parent.branch--;
}

func (trie *Trie) Del(key string) interface{} {
    trie, b := trie.Find(key);
    if (!b) {
        return nil;
    }

    uptr := trie.Uptr;
    trie.Uptr = nil;
    var intp = QuickBytes(key);
    if (trie.branch > 0) {
        trie.leaf = trie.parent.leaf;
        trie.leaf_adjust(trie.leaf);
        return uptr;
    }

    for i := len(intp) - 1; i >= 0; i-- {
        trie = trie.parent;
        trie.unlink(intp[i]);
        if (trie.branch > 0 || trie.leaf == trie) {
            break;
        }
    }
    return uptr;
}

func (trie *Trie) clear() {
    if (trie.branch == 0) {
        return;
    }

    ListDelInit(&trie.link);
    trie.parent = nil;
    trie.leaf = nil;

    for i := 0; i < 256; i++ {
        trie.child[i] = nil;
    }
    trie.branch = 0;
}

func (trie *Trie) Clear() {
    for (!ListEmpty(&trie.link)) {
        ent := trie.link.Next;
        node := ListEntry(ent).(*Trie);
        node.clear();

        if (node.Uptr == nil) {
            continue;
        }

        it, ok := node.Uptr.(*TrieUptr);
        if (ok) {
            (*it).Free(node.Uptr);
        }
    }
    trie.clear();
}
