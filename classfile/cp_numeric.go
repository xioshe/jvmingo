package classfile

import "math"

type ConstantIntegerInfo struct {
	val int32
}

func (integer *ConstantIntegerInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint32()
	integer.val = int32(bytes)
}

func (integer *ConstantIntegerInfo) Value() int32 {
	return integer.val
}

type ConstantFloatInfo struct {
	val float32
}

func (float *ConstantFloatInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint32()
	float.val = math.Float32frombits(bytes)
}

func (float *ConstantFloatInfo) Value() float32 {
	return float.val
}

type ConstantLongInfo struct {
	val int64
}

func (long *ConstantLongInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint64()
	long.val = int64(bytes)
}

func (long *ConstantLongInfo) Value() int64 {
	return long.val
}

type ConstantDoubleInfo struct {
	val float64
}

func (double *ConstantDoubleInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint64()
	double.val = math.Float64frombits(bytes)
}

func (double *ConstantDoubleInfo) Value() float64 {
	return double.val
}
