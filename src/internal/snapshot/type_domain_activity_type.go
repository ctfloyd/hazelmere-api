package snapshot

type ActivityType string

const (
	ActivityTypeUnknown                      ActivityType = "UNKNOWN"
	ActivityTypeOverall                      ActivityType = "OVERALL"
	ActivityTypeAttack                       ActivityType = "ATTACK"
	ActivityTypeDefence                      ActivityType = "DEFENCE"
	ActivityTypeStrength                     ActivityType = "STRENGTH"
	ActivityTypeHitpoints                    ActivityType = "HITPOINTS"
	ActivityTypeRanged                       ActivityType = "RANGED"
	ActivityTypePrayer                       ActivityType = "PRAYER"
	ActivityTypeMagic                        ActivityType = "MAGIC"
	ActivityTypeCooking                      ActivityType = "COOKING"
	ActivityTypeWoodcutting                  ActivityType = "WOODCUTTING"
	ActivityTypeFletching                    ActivityType = "FLETCHING"
	ActivityTypeFishing                      ActivityType = "FISHING"
	ActivityTypeFiremaking                   ActivityType = "FIREMAKING"
	ActivityTypeCrafting                     ActivityType = "CRAFTING"
	ActivityTypeSmithing                     ActivityType = "SMITHING"
	ActivityTypeMining                       ActivityType = "MINING"
	ActivityTypeHerblore                     ActivityType = "HERBLORE"
	ActivityTypeAgility                      ActivityType = "AGILITY"
	ActivityTypeThieving                     ActivityType = "THIEVING"
	ActivityTypeSlayer                       ActivityType = "SLAYER"
	ActivityTypeFarming                      ActivityType = "FARMING"
	ActivityTypeRunecraft                    ActivityType = "RUNECRAFT"
	ActivityTypeHunter                       ActivityType = "HUNTER"
	ActivityTypeConstruction                 ActivityType = "CONSTRUCTION"
	ActivityTypeLeaguePoints                 ActivityType = "LEAGUE_POINTS"
	ActivityTypeDeadmanPoints                ActivityType = "DEADMAN_POINTS"
	ActivityTypeBountyHunterHunter           ActivityType = "BOUNTY_HUNTER__HUNTER"
	ActivityTypeBountyHunterRogue            ActivityType = "BOUNTY_HUNTER__ROGUE"
	ActivityTypeBountyHunterLegacyHunter     ActivityType = "BOUNTY_HUNTER_LEGACY__HUNTER"
	ActivityTypeBountyHunterLegacyRogue      ActivityType = "BOUNTY_HUNTER_LEGACY__ROGUE"
	ActivityTypeClueScrollsall               ActivityType = "CLUE_SCROLLS_ALL"
	ActivityTypeClueScrollsbeginner          ActivityType = "CLUE_SCROLLS_BEGINNER"
	ActivityTypeClueScrollseasy              ActivityType = "CLUE_SCROLLS_EASY"
	ActivityTypeClueScrollsmedium            ActivityType = "CLUE_SCROLLS_MEDIUM"
	ActivityTypeClueScrollshard              ActivityType = "CLUE_SCROLLS_HARD"
	ActivityTypeClueScrollselite             ActivityType = "CLUE_SCROLLS_ELITE"
	ActivityTypeClueScrollsmaster            ActivityType = "CLUE_SCROLLS_MASTER"
	ActivityTypeLMSRank                      ActivityType = "LMS__RANK"
	ActivityTypePvPArenaRank                 ActivityType = "PVP_ARENA__RANK"
	ActivityTypeSoulWarsZeal                 ActivityType = "SOUL_WARS_ZEAL"
	ActivityTypeRiftsclosed                  ActivityType = "RIFTS_CLOSED"
	ActivityTypeColosseumGlory               ActivityType = "COLOSSEUM_GLORY"
	ActivityTypeCollectionsLogged            ActivityType = "COLLECTIONS_LOGGED"
	ActivityTypeAbyssalSire                  ActivityType = "ABYSSAL_SIRE"
	ActivityTypeAlchemicalHydra              ActivityType = "ALCHEMICAL_HYDRA"
	ActivityTypeAmoxliatl                    ActivityType = "AMOXLIATL"
	ActivityTypeAraxxor                      ActivityType = "ARAXXOR"
	ActivityTypeArtio                        ActivityType = "ARTIO"
	ActivityTypeBarrowsChests                ActivityType = "BARROWS_CHESTS"
	ActivityTypeBryophyta                    ActivityType = "BRYOPHYTA"
	ActivityTypeCallisto                     ActivityType = "CALLISTO"
	ActivityTypeCalvarion                    ActivityType = "CALVARION"
	ActivityTypeCerberus                     ActivityType = "CERBERUS"
	ActivityTypeChambersofXeric              ActivityType = "CHAMBERS_OF_XERIC"
	ActivityTypeChambersofXericChallengeMode ActivityType = "CHAMBERS_OF_XERIC_CHALLENGE_MODE"
	ActivityTypeChaosElemental               ActivityType = "CHAOS_ELEMENTAL"
	ActivityTypeChaosFanatic                 ActivityType = "CHAOS_FANATIC"
	ActivityTypeCommanderZilyana             ActivityType = "COMMANDER_ZILYANA"
	ActivityTypeCorporealBeast               ActivityType = "CORPOREAL_BEAST"
	ActivityTypeCrazyArchaeologist           ActivityType = "CRAZY_ARCHAEOLOGIST"
	ActivityTypeDagannothPrime               ActivityType = "DAGANNOTH_PRIME"
	ActivityTypeDagannothRex                 ActivityType = "DAGANNOTH_REX"
	ActivityTypeDagannothSupreme             ActivityType = "DAGANNOTH_SUPREME"
	ActivityTypeDerangedArchaeologist        ActivityType = "DERANGED_ARCHAEOLOGIST"
	ActivityTypeDukeSucellus                 ActivityType = "DUKE_SUCELLUS"
	ActivityTypeGeneralGraardor              ActivityType = "GENERAL_GRAARDOR"
	ActivityTypeGiantMole                    ActivityType = "GIANT_MOLE"
	ActivityTypeGrotesqueGuardians           ActivityType = "GROTESQUE_GUARDIANS"
	ActivityTypeHespori                      ActivityType = "HESPORI"
	ActivityTypeKalphiteQueen                ActivityType = "KALPHITE_QUEEN"
	ActivityTypeKingBlackDragon              ActivityType = "KING_BLACK_DRAGON"
	ActivityTypeKraken                       ActivityType = "KRAKEN"
	ActivityTypeKreeArra                     ActivityType = "KREEARRA"
	ActivityTypeKrilTsutsaroth               ActivityType = "KRIL_TSUTSAROTH"
	ActivityTypeLunarChests                  ActivityType = "LUNAR_CHESTS"
	ActivityTypeMimic                        ActivityType = "MIMIC"
	ActivityTypeNex                          ActivityType = "NEX"
	ActivityTypeNightmare                    ActivityType = "NIGHTMARE"
	ActivityTypePhosanisNightmare            ActivityType = "PHOSANIS_NIGHTMARE"
	ActivityTypeObor                         ActivityType = "OBOR"
	ActivityTypePhantomMuspah                ActivityType = "PHANTOM_MUSPAH"
	ActivityTypeSarachnis                    ActivityType = "SARACHNIS"
	ActivityTypeScorpia                      ActivityType = "SCORPIA"
	ActivityTypeScurrius                     ActivityType = "SCURRIUS"
	ActivityTypeSkotizo                      ActivityType = "SKOTIZO"
	ActivityTypeSolHeredit                   ActivityType = "SOL_HEREDIT"
	ActivityTypeSpindel                      ActivityType = "SPINDEL"
	ActivityTypeTempoross                    ActivityType = "TEMPOROSS"
	ActivityTypeTheGauntlet                  ActivityType = "THE_GAUNTLET"
	ActivityTypeTheCorruptedGauntlet         ActivityType = "THE_CORRUPTED_GAUNTLET"
	ActivityTypeTheHueycoatl                 ActivityType = "THE_HUEYCOATL"
	ActivityTypeTheLeviathan                 ActivityType = "THE_LEVIATHAN"
	ActivityTypeTheRoyalTitans               ActivityType = "THE_ROYAL_TITANS"
	ActivityTypeTheWhisperer                 ActivityType = "THE_WHISPERER"
	ActivityTypeTheatreOfBlood               ActivityType = "THEATRE_OF_BLOOD"
	ActivityTypeTheatreOfBloodHardMode       ActivityType = "THEATRE_OF_BLOOD_HARD_MODE"
	ActivityTypeThermonuclearSmokeDevil      ActivityType = "THERMONUCLEAR_SMOKE_DEVIL"
	ActivityTypeTombsOfAmascut               ActivityType = "TOMBS_OF_AMASCUT"
	ActivityTypeTombsOfAmascutExpertMode     ActivityType = "TOMBS_OF_AMASCUT_EXPERT_MODE"
	ActivityTypeTzKalZuk                     ActivityType = "TZKALZUK"
	ActivityTypeTzTokJad                     ActivityType = "TZTOKJAD"
	ActivityTypeVardorvis                    ActivityType = "VARDORVIS"
	ActivityTypeVenenatis                    ActivityType = "VENENATIS"
	ActivityTypeVetion                       ActivityType = "VETION"
	ActivityTypeVorkath                      ActivityType = "VORKATH"
	ActivityTypeWintertodt                   ActivityType = "WINTERTODT"
	ActivityTypeYama                         ActivityType = "YAMA"
	ActivityTypeZalcano                      ActivityType = "ZALCANO"
	ActivityTypeZulrah                       ActivityType = "ZULRAH"
)

