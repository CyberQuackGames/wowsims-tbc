package elemental

import (
	"github.com/wowsims/tbc/sim/core/items"
	"github.com/wowsims/tbc/sim/core/proto"
)

var BasicRaidBuffs = &proto.RaidBuffs{}
var BasicPartyBuffs = &proto.PartyBuffs{
	Bloodlust: 1,
}
var BasicIndividualBuffs = &proto.IndividualBuffs{}

var StandardTalents = &proto.ShamanTalents{
	Convection:         5,
	Concussion:         5,
	ElementalFocus:     true,
	CallOfThunder:      5,
	ElementalFury:      true,
	UnrelentingStorm:   3,
	ElementalPrecision: 3,
	LightningMastery:   5,
	ElementalMastery:   true,
	LightningOverload:  5,
	TotemOfWrath:       true,

	TotemicFocus:    5,
	NaturesGuidance: 3,
	TidalMastery:    5,
}

var eleShamOptionsNoBuffs = &proto.ElementalShaman_Options{
	WaterShield: true,
	// Bloodlust:       true,
	// ManaSpringTotem: true,
	// TotemOfWrath:    true,
	// WrathOfAirTotem: true,
}
var PlayerOptionsAdaptiveNoBuffs = &proto.Player_ElementalShaman{
	ElementalShaman: &proto.ElementalShaman{
		Talents: StandardTalents,
		Options: eleShamOptionsNoBuffs,
		Rotation: &proto.ElementalShaman_Rotation{
			Type: proto.ElementalShaman_Rotation_Adaptive,
		},
	},
}

var eleShamOptions = &proto.ElementalShaman_Options{
	WaterShield:     true,
	Bloodlust:       true,
	ManaSpringTotem: true,
	TotemOfWrath:    true,
	WrathOfAirTotem: true,
}
var PlayerOptionsAdaptive = &proto.Player_ElementalShaman{
	ElementalShaman: &proto.ElementalShaman{
		Talents: StandardTalents,
		Options: eleShamOptions,
		Rotation: &proto.ElementalShaman_Rotation{
			Type: proto.ElementalShaman_Rotation_Adaptive,
		},
	},
}

var PlayerOptionsLBOnly = &proto.Player_ElementalShaman{
	ElementalShaman: &proto.ElementalShaman{
		Talents: StandardTalents,
		Options: eleShamOptions,
		Rotation: &proto.ElementalShaman_Rotation{
			Type: proto.ElementalShaman_Rotation_LBOnly,
		},
	},
}

var PlayerOptionsCLOnClearcast = &proto.Player_ElementalShaman{
	ElementalShaman: &proto.ElementalShaman{
		Talents: StandardTalents,
		Options: eleShamOptions,
		Rotation: &proto.ElementalShaman_Rotation{
			Type: proto.ElementalShaman_Rotation_CLOnClearcast,
		},
	},
}

var FullRaidBuffs = &proto.RaidBuffs{
	ArcaneBrilliance: true,
	GiftOfTheWild:    proto.TristateEffect_TristateEffectImproved,
}
var FullPartyBuffs = &proto.PartyBuffs{
	MoonkinAura: proto.TristateEffect_TristateEffectRegular,
}
var FullIndividualBuffs = &proto.IndividualBuffs{
	BlessingOfKings:  true,
	BlessingOfWisdom: proto.TristateEffect_TristateEffectImproved,
	ShadowPriestDps:  500,
}

var FullConsumes = &proto.Consumes{
	FlaskOfBlindingLight: true,
	BrilliantWizardOil:   true,
	BlackenedBasilisk:    true,
	DefaultPotion:        proto.Potions_SuperManaPotion,
	StartingPotion:       proto.Potions_DestructionPotion,
	NumStartingPotions:   1,
	DarkRune:             true,
	Drums:                proto.Drums_DrumsOfBattle,
}

var NoDebuffTarget = &proto.Target{
	Debuffs: &proto.Debuffs{},
}

var FullDebuffTarget = &proto.Target{
	Debuffs: &proto.Debuffs{
		ImprovedSealOfTheCrusader: true,
		JudgementOfWisdom:         true,
		Misery:                    true,
	},
}

var PreRaidGear = items.EquipmentSpecFromStrings([]items.ItemStringSpec{
	items.ItemStringSpec{
		Name:    "Tidefury Helm",
		Enchant: "Glyph of Power",
		Gems: []string{
			"Runed Living Ruby",
			"Insightful Earthstorm Diamond",
		},
	},
	items.ItemStringSpec{
		Name:    "Brooch of Heightened Potential",
		Enchant: "Zandalar Signet of Mojo",
	},
	items.ItemStringSpec{
		Name: "Tidefury Shoulderguards",
	},
	items.ItemStringSpec{
		Name: "Cloak of the Black Void",
	},
	items.ItemStringSpec{
		Name: "Tidefury Chestpiece",
	},
	items.ItemStringSpec{
		Name: "Shattrath Wraps",
	},
	items.ItemStringSpec{
		Name: "Tidefury Gauntlets",
	},
	items.ItemStringSpec{
		Name: "Moonrage Girdle",
	},
	items.ItemStringSpec{
		Name:    "Tidefury Kilt",
		Enchant: "Mystic Spellthread",
	},
	items.ItemStringSpec{
		Name: "Earthbreaker's Greaves",
	},
	items.ItemStringSpec{
		Name: "Seal of the Exorcist",
	},
	items.ItemStringSpec{
		Name: "Spectral Band of Innervation",
	},
	items.ItemStringSpec{
		Name: "Xi'ri's Gift",
	},
	items.ItemStringSpec{
		Name: "Quagmirran's Eye",
	},
	items.ItemStringSpec{
		Name: "Totem of the Void",
	},
	items.ItemStringSpec{
		Name: "Sky Breaker",
	},
	items.ItemStringSpec{
		Name: "Silvermoon Crest Shield",
	},
})

