package delta

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/user"
	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_logger"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)


const MaxIntervalDuration = 5 * 365 * 24 * time.Hour

var ErrDeltaGeneric = errors.New("an unexpected error occurred while performing delta operation")
var ErrDeltaNotFound = errors.New("delta not found")
var ErrInvalidDeltaRequest = errors.New("invalid delta request")

type DeltaIntervalResponse struct {
	Deltas      []HiscoreDelta
	TotalDeltas int
}

type DeltaService interface {
	CreateDelta(ctx context.Context, delta HiscoreDelta) (HiscoreDelta, error)
	GetLatestDeltaForUser(ctx context.Context, userId string) (HiscoreDelta, error)
	GetDeltasInRange(ctx context.Context, userId string, startTime, endTime time.Time) (DeltaIntervalResponse, error)
	GetDeltaSummary(ctx context.Context, userId string, startTime, endTime time.Time) (api.GetDeltaSummaryResponse, error)
	PrimeCache(ctx context.Context) error
}

type deltaService struct {
	logger         hz_logger.Logger
	repository     DeltaRepository
	userRepository user.UserRepository
	cache          *DeltaCache
}

func NewDeltaService(logger hz_logger.Logger, repository DeltaRepository, cache *DeltaCache, userRepository user.UserRepository) DeltaService {
	return &deltaService{
		logger:         logger,
		repository:     repository,
		userRepository: userRepository,
		cache:          cache,
	}
}

func (ds *deltaService) CreateDelta(ctx context.Context, delta HiscoreDelta) (HiscoreDelta, error) {
	ctx, span := otel.Tracer("hazelmere").Start(ctx, "deltaService.CreateDelta")
	defer span.End()

	if delta.Id == "" {
		delta.Id = uuid.New().String()
	}

	// Only insert if there are any changes
	if len(delta.Skills) == 0 && len(delta.Bosses) == 0 && len(delta.Activities) == 0 {
		ds.logger.DebugArgs(ctx, "No changes in delta, skipping creation for user %s", delta.UserId)
		return HiscoreDelta{}, nil
	}

	data, err := ds.repository.InsertDelta(ctx, delta.ToData())
	if err != nil {
		return HiscoreDelta{}, errors.Join(ErrDeltaGeneric, err)
	}

	// Append to cache (doesn't invalidate existing data)
	ds.cache.AppendDelta(delta.UserId, data)

	ds.logger.DebugArgs(ctx, "Created delta %s for user %s with %d skill changes, %d boss changes, %d activity changes",
		delta.Id, delta.UserId, len(delta.Skills), len(delta.Bosses), len(delta.Activities))

	return HiscoreDelta{}.FromData(data), nil
}

func (ds *deltaService) GetLatestDeltaForUser(ctx context.Context, userId string) (HiscoreDelta, error) {
	ctx, span := otel.Tracer("hazelmere").Start(ctx, "deltaService.GetLatestDeltaForUser")
	defer span.End()

	// Check cache first - returns domain type directly
	if delta, found := ds.cache.GetLatestDelta(userId); found {
		ds.logger.DebugArgs(ctx, "Cache hit for latest delta for user %s", userId)
		return delta, nil
	}

	// Fall back to repository - returns data type, needs conversion
	ds.logger.DebugArgs(ctx, "Cache miss for latest delta for user %s, falling back to repository", userId)
	data, err := ds.repository.GetLatestDeltaForUser(ctx, userId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return HiscoreDelta{}, ErrDeltaNotFound
		}
		return HiscoreDelta{}, errors.Join(ErrDeltaGeneric, err)
	}
	return HiscoreDelta{}.FromData(data), nil
}

func (ds *deltaService) GetDeltasInRange(ctx context.Context, userId string, startTime, endTime time.Time) (DeltaIntervalResponse, error) {
	ctx, span := otel.Tracer("hazelmere").Start(ctx, "deltaService.GetDeltasInRange")
	defer span.End()

	startTime, endTime, err := validateDeltaInterval(startTime, endTime)
	if err != nil {
		return DeltaIntervalResponse{}, err
	}

	// Check cache first - returns domain types directly (pre-aggregated by day)
	if deltas, found := ds.cache.GetDeltasInRange(userId, startTime, endTime); found {
		ds.logger.DebugArgs(ctx, "Cache hit for deltas in range for user %s", userId)
		return DeltaIntervalResponse{
			Deltas:      deltas,
			TotalDeltas: len(deltas),
		}, nil
	}

	// Fall back to repository - returns data types, needs conversion
	ds.logger.DebugArgs(ctx, "Cache miss for deltas in range for user %s, falling back to repository", userId)
	data, err := ds.repository.GetDeltasInRange(ctx, userId, startTime, endTime)
	if err != nil {
		return DeltaIntervalResponse{}, errors.Join(ErrDeltaGeneric, err)
	}

	return DeltaIntervalResponse{
		Deltas:      HiscoreDelta{}.ManyFromData(data),
		TotalDeltas: len(data),
	}, nil
}

