package domain

type LLMLimits struct {
	DailyLimit int
	Used       int
	Reserved   int
}

func NewLLMLimits(daily, used, reserved int) LLMLimits {
	return LLMLimits{DailyLimit: daily, Used: used, Reserved: reserved}
}

// Remaining возвращает оставшееся количество токенов
func (l LLMLimits) Remaining() int {
	return l.DailyLimit - l.Used - l.Reserved
}
