package hiscore

import (
	"bytes"
	"encoding/binary"

	"github.com/ctfloyd/hazelmere-api/src/internal/core/delta"
	"github.com/ctfloyd/hazelmere-api/src/internal/core/snapshot"
)

// BinaryContentType is the content type for the compact binary format.
// Client should send: Accept: application/x-hazelmere-binary
const BinaryContentType = "application/x-hazelmere-binary"

// Binary format version
const binaryVersion uint8 = 1

/*
EncodeDeltaSummaryBinary encodes a DeltaSummaryResponse to a compact binary format.

# Content Negotiation

Request:
  - POST /v1/summary/delta
  - Header: Accept: application/x-hazelmere-binary
  - Body: JSON { "userId": "...", "startTime": "...", "endTime": "..." }

Response:
  - Content-Type: application/x-hazelmere-binary
  - Body: Binary data as specified below

# Binary Format Specification (v1)

All integers are big-endian. Activity types are encoded as single-byte indices
that map to the AllActivityTypes array (indices 0-113).

	Header (2 bytes):
	  [0] version: uint8 (currently 1)
	  [1] flags: uint8 (reserved, always 0)

	Snapshot (baseline values):
	  [2-9] timestamp: int64 (unix milliseconds)
	  [10] skillCount: uint8
	  skills[skillCount]:
	    [0] activityTypeIndex: uint8
	    [1-4] experience: int32
	    [5-6] level: int16
	  (7 bytes per skill)
	  [next] bossCount: uint8
	  bosses[bossCount]:
	    [0] activityTypeIndex: uint8
	    [1-4] killCount: int32
	  (5 bytes per boss)
	  [next] activityCount: uint8
	  activities[activityCount]:
	    [0] activityTypeIndex: uint8
	    [1-4] score: int32
	  (5 bytes per activity)

	Deltas:
	  [next 2 bytes] deltaCount: uint16
	  deltas[deltaCount]:
	    [0-7] timestamp: int64 (unix milliseconds)
	    [8] skillDeltaCount: uint8
	    skillDeltas[skillDeltaCount]:
	      [0] activityTypeIndex: uint8
	      [1-4] experienceGain: int32
	      [5-6] levelGain: int16
	    (7 bytes per skill delta)
	    [next] bossDeltaCount: uint8
	    bossDeltas[bossDeltaCount]:
	      [0] activityTypeIndex: uint8
	      [1-4] killCountGain: int32
	    (5 bytes per boss delta)
	    [next] activityDeltaCount: uint8
	    activityDeltas[activityDeltaCount]:
	      [0] activityTypeIndex: uint8
	      [1-4] scoreGain: int32
	    (5 bytes per activity delta)

# Activity Type Index Mapping (0-113)

Skills (indices 0-25):
  0=UNKNOWN, 1=OVERALL, 2=ATTACK, 3=DEFENCE, 4=STRENGTH, 5=HITPOINTS,
  6=RANGED, 7=PRAYER, 8=MAGIC, 9=COOKING, 10=WOODCUTTING, 11=FLETCHING,
  12=FISHING, 13=FIREMAKING, 14=CRAFTING, 15=SMITHING, 16=MINING,
  17=HERBLORE, 18=AGILITY, 19=THIEVING, 20=SLAYER, 21=FARMING,
  22=RUNECRAFT, 23=HUNTER, 24=CONSTRUCTION, 25=SAILING

Activities (indices 26-45):
  26=LEAGUE_POINTS, 27=DEADMAN_POINTS, 28=BOUNTY_HUNTER__HUNTER,
  29=BOUNTY_HUNTER__ROGUE, 30=BOUNTY_HUNTER_LEGACY__HUNTER,
  31=BOUNTY_HUNTER_LEGACY__ROGUE, 32=CLUE_SCROLLS_ALL,
  33=CLUE_SCROLLS_BEGINNER, 34=CLUE_SCROLLS_EASY, 35=CLUE_SCROLLS_MEDIUM,
  36=CLUE_SCROLLS_HARD, 37=CLUE_SCROLLS_ELITE, 38=CLUE_SCROLLS_MASTER,
  39=GRID_POINTS, 40=LMS__RANK, 41=PVP_ARENA__RANK, 42=SOUL_WARS_ZEAL,
  43=RIFTS_CLOSED, 44=COLOSSEUM_GLORY, 45=COLLECTIONS_LOGGED

Bosses (indices 46-113):
  46=ABYSSAL_SIRE, 47=ALCHEMICAL_HYDRA, 48=AMOXLIATL, 49=ARAXXOR,
  50=ARTIO, 51=BARROWS_CHESTS, 52=BRYOPHYTA, 53=CALLISTO, 54=CALVARION,
  55=CERBERUS, 56=CHAMBERS_OF_XERIC, 57=CHAMBERS_OF_XERIC_CHALLENGE_MODE,
  58=CHAOS_ELEMENTAL, 59=CHAOS_FANATIC, 60=COMMANDER_ZILYANA,
  61=CORPOREAL_BEAST, 62=CRAZY_ARCHAEOLOGIST, 63=DAGANNOTH_PRIME,
  64=DAGANNOTH_REX, 65=DAGANNOTH_SUPREME, 66=DERANGED_ARCHAEOLOGIST,
  67=DOOM_OF_MOKHAIOTL, 68=DUKE_SUCELLUS, 69=GENERAL_GRAARDOR, 70=GIANT_MOLE,
  71=GROTESQUE_GUARDIANS, 72=HESPORI, 73=KALPHITE_QUEEN, 74=KING_BLACK_DRAGON,
  75=KRAKEN, 76=KREEARRA, 77=KRIL_TSUTSAROTH, 78=LUNAR_CHESTS, 79=MIMIC,
  80=NEX, 81=NIGHTMARE, 82=PHOSANIS_NIGHTMARE, 83=OBOR, 84=PHANTOM_MUSPAH,
  85=SARACHNIS, 86=SCORPIA, 87=SCURRIUS, 88=SHELLBANE_GRYPHON, 89=SKOTIZO,
  90=SOL_HEREDIT, 91=SPINDEL, 92=TEMPOROSS, 93=THE_GAUNTLET,
  94=THE_CORRUPTED_GAUNTLET, 95=THE_HUEYCOATL, 96=THE_LEVIATHAN,
  97=THE_ROYAL_TITANS, 98=THE_WHISPERER, 99=THEATRE_OF_BLOOD,
  100=THEATRE_OF_BLOOD_HARD_MODE, 101=THERMONUCLEAR_SMOKE_DEVIL,
  102=TOMBS_OF_AMASCUT, 103=TOMBS_OF_AMASCUT_EXPERT_MODE, 104=TZKALZUK,
  105=TZTOKJAD, 106=VARDORVIS, 107=VENENATIS, 108=VETION, 109=VORKATH,
  110=WINTERTODT, 111=YAMA, 112=ZALCANO, 113=ZULRAH

# Size Comparison

JSON response (typical): ~15-25 KB
Binary response (typical): ~400-800 bytes
Savings: ~95-97%

# Notes for Client Implementation

- User ID is known from request, not included in response
- Snapshot IDs are omitted (not needed for display)
- Activity names can be hardcoded based on activity type index
- Ranks are omitted from snapshot (add back if needed)
- All timestamps are Unix milliseconds (int64)
*/
func EncodeDeltaSummaryBinary(resp DeltaSummaryResponse) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Header
	buf.WriteByte(binaryVersion)
	buf.WriteByte(0) // flags reserved

	// Encode snapshot
	encodeSnapshot(buf, resp.Snapshot)

	// Encode deltas
	encodeDeltas(buf, resp.Deltas)

	return buf.Bytes(), nil
}

