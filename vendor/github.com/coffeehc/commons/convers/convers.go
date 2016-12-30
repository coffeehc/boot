package convers

//以下方法都很危险,小心使用,不然会 crash
import (
	"time"
	"unsafe"
)

//BytesToString []byte转换为 String,危险操作
func BytesToString(v []byte) string {
	return *(*string)(unsafe.Pointer(&v))
}

//StringToBytes String转换为 []byte,危险操作
func StringToBytes(v string) []byte {
	return *(*[]byte)(unsafe.Pointer(&v))
}

//DurationToInt64 Duration转换为Int64,危险操作
func DurationToInt64(v time.Duration) int64 {
	return *(*int64)(unsafe.Pointer(&v))
}

//Int64ToDuration Int64转换为Duration,危险操作
func Int64ToDuration(v int64) time.Duration {
	return *(*time.Duration)(unsafe.Pointer(&v))
}

//ArrayToInts Array转换为Ints,危险操作
func ArrayToInts(v []interface{}) []int {
	return *(*[]int)(unsafe.Pointer(&v))
}

//ArrayToInt32s Array转换为Int32s,危险操作
func ArrayToInt32s(v []interface{}) []int32 {
	return *(*[]int32)(unsafe.Pointer(&v))
}

//ArrayToInt64s Array转换为Int64s,危险操作
func ArrayToInt64s(v []interface{}) []int64 {
	return *(*[]int64)(unsafe.Pointer(&v))
}

//ArrayToStrings Array转换为Strings,危险操作
func ArrayToStrings(v []interface{}) []string {
	return *(*[]string)(unsafe.Pointer(&v))
}
