package otelconfig

import (
	"net/url"
	"strings"
	"sync"
)

const blockAttbs = ",authorization"

type HttpRequest struct {
	Header             map[string]string `json:"header" validate:"required"`
	Method             string            `json:"method" validate:"required"`
	RemoteAddr         string            `json:"remoteAddr" validate:"required"`
	URL                *url.URL          `json:"url" validate:"required"`
	headerAttbsIsLower bool
	mu                 sync.Mutex
}

func (c *HttpRequest) headerAttbsToLower() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.headerAttbsIsLower {
		for k, v := range c.Header {
			delete(c.Header, k)
			lowerK := strings.ToLower(k)
			if !strings.Contains(blockAttbs, ","+lowerK) && !strings.HasSuffix(lowerK, "key") {
				c.Header[lowerK] = v
			}
		}
		c.headerAttbsIsLower = true
	}
}

func (c *HttpRequest) GetHeader(name string) string {
	if !c.headerAttbsIsLower {
		c.headerAttbsToLower()
	}
	return c.Header[strings.ToLower(name)]
}