func encodeSnapshot(buf *bytes.Buffer, snap snapshot.HiscoreSnapshot) {
	// Timestamp (unix millis)
	binary.Write(buf, binary.BigEndian, snap.Timestamp.UnixMilli())

	// Skills (no rank - client doesn't need it)
	buf.WriteByte(uint8(len(snap.Skills)))
	for _, s := range snap.Skills {
		buf.WriteByte(s.ActivityType.ToIndex())
		binary.Write(buf, binary.BigEndian, int32(s.Experience))
		binary.Write(buf, binary.BigEndian, int16(s.Level))
	}

	// Bosses (no rank)
	buf.WriteByte(uint8(len(snap.Bosses)))
	for _, b := range snap.Bosses {
		buf.WriteByte(b.ActivityType.ToIndex())
		binary.Write(buf, binary.BigEndian, int32(b.KillCount))
	}

	// Activities (no rank)
	buf.WriteByte(uint8(len(snap.Activities)))
	for _, a := range snap.Activities {
		buf.WriteByte(a.ActivityType.ToIndex())
		binary.Write(buf, binary.BigEndian, int32(a.Score))
	}
}

func encodeDeltas(buf *bytes.Buffer, deltas []delta.HiscoreDelta) {
	// Delta count
	binary.Write(buf, binary.BigEndian, uint16(len(deltas)))

	for _, d := range deltas {
		// Timestamp
		binary.Write(buf, binary.BigEndian, d.Timestamp.UnixMilli())

		// Skill deltas
		buf.WriteByte(uint8(len(d.Skills)))
		for _, s := range d.Skills {
			buf.WriteByte(s.ActivityType.ToIndex())
			binary.Write(buf, binary.BigEndian, int32(s.ExperienceGain))
			binary.Write(buf, binary.BigEndian, int16(s.LevelGain))
		}

		// Boss deltas
		buf.WriteByte(uint8(len(d.Bosses)))
		for _, b := range d.Bosses {
			buf.WriteByte(b.ActivityType.ToIndex())
			binary.Write(buf, binary.BigEndian, int32(b.KillCountGain))
		}

		// Activity deltas
		buf.WriteByte(uint8(len(d.Activities)))
		for _, a := range d.Activities {
			buf.WriteByte(a.ActivityType.ToIndex())
			binary.Write(buf, binary.BigEndian, int32(a.ScoreGain))
		}
	}
}
