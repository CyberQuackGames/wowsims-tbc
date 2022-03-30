package core

import (
	"fmt"
	"math"

	"github.com/wowsims/tbc/sim/core/stats"
)

// Callback for after a spell hits the target and after damage is calculated. Use it for proc effects
// or anything that comes from the final result of the spell.
type OnSpellHit func(sim *Simulation, spellCast *SpellCast, spellEffect *SpellEffect)

// OnPeriodicDamage is called when dots tick, after damage is calculated. Use it for proc effects
// or anything that comes from the final result of a tick.
type OnPeriodicDamage func(sim *Simulation, spellCast *SpellCast, spellEffect *SpellEffect, tickDamage float64)

// A Spell is a type of cast that can hit/miss using spell stats, and has a spell school.
type SpellCast struct {
	// Embedded Cast
	Cast

	// Whether this is a phantom cast. Phantom casts are usually casts triggered by some effect,
	// like The Lightning Capacitor or Shaman Flametongue Weapon. Many on-hit effects do not
	// proc from phantom casts, only regular casts.
	IsPhantom bool

	OutcomeRollCategory OutcomeRollCategory
	CritRollCategory    CritRollCategory

	// How much to multiply damage by, if this cast crits.
	CritMultiplier float64

	// Results from the spell cast. Spell casts can have multiple effects (e.g.
	// Chain Lightning, Moonfire) so these are totals from all the effects.
	Hits               int32
	Misses             int32
	Crits              int32
	PartialResists_1_4 int32   // 1/4 of the spell was resisted
	PartialResists_2_4 int32   // 2/4 of the spell was resisted
	PartialResists_3_4 int32   // 3/4 of the spell was resisted
	TotalDamage        float64 // Damage done by this cast.
	TotalThreat        float64 // Threat generated by this cast.

	// Melee only stats
	Dodges  int32
	Glances int32
	Parries int32
	Blocks  int32
}

type SpellEffect struct {
	// Target of the spell.
	Target *Target

	BaseDamage BaseDamageConfig
	DotInput   DotDamageInput

	// Bonus stats to be added to the spell.
	BonusSpellHitRating  float64
	BonusSpellPower      float64
	BonusSpellCritRating float64

	BonusAttackPower float64
	BonusCritRating  float64

	// Additional multiplier that is always applied.
	DamageMultiplier float64

	// Multiplier for all threat generated by this effect.
	ThreatMultiplier float64

	// Adds a fixed amount of threat to this spell, before multipliers.
	FlatThreatBonus float64

	// Controls which effects can proc from this effect.
	ProcMask ProcMask

	// Causes the first roll for this hit to be copied from ActiveMeleeAbility.Effects[0].HitType.
	// This is only used by Shaman Stormstrike.
	ReuseMainHitRoll bool

	// Callbacks for providing additional custom behavior.
	OnSpellHit OnSpellHit

	// Results
	Outcome HitOutcome
	Damage  float64 // Damage done by this cast.
	Threat  float64

	// Certain damage multiplier, such as target debuffs and crit multipliers, do
	// not count towards the AOE cap. Store them here to they can be subtracted
	// later when calculating AOE cap.
	BeyondAOECapMultiplier float64
}

func (spellEffect *SpellEffect) Landed() bool {
	return spellEffect.Outcome.Matches(OutcomeLanded)
}

func (spellEffect *SpellEffect) TotalThreatMultiplier(spellCast *SpellCast) float64 {
	return spellEffect.ThreatMultiplier * spellCast.Character.PseudoStats.ThreatMultiplier
}

func (spellEffect *SpellEffect) calcThreat(spellCast *SpellCast) float64 {
	if spellEffect.Landed() {
		return (spellEffect.Damage + spellEffect.FlatThreatBonus) * spellEffect.TotalThreatMultiplier(spellCast)
	} else {
		return 0
	}
}

func (spellEffect *SpellEffect) MeleeAttackPower(spellCast *SpellCast) float64 {
	return spellCast.Character.stats[stats.AttackPower] + spellCast.Character.PseudoStats.MobTypeAttackPower + spellEffect.BonusAttackPower
}

