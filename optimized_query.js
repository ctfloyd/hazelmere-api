// Optimized MongoDB Query to fetch only snapshots with changed OVERALL experience
// This query should be implemented in the repository layer

db.snapshots.aggregate([
  // Stage 1: Match snapshots for the user within time range
  {
    $match: {
      userId: "USER_ID_PLACEHOLDER",
      timestamp: {
        $gte: new Date("START_TIME_PLACEHOLDER"),
        $lte: new Date("END_TIME_PLACEHOLDER")
      }
    }
  },
  
  // Stage 2: Sort by timestamp to ensure proper ordering
  {
    $sort: { timestamp: 1 }
  },
  
  // Stage 3: Add field with OVERALL skill experience value
  {
    $addFields: {
      overallExperience: {
        $let: {
          vars: {
            overallSkill: {
              $arrayElemAt: [
                {
                  $filter: {
                    input: "$skills",
                    cond: { $eq: ["$$this.activityType", "OVERALL"] }
                  }
                },
                0
              ]
            }
          },
          in: "$$overallSkill.experience"
        }
      }
    }
  },
  
  // Stage 4: Add previous document using $setWindowFields (MongoDB 5.0+)
  {
    $setWindowFields: {
      sortBy: { timestamp: 1 },
      output: {
        prevOverallExperience: {
          $shift: {
            output: "$overallExperience",
            by: -1
          }
        }
      }
    }
  },
  
  // Stage 5: Filter to only include documents where experience changed
  {
    $match: {
      $or: [
        { prevOverallExperience: { $exists: false } }, // First document
        { $expr: { $ne: ["$overallExperience", "$prevOverallExperience"] } } // Changed experience
      ]
    }
  },
  
  // Stage 6: Remove temporary fields
  {
    $project: {
      overallExperience: 0,
      prevOverallExperience: 0
    }
  }
]);

// Alternative approach for MongoDB versions < 5.0 (without $setWindowFields):
// This would require a more complex approach using multiple stages or client-side processing
// Since the current implementation already does client-side filtering, 
// the MongoDB 5.0+ version above is the most efficient solution.