package goc

import (
    "reflect"
    "unsafe"
)

type TrieUptr interface {
    Free(interface{});
}

type Trie struct {
    link ListHead;
    tree, treentry ListHead;
    leaf *Trie;
    parent *Trie;
    child [256]*Trie;
    branch, Level int;
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
    InitListHead(&trie.tree, trie);
    InitListHead(&trie.treentry, trie);
    return trie;
}

func ResetTree(trie *Trie) {
    if (trie.parent != nil) {
        return;
    }
    trie.reset();
}

func (trie *Trie) Children() *ListHead {
    leaf := trie.leaf;
    if (leaf == nil) {
        leaf = trie;
    }
    return &leaf.tree;
}

func (trie *Trie) leavesJoin(leaf *Trie, leaves *ListHead) {
    if (leaf == nil) {
        leaf = trie;
    }
    ListJoin(leaves, &leaf.tree);
    ListDelInit(leaves);
}

func (trie *Trie) Replace(key string, uptr interface{}) interface{} {
    cursor := trie;

    intp := QuickBytes(key);
    for i := 0; i < len(intp); i++ {
        key := intp[i];
        if (cursor.child[key] == nil) {
            cursor.branch += 1;
            cursor.child[key] = NewTrie();
            cursor.child[key].Level = cursor.Level + 1;
            cursor.child[key].leaf = cursor.leaf;
            cursor.child[key].parent = cursor;
            ListAddTail(&cursor.child[key].link, &trie.link);
        }
        cursor = cursor.child[key];
    }

    var ptr interface{};
    if (cursor != trie) {
        ptr = cursor.Uptr;
        cursor.Uptr = uptr;

        if (cursor.leaf != cursor) {
            leaf := cursor.leaf;
            if (leaf == nil) {
                leaf = trie;
            }
            ListAdd(&cursor.treentry,  &leaf.tree);
            cursor.leaf = cursor;
            cursor.leaf_adjust(cursor);
        }
    }
    return ptr;
}

func (trie *Trie) leaf_adjust(leaf *Trie) {
    for i := 0; i < 256; i++ {
        child := trie.child[i];
        if (child == nil) {
            continue;
        }

        if (child.leaf == child) {
            if (leaf == nil) {
                leaf = trie;
            }
            ListDel(&child.treentry);
            ListAdd(&child.treentry, &leaf.tree);
            continue;
        }
        child.leaf = leaf;
        child.leaf_adjust(leaf);
    }
}

func (trie *Trie) Find(key string) (*Trie, *Trie) {
    intp := QuickBytes(key);
    for i := 0; i < len(key); i++ {
        key := intp[i];
        if (trie.child[key] == nil) {
            break;
        }
        trie = trie.child[key];
    }
    return trie, trie.leaf;
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
    found, leaf := trie.Find(key);
    if (found != leaf) {
        return nil;
    }

    uptr := found.Uptr;
    found.Uptr = nil;
    ListDelInit(&found.treentry);
    trie.leavesJoin(found.parent.leaf, &found.tree);

    var intp = QuickBytes(key);
    if (found.branch > 0) {
        found.leaf = found.parent.leaf;
        found.leaf_adjust(found.leaf);
        return uptr;
    }

    for i := len(intp) - 1; i >= 0; i-- {
        parent := found.parent;
        parent.unlink(intp[i]);
        if (parent.branch > 0 || parent.leaf == parent) {
            break;
        }
        found = parent;
    }
    return uptr;
}

func (trie *Trie) clear() {
    if (trie.branch == 0) {
        return;
    }

    ListDelInit(&trie.link);
    ListDelInit(&trie.tree);
    ListDelInit(&trie.treentry);
    trie.parent = nil;
    trie.leaf = nil;

    for i := 0; i < 256; i++ {
        trie.child[i] = nil;
    }
    trie.branch = 0;
}

func (trie *Trie) reset() {
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