func (spellEffect *SpellEffect) MeleeAttackPowerOnTarget() float64 {
	return spellEffect.Target.PseudoStats.BonusMeleeAttackPower
}

func (spellEffect *SpellEffect) RangedAttackPower(spellCast *SpellCast) float64 {
	return spellCast.Character.stats[stats.RangedAttackPower] + spellCast.Character.PseudoStats.MobTypeAttackPower + spellEffect.BonusAttackPower
}

func (spellEffect *SpellEffect) RangedAttackPowerOnTarget() float64 {
	return spellEffect.Target.PseudoStats.BonusRangedAttackPower
}

func (spellEffect *SpellEffect) BonusWeaponDamage(spellCast *SpellCast) float64 {
	return spellCast.Character.PseudoStats.BonusDamage
}

func (spellEffect *SpellEffect) PhysicalHitChance(character *Character, spellCast *SpellCast) float64 {
	hitRating := character.stats[stats.MeleeHit] + spellEffect.Target.PseudoStats.BonusMeleeHitRating

	if spellCast.OutcomeRollCategory.Matches(OutcomeRollCategoryRanged) {
		hitRating += character.PseudoStats.BonusRangedHitRating
	}

	return (hitRating / (MeleeHitRatingPerHitChance * 100)) - spellEffect.Target.HitSuppression
}

func (spellEffect *SpellEffect) PhysicalCritChance(character *Character, spellCast *SpellCast) float64 {
	critRating := character.stats[stats.MeleeCrit] + spellEffect.BonusCritRating + spellEffect.Target.PseudoStats.BonusCritRating

	if spellCast.OutcomeRollCategory.Matches(OutcomeRollCategoryRanged) {
		critRating += character.PseudoStats.BonusRangedCritRating
	} else {
		critRating += character.PseudoStats.BonusMeleeCritRating
	}
	if spellCast.SpellExtras.Matches(SpellExtrasAgentReserved1) {
		critRating += character.PseudoStats.BonusCritRatingAgentReserved1
	}
	if spellEffect.ProcMask.Matches(ProcMaskMeleeMH) {
		spellEffect.BonusCritRating += character.PseudoStats.BonusMHCritRating
	} else if spellEffect.ProcMask.Matches(ProcMaskMeleeOH) {
		spellEffect.BonusCritRating += character.PseudoStats.BonusOHCritRating
	}

	return (critRating / (MeleeCritRatingPerCritChance * 100)) - spellEffect.Target.CritSuppression
}

func (spellEffect *SpellEffect) SpellPower(character *Character, spellCast *SpellCast) float64 {
	return character.GetStat(stats.SpellPower) + character.GetStat(spellCast.SpellSchool.Stat()) + character.PseudoStats.MobTypeSpellPower + spellEffect.BonusSpellPower
}

func (spellEffect *SpellEffect) SpellCritChance(character *Character, spellCast *SpellCast) float64 {
	critRating := (character.GetStat(stats.SpellCrit) + spellEffect.BonusSpellCritRating + spellEffect.Target.PseudoStats.BonusCritRating)
	if spellCast.SpellSchool.Matches(SpellSchoolFire) {
		critRating += character.PseudoStats.BonusFireCritRating
	} else if spellCast.SpellSchool.Matches(SpellSchoolFrost) {
		critRating += spellEffect.Target.PseudoStats.BonusFrostCritRating
	}
	return critRating / (SpellCritRatingPerCritChance * 100)
}

func (hitEffect *SpellEffect) directCalculations(sim *Simulation, spell *SimpleSpell) {
	damage := hitEffect.calculateBaseDamage(sim, &spell.SpellCast)

	damage *= hitEffect.DamageMultiplier
	hitEffect.applyAttackerModifiers(sim, &spell.SpellCast, false, &damage)
	hitEffect.applyTargetModifiers(sim, &spell.SpellCast, false, hitEffect.BaseDamage.TargetSpellCoefficient, &damage)
	hitEffect.applyResistances(sim, &spell.SpellCast, &damage)
	hitEffect.applyOutcome(sim, &spell.SpellCast, &damage)

	hitEffect.Damage = damage
}

