import { RaidBuffs } from '/tbc/core/proto/common.js';
import { PartyBuffs } from '/tbc/core/proto/common.js';
import { IndividualBuffs } from '/tbc/core/proto/common.js';
import { Class } from '/tbc/core/proto/common.js';
import { Consumes } from '/tbc/core/proto/common.js';
import { Debuffs } from '/tbc/core/proto/common.js';
import { Drums } from '/tbc/core/proto/common.js';
import { Encounter } from '/tbc/core/proto/common.js';
import { ItemSlot } from '/tbc/core/proto/common.js';
import { MobType } from '/tbc/core/proto/common.js';
import { Potions } from '/tbc/core/proto/common.js';
import { RaidTarget } from '/tbc/core/proto/common.js';
import { Spec } from '/tbc/core/proto/common.js';
import { Stat } from '/tbc/core/proto/common.js';
import { TristateEffect } from '/tbc/core/proto/common.js'
import { Stats } from '/tbc/core/proto_utils/stats.js';
import { Sim } from '/tbc/core/sim.js';
import { DefaultTheme } from '/tbc/core/themes/default.js';

import { BalanceDruid, BalanceDruid_Rotation as BalanceDruidRotation, DruidTalents as DruidTalents, BalanceDruid_Options as BalanceDruidOptions } from '/tbc/core/proto/druid.js';
import { BalanceDruid_Rotation_PrimarySpell as PrimarySpell } from '/tbc/core/proto/druid.js';

import * as IconInputs from '/tbc/core/components/icon_inputs.js';
import * as OtherInputs from '/tbc/core/components/other_inputs.js';
import * as Gems from '/tbc/core/constants/gems.js';
import * as Tooltips from '/tbc/core/constants/tooltips.js';

import * as DruidInputs from './inputs.js';
import * as Presets from './presets.js';