var AllSkillActivityTypes = []ActivityType{
	ActivityTypeOverall,
	ActivityTypeAttack,
	ActivityTypeDefence,
	ActivityTypeStrength,
	ActivityTypeHitpoints,
	ActivityTypeRanged,
	ActivityTypePrayer,
	ActivityTypeMagic,
	ActivityTypeCooking,
	ActivityTypeWoodcutting,
	ActivityTypeFletching,
	ActivityTypeFishing,
	ActivityTypeFiremaking,
	ActivityTypeCrafting,
	ActivityTypeSmithing,
	ActivityTypeMining,
	ActivityTypeHerblore,
	ActivityTypeAgility,
	ActivityTypeThieving,
	ActivityTypeSlayer,
	ActivityTypeFarming,
	ActivityTypeRunecraft,
	ActivityTypeHunter,
	ActivityTypeConstruction,
}

var AllActivityActivityTypes = []ActivityType{
	ActivityTypeLeaguePoints,
	ActivityTypeDeadmanPoints,
	ActivityTypeBountyHunterHunter,
	ActivityTypeBountyHunterRogue,
	ActivityTypeBountyHunterLegacyHunter,
	ActivityTypeBountyHunterLegacyRogue,
	ActivityTypeClueScrollsall,
	ActivityTypeClueScrollsbeginner,
	ActivityTypeClueScrollseasy,
	ActivityTypeClueScrollsmedium,
	ActivityTypeClueScrollshard,
	ActivityTypeClueScrollselite,
	ActivityTypeClueScrollsmaster,
	ActivityTypeLMSRank,
	ActivityTypePvPArenaRank,
	ActivityTypeSoulWarsZeal,
	ActivityTypeRiftsclosed,
	ActivityTypeColosseumGlory,
	ActivityTypeCollectionsLogged,
}

