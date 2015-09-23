package gopentsdb

import (
	"errors"
	"log"
	"sync"
	"time"
)

// container for points
type container struct {
	sync.Mutex
	points []Point
}

// push a new point in container
// if MaxPoint is reached, points are pushed
func (c *container) add(point Point) {
	c.Lock()
	c.points = append(c.points, point)
	c.Unlock()
}

// TemporizedClientConfig is the configuration structure for a temporized pusher
type TemporizedClientConfig struct {
	// wait Period seconds between two push
	Period uint
	// if MaxPoints is reached, points are pushed
	// default=0, no limits.
	MaxPoints uint
	// Client configuration
	CConfig *ClientConfig
}

// TemporizedClient is an openstdb client which collect data points and push
// them to openstdb server each period
type TemporizedClient struct {
	client    *Client
	container *container
	period    uint
	maxPoints uint
	timer     *time.Timer
}

// NewTemporizedClient returns a new TemporizedClient
func NewTemporizedClient(config TemporizedClientConfig) (*TemporizedClient, error) {
	var err error
	tClient := &TemporizedClient{
		container: &container{
			points: []Point{},
		},
		maxPoints: config.MaxPoints,
	}
	if config.Period == 0 {
		return nil, errors.New("period can not be 0")
	}
	tClient.period = config.Period

	if config.CConfig == nil {
		return nil, errors.New("clientConfig is nil")
	}
	tClient.client, err = NewClient(*config.CConfig)
	if err != nil {
		return nil, err
	}
	tClient.push()
	return tClient, nil
}

// Add add a new point to current container
func (c *TemporizedClient) Add(point Point) error {
	c.container.add(point)
	if c.maxPoints > 0 && uint(len(c.container.points)) >= c.maxPoints {
		c.push()
	}
	return nil
}

// push current container points sto openstdb server
func (c *TemporizedClient) push() {
	log.Println("On push")
	if c.timer != nil {
		c.timer.Stop()
	}
	c.container.Lock()
	points := c.container.points
	c.container.points = []Point{}
	c.container.Unlock()
	log.Println("Nombre de points", len(points))
	if len(points) > 0 {
		if err := c.client.Push(points); err != nil {
			log.Println(err)
		}
	}
	c.timer = time.AfterFunc(time.Duration(c.period)*time.Second, func() {
		c.push()
	})
}
