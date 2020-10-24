package dagger

import (
	"fmt"
	"strconv"
)

func parseInt(obj interface{}) int {
	switch obj.(type) {
	case string:
		val, _ := strconv.Atoi(obj.(string))
		return val
	case int:
		return obj.(int)
	case int32:
		return int(obj.(int32))
	case int64:
		return int(obj.(int64))
	case float32:
		return int(obj.(float32))
	case float64:
		return int(obj.(float64))
	default:
		return 0
	}
}

func parseString(obj interface{}) string {
	switch obj.(type) {
	case string:
		return obj.(string)
	default:
		return fmt.Sprint(obj)
	}
}

func parseBool(obj interface{}) bool {
	switch obj.(type) {
	case bool:
		return obj.(bool)
	case string:
		val, _ := strconv.ParseBool(obj.(string))
		return val
	default:
		return false
	}
}
