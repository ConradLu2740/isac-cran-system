package scheduling

import (
	"sort"
	"sync"
	"time"

	"isac-cran-system/pkg/logger"

	"go.uber.org/zap"
)

type UserPriority int

const (
	PriorityLow      UserPriority = 1
	PriorityNormal   UserPriority = 2
	PriorityHigh     UserPriority = 3
	PriorityCritical UserPriority = 4
)

type User struct {
	ID          int
	Priority    UserPriority
	DataRate    float64
	BufferSize  int
	ChannelGain float64
	LastServed  time.Time
	WaitTime    time.Duration
}

type Resource struct {
	ID          int
	Bandwidth   float64
	Power       float64
	Allocated   bool
	AllocatedTo int
}

type Scheduler struct {
	users     map[int]*User
	resources []*Resource
	mu        sync.RWMutex
	algorithm SchedulingAlgorithm
}

type SchedulingAlgorithm string

const (
	AlgorithmRoundRobin       SchedulingAlgorithm = "round_robin"
	AlgorithmPriority         SchedulingAlgorithm = "priority"
	AlgorithmProportionalFair SchedulingAlgorithm = "proportional_fair"
)

func NewScheduler(algorithm SchedulingAlgorithm, numResources int, bandwidth, power float64) *Scheduler {
	resources := make([]*Resource, numResources)
	for i := 0; i < numResources; i++ {
		resources[i] = &Resource{
			ID:        i,
			Bandwidth: bandwidth / float64(numResources),
			Power:     power / float64(numResources),
		}
	}

	return &Scheduler{
		users:     make(map[int]*User),
		resources: resources,
		algorithm: algorithm,
	}
}

func (s *Scheduler) AddUser(user *User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user.LastServed = time.Now()
	s.users[user.ID] = user
	logger.Info("User added to scheduler", zap.Int("user_id", user.ID))
}

func (s *Scheduler) RemoveUser(userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, userID)
	logger.Info("User removed from scheduler", zap.Int("user_id", userID))
}

func (s *Scheduler) Schedule() map[int]int {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.resources {
		r.Allocated = false
		r.AllocatedTo = -1
	}

	allocation := make(map[int]int)

	switch s.algorithm {
	case AlgorithmRoundRobin:
		allocation = s.roundRobin()
	case AlgorithmPriority:
		allocation = s.priorityBased()
	case AlgorithmProportionalFair:
		allocation = s.proportionalFair()
	default:
		allocation = s.roundRobin()
	}

	return allocation
}

func (s *Scheduler) roundRobin() map[int]int {
	allocation := make(map[int]int)

	userIDs := make([]int, 0, len(s.users))
	for id := range s.users {
		userIDs = append(userIDs, id)
	}
	sort.Ints(userIDs)

	resourceIdx := 0
	for _, userID := range userIDs {
		if resourceIdx >= len(s.resources) {
			break
		}

		user := s.users[userID]
		for resourceIdx < len(s.resources) && s.resources[resourceIdx].Allocated {
			resourceIdx++
		}

		if resourceIdx < len(s.resources) {
			s.resources[resourceIdx].Allocated = true
			s.resources[resourceIdx].AllocatedTo = userID
			allocation[userID] = resourceIdx
			user.LastServed = time.Now()
			resourceIdx++
		}
	}

	return allocation
}

func (s *Scheduler) priorityBased() map[int]int {
	allocation := make(map[int]int)

	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	sort.Slice(users, func(i, j int) bool {
		if users[i].Priority != users[j].Priority {
			return users[i].Priority > users[j].Priority
		}
		return users[i].WaitTime > users[j].WaitTime
	})

	for _, user := range users {
		for _, resource := range s.resources {
			if !resource.Allocated {
				resource.Allocated = true
				resource.AllocatedTo = user.ID
				allocation[user.ID] = resource.ID
				user.LastServed = time.Now()
				user.WaitTime = 0
				break
			}
		}
	}

	return allocation
}

func (s *Scheduler) proportionalFair() map[int]int {
	allocation := make(map[int]int)

	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	sort.Slice(users, func(i, j int) bool {
		metricI := users[i].ChannelGain / (users[i].DataRate + 1e-6)
		metricJ := users[j].ChannelGain / (users[j].DataRate + 1e-6)
		return metricI > metricJ
	})

	for _, user := range users {
		bestResource := -1
		bestMetric := 0.0

		for _, resource := range s.resources {
			if !resource.Allocated {
				metric := user.ChannelGain * resource.Bandwidth
				if metric > bestMetric {
					bestMetric = metric
					bestResource = resource.ID
				}
			}
		}

		if bestResource >= 0 {
			s.resources[bestResource].Allocated = true
			s.resources[bestResource].AllocatedTo = user.ID
			allocation[user.ID] = bestResource
			user.DataRate += bestMetric
			user.LastServed = time.Now()
		}
	}

	return allocation
}

func (s *Scheduler) GetResourceStatus() []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make([]*Resource, len(s.resources))
	for i, r := range s.resources {
		status[i] = &Resource{
			ID:          r.ID,
			Bandwidth:   r.Bandwidth,
			Power:       r.Power,
			Allocated:   r.Allocated,
			AllocatedTo: r.AllocatedTo,
		}
	}
	return status
}

func (s *Scheduler) GetUserStatus() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, &User{
			ID:          u.ID,
			Priority:    u.Priority,
			DataRate:    u.DataRate,
			BufferSize:  u.BufferSize,
			ChannelGain: u.ChannelGain,
			LastServed:  u.LastServed,
			WaitTime:    u.WaitTime,
		})
	}
	return users
}
