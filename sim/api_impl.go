// Proto based function interface for the simulator
package sim

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/items"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	RegisterAll()
}

func getGearListImpl(request *proto.GearListRequest) *proto.GearListResult {
	result := &proto.GearListResult{}

	for i := range items.Items {
		item := items.Items[i]
		result.Items = append(result.Items,
			&proto.Item{
				Id:               item.ID,
				Type:             proto.ItemType(item.Type),
				ArmorType:        proto.ArmorType(item.ArmorType),
				WeaponType:       proto.WeaponType(item.WeaponType),
				HandType:         proto.HandType(item.HandType),
				RangedWeaponType: proto.RangedWeaponType(item.RangedWeaponType),
				Name:             item.Name,
				Stats:            item.Stats[:],
				Phase:            int32(item.Phase),
				Quality:          item.Quality, // Hack until we use generated items
				GemSockets:       item.GemSockets,
				SocketBonus:      item.SocketBonus[:],
			},
		)
	}
	for i := range items.Gems {
		gem := items.Gems[i]
		result.Gems = append(result.Gems, &proto.Gem{
			Id:      gem.ID,
			Name:    gem.Name,
			Stats:   gem.Stats[:],
			Color:   gem.Color,
			Phase:   int32(gem.Phase),
			Quality: gem.Quality, // Hack until we use generated items
			Unique:  gem.Unique,
		})
	}
	for i := range items.Enchants {
		enchant := items.Enchants[i]
		result.Enchants = append(result.Enchants, &proto.Enchant{
			Id:       enchant.ID,
			EffectId: enchant.EffectID,
			Name:     enchant.Name,
			Type:     enchant.ItemType,
			Stats:    enchant.Bonus[:],
			Quality:  proto.ItemQuality(4),
		})
	}

	return result
}

func computeStatsImpl(request *proto.ComputeStatsRequest) *proto.ComputeStatsResult {
	return statsFromIndSimRequest(&proto.IndividualSimRequest{Player: request.Player, Buffs: request.Buffs})
}

func statsFromIndSimRequest(isr *proto.IndividualSimRequest) *proto.ComputeStatsResult {
	sim := createSim(isr)
	gearStats := sim.Raid.Parties[0].Players[0].GetCharacter().Equip.Stats()
	return &proto.ComputeStatsResult{
		GearOnly:   gearStats[:],
		FinalStats: sim.Raid.Parties[0].Players[0].GetCharacter().Stats[:], // createSim includes a call to buff up all party members.
		Sets:       []string{},
	}
}

func statWeightsImpl(request *proto.StatWeightsRequest) *proto.StatWeightsResult {
	statsToWeight := make([]stats.Stat, len(request.StatsToWeigh))
	for i, v := range request.StatsToWeigh {
		statsToWeight[i] = stats.Stat(v)
	}
	result := core.CalcStatWeight(convertSimParams(request.Options), statsToWeight, stats.Stat(request.EpReferenceStat))
	return &proto.StatWeightsResult{
		Weights:       result.Weights[:],
		WeightsStdev:  result.WeightsStdev[:],
		EpValues:      result.EpValues[:],
		EpValuesStdev: result.EpValuesStdev[:],
	}
}

func convertSimParams(request *proto.IndividualSimRequest) core.IndividualParams {
	options := core.Options{
		Iterations: int(request.Iterations),
		RSeed:      request.RandomSeed,
		ExitOnOOM:  request.ExitOnOom,
		GCDMin:     time.Duration(request.GcdMin),
		Debug:      request.Debug,
	}
	if request.Encounter != nil {
		options.Encounter = core.Encounter{
			Duration:   request.Encounter.Duration,
			NumTargets: int(request.Encounter.NumTargets),
			Armor:      request.Encounter.TargetArmor,
		}
	}

	params := core.IndividualParams{
		Equip:    convertEquip(request.Player.Equipment),
		Race:     core.RaceBonusType(request.Player.Options.Race),
		Consumes: convertConsumes(request.Player.Options.Consumes),
		Buffs:    convertBuffs(request.Buffs),
		Options:  options,

		PlayerOptions: request.Player.Options,
	}
	copy(params.CustomStats[:], request.Player.CustomStats[:])

	return params
}

func createSim(request *proto.IndividualSimRequest) *core.Simulation {
	params := convertSimParams(request)
	sim := core.NewIndividualSim(params)
	return sim
}