func (hitEffect *SpellEffect) calculateBaseDamage(sim *Simulation, spellCast *SpellCast) float64 {
	if hitEffect.BaseDamage.Calculator == nil {
		return 0
	} else {
		return hitEffect.BaseDamage.Calculator(sim, hitEffect, spellCast)
	}
}

func (spellEffect *SpellEffect) determineOutcome(sim *Simulation, spell *SimpleSpell) {
	if spell.OutcomeRollCategory == OutcomeRollCategoryNone || spell.SpellExtras.Matches(SpellExtrasAlwaysHits) {
		spellEffect.Outcome = OutcomeHit
		if spellEffect.critCheck(sim, &spell.SpellCast) {
			spellEffect.Outcome = OutcomeCrit
		}
	} else if spellEffect.ReuseMainHitRoll { // TODO: can we remove this.
		spellEffect.Outcome = spell.Effects[0].Outcome
	} else if spell.OutcomeRollCategory.Matches(OutcomeRollCategoryMagic) {
		if spellEffect.hitCheck(sim, &spell.SpellCast) {
			spellEffect.Outcome = OutcomeHit
			if spellEffect.critCheck(sim, &spell.SpellCast) {
				spellEffect.Outcome = OutcomeCrit
			}
		} else {
			spellEffect.Outcome = OutcomeMiss
		}
	} else if spell.OutcomeRollCategory.Matches(OutcomeRollCategoryPhysical) {
		spellEffect.Outcome = spellEffect.WhiteHitTableResult(sim, spell)
		if spellEffect.Landed() && spellEffect.critCheck(sim, &spell.SpellCast) {
			spellEffect.Outcome = OutcomeCrit
		}
	}
}

// Computes an attack result using the white-hit table formula (single roll).
func (ahe *SpellEffect) WhiteHitTableResult(sim *Simulation, ability *SimpleSpell) HitOutcome {
	character := ability.Character

	roll := sim.RandomFloat("White Hit Table")

	// Miss
	missChance := ahe.Target.MissChance - ahe.PhysicalHitChance(character, &ability.SpellCast)
	if character.AutoAttacks.IsDualWielding && ability.OutcomeRollCategory == OutcomeRollCategoryWhite {
		missChance += 0.19
	}
	missChance = MaxFloat(0, missChance)

	chance := missChance
	if roll < chance {
		return OutcomeMiss
	}

	if !ability.OutcomeRollCategory.Matches(OutcomeRollCategoryRanged) { // Ranged hits can't be dodged/glance, and are always 2-roll
		// Dodge
		if !ability.SpellExtras.Matches(SpellExtrasCannotBeDodged) {
			dodge := ahe.Target.Dodge

			expertiseRating := character.stats[stats.Expertise]
			if ahe.ProcMask.Matches(ProcMaskMeleeMH) {
				expertiseRating += character.PseudoStats.BonusMHExpertiseRating
			} else if ahe.ProcMask.Matches(ProcMaskMeleeOH) {
				expertiseRating += character.PseudoStats.BonusOHExpertiseRating
			}
			expertisePercentage := MinFloat(math.Floor(expertiseRating/ExpertisePerQuarterPercentReduction)/400, dodge)

			chance += dodge - expertisePercentage
			if roll < chance {
				return OutcomeDodge
			}
		}

		// Parry (if in front)
		// If the target is a mob and defense minus weapon skill is 11 or more:
		// ParryChance = 5% + (TargetLevel*5 - AttackerSkill) * 0.6%

		// If the target is a mob and defense minus weapon skill is 10 or less:
		// ParryChance = 5% + (TargetLevel*5 - AttackerSkill) * 0.1%

		// Block (if in front)
		// If the target is a mob:
		// BlockChance = MIN(5%, 5% + (TargetLevel*5 - AttackerSkill) * 0.1%)
		// If we actually implement blocks, ranged hits can be blocked.

		// No need to crit/glance roll if we are not a white hit
		if ability.OutcomeRollCategory.Matches(OutcomeRollCategorySpecial | OutcomeRollCategoryRanged) {
			return OutcomeHit
		}

		// Glance
		chance += ahe.Target.Glance
		if roll < chance {
			return OutcomeGlance
		}

		// Crit
		chance += ahe.PhysicalCritChance(character, &ability.SpellCast)
		if roll < chance {
			return OutcomeCrit
		}
	}

	return OutcomeHit
}

