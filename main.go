package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	kvsName     = "eagles_scores"
	updateTopic = "eagles.updates"
)

func fetchAndProcessGames(ctx context.Context, nc *nats.Conn, kv nats.KeyValue, teamFilter string) {
	sCtx, span := tracer.Start(ctx, "fetch_and_process_games")
	defer span.End()

	// Add 1 second timeout if no deadline is set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	// Create child span for HTTP request
	reqCtx, reqSpan := tracer.Start(sCtx, "espn_api_request")

	// Create request with context
	// yesterday := time.Now().AddDate(0, 0, -1).Format("20060102")
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?dates=%s", "20250119")

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		reqSpan.RecordError(err)
		reqSpan.End()
		logger.Error("failed to create request",
			zap.Error(err),
		)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		reqSpan.RecordError(err)
		reqSpan.End()
		apiPollFailure.Inc()
		logger.Error("failed to fetch data",
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		reqSpan.RecordError(fmt.Errorf("invalid status code from API: %d", resp.StatusCode))
		reqSpan.End()
		apiPollFailure.Inc()
		logger.Error("invalid status code from API",
			zap.Int("status_code", resp.StatusCode),
		)
		return
	}
	apiPollSuccess.Inc()
	reqSpan.End()

	var espnResp ESPNResponse
	if err := json.NewDecoder(resp.Body).Decode(&espnResp); err != nil {
		apiDecodeFailure.Inc()
		logger.Error("failed to parse JSON", zap.Error(err))
		return
	}
	apiDecodeSuccess.Inc()

	if len(espnResp.Events) == 0 {
		logger.Warn("no games found in response")
		return
	}

	gamesFound := false
	for _, event := range espnResp.Events {
		if len(event.Competitions) == 0 {
			continue
		}
		gamesFound = true

		for _, comp := range event.Competitions {
			hasEagles := false
			var currentCompetitor *Competitor

			for _, competitor := range comp.Competitors {
				team := strings.ToLower(competitor.Team.DisplayName)
				if strings.Contains(team, teamFilter) {
					hasEagles = true
					currentCompetitor = &competitor
					break
				}
			}

			if !hasEagles {
				continue
			}

			_, getScoreSpan := tracer.Start(sCtx, "get_score")

			// Check previous score
			scoreKey := fmt.Sprintf("%s-%s", comp.ID, currentCompetitor.Team.ID)
			prevScore, err := kv.Get(scoreKey)
			if err != nil && err != nats.ErrKeyNotFound {
				getScoreSpan.RecordError(err)
				getScoreSpan.End()
				logger.Error("failed to get previous score",
					zap.Error(err),
					zap.String("game_id", comp.ID))
				continue
			}

			getScoreSpan.End()

			// Convert quarter scores to integers
			quarterScores := make([]int64, len(currentCompetitor.Linescores))
			for i, score := range currentCompetitor.Linescores {
				quarterScores[i] = int64(score.Value)
			}

			// If score changed, publish update
			if err == nats.ErrKeyNotFound || string(prevScore.Value()) != currentCompetitor.Score {

				_, putScoreSpan := tracer.Start(sCtx, "put_score")
				// Store new score
				if _, err := kv.Put(scoreKey, []byte(currentCompetitor.Score)); err != nil {
					putScoreSpan.RecordError(err)
					putScoreSpan.End()
					natsKVPutFailure.Inc()
					logger.Error("failed to store score",
						zap.Error(err),
						zap.String("game_id", comp.ID))
					continue
				}
				putScoreSpan.End()
				natsKVPutSuccess.Inc()

				// If it's not the first time we're seeing this score, publish update
				if err != nats.ErrKeyNotFound {
					scoreChangeTotal.Inc()
					update := ScoreData{
						GameID:        comp.ID,
						Score:         currentCompetitor.Score,
						QuarterScores: quarterScores,
					}
					updateJSON, err := json.Marshal(update)
					if err != nil {
						logger.Error("failed to marshal score update", zap.Error(err))
						continue
					}

					_, pubSpan := tracer.Start(sCtx, "nats_publish")
					if err := nc.Publish(updateTopic, updateJSON); err != nil {
						pubSpan.RecordError(err)
						pubSpan.End()
						logger.Error("failed to publish score update", zap.Error(err))
						continue
					}
					pubSpan.End()

					logger.Info("score update published",
						zap.String("game_id", comp.ID),
						zap.String("old_score", string(prevScore.Value())),
						zap.String("new_score", currentCompetitor.Score))
				}
			}

			logger.Info("game details",
				zap.String("game", event.Name),
				zap.String("date", event.Date),
				zap.String("competition_id", comp.ID),
				zap.String("status", event.Status.Type.Description),
				zap.String("team", currentCompetitor.Team.DisplayName),
				zap.String("score", currentCompetitor.Score),
				zap.Int64s("quarter_scores", quarterScores),
			)
		}
	}

	if !gamesFound {
		logger.Warn("no competitions found in any events")
	}
}

func watchScores(kv nats.KeyValue) {
	watcher, err := kv.WatchAll()
	if err != nil {
		logger.Error("failed to create watcher", zap.Error(err))
		return
	}

	for entry := range watcher.Updates() {
		if entry == nil {
			continue
		}
		logger.Info("score changed",
			zap.String("key", entry.Key()),
			zap.String("value", string(entry.Value())),
			zap.Uint64("revision", entry.Revision()))
	}
}

func main() {
	ctx := context.Background()

	if err := initLogger(); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	shutdownTracer, err := initTracer()
	if err != nil {
		logger.Error("failed to initialize tracer", zap.Error(err))
	}
	defer shutdownTracer(ctx)

	// Start NATS server
	ns, err := setupNatsServer()
	if err != nil {
		logger.Fatal("failed to start NATS server", zap.Error(err))
	}
	defer ns.Shutdown()

	// Connect to NATS
	nc, err := connectToNats()
	if err != nil {
		logger.Fatal("failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	// Setup KV store
	kv, err := setupKVStore(nc)
	if err != nil {
		logger.Fatal("failed to setup KV store", zap.Error(err))
	}

	// Start watching scores in background
	go watchScores(kv)

	teamFilter := "Eagles"
	teamFilter = strings.ToLower(teamFilter)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// First immediate fetch
	fetchAndProcessGames(ctx, nc, kv, teamFilter)

	// Then fetch every 5 seconds
	for range ticker.C {
		fetchAndProcessGames(ctx, nc, kv, teamFilter)
	}

	// Add metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":2112", nil); err != nil {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()
}
