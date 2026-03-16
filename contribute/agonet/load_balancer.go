package agonet

import (
	"net"
)

type (
	// loadBalancer is an interface which manipulates the event-loop set.
	loadBalancer interface {
		register(*eventloop)
		next(net.Addr) *eventloop
		index(int) *eventloop
		iterate(func(int, *eventloop) bool)
		len() int
	}

	// baseLoadBalancer with base lb.
	baseLoadBalancer struct {
		eventLoops []*eventloop
		size       int
	}

	// roundRobinLoadBalancer with Round-Robin algorithm.
	roundRobinLoadBalancer struct {
		baseLoadBalancer
		nextIndex uint64
	}

	// leastConnectionsLoadBalancer with Least-Connections algorithm.
	leastConnectionsLoadBalancer struct {
		baseLoadBalancer
	}
)

// ==================================== Implementation of base load-balancer ====================================

// register adds a new eventloop into load-balancer.
func (lb *baseLoadBalancer) register(el *eventloop) {
	el.idx = lb.size
	lb.eventLoops = append(lb.eventLoops, el)
	lb.size++
}

// index returns the eligible eventloop by index.
func (lb *baseLoadBalancer) index(i int) *eventloop {
	if i >= lb.size {
		return nil
	}
	return lb.eventLoops[i]
}

// iterate iterates all the eventloops.
func (lb *baseLoadBalancer) iterate(f func(int, *eventloop) bool) {
	for i, el := range lb.eventLoops {
		if !f(i, el) {
			break
		}
	}
}

// len returns the length of event-loop list.
func (lb *baseLoadBalancer) len() int {
	return lb.size
}

// ==================================== Implementation of Round-Robin load-balancer ====================================

// next returns the eligible event-loop based on Round-Robin algorithm.
func (lb *roundRobinLoadBalancer) next(_ net.Addr) (el *eventloop) {
	el = lb.eventLoops[lb.nextIndex%uint64(lb.size)]
	lb.nextIndex++
	return
}

// ================================= Implementation of Least-Connections load-balancer =================================

func (lb *leastConnectionsLoadBalancer) next(_ net.Addr) (el *eventloop) {
	el = lb.eventLoops[0]
	minN := el.countConn()
	for _, v := range lb.eventLoops[1:] {
		if n := v.countConn(); n < minN {
			minN = n
			el = v
		}
	}
	return
}
