package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"regexp"
	"strings"
)

func main() {
	type skill struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Rank  int    `json:"rank"`
		Level int    `json:"level"`
		Xp    int    `json:"xp"`
	}

	type activity struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Rank  int    `json:"rank"`
		Score int    `json:"score"`
	}

	type hiscore struct {
		Skills     []skill    `json:"skills"`
		Activities []activity `json:"activities"`
	}

	json := "{\"skills\":[{\"id\":0,\"name\":\"Overall\",\"rank\":244735,\"level\":2076,\"xp\":154609623},{\"id\":1,\"name\":\"Attack\",\"rank\":591988,\"level\":91,\"xp\":6380016},{\"id\":2,\"name\":\"Defence\",\"rank\":602493,\"level\":90,\"xp\":5420418},{\"id\":3,\"name\":\"Strength\",\"rank\":456814,\"level\":99,\"xp\":13362733},{\"id\":4,\"name\":\"Hitpoints\",\"rank\":456825,\"level\":99,\"xp\":16503312},{\"id\":5,\"name\":\"Ranged\",\"rank\":593949,\"level\":97,\"xp\":11476438},{\"id\":6,\"name\":\"Prayer\",\"rank\":412029,\"level\":82,\"xp\":2621330},{\"id\":7,\"name\":\"Magic\",\"rank\":500925,\"level\":97,\"xp\":11426434},{\"id\":8,\"name\":\"Cooking\",\"rank\":513417,\"level\":92,\"xp\":6646815},{\"id\":9,\"name\":\"Woodcutting\",\"rank\":345641,\"level\":90,\"xp\":5389093},{\"id\":10,\"name\":\"Fletching\",\"rank\":434470,\"level\":90,\"xp\":5358770},{\"id\":11,\"name\":\"Fishing\",\"rank\":649034,\"level\":81,\"xp\":2192820},{\"id\":12,\"name\":\"Firemaking\",\"rank\":903162,\"level\":80,\"xp\":2051312},{\"id\":13,\"name\":\"Crafting\",\"rank\":98731,\"level\":99,\"xp\":13422233},{\"id\":14,\"name\":\"Smithing\",\"rank\":208253,\"level\":90,\"xp\":5351046},{\"id\":15,\"name\":\"Mining\",\"rank\":294435,\"level\":89,\"xp\":5122293},{\"id\":16,\"name\":\"Herblore\",\"rank\":378498,\"level\":84,\"xp\":3093792},{\"id\":17,\"name\":\"Agility\",\"rank\":138155,\"level\":92,\"xp\":6633400},{\"id\":18,\"name\":\"Thieving\",\"rank\":217604,\"level\":92,\"xp\":6923796},{\"id\":19,\"name\":\"Slayer\",\"rank\":338247,\"level\":93,\"xp\":7739923},{\"id\":20,\"name\":\"Farming\",\"rank\":511661,\"level\":87,\"xp\":4360989},{\"id\":21,\"name\":\"Runecraft\",\"rank\":274383,\"level\":82,\"xp\":2426097},{\"id\":22,\"name\":\"Hunter\",\"rank\":210263,\"level\":90,\"xp\":5347358},{\"id\":23,\"name\":\"Construction\",\"rank\":167572,\"level\":90,\"xp\":5359205}],\"activities\":[{\"id\":0,\"name\":\"League Points\",\"rank\":-1,\"score\":-1},{\"id\":1,\"name\":\"Deadman Points\",\"rank\":-1,\"score\":-1},{\"id\":2,\"name\":\"Bounty Hunter - Hunter\",\"rank\":-1,\"score\":-1},{\"id\":3,\"name\":\"Bounty Hunter - Rogue\",\"rank\":-1,\"score\":-1},{\"id\":4,\"name\":\"Bounty Hunter (Legacy) - Hunter\",\"rank\":-1,\"score\":-1},{\"id\":5,\"name\":\"Bounty Hunter (Legacy) - Rogue\",\"rank\":-1,\"score\":-1},{\"id\":6,\"name\":\"Clue Scrolls (all)\",\"rank\":556277,\"score\":78},{\"id\":7,\"name\":\"Clue Scrolls (beginner)\",\"rank\":1881983,\"score\":1},{\"id\":8,\"name\":\"Clue Scrolls (easy)\",\"rank\":1306433,\"score\":2},{\"id\":9,\"name\":\"Clue Scrolls (medium)\",\"rank\":702461,\"score\":10},{\"id\":10,\"name\":\"Clue Scrolls (hard)\",\"rank\":383891,\"score\":55},{\"id\":11,\"name\":\"Clue Scrolls (elite)\",\"rank\":339462,\"score\":9},{\"id\":12,\"name\":\"Clue Scrolls (master)\",\"rank\":435861,\"score\":1},{\"id\":13,\"name\":\"LMS - Rank\",\"rank\":195196,\"score\":605},{\"id\":14,\"name\":\"PvP Arena - Rank\",\"rank\":-1,\"score\":-1},{\"id\":15,\"name\":\"Soul Wars Zeal\",\"rank\":-1,\"score\":-1},{\"id\":16,\"name\":\"Rifts closed\",\"rank\":143619,\"score\":210},{\"id\":17,\"name\":\"Colosseum Glory\",\"rank\":-1,\"score\":-1},{\"id\":18,\"name\":\"Collections Logged\",\"rank\":-1,\"score\":-1},{\"id\":19,\"name\":\"Abyssal Sire\",\"rank\":-1,\"score\":-1},{\"id\":20,\"name\":\"Alchemical Hydra\",\"rank\":-1,\"score\":-1},{\"id\":21,\"name\":\"Amoxliatl\",\"rank\":-1,\"score\":-1},{\"id\":22,\"name\":\"Araxxor\",\"rank\":-1,\"score\":-1},{\"id\":23,\"name\":\"Artio\",\"rank\":137810,\"score\":7},{\"id\":24,\"name\":\"Barrows Chests\",\"rank\":258446,\"score\":233},{\"id\":25,\"name\":\"Bryophyta\",\"rank\":-1,\"score\":-1},{\"id\":26,\"name\":\"Callisto\",\"rank\":111614,\"score\":101},{\"id\":27,\"name\":\"Calvar'ion\",\"rank\":-1,\"score\":-1},{\"id\":28,\"name\":\"Cerberus\",\"rank\":185234,\"score\":285},{\"id\":29,\"name\":\"Chambers of Xeric\",\"rank\":-1,\"score\":-1},{\"id\":30,\"name\":\"Chambers of Xeric: Challenge Mode\",\"rank\":-1,\"score\":-1},{\"id\":31,\"name\":\"Chaos Elemental\",\"rank\":-1,\"score\":-1},{\"id\":32,\"name\":\"Chaos Fanatic\",\"rank\":-1,\"score\":-1},{\"id\":33,\"name\":\"Commander Zilyana\",\"rank\":146497,\"score\":92},{\"id\":34,\"name\":\"Corporeal Beast\",\"rank\":-1,\"score\":-1},{\"id\":35,\"name\":\"Crazy Archaeologist\",\"rank\":179205,\"score\":42},{\"id\":36,\"name\":\"Dagannoth Prime\",\"rank\":278255,\"score\":56},{\"id\":37,\"name\":\"Dagannoth Rex\",\"rank\":376210,\"score\":79},{\"id\":38,\"name\":\"Dagannoth Supreme\",\"rank\":285793,\"score\":51},{\"id\":39,\"name\":\"Deranged Archaeologist\",\"rank\":-1,\"score\":-1},{\"id\":40,\"name\":\"Duke Sucellus\",\"rank\":-1,\"score\":-1},{\"id\":41,\"name\":\"General Graardor\",\"rank\":136545,\"score\":306},{\"id\":42,\"name\":\"Giant Mole\",\"rank\":-1,\"score\":-1},{\"id\":43,\"name\":\"Grotesque Guardians\",\"rank\":-1,\"score\":-1},{\"id\":44,\"name\":\"Hespori\",\"rank\":281141,\"score\":39},{\"id\":45,\"name\":\"Kalphite Queen\",\"rank\":-1,\"score\":-1},{\"id\":46,\"name\":\"King Black Dragon\",\"rank\":658393,\"score\":10},{\"id\":47,\"name\":\"Kraken\",\"rank\":459361,\"score\":168},{\"id\":48,\"name\":\"Kree'Arra\",\"rank\":-1,\"score\":-1},{\"id\":49,\"name\":\"K'ril Tsutsaroth\",\"rank\":202296,\"score\":38},{\"id\":50,\"name\":\"Lunar Chests\",\"rank\":-1,\"score\":-1},{\"id\":51,\"name\":\"Mimic\",\"rank\":-1,\"score\":-1},{\"id\":52,\"name\":\"Nex\",\"rank\":-1,\"score\":-1},{\"id\":53,\"name\":\"Nightmare\",\"rank\":-1,\"score\":-1},{\"id\":54,\"name\":\"Phosani's Nightmare\",\"rank\":-1,\"score\":-1},{\"id\":55,\"name\":\"Obor\",\"rank\":-1,\"score\":-1},{\"id\":56,\"name\":\"Phantom Muspah\",\"rank\":101954,\"score\":92},{\"id\":57,\"name\":\"Sarachnis\",\"rank\":304454,\"score\":22},{\"id\":58,\"name\":\"Scorpia\",\"rank\":-1,\"score\":-1},{\"id\":59,\"name\":\"Scurrius\",\"rank\":-1,\"score\":-1},{\"id\":60,\"name\":\"Skotizo\",\"rank\":-1,\"score\":-1},{\"id\":61,\"name\":\"Sol Heredit\",\"rank\":-1,\"score\":-1},{\"id\":62,\"name\":\"Spindel\",\"rank\":94195,\"score\":35},{\"id\":63,\"name\":\"Tempoross\",\"rank\":765980,\"score\":8},{\"id\":64,\"name\":\"The Gauntlet\",\"rank\":229000,\"score\":11},{\"id\":65,\"name\":\"The Corrupted Gauntlet\",\"rank\":45984,\"score\":497},{\"id\":66,\"name\":\"The Hueycoatl\",\"rank\":-1,\"score\":-1},{\"id\":67,\"name\":\"The Leviathan\",\"rank\":-1,\"score\":-1},{\"id\":68,\"name\":\"The Royal Titans\",\"rank\":117117,\"score\":23},{\"id\":69,\"name\":\"The Whisperer\",\"rank\":-1,\"score\":-1},{\"id\":70,\"name\":\"Theatre of Blood\",\"rank\":-1,\"score\":-1},{\"id\":71,\"name\":\"Theatre of Blood: Hard Mode\",\"rank\":-1,\"score\":-1},{\"id\":72,\"name\":\"Thermonuclear Smoke Devil\",\"rank\":189768,\"score\":203},{\"id\":73,\"name\":\"Tombs of Amascut\",\"rank\":127556,\"score\":41},{\"id\":74,\"name\":\"Tombs of Amascut: Expert Mode\",\"rank\":93810,\"score\":33},{\"id\":75,\"name\":\"TzKal-Zuk\",\"rank\":-1,\"score\":-1},{\"id\":76,\"name\":\"TzTok-Jad\",\"rank\":99091,\"score\":12},{\"id\":77,\"name\":\"Vardorvis\",\"rank\":-1,\"score\":-1},{\"id\":78,\"name\":\"Venenatis\",\"rank\":113269,\"score\":62},{\"id\":79,\"name\":\"Vet'ion\",\"rank\":-1,\"score\":-1},{\"id\":80,\"name\":\"Vorkath\",\"rank\":595389,\"score\":27},{\"id\":81,\"name\":\"Wintertodt\",\"rank\":806006,\"score\":109},{\"id\":82,\"name\":\"Zalcano\",\"rank\":289401,\"score\":5},{\"id\":83,\"name\":\"Zulrah\",\"rank\":383641,\"score\":58}]}"

	var hs hiscore
	err := jsoniter.Unmarshal([]byte(json), &hs)
	if err != nil {
		panic(err)
	}

	for _, s := range hs.Skills {
		fmt.Printf("ActivityType%s ActivityType = \"%s\"\n", sanitizeName(s.Name), sanitizeNameUpper(s.Name))
	}

	for _, s := range hs.Activities {
		fmt.Printf("ActivityType%s ActivityType = \"%s\"\n", sanitizeName(s.Name), sanitizeNameUpper(s.Name))
	}

	for _, s := range hs.Skills {
		fmt.Printf("ActivityType%s,\n", sanitizeName(s.Name))
	}

	for _, s := range hs.Activities {
		fmt.Printf("ActivityType%s,\n", sanitizeName(s.Name))
	}
}

func removeIllegalChars(str string) string {
	remove := regexp.MustCompile("['\\-_:()]")
	return remove.ReplaceAllString(str, "")
}

func sanitizeName(name string) string {
	name = removeIllegalChars(name)
	return strings.ReplaceAll(name, " ", "")
}

func sanitizeNameUpper(name string) string {
	name = removeIllegalChars(name)
	return strings.ToUpper(strings.ReplaceAll(name, " ", "_"))
}
