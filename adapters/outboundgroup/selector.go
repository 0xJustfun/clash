package outboundgroup

import (
	"encoding/json"
	"errors"

	"github.com/Dreamacro/clash/adapters/outbound"
	"github.com/Dreamacro/clash/adapters/provider"
	"github.com/Dreamacro/clash/common/singledo"
	C "github.com/Dreamacro/clash/constant"
)

type Selector struct {
	*outbound.Base
	single    *singledo.Single
	selected  string
	providers []provider.ProxyProvider
}

func (s *Selector) Dialer(parent C.ProxyDialer) C.ProxyDialer {
	return newGroupDialer(s, s.selectedProxy().Dialer(parent))
}

func (s *Selector) SupportUDP() bool {
	return s.selectedProxy().SupportUDP()
}

func (s *Selector) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range getProvidersProxies(s.providers) {
		all = append(all, proxy.Name())
	}

	return json.Marshal(map[string]interface{}{
		"type": s.Type().String(),
		"now":  s.Now(),
		"all":  all,
	})
}

func (s *Selector) Now() string {
	return s.selectedProxy().Name()
}

func (s *Selector) Set(name string) error {
	for _, proxy := range getProvidersProxies(s.providers) {
		if proxy.Name() == name {
			s.selected = name
			return nil
		}
	}

	return errors.New("Proxy does not exist")
}

func (s *Selector) selectedProxy() C.Proxy {
	elm, _, _ := s.single.Do(func() (interface{}, error) {
		proxies := getProvidersProxies(s.providers)
		for _, proxy := range proxies {
			if proxy.Name() == s.selected {
				return proxy, nil
			}
		}

		return proxies[0], nil
	})

	return elm.(C.Proxy)
}

func NewSelector(name string, providers []provider.ProxyProvider) *Selector {
	selected := providers[0].Proxies()[0].Name()
	return &Selector{
		Base:      outbound.NewBase(name, "", C.Selector, false),
		single:    singledo.NewSingle(defaultGetProxiesDuration),
		providers: providers,
		selected:  selected,
	}
}
