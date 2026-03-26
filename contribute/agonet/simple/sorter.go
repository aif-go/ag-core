package simple

const (
	HandlerPriorityHighest = 0
	HandlerPriorityHigh    = 1000
	HandlerPriorityNormal  = 2000
	HandlerPriorityLow     = 3000
	HandlerPriorityLowest  = 4000
)

type OrderAware interface {
	GetOrder() int
}

type ByPriority []Handler

func (p ByPriority) Len() int      { return len(p) }
func (p ByPriority) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ByPriority) Less(i, j int) bool {
	oi, oj := HandlerPriorityNormal, HandlerPriorityNormal
	ori, ok := p[i].(OrderAware)
	if ok {
		oi = ori.GetOrder()
	}

	orj, ok := p[j].(OrderAware)
	if ok {
		oj = orj.GetOrder()
	}
	return oi < oj
}
