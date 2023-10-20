package models

type Pagination struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func maskPhoneNumber(phone string) string {
	// 实现你的手机号掩码逻辑
	// 这里的示例只保留前三位和后四位，其他位用*替代
	masked := phone[:3] + "****" + phone[len(phone)-4:]
	return masked
}