func runSimulationImpl(request *proto.IndividualSimRequest) *proto.IndividualSimResult {
	sim := createSim(request)
	result := sim.Run()

	castMetrics := map[int32]*proto.CastMetric{}
	for k, v := range result.Casts {
		castMetrics[k] = &proto.CastMetric{
			Casts:  v.Casts,
			Crits:  v.Crits,
			Misses: v.Misses,
			Dmgs:   v.Dmgs,
		}
	}
	isr := &proto.IndividualSimResult{
		DpsAvg:              result.DpsAvg,
		DpsStdev:            result.DpsStDev,
		DpsHist:             result.DpsHist,
		Logs:                result.Logs,
		DpsMax:              result.DpsMax,
		ExecutionDurationMs: result.ExecutionDurationMs,
		NumOom:              int32(result.NumOom),
		OomAtAvg:            result.OomAtAvg,
		DpsAtOomAvg:         result.DpsAtOomAvg,
		Casts:               castMetrics,
	}
	return isr
}

func convertConsumes(c *proto.Consumes) core.Consumes {
	cconsume := core.Consumes{
		FlaskOfBlindingLight:     c.FlaskOfBlindingLight,
		FlaskOfMightyRestoration: c.FlaskOfMightyRestoration,
		FlaskOfPureDeath:         c.FlaskOfPureDeath,
		FlaskOfSupremePower:      c.FlaskOfSupremePower,
		AdeptsElixir:             c.AdeptsElixir,
		ElixirOfMajorFirePower:   c.ElixirOfMajorFirePower,
		ElixirOfMajorFrostPower:  c.ElixirOfMajorFrostPower,
		ElixirOfMajorShadowPower: c.ElixirOfMajorShadowPower,
		ElixirOfDraenicWisdom:    c.ElixirOfDraenicWisdom,
		ElixirOfMajorMageblood:   c.ElixirOfMajorMageblood,
		BrilliantWizardOil:       c.BrilliantWizardOil,
		SuperiorWizardOil:        c.SuperiorWizardOil,
		BlackenedBasilisk:        c.BlackenedBasilisk,
		SkullfishSoup:            c.SkullfishSoup,
		DestructionPotion:        c.DestructionPotion,
		SuperManaPotion:          c.SuperManaPotion,
		DarkRune:                 c.DarkRune,
		DrumsOfBattle:            c.DrumsOfBattle,
		DrumsOfRestoration:       c.DrumsOfRestoration,
	}

	return cconsume
}

func convertEquip(es *proto.EquipmentSpec) items.EquipmentSpec {
	coreEquip := items.EquipmentSpec{}

	for i, item := range es.Items {
		spec := items.ItemSpec{
			ID: item.Id,
		}
		spec.Gems = item.Gems
		spec.Enchant = item.Enchant
		coreEquip[i] = spec
	}

	return coreEquip
}

func convertBuffs(inBuff *proto.Buffs) core.Buffs {
	// TODO: support tri-state better
	return core.Buffs{
		ArcaneBrilliance:          inBuff.ArcaneBrilliance,
		GiftOfTheWild:             inBuff.GiftOfTheWild,
		BlessingOfKings:           inBuff.BlessingOfKings,
		BlessingOfWisdom:          inBuff.BlessingOfWisdom,
		DivineSpirit:              inBuff.DivineSpirit,
		MoonkinAura:               inBuff.MoonkinAura,
		ShadowPriestDPS:           uint16(inBuff.ShadowPriestDps),

		JudgementOfWisdom:         inBuff.JudgementOfWisdom,
		ImprovedSealOfTheCrusader: inBuff.ImprovedSealOfTheCrusader,
		Misery:                    inBuff.Misery,

		ManaSpringTotem:           inBuff.ManaSpringTotem,
		ManaTideTotem:             inBuff.ManaTideTotem,
		TotemOfWrath:              inBuff.TotemOfWrath,
		WrathOfAirTotem:           inBuff.WrathOfAirTotem,

		AtieshMage:                inBuff.AtieshMage,
		AtieshWarlock:             inBuff.AtieshWarlock,
		BraidedEterniumChain:      inBuff.BraidedEterniumChain,
		ChainOfTheTwilightOwl:     inBuff.ChainOfTheTwilightOwl,
		EyeOfTheNight:             inBuff.EyeOfTheNight,
		JadePendantOfBlasting:     inBuff.JadePendantOfBlasting,
	}
}