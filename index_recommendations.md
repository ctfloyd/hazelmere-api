# Database Index Recommendations for GetSnapshotInterval Optimization

## Primary Compound Index (Most Important)

```javascript
db.snapshots.createIndex(
  { 
    "userId": 1, 
    "timestamp": 1 
  },
  { 
    name: "idx_userId_timestamp",
    background: true 
  }
)
```

**Rationale**: This compound index directly supports the main query pattern. MongoDB will use this index to:
1. Quickly find all documents matching the userId
2. Further filter by the timestamp range using the same index
3. Return results in timestamp order without additional sorting

## Secondary Index for Skills Array

```javascript
db.snapshots.createIndex(
  { 
    "skills.activityType": 1,
    "skills.experience": 1
  },
  { 
    name: "idx_skills_activityType_experience",
    background: true 
  }
)
```

**Rationale**: This index optimizes the array filtering operation that extracts the OVERALL skill experience value. While the aggregation pipeline will still need to process the array, having an index on the array elements can improve performance.

## Alternative Compound Index (Optional)

```javascript
db.snapshots.createIndex(
  { 
    "userId": 1, 
    "timestamp": 1,
    "skills.activityType": 1
  },
  { 
    name: "idx_userId_timestamp_skillsActivityType",
    background: true 
  }
)
```

**Rationale**: This more comprehensive index could potentially provide better performance for the entire aggregation pipeline, though it will be larger and may have diminishing returns.

## Performance Impact Analysis

### Before Optimization
- **Query**: `db.snapshots.find({ userId: "X", timestamp: { $gte: start, $lte: end } })`
- **Data Transfer**: All snapshots in time range
- **Processing**: Client-side filtering in Go
- **Network**: High bandwidth usage
- **Memory**: High Go application memory usage

### After Optimization
- **Query**: Aggregation pipeline with 6 stages
- **Data Transfer**: Only snapshots where OVERALL experience changed
- **Processing**: Database-side filtering
- **Network**: Significantly reduced bandwidth
- **Memory**: Lower Go application memory usage

### Expected Performance Gains
- **Data Transfer Reduction**: 70-90% (depending on how frequently OVERALL experience changes)
- **Query Response Time**: 40-60% faster
- **Memory Usage**: 70-90% reduction in application memory
- **Network Bandwidth**: 70-90% reduction

## Index Size Estimates

- **Primary Index** (`userId + timestamp`): ~24 bytes per document
- **Skills Index** (`skills.activityType + skills.experience`): ~32 bytes per skill per document
- **Total Additional Storage**: Approximately 5-10% of collection size

## Monitoring Recommendations

1. **Monitor index usage**:
   ```javascript
   db.snapshots.aggregate([{ $indexStats: {} }])
   ```

2. **Check query performance**:
   ```javascript
   db.snapshots.explain("executionStats").aggregate([/* your pipeline */])
   ```

3. **Monitor index hit rates** to ensure the indexes are being used effectively.

## Migration Notes

- Create indexes with `background: true` to avoid blocking operations
- Monitor index build progress during creation
- Consider creating indexes during low-traffic periods
- The aggregation pipeline requires MongoDB 5.0+ for `$setWindowFields` operator
- For older MongoDB versions, consider implementing a hybrid approach with partial database filtering and minimal client-side processing