// Calculates a hit check using the stats from this spell.
func (spellEffect *SpellEffect) hitCheck(sim *Simulation, spellCast *SpellCast) bool {
	hit := 0.83 + (spellCast.Character.GetStat(stats.SpellHit)+spellEffect.BonusSpellHitRating)/(SpellHitRatingPerHitChance*100)
	hit = MinFloat(hit, 0.99) // can't get away from the 1% miss

	return sim.RandomFloat("Magical Hit Roll") < hit
}

// Calculates a crit check using the stats from this spell.
func (spellEffect *SpellEffect) critCheck(sim *Simulation, spellCast *SpellCast) bool {
	switch spellCast.CritRollCategory {
	case CritRollCategoryMagical:
		critChance := spellEffect.SpellCritChance(spellCast.Character, spellCast)
		return sim.RandomFloat("Magical Crit Roll") < critChance
	case CritRollCategoryPhysical:
		return sim.RandomFloat("Physical Crit Roll") < spellEffect.PhysicalCritChance(spellCast.Character, spellCast)
	default:
		return false
	}
}

func (spellEffect *SpellEffect) triggerSpellProcs(sim *Simulation, spell *SimpleSpell) {
	if spellEffect.OnSpellHit != nil {
		spellEffect.OnSpellHit(sim, &spell.SpellCast, spellEffect)
	}
	spell.Character.OnSpellHit(sim, &spell.SpellCast, spellEffect)
	spellEffect.Target.OnSpellHit(sim, &spell.SpellCast, spellEffect)
}

func (spellEffect *SpellEffect) afterCalculations(sim *Simulation, spell *SimpleSpell) {
	if sim.Log != nil && !spell.SpellExtras.Matches(SpellExtrasAlwaysHits) {
		spell.Character.Log(sim, "%s %s. (Threat: %0.3f)", spell.ActionID, spellEffect, spellEffect.calcThreat(&spell.SpellCast))
	}

	spellEffect.triggerSpellProcs(sim, spell)
}

func (spellEffect *SpellEffect) applyResultsToCast(spellCast *SpellCast) {
	if spellEffect.Outcome.Matches(OutcomeHit) {
		spellCast.Hits++
	}
	if spellEffect.Outcome.Matches(OutcomeGlance) {
		spellCast.Glances++
	}
	if spellEffect.Outcome.Matches(OutcomeCrit) {
		spellCast.Crits++
	}
	if spellEffect.Outcome.Matches(OutcomeBlock) {
		spellCast.Blocks++
	}

	if spellEffect.Landed() {
		if spellEffect.Outcome.Matches(OutcomePartial1_4) {
			spellCast.PartialResists_1_4++
		} else if spellEffect.Outcome.Matches(OutcomePartial2_4) {
			spellCast.PartialResists_2_4++
		} else if spellEffect.Outcome.Matches(OutcomePartial3_4) {
			spellCast.PartialResists_3_4++
		}
	} else {
		if spellEffect.Outcome == OutcomeMiss {
			spellCast.Misses++
		} else if spellEffect.Outcome == OutcomeDodge {
			spellCast.Dodges++
		} else if spellEffect.Outcome == OutcomeParry {
			spellCast.Parries++
		}
	}

	spellCast.TotalDamage += spellEffect.Damage
	spellCast.TotalThreat += spellEffect.calcThreat(spellCast)
}

func (spellEffect *SpellEffect) String() string {
	outcomeStr := spellEffect.Outcome.String()
	if !spellEffect.Landed() {
		return outcomeStr
	}
	return fmt.Sprintf("%s for %0.3f damage", outcomeStr, spellEffect.Damage)
}