var AllBossActivityTypes = []ActivityType{
	ActivityTypeAbyssalSire,
	ActivityTypeAlchemicalHydra,
	ActivityTypeAmoxliatl,
	ActivityTypeAraxxor,
	ActivityTypeArtio,
	ActivityTypeBarrowsChests,
	ActivityTypeBryophyta,
	ActivityTypeCallisto,
	ActivityTypeCalvarion,
	ActivityTypeCerberus,
	ActivityTypeChambersofXeric,
	ActivityTypeChambersofXericChallengeMode,
	ActivityTypeChaosElemental,
	ActivityTypeChaosFanatic,
	ActivityTypeCommanderZilyana,
	ActivityTypeCorporealBeast,
	ActivityTypeCrazyArchaeologist,
	ActivityTypeDagannothPrime,
	ActivityTypeDagannothRex,
	ActivityTypeDagannothSupreme,
	ActivityTypeDerangedArchaeologist,
	ActivityTypeDukeSucellus,
	ActivityTypeGeneralGraardor,
	ActivityTypeGiantMole,
	ActivityTypeGrotesqueGuardians,
	ActivityTypeHespori,
	ActivityTypeKalphiteQueen,
	ActivityTypeKingBlackDragon,
	ActivityTypeKraken,
	ActivityTypeKreeArra,
	ActivityTypeKrilTsutsaroth,
	ActivityTypeLunarChests,
	ActivityTypeMimic,
	ActivityTypeNex,
	ActivityTypeNightmare,
	ActivityTypePhosanisNightmare,
	ActivityTypeObor,
	ActivityTypePhantomMuspah,
	ActivityTypeSarachnis,
	ActivityTypeScorpia,
	ActivityTypeScurrius,
	ActivityTypeSkotizo,
	ActivityTypeSolHeredit,
	ActivityTypeSpindel,
	ActivityTypeTempoross,
	ActivityTypeTheGauntlet,
	ActivityTypeTheCorruptedGauntlet,
	ActivityTypeTheHueycoatl,
	ActivityTypeTheLeviathan,
	ActivityTypeTheRoyalTitans,
	ActivityTypeTheWhisperer,
	ActivityTypeTheatreOfBlood,
	ActivityTypeTheatreOfBloodHardMode,
	ActivityTypeThermonuclearSmokeDevil,
	ActivityTypeTombsOfAmascut,
	ActivityTypeTombsOfAmascutExpertMode,
	ActivityTypeTzKalZuk,
	ActivityTypeTzTokJad,
	ActivityTypeVardorvis,
	ActivityTypeVenenatis,
	ActivityTypeVetion,
	ActivityTypeVorkath,
	ActivityTypeWintertodt,
	ActivityTypeYama,
	ActivityTypeZalcano,
	ActivityTypeZulrah,
}

