package base

import (
	"jvmingo/rtda"
	"jvmingo/rtda/heap"
)

func InvokeMethod(invokerFrame *rtda.Frame, method *heap.Method) {
	thread := invokerFrame.Thread()
	newFrame := thread.NewFrame(method)
	thread.PushFrame(newFrame)

	// 传递参数 stack -> localVars
	argSlotCount := int(method.ArgSlotCount())
	if argSlotCount > 0 {
		for i := argSlotCount - 1; i >= 0; i-- {
			slot := invokerFrame.OperandStack().PopSlot()
			newFrame.LocalVars().SetSlot(uint(i), slot)
		}
	}
	// hack!
	//if method.IsNative() {
	//	if method.Name() == "registerNatives" {
	//		thread.PopFrame()
	//	} else {
	//		panic(fmt.Sprintf("native method: %v.%v%v\n",
	//			method.Class().Name(), method.Name(), method.Descriptor()))
	//	}
	//}
}
