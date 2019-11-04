package tightrope

type Pool []*Worker

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool {
	return p[i].Pending < p[j].Pending
}

func (p *Pool) Swap(i, j int) {
	a := *p
	a[i], a[j] = a[j], a[i]
	a[i].Index = i
	a[j].Index = j
}

func (p *Pool) Push(x interface{}) {
	n := len(*p)
	item := x.(*Worker)
	item.Index = n
	*p = append(*p, item)
}
func (p *Pool) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	item.Index = -1
	*p = old[0 : n-1]
	return item
}
