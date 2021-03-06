package heap

import (
	"fmt"
	"jvmingo/classfile"
	"jvmingo/classpath"
)

type ClassLoader struct {
	cp          *classpath.Classpath
	classMap    map[string]*Class // Class cache
	verboseFlag bool
}

func NewClassLoader(cp *classpath.Classpath, verboseFlag bool) *ClassLoader {
	loader := &ClassLoader{
		cp:          cp,
		classMap:    make(map[string]*Class),
		verboseFlag: verboseFlag,
	}
	loader.loadBasicClasses()
	loader.loadPrimitiveClasses()
	return loader
}

// Find class file and parse into Class, then save to method area

func (cl *ClassLoader) LoadClass(name string) *Class {
	if class, ok := cl.classMap[name]; ok {
		return class
	}
	var class *Class
	if name[0] == '[' {
		class = cl.loadArrayClass(name)
	} else {
		class = cl.loadNonArray(name)
	}
	if jlClassClass, ok := cl.classMap["java/lang/Class"]; ok {
		class.jClass = jlClassClass.NewObject()
		class.jClass.extra = class
	}

	return class
}

// 类加载器加载类信息入口

func (cl *ClassLoader) loadNonArray(name string) *Class {
	data, entry := cl.readClass(name)
	class := cl.defineClass(data)
	link(class)
	if cl.verboseFlag {
		fmt.Printf("[Loaded %s from %s]\n", name, entry)
	}
	return class
}

// Return Entry for print class loading message

func (cl *ClassLoader) readClass(name string) ([]byte, classpath.Entry) {
	data, entry, err := cl.cp.ReadClass(name)
	if err != nil {
		panic("java.lang.ClassNotFoundException: " + name)
	}
	return data, entry
}

func (cl *ClassLoader) defineClass(data []byte) *Class {
	class := parseClass(data)
	class.loader = cl
	resolveSuper(class)
	resolveInterfaces(class)
	cl.classMap[class.name] = class
	return class
}

func (cl *ClassLoader) loadArrayClass(name string) *Class {
	class := &Class{
		accessFlags: ACC_PUBLIC,
		name:        name,
		loader:      cl,
		superClass:  cl.LoadClass("java/lang/Object"),
		interfaces: []*Class{
			cl.LoadClass("java/lang/Cloneable"),
			cl.LoadClass("java/io/Serializable"),
		},
		initStarted: true,
	}
	cl.classMap[name] = class
	return class
}

func (cl *ClassLoader) loadBasicClasses() {
	jlClassClass := cl.LoadClass("java/lang/Class")
	for _, class := range cl.classMap {
		if class.jClass == nil {
			class.jClass = jlClassClass.NewObject()
			class.jClass.extra = class
		}
	}
}

func (cl *ClassLoader) loadPrimitiveClasses() {
	for primitiveType := range primitiveTypes {
		cl.loadPrimitiveClass(primitiveType)
	}
}

func (cl *ClassLoader) loadPrimitiveClass(className string) {
	class := &Class{
		accessFlags: ACC_PUBLIC,
		name:        className,
		loader:      cl,
		initStarted: true,
	}
	class.jClass = cl.classMap["java/lang/Class"].NewObject()
	class.jClass.extra = class
	cl.classMap[className] = class
}

func parseClass(data []byte) *Class {
	cf, err := classfile.Parse(data)
	if err != nil {
		//panic("java.lang.ClassFormatError")
		panic(err)
	}
	return newClass(cf)
}

// Resolve superClassName to Class object

func resolveSuper(class *Class) {
	if class.name != "java/lang/Object" {
		// Super Class use same classloader
		class.superClass = class.loader.LoadClass(class.superClassName)
	}
}

// Resolve interface names to Class array

func resolveInterfaces(class *Class) {
	interfaceCount := len(class.interfaceNames)
	if interfaceCount > 0 {
		class.interfaces = make([]*Class, interfaceCount)
		for i, interfaceName := range class.interfaceNames {
			class.interfaces[i] = class.loader.LoadClass(interfaceName)
		}
	}
}

func link(class *Class) {
	verify(class)
	prepare(class)
}

func verify(class *Class) {
	// Verify class for safe
}

// init space for class vars
func prepare(class *Class) {
	calcInstanceFieldSlotIds(class)
	calcStaticFieldSlotIds(class)
	allocAndInitStaticVars(class)
}

// 计算示例字段个数，包括父类继承
func calcInstanceFieldSlotIds(class *Class) {
	slotId := uint(0)
	if class.superClass != nil {
		slotId = class.superClass.instanceSlotCount
	}
	for _, field := range class.fields {
		if !field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.instanceSlotCount = slotId
}

func calcStaticFieldSlotIds(class *Class) {
	slotId := uint(0)
	for _, field := range class.fields {
		if field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.staticSlotCount = slotId
}

// 变量初始值为默认值，常量初始值需要指定
func allocAndInitStaticVars(class *Class) {
	class.staticVars = newSlots(class.staticSlotCount)

	for _, field := range class.fields {
		if field.IsStatic() && field.IsFinal() {
			initStaticFinalVar(class, field)
		}
	}
}

func initStaticFinalVar(class *Class, field *Field) {
	vars := class.staticVars
	cp := class.constantPool
	cpIndex := field.ConstValueIndex()
	slotId := field.SlotId()
	if cpIndex > 0 {
		switch field.Descriptor() {
		case "Z", "B", "C", "S", "I":
			val := cp.GetConstant(cpIndex).(int32)
			vars.SetInt(slotId, val)
		case "J":
			val := cp.GetConstant(cpIndex).(int64)
			vars.SetLong(slotId, val)
		case "F":
			val := cp.GetConstant(cpIndex).(float32)
			vars.SetFloat(slotId, val)
		case "D":
			val := cp.GetConstant(cpIndex).(float64)
			vars.SetDouble(slotId, val)
		case "Ljava/lang/String;":
			goStr := cp.GetConstant(cpIndex).(string)
			jStr := JString(class.Loader(), goStr)
			vars.SetRef(slotId, jStr)
		}
	}
}