var P1Gear = items.EquipmentSpecFromStrings([]items.ItemStringSpec{
	items.ItemStringSpec{
		Name:    "Cyclone Faceguard",
		Enchant: "Glyph of Power",
		Gems: []string{
			"Chaotic Skyfire Diamond",
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name: "Adornment of Stolen Souls",
	},
	items.ItemStringSpec{
		Name:    "Cyclone Shoulderguards",
		Enchant: "Greater Inscription of Discipline",
		Gems: []string{
			"Potent Noble Topaz",
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name: "Ruby Drape of the Mysticant",
	},
	items.ItemStringSpec{
		Name:    "Netherstrike Breastplate",
		Enchant: "Chest - Exceptional Stats",
		Gems: []string{
			"Runed Living Ruby",
			"Runed Living Ruby",
			"Runed Living Ruby",
		},
	},
	items.ItemStringSpec{
		Name:    "Netherstrike Bracers",
		Enchant: "Bracer - Spellpower",
		Gems: []string{
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name:    "Soul-Eater's Handwraps",
		Enchant: "Gloves - Major Spellpower",
		Gems: []string{
			"Potent Noble Topaz",
			"Glowing Nightseye",
		},
	},
	items.ItemStringSpec{
		Name: "Netherstrike Belt",
		Gems: []string{
			"Glowing Nightseye",
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name:    "Stormsong Kilt",
		Enchant: "Runic Spellthread",
		Gems: []string{
			"Potent Noble Topaz",
			"Runed Living Ruby",
			"Glowing Nightseye",
		},
	},
	items.ItemStringSpec{
		Name: "Windshear Boots",
	},
	items.ItemStringSpec{
		Name:    "Ring of Unrelenting Storms",
		Enchant: "Ring - Spellpower",
	},
	items.ItemStringSpec{
		Name:    "Ring of Recurrence",
		Enchant: "Ring - Spellpower",
	},
	items.ItemStringSpec{
		Name: "The Lightning Capacitor",
	},
	items.ItemStringSpec{
		Name: "Icon of the Silver Crescent",
	},
	items.ItemStringSpec{
		Name: "Totem of the Void",
	},
	items.ItemStringSpec{
		Name:    "Nathrezim Mindblade",
		Enchant: "Weapon - Major Spellpower",
	},
	items.ItemStringSpec{
		Name:    "Mazthoril Honor Shield",
		Enchant: "Shield - Intellect",
	},
})

var P1Tidefury = items.EquipmentSpecFromStrings([]items.ItemStringSpec{
	items.ItemStringSpec{
		Name:    "Cyclone Faceguard",
		Enchant: "Glyph of Power",
		Gems: []string{
			"Chaotic Skyfire Diamond",
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name: "Adornment of Stolen Souls",
	},
	items.ItemStringSpec{
		Name:    "Cyclone Shoulderguards",
		Enchant: "Greater Inscription of Discipline",
		Gems: []string{
			"Potent Noble Topaz",
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name: "Ruby Drape of the Mysticant",
	},
	items.ItemStringSpec{
		Name:    "Netherstrike Breastplate",
		Enchant: "Chest - Exceptional Stats",
		Gems: []string{
			"Runed Living Ruby",
			"Runed Living Ruby",
			"Runed Living Ruby",
		},
	},
	items.ItemStringSpec{
		Name:    "Netherstrike Bracers",
		Enchant: "Bracer - Spellpower",
		Gems: []string{
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name:    "Tidefury Gauntlets",
		Enchant: "Gloves - Major Spellpower",
		// Gems: []string{
		// 	"Potent Noble Topaz",
		// 	"Glowing Nightseye",
		// },
	},
	items.ItemStringSpec{
		Name: "Netherstrike Belt",
		Gems: []string{
			"Glowing Nightseye",
			"Potent Noble Topaz",
		},
	},
	items.ItemStringSpec{
		Name:    "Tidefury Kilt",
		Enchant: "Runic Spellthread",
		// Gems: []string{
		// 	"Potent Noble Topaz",
		// 	"Runed Living Ruby",
		// 	"Glowing Nightseye",
		// },
	},
	items.ItemStringSpec{
		Name: "Windshear Boots",
	},
	items.ItemStringSpec{
		Name:    "Ring of Unrelenting Storms",
		Enchant: "Ring - Spellpower",
	},
	items.ItemStringSpec{
		Name:    "Ring of Recurrence",
		Enchant: "Ring - Spellpower",
	},
	items.ItemStringSpec{
		Name: "The Lightning Capacitor",
	},
	items.ItemStringSpec{
		Name: "Icon of the Silver Crescent",
	},
	items.ItemStringSpec{
		Name: "Totem of the Void",
	},
	items.ItemStringSpec{
		Name:    "Nathrezim Mindblade",
		Enchant: "Weapon - Major Spellpower",
	},
	items.ItemStringSpec{
		Name:    "Mazthoril Honor Shield",
		Enchant: "Shield - Intellect",
	},
})