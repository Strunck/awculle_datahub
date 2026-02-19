package globals

import (
	"time"

	"github.com/awcullen/opcua/ua"
)

func ConvertSymbol(v ua.Variant, typ string) (res any) {

	switch typ {

	case "T_BOOL":
		res = v.(bool)
	case "T_ARRAY__1__2__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__4__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__5__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__6__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__7__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__8__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__10__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__16__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__20__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__24__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__36__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__40__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__46__OF_BOOL":
		res = v.([]bool)
	case "T_ARRAY__1__160__OF_BOOL":
		res = v.([]bool)
	case "T_BYTE":
		res = uint8(v.(byte))
	case "T_ARRAY__1__2__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__3__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__4__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__5__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__6__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__10__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__20__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__23__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__40__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__100__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__20__1__4__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__4__1__40__OF_BYTE":
		res = handleByteArray(v)
	case "T_ARRAY__1__6__1__20__OF_BYTE":
		res = handleByteArray(v)
	case "T_WORD":
		res = v.(uint16)
	case "T_ARRAY__1__2__OF_WORD":
		res = v.([]uint16)
	case "T_ARRAY__1__6__OF_WORD":
		res = v.([]uint16)
	case "T_ARRAY__1__7__OF_WORD":
		res = v.([]uint16)
	case "T_ARRAY__1__20__OF_WORD":
		res = v.([]uint16)
	case "T_ARRAY__1__40__OF_WORD":
		res = v.([]uint16)
	case "T_ARRAY__1__20__1__4__OF_WORD":
		res = v.([][]uint16)
	case "T_ARRAY__1__6__1__20__OF_WORD":
		res = v.([][]uint16)
	case "T_INT":
		res = v.(int16)
	case "T_ARRAY__1__2__OF_INT":
		res = v.([]int16)
	case "T_ARRAY__1__20__OF_INT":
		res = v.([]int16)
	case "T_ARRAY__1__40__OF_INT":
		res = v.([]int16)
	case "T_ARRAY__1__46__OF_INT":
		res = v.([]int16)
	case "T_ARRAY__1__6__1__2__OF_INT":
		res = v.([][]int16)
	case "T_ARRAY__1__6__OF_INT":
		res = v.([]int16)
	case "T_UINT":
		res = v.(uint16)
	case "T_ARRAY__1__40__OF_UINT":
		res = v.([]uint16)
	case "T_DWORD":
		res = v.(uint32)
	case "T_ARRAY__1__20__OF_DWORD":
		res = v.([]uint32)
	case "T_ARRAY__1__6__OF_DWORD":
		res = v.([]uint32)
	case "T_ARRAY__1__40__OF_DWORD":
		res = v.([]uint32)
	case "T_REAL":
		res = v.(float32)
	case "T_ARRAY__1__6__OF_REAL":
		res = v.([]float32)
	case "T_ARRAY__1__24__OF_LREAL":
		res = v.([]float64)
	case "T_DATE_AND_TIME":
		res = v.(time.Time)
	case "T_ARRAY__1__40__OF_DATE_AND_TIME":
		res = v.([]time.Time)
	case "T_RTS_SYSTIMEDATE":
		res = v.(time.Time)
	case "T_DATE":
		res = v.(time.Time)
	case "T_ARRAY__1__4__OF_TIME":
		res = v.([]time.Duration)
	case "T_STRING":
		res = v.(string)
	case "T_STRING_255_":
		res = v.(string)
	case "T_ARRAY__1__6__OF_STRING":
		res = v.([]string)
	case "T_ARRAY__1__21__1__41__OF_STRING_255_":
		res = v.([][]string)
	case "T_ARRAY__1__24__OF_TYPE_EMeterKonfig":
		//res = v.([]EMeterKonfig)
		res = v
	case "T_ARRAY__1__40__OF_TYPE_Rmm":
		res = v
	case "T_ARRAY__1__40__OF_PVMOD":
		res = v.([]int)
	case "T_ARRAY__1__20__OF_TYPE_Warnung":
		res = v
	case "T_ARRAY__1__100__OF_TYPE_Verbraucher":
		res = v
	case "T_ARRAY__1__14__OF_struct_Messwerte":
		res = v
	case "T_ARRAY__1__160__OF_TYPE_Luefterkonfig":
		res = v
	case "T_ARRAY__1__2__OF_Struct_ArrOfPoints":
		res = v
	case "T_ARRAY__1__6__OF_STRUCT_VerdichterVisu":
		res = v
	case "T_ARRAY__1__6__OF_STRUCT_VerflussigerVisu":
		res = v
	}

	return res
}

func handleEMeterKonfig(v ua.Variant) (res any) {
	res = v
	return res
}

func handleByteArray(v ua.Variant) (res []int8) {
	if v == nil {
		return nil
	}
	// cast every byte to int8
	for _, b := range v.([]byte) {
		res = append(res, int8(b))
	}
	return res
}