func (ds *deltaService) GetDeltaSummary(ctx context.Context, userId string, startTime, endTime time.Time) (api.GetDeltaSummaryResponse, error) {
	ctx, span := otel.Tracer("hazelmere").Start(ctx, "deltaService.GetDeltaSummary")
	defer span.End()

	startTime, endTime, err := validateDeltaInterval(startTime, endTime)
	if err != nil {
		return api.GetDeltaSummaryResponse{}, err
	}

	// Check cache first, fall back to repository
	var deltas []HiscoreDelta
	if cachedDeltas, found := ds.cache.GetDeltasInRange(userId, startTime, endTime); found {
		ds.logger.DebugArgs(ctx, "Cache hit for delta summary for user %s", userId)
		deltas = cachedDeltas
	} else {
		ds.logger.DebugArgs(ctx, "Cache miss for delta summary for user %s, falling back to repository", userId)
		repoData, err := ds.repository.GetDeltasInRange(ctx, userId, startTime, endTime)
		if err != nil {
			return api.GetDeltaSummaryResponse{}, errors.Join(ErrDeltaGeneric, err)
		}
		deltas = HiscoreDelta{}.ManyFromData(repoData)
	}

	// Aggregate the deltas using activity type string as key
	skillSummaries := make(map[string]*api.SkillDeltaSummary)
	bossSummaries := make(map[string]*api.BossDeltaSummary)
	activitySummaries := make(map[string]*api.ActivityDeltaSummary)
	totalXpGain := 0

	for _, delta := range deltas {
		for _, skill := range delta.Skills {
			at := string(skill.ActivityType)
			if _, exists := skillSummaries[at]; !exists {
				skillSummaries[at] = &api.SkillDeltaSummary{
					ActivityType: skill.ActivityType.ToAPI(),
					Name:         skill.Name,
				}
			}
			skillSummaries[at].TotalExperienceGain += skill.ExperienceGain
			skillSummaries[at].TotalLevelGain += skill.LevelGain
			// Sum up experience gains for total
			if skill.ExperienceGain > 0 {
				totalXpGain += skill.ExperienceGain
			}
		}

		for _, boss := range delta.Bosses {
			at := string(boss.ActivityType)
			if _, exists := bossSummaries[at]; !exists {
				bossSummaries[at] = &api.BossDeltaSummary{
					ActivityType: boss.ActivityType.ToAPI(),
					Name:         boss.Name,
				}
			}
			bossSummaries[at].TotalKillCountGain += boss.KillCountGain
		}

		for _, activity := range delta.Activities {
			at := string(activity.ActivityType)
			if _, exists := activitySummaries[at]; !exists {
				activitySummaries[at] = &api.ActivityDeltaSummary{
					ActivityType: activity.ActivityType.ToAPI(),
					Name:         activity.Name,
				}
			}
			activitySummaries[at].TotalScoreGain += activity.ScoreGain
		}
	}

	// Convert maps to slices
	skills := make([]api.SkillDeltaSummary, 0, len(skillSummaries))
	for _, s := range skillSummaries {
		skills = append(skills, *s)
	}

	bosses := make([]api.BossDeltaSummary, 0, len(bossSummaries))
	for _, b := range bossSummaries {
		bosses = append(bosses, *b)
	}

	activities := make([]api.ActivityDeltaSummary, 0, len(activitySummaries))
	for _, a := range activitySummaries {
		activities = append(activities, *a)
	}

	return api.GetDeltaSummaryResponse{
		UserId:              userId,
		StartTime:           startTime,
		EndTime:             endTime,
		TotalExperienceGain: totalXpGain,
		Skills:              skills,
		Bosses:              bosses,
		Activities:          activities,
		DeltaCount:          len(deltas),
	}, nil
}

func (ds *deltaService) PrimeCache(ctx context.Context) error {
	ctx, span := otel.Tracer("hazelmere").Start(ctx, "deltaService.PrimeCache")
	defer span.End()

	users, err := ds.userRepository.GetUsersWithTrackingEnabled(ctx)
	if err != nil {
		return err
	}

	ds.logger.InfoArgs(ctx, "Priming delta cache for %d users with tracking enabled", len(users))

	var wg sync.WaitGroup
	for _, u := range users {
		wg.Add(1)
		go func(userId string) {
			defer wg.Done()
			ds.primeUserCache(ctx, userId)
		}(u.Id)
	}
	wg.Wait()

	ds.logger.InfoArgs(ctx, "Delta cache priming complete. Total users cached: %d", ds.cache.GetTotalCachedUsers())
	return nil
}

func (ds *deltaService) primeUserCache(ctx context.Context, userId string) {
	deltas, err := ds.repository.GetAllDeltasForUser(ctx, userId)
	if err != nil {
		ds.logger.WarnArgs(ctx, "Failed to load deltas for user %s: %v", userId, err)
		return
	}

	if len(deltas) == 0 {
		ds.logger.DebugArgs(ctx, "No deltas found for user %s, skipping cache prime", userId)
		return
	}

	ds.cache.SetUserDeltas(userId, deltas)
	ds.logger.DebugArgs(ctx, "Primed cache for user %s with %d deltas", userId, len(deltas))
}

func validateDeltaInterval(startTime, endTime time.Time) (time.Time, time.Time, error) {
	if startTime.Equal(endTime) {
		return time.Time{}, time.Time{}, errors.Join(ErrInvalidDeltaRequest, errors.New("start time must not equal end time"))
	}

	if endTime.After(time.Now()) {
		endTime = time.Now()
	}

	if endTime.Before(startTime) {
		tmp := endTime
		endTime = startTime
		startTime = tmp
	}

	if endTime.Sub(startTime) > MaxIntervalDuration {
		return time.Time{}, time.Time{}, errors.Join(ErrInvalidDeltaRequest, errors.New("maximum time interval exceeded"))
	}

	return startTime, endTime, nil
}