func (hitEffect *SpellEffect) applyAttackerModifiers(sim *Simulation, spellCast *SpellCast, isPeriodic bool, damage *float64) {
	attacker := spellCast.Character

	if spellCast.OutcomeRollCategory.Matches(OutcomeRollCategoryRanged) {
		*damage *= attacker.PseudoStats.RangedDamageDealtMultiplier
	}
	if spellCast.SpellExtras.Matches(SpellExtrasAgentReserved1) {
		*damage *= attacker.PseudoStats.AgentReserved1DamageDealtMultiplier
	}

	*damage *= attacker.PseudoStats.DamageDealtMultiplier
	if spellCast.SpellSchool.Matches(SpellSchoolPhysical) {
		*damage *= attacker.PseudoStats.PhysicalDamageDealtMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolArcane) {
		*damage *= attacker.PseudoStats.ArcaneDamageDealtMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolFire) {
		*damage *= attacker.PseudoStats.FireDamageDealtMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolFrost) {
		*damage *= attacker.PseudoStats.FrostDamageDealtMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolHoly) {
		*damage *= attacker.PseudoStats.HolyDamageDealtMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolNature) {
		*damage *= attacker.PseudoStats.NatureDamageDealtMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolShadow) {
		*damage *= attacker.PseudoStats.ShadowDamageDealtMultiplier
	}
}

func (hitEffect *SpellEffect) applyTargetModifiers(sim *Simulation, spellCast *SpellCast, isPeriodic bool, targetCoeff float64, damage *float64) {
	target := hitEffect.Target

	*damage *= target.PseudoStats.DamageTakenMultiplier
	if spellCast.SpellSchool.Matches(SpellSchoolPhysical) {
		if targetCoeff > 0 {
			*damage += target.PseudoStats.BonusPhysicalDamageTaken
		}
		*damage *= target.PseudoStats.PhysicalDamageTakenMultiplier
		if isPeriodic {
			*damage *= target.PseudoStats.PeriodicPhysicalDamageTakenMultiplier
		}
	} else if spellCast.SpellSchool.Matches(SpellSchoolArcane) {
		*damage *= target.PseudoStats.ArcaneDamageTakenMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolFire) {
		*damage *= target.PseudoStats.FireDamageTakenMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolFrost) {
		*damage *= target.PseudoStats.FrostDamageTakenMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolHoly) {
		*damage += target.PseudoStats.BonusHolyDamageTaken * targetCoeff
		*damage *= target.PseudoStats.HolyDamageTakenMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolNature) {
		*damage *= target.PseudoStats.NatureDamageTakenMultiplier
	} else if spellCast.SpellSchool.Matches(SpellSchoolShadow) {
		*damage *= target.PseudoStats.ShadowDamageTakenMultiplier
	}
}

// Modifies damage based on Armor or Magic resistances, depending on the damage type.
func (hitEffect *SpellEffect) applyResistances(sim *Simulation, spellCast *SpellCast, damage *float64) {
	if spellCast.SpellExtras.Matches(SpellExtrasIgnoreResists) {
		return
	}

	if spellCast.SpellSchool.Matches(SpellSchoolPhysical) {
		// Physical resistance (armor).
		*damage *= 1 - hitEffect.Target.ArmorDamageReduction(spellCast.Character.stats[stats.ArmorPenetration])
	} else if !spellCast.SpellExtras.Matches(SpellExtrasBinary) {
		// Magical resistance.
		// https://royalgiraffe.github.io/resist-guide

		resistanceRoll := sim.RandomFloat("Partial Resist")
		if resistanceRoll > 0.18 { // 13% chance for 25% resist, 4% for 50%, 1% for 75%
			// No partial resist.
		} else if resistanceRoll > 0.05 {
			hitEffect.Outcome |= OutcomePartial1_4
			*damage *= 0.75
		} else if resistanceRoll > 0.01 {
			hitEffect.Outcome |= OutcomePartial2_4
			*damage *= 0.5
		} else {
			hitEffect.Outcome |= OutcomePartial3_4
			*damage *= 0.25
		}
	}
}

func (hitEffect *SpellEffect) applyOutcome(sim *Simulation, spellCast *SpellCast, damage *float64) {
	if !hitEffect.Landed() {
		*damage = 0
	} else if hitEffect.Outcome.Matches(OutcomeCrit) {
		*damage *= spellCast.CritMultiplier
	} else if hitEffect.Outcome == OutcomeGlance {
		// TODO glancing blow damage reduction is actually a range ([65%, 85%] vs. 73)
		*damage *= 0.75
	}
}
