package module

import (
	"fmt"
	"testing"
)

func mergeMap(target, src map[string]interface{}) map[string]interface{} {
	for key, value := range src {
		if _, exist := target[key]; !exist {
			target[key] = value
		} else {
			// key existed in both src,target
			switch src[key].(type) {
			case []interface{}:
				sourceSlice := value.([]interface{})
				targetSlice := make([]interface{}, len(sourceSlice))
				copy(targetSlice, target[key].([]interface{}))

				for index, val := range sourceSlice {
					switch val.(type) {
					case string, int, bool:
						targetSlice[index] = val
					case map[string]interface{}:
						targetMap, ok := targetSlice[index].(map[string]interface{})
						if ok {
							targetSlice[index] = mergeMap(targetMap, val.(map[string]interface{}))
						} else {
							targetSlice[index] = mergeMap(map[string]interface{}{}, val.(map[string]interface{}))
						}
					}
				}
				target[key] = targetSlice
			case map[string]interface{}:
				target[key] = mergeMap(target[key].(map[string]interface{}), src[key].(map[string]interface{}))
			default:
				target[key] = src[key]
			}
		}
	}
	return target
}

func TestMergeMap(t *testing.T) {
	ReadMockResponse := map[string]interface{}{
		//"IsPtr":        "is_ptr",
		//"ProxyPattern": "proxy_pattern",
		//"RecordCount":  1,
		//"Remark":       "remark",
		//"ZoneName":     "zone_name",
		//"Name":         "zone_name",
		//"EcsRegions": map[string]interface{}{
		//	"EcsRegion": []interface{}{
		//		map[string]interface{}{
		//			"UserId": "user_id",
		//			"RegionIds": map[string]interface{}{
		//				"RegionId": []interface{}{
		//					"cn-beijing",
		//					"cn-hangzhou",
		//				},
		//			},
		//		},
		//	},
		//},
		"UserInfo": []interface{}{
			map[string]interface{}{
				"UserId":    "user_id",
				"RegionIds": []string{"cn-beijing", "cn-hangzhou", "cn-chengdu"},
			},
		},
		"Status": "sync_status",
	}
	ReadDiff := map[string]interface{}{
		//"IsPtr":        "is_ptr_update",
		//"Status": "sync_status_update",
		//"EcsRegions": map[string]interface{}{
		//	"EcsRegion": []interface{}{
		//		map[string]interface{}{
		//			"UserId": "user_id_update",
		//			"RegionIds": map[string]interface{}{
		//				"RegionId": []interface{}{
		//					"cn-beijing",
		//					"cn-hangzhou",
		//					"cn-tianjin",
		//				},
		//			},
		//		},
		//	},
		//},
		"UserInfo": []map[string]interface{}{
			{
				"UserId":    "user_id",
				"RegionIds": []string{"cn-beijing", "cn-hangzhou", "cn-chengdu111"},
			},
		},
	}
	v := mergeMap(ReadMockResponse, ReadDiff)
	fmt.Println(v)
}
