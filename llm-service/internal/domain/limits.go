package domain

type LLMLimits struct {
	DailyLimit int
	Used       int
	Reserved   int
}

func NewLLMLimits(daily, used, reserved int) LLMLimits {
	return LLMLimits{DailyLimit: daily, Used: used, Reserved: reserved}
}
