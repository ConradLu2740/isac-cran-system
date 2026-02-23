package discovery

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

type ServiceInfo struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
}

type ServiceDiscovery struct {
	client    *api.Client
	serviceID string
}

func NewServiceDiscovery(consulAddr string) (*ServiceDiscovery, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ServiceDiscovery{client: client}, nil
}

func (d *ServiceDiscovery) Register(service *ServiceInfo) error {
	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err := d.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	d.serviceID = service.ID
	log.Printf("Service registered: %s (%s:%d)", service.Name, service.Address, service.Port)
	return nil
}

func (d *ServiceDiscovery) Deregister() error {
	if d.serviceID == "" {
		return nil
	}

	err := d.client.Agent().ServiceDeregister(d.serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	log.Printf("Service deregistered: %s", d.serviceID)
	return nil
}

func (d *ServiceDiscovery) Discover(serviceName string) ([]*ServiceInfo, error) {
	services, _, err := d.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}

	var results []*ServiceInfo
	for _, service := range services {
		results = append(results, &ServiceInfo{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
		})
	}

	return results, nil
}

func (d *ServiceDiscovery) Watch(serviceName string, callback func([]*ServiceInfo)) {
	go func() {
		var lastIndex uint64
		for {
			services, meta, err := d.client.Health().Service(serviceName, "", true, &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second,
			})
			if err != nil {
				log.Printf("Watch error: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if meta.LastIndex > lastIndex {
				lastIndex = meta.LastIndex
				var results []*ServiceInfo
				for _, service := range services {
					results = append(results, &ServiceInfo{
						ID:      service.Service.ID,
						Name:    service.Service.Service,
						Address: service.Service.Address,
						Port:    service.Service.Port,
						Tags:    service.Service.Tags,
					})
				}
				callback(results)
			}
		}
	}()
}

func (d *ServiceDiscovery) HealthCheck(ctx context.Context) error {
	_, err := d.client.Agent().Self()
	if err != nil {
		return fmt.Errorf("consul health check failed: %w", err)
	}
	return nil
}