var AllActivityTypes = []ActivityType{
	ActivityTypeUnknown,
	ActivityTypeOverall,
	ActivityTypeAttack,
	ActivityTypeDefence,
	ActivityTypeStrength,
	ActivityTypeHitpoints,
	ActivityTypeRanged,
	ActivityTypePrayer,
	ActivityTypeMagic,
	ActivityTypeCooking,
	ActivityTypeWoodcutting,
	ActivityTypeFletching,
	ActivityTypeFishing,
	ActivityTypeFiremaking,
	ActivityTypeCrafting,
	ActivityTypeSmithing,
	ActivityTypeMining,
	ActivityTypeHerblore,
	ActivityTypeAgility,
	ActivityTypeThieving,
	ActivityTypeSlayer,
	ActivityTypeFarming,
	ActivityTypeRunecraft,
	ActivityTypeHunter,
	ActivityTypeConstruction,
	ActivityTypeLeaguePoints,
	ActivityTypeDeadmanPoints,
	ActivityTypeBountyHunterHunter,
	ActivityTypeBountyHunterRogue,
	ActivityTypeBountyHunterLegacyHunter,
	ActivityTypeBountyHunterLegacyRogue,
	ActivityTypeClueScrollsall,
	ActivityTypeClueScrollsbeginner,
	ActivityTypeClueScrollseasy,
	ActivityTypeClueScrollsmedium,
	ActivityTypeClueScrollshard,
	ActivityTypeClueScrollselite,
	ActivityTypeClueScrollsmaster,
	ActivityTypeLMSRank,
	ActivityTypePvPArenaRank,
	ActivityTypeSoulWarsZeal,
	ActivityTypeRiftsclosed,
	ActivityTypeColosseumGlory,
	ActivityTypeCollectionsLogged,
	ActivityTypeAbyssalSire,
	ActivityTypeAlchemicalHydra,
	ActivityTypeAmoxliatl,
	ActivityTypeAraxxor,
	ActivityTypeArtio,
	ActivityTypeBarrowsChests,
	ActivityTypeBryophyta,
	ActivityTypeCallisto,
	ActivityTypeCalvarion,
	ActivityTypeCerberus,
	ActivityTypeChambersofXeric,
	ActivityTypeChambersofXericChallengeMode,
	ActivityTypeChaosElemental,
	ActivityTypeChaosFanatic,
	ActivityTypeCommanderZilyana,
	ActivityTypeCorporealBeast,
	ActivityTypeCrazyArchaeologist,
	ActivityTypeDagannothPrime,
	ActivityTypeDagannothRex,
	ActivityTypeDagannothSupreme,
	ActivityTypeDerangedArchaeologist,
	ActivityTypeDukeSucellus,
	ActivityTypeGeneralGraardor,
	ActivityTypeGiantMole,
	ActivityTypeGrotesqueGuardians,
	ActivityTypeHespori,
	ActivityTypeKalphiteQueen,
	ActivityTypeKingBlackDragon,
	ActivityTypeKraken,
	ActivityTypeKreeArra,
	ActivityTypeKrilTsutsaroth,
	ActivityTypeLunarChests,
	ActivityTypeMimic,
	ActivityTypeNex,
	ActivityTypeNightmare,
	ActivityTypePhosanisNightmare,
	ActivityTypeObor,
	ActivityTypePhantomMuspah,
	ActivityTypeSarachnis,
	ActivityTypeScorpia,
	ActivityTypeScurrius,
	ActivityTypeSkotizo,
	ActivityTypeSolHeredit,
	ActivityTypeSpindel,
	ActivityTypeTempoross,
	ActivityTypeTheGauntlet,
	ActivityTypeTheCorruptedGauntlet,
	ActivityTypeTheHueycoatl,
	ActivityTypeTheLeviathan,
	ActivityTypeTheRoyalTitans,
	ActivityTypeTheWhisperer,
	ActivityTypeTheatreOfBlood,
	ActivityTypeTheatreOfBloodHardMode,
	ActivityTypeThermonuclearSmokeDevil,
	ActivityTypeTombsOfAmascut,
	ActivityTypeTombsOfAmascutExpertMode,
	ActivityTypeTzKalZuk,
	ActivityTypeTzTokJad,
	ActivityTypeVardorvis,
	ActivityTypeVenenatis,
	ActivityTypeVetion,
	ActivityTypeVorkath,
	ActivityTypeWintertodt,
	ActivityTypeYama,
	ActivityTypeZalcano,
	ActivityTypeZulrah,
}

func ActivityTypeFromValue(value string) ActivityType {
	for _, at := range AllActivityTypes {
		if value == string(at) {
			return at
		}
	}
	return ActivityTypeUnknown
}
