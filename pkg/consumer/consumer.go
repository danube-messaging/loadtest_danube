package consumer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	danube "github.com/danube-messaging/danube-go"

	"github.com/danube-messaging/loadtest_danube/pkg/config"
	"github.com/danube-messaging/loadtest_danube/pkg/metrics"
)

type Pool struct {
	serviceURL string
	cfg        *config.Config
	metrics    *metrics.Collector
}

func NewPool(serviceURL string, cfg *config.Config, m *metrics.Collector) *Pool {
	return &Pool{serviceURL: serviceURL, cfg: cfg, metrics: m}
}

// Start launches consumers for all consumer groups.
func (p *Pool) Start(ctx context.Context, wg *sync.WaitGroup) {
	for _, cg := range p.cfg.Consumers {
		for i := 0; i < cg.Count; i++ {
			wg.Add(1)
			go func(group config.ConsumerGroup, workerIdx int) {
				defer wg.Done()
				p.runWorker(ctx, group, workerIdx)
			}(cg, i)
		}
	}
}

func (p *Pool) runWorker(ctx context.Context, cg config.ConsumerGroup, idx int) {
	subType := mapSubType(cg.SubscriptionType)

	client := danube.NewClient().ServiceURL(p.serviceURL).Build()
	baseName := cg.Name
	if baseName == "" {
		baseName = "consumer"
	}
	consName := fmt.Sprintf("%s-%d", baseName, idx)
	builder := client.NewConsumer(ctx).
		WithConsumerName(consName).
		WithTopic(cg.Topic).
		WithSubscription(cg.Subscription).
		WithSubscriptionType(subType)

	cons, err := builder.Build()
	if err != nil {
		log.Printf("consumer build error: %v", err)
		p.metrics.IncError(1)
		return
	}
	// Retry subscribe briefly to handle races where topic is not fully created yet
	{
		const (
			maxAttempts = 15
			backoffMs   = 200
		)
		var subErr error
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if ctx.Err() != nil {
				return
			}
			if subErr = cons.Subscribe(ctx); subErr == nil {
				break
			}
			if attempt == 1 || attempt%5 == 0 {
				log.Printf("consumer subscribe error (attempt %d/%d): %v", attempt, maxAttempts, subErr)
			}
			time.Sleep(backoffMs * time.Millisecond)
		}
		if subErr != nil {
			log.Printf("consumer subscribe failed after retries: %v", subErr)
			p.metrics.IncError(1)
			return
		}
	}

	stream, err := cons.Receive(ctx)
	if err != nil {
		log.Printf("consumer receive error: %v", err)
		p.metrics.IncError(1)
		return
	}

	// Latency no longer depends on schema; use PublishTime only
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-stream:
			if !ok {
				return
			}
			// Use broker/client PublishTime for E2E latency
			nowMs := time.Now().UnixMilli()
			if pub := msg.GetPublishTime(); pub > 0 {
				lat := float64(nowMs) - float64(pub)
				if lat >= 0 {
					p.metrics.RecordLatency(lat)
				}
			}

			// TODO: parse attributes if needed for additional metrics
			_ = msg.GetAttributes()

			p.metrics.IncReceived(1)
			if _, err := cons.Ack(ctx, msg); err != nil {
				log.Printf("ack error topic=%s worker=%d: %v", cg.Topic, idx, err)
				p.metrics.IncError(1)
			}
		}
	}
}

func mapSubType(s string) danube.SubType {
	switch strings.ToLower(s) {
	case "exclusive":
		return danube.Exclusive
	case "shared":
		return danube.Shared
	case "failover":
		return danube.FailOver
	default:
		return danube.Exclusive
	}
}
