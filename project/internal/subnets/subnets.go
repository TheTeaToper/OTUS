package subnets

import (
	"fmt"
	"net"
	"sync"
)

type Subnets struct {
	mutex  sync.RWMutex
	ipnets map[string]*net.IPNet
}

func NewSubnets() *Subnets {
	return &Subnets{
		ipnets: make(map[string]*net.IPNet),
	}
}

func (s *Subnets) Add(cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("Invalid CIDR: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.ipnets[ipNet.String()] = ipNet
	return nil
}

func (s *Subnets) Delete(cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("Invalid CIDR: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.ipnets, ipNet.String())
	return nil
}

func (s *Subnets) Contains(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, ipnet := range s.ipnets {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}

func (s *Subnets) ToString() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]string, 0, len(s.ipnets))
	for _, ipnet := range s.ipnets {
		result = append(result, ipnet.String())
	}
	return result
}
