package ipam

import (
	"sort"
	"strings"
)

type IPRangeList struct {
	ranges []*IPRange
}

func NewIPRangeList() *IPRangeList {
	return &IPRangeList{}
}

func NewIPRangeListFrom(x ...string) *IPRangeList {
	ret := &IPRangeList{make([]*IPRange, 0, len(x))}
	for _, s := range x {
		ips := strings.Split(s, "..")
		if len(ips) == 1 {
			ret.Add(NewIP(ips[0]))
		} else {
			n1, found1 := ret.Find(NewIP(ips[0]))
			n2, found2 := ret.Find(NewIP(ips[1]))
			if found1 {
				if found2 {
					if n1 != n2 {
						ret.ranges[n1].SetEnd(ret.ranges[n2].End())
						ret.ranges = append(ret.ranges[:n1+1], ret.ranges[n2+1:]...)
					}
				} else {
					ret.ranges[n1].SetEnd(NewIP(ips[1]))
					ret.ranges = append(ret.ranges[:n1+1], ret.ranges[n2:]...)
				}
			} else {
				if found2 {
					ret.ranges[n2].SetStart(NewIP(ips[0]))
					ret.ranges = append(ret.ranges[:n1], ret.ranges[n2:]...)
				} else {
					if n1 == n2 {
						tmp := make([]*IPRange, ret.Len()+1)
						copy(tmp, ret.ranges[:n1])
						tmp[n1] = NewIPRange(NewIP(ips[0]), NewIP(ips[1]))
						copy(tmp[n1+1:], ret.ranges[n1:])
						ret.ranges = tmp
					} else {
						ret.ranges[n1] = NewIPRange(NewIP(ips[0]), NewIP(ips[1]))
						ret.ranges = append(ret.ranges[:n1+1], ret.ranges[n2+1:]...)
					}
				}
			}
		}
	}
	return ret
}

func (r *IPRangeList) Clone() *IPRangeList {
	ret := &IPRangeList{make([]*IPRange, r.Len())}
	copy(ret.ranges, r.ranges)
	return ret
}

func (r *IPRangeList) Len() int {
	return len(r.ranges)
}

func (r *IPRangeList) At(i int) *IPRange {
	if i < len(r.ranges) {
		return r.ranges[i]
	}
	return nil
}

func (r *IPRangeList) Find(ip IP) (int, bool) {
	return sort.Find(len(r.ranges), func(i int) int {
		if r.At(i).Start().GreaterThan(ip) {
			return -1
		}
		if r.At(i).End().LessThan(ip) {
			return 1
		}
		return 0
	})
}

func (r *IPRangeList) Contains(ip IP) bool {
	_, found := r.Find(ip)
	return found
}

func (r *IPRangeList) Add(ip IP) bool {
	n, ok := r.Find(ip)
	if ok {
		return false
	}

	if (n-1 >= 0 && r.ranges[n-1].Add(ip)) ||
		(n < r.Len() && r.ranges[n].Add(ip)) {
		if n-1 >= 0 && n < r.Len() && r.ranges[n-1].End().Add(1).Equal(r.ranges[n].Start()) {
			r.ranges[n-1].SetEnd(r.ranges[n].End())
			r.ranges = append(r.ranges[:n], r.ranges[n+1:]...)
		}
		return true
	}

	tmp := make([]*IPRange, r.Len()+1)
	copy(tmp, r.ranges[:n])
	tmp[n] = NewIPRange(ip, ip)
	copy(tmp[n+1:], r.ranges[n:])
	r.ranges = tmp

	return true
}

func (r *IPRangeList) Remove(ip IP) bool {
	n, ok := r.Find(ip)
	if !ok {
		return false
	}

	v, _ := r.ranges[n].Remove(ip)
	switch len(v) {
	case 0:
		r.ranges = append(r.ranges[:n], r.ranges[n+1:]...)
	case 1:
		r.ranges[n] = v[0]
	case 2:
		tmp := make([]*IPRange, r.Len()+1)
		copy(tmp, r.ranges[:n])
		copy(tmp[n:], v)
		copy(tmp[n+2:], r.ranges[n+1:])
		r.ranges = tmp
	}

	return true
}

func (r *IPRangeList) Allocate(skipped []IP) IP {
	if r.Len() == 0 {
		return nil
	}

	if len(skipped) == 0 {
		ret := r.ranges[0].Start()
		r.Remove(ret)
		return ret
	}

	tmp := NewIPRangeList()
	for _, ip := range skipped {
		tmp.Add(ip)
	}

	filtered := r.Difference(tmp)
	if filtered.Len() == 0 {
		return nil
	}

	ret := filtered.ranges[0].Start()
	r.Remove(ret)
	return ret
}

func (r *IPRangeList) Equal(x *IPRangeList) bool {
	if r.Len() != x.Len() {
		return false
	}

	for i := 0; i < r.Len(); i++ {
		if !r.At(i).Start().Equal(x.At(i).Start()) || !r.At(i).End().Equal(x.At(i).End()) {
			return false
		}
	}

	return true
}

// Difference returns a new list which contains items which are in `r` but not in `x`
func (r *IPRangeList) Difference(x *IPRangeList) *IPRangeList {
	if r.Len() == 0 {
		return NewIPRangeList()
	}
	if x.Len() == 0 {
		return r.Clone()
	}

	var i, j int
	ret := &IPRangeList{}
	for ; i < r.Len(); i++ {
		start, end := r.At(i).Start(), r.At(i).End()
		for ; j < x.Len(); j++ {
			if x.At(j).End().LessThan(start) {
				continue
			}
			if x.At(j).Start().GreaterThan(end) {
				ret.ranges = append(ret.ranges, NewIPRange(start, end))
				break
			}
			if !x.At(j).End().LessThan(end) {
				if x.At(j).Start().GreaterThan(start) {
					ret.ranges = append(ret.ranges, NewIPRange(start, x.At(j).Start().Sub(1)))
				}
				break
			}
			if start.LessThan(x.At(j).Start()) {
				ret.ranges = append(ret.ranges, NewIPRange(start, x.At(j).Start().Sub(1)))
			}
			start = x.At(j).End().Add(1)
		}
		if j == x.Len() {
			ret.ranges = append(ret.ranges, NewIPRange(start, end))
		}
	}

	return ret
}

func (r *IPRangeList) Merge(x *IPRangeList) *IPRangeList {
	ret := &IPRangeList{make([]*IPRange, 0, r.Len()+x.Len())}

	var i, j int
	for i != r.Len() || j != x.Len() {
		if i == r.Len() {
			ret.ranges = append(ret.ranges, x.ranges[j].Clone())
			j++
			continue
		}
		if j == x.Len() {
			ret.ranges = append(ret.ranges, r.ranges[i].Clone())
			i++
			continue
		}
		if r.ranges[i].Start().LessThan(x.ranges[j].Start()) {
			ret.ranges = append(ret.ranges, r.ranges[i].Clone())
			i++
		} else {
			ret.ranges = append(ret.ranges, x.ranges[j].Clone())
			j++
		}
	}

	return ret
}

func (r *IPRangeList) String() string {
	s := make([]string, 0, r.Len())
	for i := 0; i < r.Len(); i++ {
		s = append(s, r.At(i).String())
	}
	return strings.Join(s, ",")
}
