package rtda

/* Local Var Table element, store an integer or a ref */

type Slot struct {
	num int32 // Java int is 4 byte == 32 bit
	ref *Object
}
