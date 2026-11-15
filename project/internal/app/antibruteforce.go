package app

import (
	"time"

	"github.com/TheTeaToper/OTUS/project/internal/limiter"
	"github.com/TheTeaToper/OTUS/project/internal/subnets"
)

type Config struct {
	// N
	LoginLimit int
	// M
	PasswordLimit int
	// K
	IpLimit int
}

type AntibruteforceService struct {
	config           Config
	blackList        *subnets.Subnets
	whiteList        *subnets.Subnets
	loginsManager    *limiter.Manager
	passwordsManager *limiter.Manager
	ipsManager       *limiter.Manager
}

func NewAntibruteforceService(cfg Config, bucketTtl time.Duration) *AntibruteforceService {
	return &AntibruteforceService{
		config:           cfg,
		blackList:        subnets.NewSubnets(),
		whiteList:        subnets.NewSubnets(),
		loginsManager:    limiter.NewManager(bucketTtl),
		passwordsManager: limiter.NewManager(bucketTtl),
		ipsManager:       limiter.NewManager(bucketTtl),
	}
}

func (s *AntibruteforceService) ValidateAuthorization(login, password, ip string) bool {
	// Проверка blackSubnets в приоритете
	if s.blackList.Contains(ip) {
		return false
	}
	if s.whiteList.Contains(ip) {
		return true
	}

	loginBucket := s.loginsManager.GetBucket(login, s.config.LoginLimit)
	passwordBucket := s.passwordsManager.GetBucket(password, s.config.PasswordLimit)
	ipBucket := s.ipsManager.GetBucket(ip, s.config.IpLimit)

	loginAllowed := loginBucket.Allow()
	passwordAllowed := passwordBucket.Allow()
	ipAllowed := ipBucket.Allow()

	return loginAllowed && passwordAllowed && ipAllowed
}

// Сброс бакетов для логина и IP.
func (s *AntibruteforceService) ResetBuckets(login, ip string) {
	s.loginsManager.ResetKey(login)
	s.ipsManager.ResetKey(ip)
}

// Методы-прослойки для управления списками (вызываются из API/CLI).
func (s *AntibruteforceService) AddToWhitelist(cidr string) error { return s.whiteList.Add(cidr) }

func (s *AntibruteforceService) RemoveFromWhitelist(cidr string) error {
	return s.whiteList.Delete(cidr)
}
func (s *AntibruteforceService) AddToBlacklist(cidr string) error { return s.blackList.Add(cidr) }
func (s *AntibruteforceService) RemoveFromBlacklist(cidr string) error {
	return s.blackList.Delete(cidr)
}

func (s *AntibruteforceService) Stop() {
	s.loginsManager.Stop()
	s.passwordsManager.Stop()
	s.ipsManager.Stop()
}
