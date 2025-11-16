package producer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"

	danube "github.com/danube-messaging/danube-go"

	"github.com/danube-messaging/loadtest_danube/pkg/config"
	"github.com/danube-messaging/loadtest_danube/pkg/metrics"
	"github.com/danube-messaging/loadtest_danube/pkg/workload"
)

type Pool struct {
	serviceURL string
	cfg        *config.Config
	metrics    *metrics.Collector
	stopOnce   sync.Once
}

func NewPool(serviceURL string, cfg *config.Config, m *metrics.Collector) *Pool {
	return &Pool{serviceURL: serviceURL, cfg: cfg, metrics: m}
}

// Start launches producer workers for all producer groups in the config.
// It returns a function to stop all workers (by canceling the given context).
func (p *Pool) Start(ctx context.Context, wg *sync.WaitGroup) {
	client := danube.NewClient().ServiceURL(p.serviceURL).Build()
	for _, pg := range p.cfg.Producers {
		// find topic config by name
		var topicCfg *config.Topic
		schemaType := "string"
		for i := range p.cfg.Topics {
			if p.cfg.Topics[i].Name == pg.Topic {
				topicCfg = &p.cfg.Topics[i]
				schemaType = topicCfg.SchemaType
				break
			}
		}

		for i := 0; i < pg.Count; i++ {
			wg.Add(1)
			go func(group config.ProducerGroup, workerIdx int, schema string) {
				defer wg.Done()
				p.runWorker(ctx, client, group, workerIdx, schema)
			}(pg, i, schemaType)
		}
	}
}

func (p *Pool) runWorker(ctx context.Context, _ any, pg config.ProducerGroup, idx int, schema string) {
	// init producer
	baseName := pg.Name
	if baseName == "" {
		baseName = "producer"
	}
	prodName := fmt.Sprintf("%s-%d", baseName, idx)

	// Build a client locally to avoid depending on client type names
	client := danube.NewClient().ServiceURL(p.serviceURL).Build()
	builder := client.NewProducer(ctx).
		WithName(prodName).
		WithTopic(pg.Topic)

	// Apply topic-level partitions and dispatch strategy if configured
	// Note: producers operate per topic; partitions are internal to Danube
	// and controlled via WithPartitions on the producer builder.
	// Dispatch strategy: reliable vs non_reliable.
	// We need the topic config; lookup again here.
	var topicCfg *config.Topic
	for i := range p.cfg.Topics {
		if p.cfg.Topics[i].Name == pg.Topic {
			topicCfg = &p.cfg.Topics[i]
			break
		}
	}
	if topicCfg != nil {
		if topicCfg.Partitions > 0 {
			builder = builder.WithPartitions(int32(topicCfg.Partitions))
		}
		// Apply schema: only JSON requires explicit schema configuration.
		switch topicCfg.SchemaType {
		case "json":
			// Name is arbitrary; provide configured JSON schema
			builder = builder.WithSchema("json_schema", danube.SchemaType_JSON, topicCfg.JSONSchema)
		}
		switch topicCfg.DispatchStrategy {
		case "reliable":
			builder = builder.WithDispatchStrategy(danube.NewReliableDispatchStrategy())
		default:
			// non_reliable or omitted -> default behavior (do nothing)
		}
	}

	producer, err := builder.Build()
	if err != nil {
		log.Printf("producer build error: %v", err)
		p.metrics.IncError(1)
		return
	}
	if err := producer.Create(ctx); err != nil {
		log.Printf("producer create error: %v", err)
		p.metrics.IncError(1)
		return
	}

	// rate limiter per worker
	r := pg.RatePerSecond
	if r <= 0 {
		r = 0 // unlimited for MVP when 0
	}
	var limiter *rate.Limiter
	if r > 0 {
		limiter = rate.NewLimiter(rate.Limit(r), r)
	}

	var seq uint64
	pspec := workload.PayloadSpec{SchemaType: schema, MessageSize: pg.MessageSize}

	for {
		if ctx.Err() != nil {
			return
		}
		if limiter != nil {
			if err := limiter.Wait(ctx); err != nil {
				return
			}
		}
		seq++
		payload := workload.GeneratePayload(pspec, seq)
		if _, err := producer.Send(ctx, payload, nil); err != nil {
			p.metrics.IncError(1)
			log.Printf("send error topic=%s worker=%d: %v", pg.Topic, idx, err)
			// small backoff to avoid hot loop on error
			time.Sleep(50 * time.Millisecond)
			continue
		}
		p.metrics.IncSent(1)
	}
}