const theme = new DefaultTheme<Spec.SpecBalanceDruid>(document.body, {
	spec: Spec.SpecBalanceDruid,
	// Can be 'Alpha', 'Beta', or 'Live'. Just adds a postfix to the generated title.
	releaseStatus: 'Beta',
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [
	],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatSpellPower,
		Stat.StatArcaneSpellPower,
		Stat.StatNatureSpellPower,
		Stat.StatSpellHit,
		Stat.StatSpellCrit,
		Stat.StatSpellHaste,
		Stat.StatMP5,
	],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatSpellPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: [
		Stat.StatStamina,
		Stat.StatIntellect,
		Stat.StatSpirit,
		Stat.StatSpellPower,
		Stat.StatArcaneSpellPower,
		Stat.StatNatureSpellPower,
		Stat.StatSpellHit,
		Stat.StatSpellCrit,
		Stat.StatSpellHaste,
		Stat.StatMP5,
	],

	defaults: {
		// Default equipped gear.
		gear: Presets.P1_ALLIANCE_BIS.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Stats.fromMap({
			[Stat.StatIntellect]: 0.54,
			[Stat.StatSpirit]: 0.1,
			[Stat.StatSpellPower]: 1,
			[Stat.StatArcaneSpellPower]: 1,
			[Stat.StatNatureSpellPower]: 0,
			[Stat.StatSpellCrit]: 0.84,
			[Stat.StatSpellHaste]: 1.29,
			[Stat.StatMP5]: 0.00,
		}),
		// Default consumes settings.
		consumes: Consumes.create({
			defaultPotion: Potions.SuperManaPotion,
		}),
		// Default rotation settings.
		rotation: BalanceDruidRotation.create({
			primarySpell: PrimarySpell.Adaptive,
			faerieFire: true,
		}),
		// Default talents.
		talents: Presets.StandardTalents.data,
		// Default spec-specific settings.
		specOptions: BalanceDruidOptions.create({
			innervateTarget: RaidTarget.create({
				targetIndex: 0, // In an individual sim the 0-indexed player is ourself.
			}),
		}),
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			arcaneBrilliance: true,
			divineSpirit: TristateEffect.TristateEffectImproved,
		}),
		partyBuffs: PartyBuffs.create({
			bloodlust: 1,
			manaSpringTotem: TristateEffect.TristateEffectRegular,
			totemOfWrath: 1,
			wrathOfAirTotem: TristateEffect.TristateEffectRegular,
		}),
		individualBuffs: IndividualBuffs.create({
			blessingOfKings: true,
			blessingOfWisdom: 2,
		}),
		debuffs: Debuffs.create({
			judgementOfWisdom: true,
			misery: true,
			curseOfElements: TristateEffect.TristateEffectRegular,
		}),
	},

	// IconInputs to include in the 'Self Buffs' section on the settings tab.
	selfBuffInputs: {
		tooltip: Tooltips.SELF_BUFFS_SECTION,
		icons: [
			DruidInputs.SelfInnervate,
			IconInputs.DrumsOfBattleConsume,
			IconInputs.DrumsOfRestorationConsume,
		],
	},
	// IconInputs to include in the 'Other Buffs' section on the settings tab.
	buffInputs: {
		tooltip: Tooltips.OTHER_BUFFS_SECTION,
		icons: [
			IconInputs.ArcaneBrilliance,
			IconInputs.DivineSpirit,
			IconInputs.BlessingOfKings,
			IconInputs.BlessingOfWisdom,
			IconInputs.DrumsOfBattleBuff,
			IconInputs.DrumsOfRestorationBuff,
			IconInputs.Bloodlust,
			IconInputs.WrathOfAirTotem,
			IconInputs.TotemOfWrath,
			IconInputs.ManaSpringTotem,
			IconInputs.DraeneiRacialCaster,
			IconInputs.EyeOfTheNight,
			IconInputs.ChainOfTheTwilightOwl,
			IconInputs.JadePendantOfBlasting,
			IconInputs.AtieshWarlock,
			IconInputs.AtieshMage,
		],
	},
	// IconInputs to include in the 'Debuffs' section on the settings tab.
	debuffInputs: {
		icons: [
			IconInputs.JudgementOfWisdom,
			IconInputs.ImprovedSealOfTheCrusader,
			IconInputs.CurseOfElements,
			IconInputs.Misery,
		],
	},
	// IconInputs to include in the 'Consumes' section on the settings tab.
	consumeInputs: {
		icons: [
			IconInputs.DefaultSuperManaPotion,
			IconInputs.DefaultDestructionPotion,
			IconInputs.DarkRune,
			IconInputs.FlaskOfBlindingLight,
			IconInputs.FlaskOfSupremePower,
			IconInputs.AdeptsElixir,
			IconInputs.ElixirOfMajorMageblood,
			IconInputs.ElixirOfDraenicWisdom,
			IconInputs.BrilliantWizardOil,
			IconInputs.SuperiorWizardOil,
			IconInputs.BlackenedBasilisk,
			IconInputs.SkullfishSoup,
			IconInputs.KreegsStoutBeatdown,
		],
	},
	// Inputs to include in the 'Rotation' section on the settings tab.
	rotationInputs: DruidInputs.BalanceDruidRotationConfig,
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.ShadowPriestDPS,
			OtherInputs.StartingPotion,
			OtherInputs.NumStartingPotions,
		],
	},
	encounterPicker: {
		// Whether to include 'Target Armor' in the 'Encounter' section of the settings tab.
		showTargetArmor: false,
		// Whether to include 'Num Targets' in the 'Encounter' section of the settings tab.
		showNumTargets: true,
	},

	// If true, the talents on the talents tab will not be individually modifiable by the user.
	// Note that the use can still pick between preset talents, if there is more than 1.
	freezeTalents: false,

  presets: {
		// Preset talents that the user can quickly select.
    talents: [
			Presets.StandardTalents,
		],
		// Preset gear configurations that the user can quickly select.
    gear: [
      Presets.P1_ALLIANCE_BIS,
      Presets.P2_ALLIANCE_BIS,
      Presets.P1_HORDE_BIS,
      Presets.P2_HORDE_BIS,
    ],
		// Preset encounter settings that the user can quickly select.
		encounters: [
		],
  },
});
theme.init();