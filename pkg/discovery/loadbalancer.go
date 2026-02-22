package discovery

import (
	"sync/atomic"
)

type LoadBalancer interface {
	Select(services []*ServiceInfo) *ServiceInfo
}

type RoundRobinLoadBalancer struct {
	counter uint64
}

func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

func (lb *RoundRobinLoadBalancer) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	index := atomic.AddUint64(&lb.counter, 1) - 1
	return services[index%uint64(len(services))]
}

type RandomLoadBalancer struct{}

func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{}
}

func (lb *RandomLoadBalancer) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	return services[0]
}

type WeightedLoadBalancer struct {
	weights map[string]int
}

func NewWeightedLoadBalancer(weights map[string]int) *WeightedLoadBalancer {
	return &WeightedLoadBalancer{weights: weights}
}

func (lb *WeightedLoadBalancer) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	var totalWeight int
	for _, s := range services {
		w := lb.weights[s.ID]
		if w == 0 {
			w = 1
		}
		totalWeight += w
	}

	var sum int
	for _, s := range services {
		w := lb.weights[s.ID]
		if w == 0 {
			w = 1
		}
		sum += w
		if sum >= totalWeight/2 {
			return s
		}
	}

	return services[0]
}
