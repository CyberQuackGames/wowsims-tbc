package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/wowsims/tbc/sim/core/stats"
)

// Callback for after a spell hits the target, before damage has been calculated.
// Use it to modify the spell damage or results.
type OnBeforeSpellHit func(sim *Simulation, spellCast *SpellCast, spellEffect *SpellHitEffect)

// Callback for after a spell hits the target and after damage is calculated. Use it for proc effects
// or anything that comes from the final result of the spell.
type OnSpellHit func(sim *Simulation, spellCast *SpellCast, spellEffect *SpellEffect)

// OnBeforePeriodicDamage is called when dots tick, before damage is calculated.
// Use it to modify the spell damage or results.
type OnBeforePeriodicDamage func(sim *Simulation, spellCast *SpellCast, spellEffect *SpellEffect, tickDamage *float64)

// OnPeriodicDamage is called when dots tick, after damage is calculated. Use it for proc effects
// or anything that comes from the final result of a tick.
type OnPeriodicDamage func(sim *Simulation, spellCast *SpellCast, spellEffect *SpellEffect, tickDamage float64)

// A Spell is a type of cast that can hit/miss using spell stats, and has a spell school.
type SpellCast struct {
	// Embedded Cast
	Cast

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

	// Bonus stats to be added to the spell.
	BonusSpellHitRating  float64
	BonusSpellPower      float64
	BonusSpellCritRating float64

	BonusHitRating        float64
	BonusAttackPower      float64
	BonusCritRating       float64
	BonusExpertiseRating  float64
	BonusArmorPenetration float64
	BonusWeaponDamage     float64

	// Additional multiplier that is always applied.
	DamageMultiplier float64

	// applies fixed % increases to damage at cast time.
	//  Only use multipliers that don't change for the lifetime of the sim.
	//  This should probably only be mutated in a template and not changed in auras.
	StaticDamageMultiplier float64

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

func (she *SpellHitEffect) beforeCalculations(sim *Simulation, spell *SimpleSpell) {
	she.SpellEffect.beforeCalculations(sim, spell, she)
}

func (spellEffect *SpellEffect) beforeCalculations(sim *Simulation, spell *SimpleSpell, she *SpellHitEffect) {
	spellEffect.BeyondAOECapMultiplier = 1
	multiplierBeforeTargetEffects := spellEffect.DamageMultiplier

	spell.Character.OnBeforeSpellHit(sim, &spell.SpellCast, she)
	spellEffect.Target.OnBeforeSpellHit(sim, &spell.SpellCast, she)

	spellEffect.BeyondAOECapMultiplier *= spellEffect.DamageMultiplier / multiplierBeforeTargetEffects

	if spell.OutcomeRollCategory == OutcomeRollCategoryNone || spell.SpellExtras.Matches(SpellExtrasAlwaysHits) {
		spellEffect.Outcome = OutcomeHit
	} else if spellEffect.ReuseMainHitRoll { // TODO: can we remove this.
		spellEffect.Outcome = spell.Effects[0].Outcome
	} else if spell.OutcomeRollCategory.Matches(OutcomeRollCategoryMagic) {
		if spellEffect.hitCheck(sim, &spell.SpellCast) {
			spellEffect.Outcome = OutcomeHit
		} else {
			spellEffect.Outcome = OutcomeMiss
		}
	} else if spell.OutcomeRollCategory.Matches(OutcomeRollCategoryPhysical) {
		spellEffect.Outcome = spellEffect.WhiteHitTableResult(sim, spell)
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
		spell.Character.Log(sim, "%s %s.", spell.ActionID, spellEffect)
	}
	if spellEffect.Landed() && spellEffect.FlatThreatBonus > 0 {
		spell.TotalThreat += spellEffect.FlatThreatBonus * spellEffect.TotalThreatMultiplier(&spell.SpellCast)
	}

	spellEffect.triggerSpellProcs(sim, spell)
}

// Calculates a hit check using the stats from this spell.
func (spellEffect *SpellEffect) hitCheck(sim *Simulation, spellCast *SpellCast) bool {
	hit := 0.83 + (spellCast.Character.GetStat(stats.SpellHit)+spellEffect.BonusSpellHitRating)/(SpellHitRatingPerHitChance*100)
	hit = MinFloat(hit, 0.99) // can't get away from the 1% miss

	return sim.RandomFloat("SpellCast Hit") < hit
}

// Calculates a crit check using the stats from this spell.
func (spellEffect *SpellEffect) critCheck(sim *Simulation, spellCast *SpellCast) bool {
	switch spellCast.CritRollCategory {
	case CritRollCategoryMagical:
		critChance := (spellCast.Character.GetStat(stats.SpellCrit) + spellCast.BonusCritRating + spellEffect.BonusSpellCritRating) / (SpellCritRatingPerCritChance * 100)
		return sim.RandomFloat("DirectSpell Crit") < critChance
	case CritRollCategoryPhysical:
		critChance := (spellCast.Character.GetStat(stats.MeleeCrit)+spellCast.BonusCritRating+spellEffect.BonusCritRating)/(MeleeCritRatingPerCritChance*100) - spellEffect.Target.CritSuppression
		return sim.RandomFloat("weapon swing") < critChance
	default:
		return false
	}
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
	spellCast.TotalThreat += spellEffect.Damage * spellEffect.TotalThreatMultiplier(spellCast)
}

// Only applies the results from the ticks, not the initial dot application.
func (hitEffect *SpellHitEffect) applyDotTickResultsToCast(spellCast *SpellCast) {
	if hitEffect.DotInput.TicksCanMissAndCrit {
		if hitEffect.Landed() {
			spellCast.Hits++
			if hitEffect.Outcome.Matches(OutcomeCrit) {
				spellCast.Crits++
			}

			if hitEffect.Outcome.Matches(OutcomePartial1_4) {
				spellCast.PartialResists_1_4++
			} else if hitEffect.Outcome.Matches(OutcomePartial2_4) {
				spellCast.PartialResists_2_4++
			} else if hitEffect.Outcome.Matches(OutcomePartial3_4) {
				spellCast.PartialResists_3_4++
			}
		} else {
			spellCast.Misses++
		}
	}

	spellCast.TotalDamage += hitEffect.Damage
	spellCast.TotalThreat += hitEffect.Damage * hitEffect.TotalThreatMultiplier(spellCast)
}

func (hitEffect *SpellHitEffect) calculateDirectDamage(sim *Simulation, spellCast *SpellCast) {
	character := spellCast.Character

	baseDamage := hitEffect.DirectInput.MinBaseDamage + sim.RandomFloat("DirectSpell Base Damage")*(hitEffect.DirectInput.MaxBaseDamage-hitEffect.DirectInput.MinBaseDamage)

	schoolBonus := 0.0
	// Use outcome roll to decide if it should use AP or spell school for bonus damage.
	isPhysical := spellCast.OutcomeRollCategory.Matches(OutcomeRollCategoryPhysical)
	if isPhysical {
		if spellCast.OutcomeRollCategory.Matches(OutcomeRollCategoryRanged) {
			schoolBonus = character.stats[stats.RangedAttackPower]
		} else if spellCast.SpellSchool == SpellSchoolPhysical {
			schoolBonus = character.stats[stats.AttackPower]
		}
		schoolBonus += hitEffect.BonusAttackPower
	} else {
		schoolBonus = character.GetStat(stats.SpellPower) + character.GetStat(spellCast.SpellSchool.Stat()) + hitEffect.SpellEffect.BonusSpellPower
	}
	damage := baseDamage + (schoolBonus * hitEffect.DirectInput.SpellCoefficient) + hitEffect.DirectInput.FlatDamageBonus
	damage *= hitEffect.SpellEffect.DamageMultiplier * hitEffect.SpellEffect.StaticDamageMultiplier

	// Use spell school to determine damage reduction type.
	if spellCast.SpellSchool.Matches(SpellSchoolPhysical) {
		if !spellCast.SpellExtras.Matches(SpellExtrasIgnoreResists) {
			damage *= 1 - hitEffect.Target.ArmorDamageReduction(character.stats[stats.ArmorPenetration]+hitEffect.BonusArmorPenetration)
		}
	} else if !spellCast.SpellExtras.Matches(SpellExtrasBinary | SpellExtrasIgnoreResists) {
		damage = calculateResists(sim, damage, &hitEffect.SpellEffect)
	}

	if hitEffect.SpellEffect.critCheck(sim, spellCast) {
		hitEffect.Outcome |= OutcomeCrit
		damage *= spellCast.CritMultiplier
		hitEffect.SpellEffect.BeyondAOECapMultiplier *= spellCast.CritMultiplier
	}

	hitEffect.SpellEffect.Damage = damage
}

// Snapshots a few values at the start of a dot.
func (hitEffect *SpellHitEffect) takeDotSnapshot(sim *Simulation, spellCast *SpellCast) {
	totalSpellPower := spellCast.Character.GetStat(stats.SpellPower) + spellCast.Character.GetStat(spellCast.SpellSchool.Stat()) + hitEffect.BonusSpellPower

	// snapshot total damage per tick, including any static damage multipliers
	hitEffect.DotInput.startTime = sim.CurrentTime
	hitEffect.DotInput.finalTickTime = sim.CurrentTime + time.Duration(hitEffect.DotInput.NumberOfTicks)*hitEffect.DotInput.TickLength
	hitEffect.DotInput.damagePerTick = (hitEffect.DotInput.TickBaseDamage + totalSpellPower*hitEffect.DotInput.TickSpellCoefficient) * hitEffect.StaticDamageMultiplier
	hitEffect.SpellEffect.BeyondAOECapMultiplier = 1
}

func (hitEffect *SpellHitEffect) calculateDotDamage(sim *Simulation, spellCast *SpellCast) {
	// fmt.Printf("DOT (%s) Ticking, Time Remaining: %0.2f\n", spellCast.Name, hitEffect.DotInput.TimeRemaining(sim).Seconds())
	damage := hitEffect.DotInput.damagePerTick

	spellCast.Character.OnBeforePeriodicDamage(sim, spellCast, &hitEffect.SpellEffect, &damage)

	damageBeforeTargetEffects := damage
	hitEffect.Target.OnBeforePeriodicDamage(sim, spellCast, &hitEffect.SpellEffect, &damage)
	hitEffect.SpellEffect.BeyondAOECapMultiplier *= damage / damageBeforeTargetEffects

	if hitEffect.DotInput.OnBeforePeriodicDamage != nil {
		hitEffect.DotInput.OnBeforePeriodicDamage(sim, spellCast, &hitEffect.SpellEffect, &damage)
	}
	if hitEffect.DotInput.IgnoreDamageModifiers {
		damage = hitEffect.DotInput.damagePerTick
	}

	hitEffect.Outcome = OutcomeEmpty
	if !hitEffect.DotInput.TicksCanMissAndCrit || hitEffect.hitCheck(sim, spellCast) {
		hitEffect.Outcome = OutcomeHit
	} else {
		hitEffect.Outcome = OutcomeMiss
	}

	if hitEffect.Outcome == OutcomeHit {
		if !spellCast.SpellExtras.Matches(SpellExtrasBinary | SpellExtrasIgnoreResists) {
			damage = calculateResists(sim, damage, &hitEffect.SpellEffect)
		}

		if hitEffect.DotInput.TicksCanMissAndCrit && hitEffect.critCheck(sim, spellCast) {
			hitEffect.Outcome |= OutcomeCrit
			damage *= spellCast.CritMultiplier
			hitEffect.SpellEffect.BeyondAOECapMultiplier *= spellCast.CritMultiplier
		}
	} else {
		damage = 0
	}

	hitEffect.SpellEffect.Damage = damage
}

// This should be called on each dot tick.
func (hitEffect *SpellHitEffect) afterDotTick(sim *Simulation, spell *SimpleSpell) {
	if sim.Log != nil {
		spell.Character.Log(sim, "%s %s.", spell.ActionID, hitEffect.SpellEffect.DotResultString())
	}

	hitEffect.applyDotTickResultsToCast(&spell.SpellCast)

	if hitEffect.DotInput.TicksProcSpellHitEffects {
		hitEffect.SpellEffect.triggerSpellProcs(sim, spell)
	}

	spell.Character.OnPeriodicDamage(sim, &spell.SpellCast, &hitEffect.SpellEffect, hitEffect.Damage)
	hitEffect.Target.OnPeriodicDamage(sim, &spell.SpellCast, &hitEffect.SpellEffect, hitEffect.Damage)
	if hitEffect.DotInput.OnPeriodicDamage != nil {
		hitEffect.DotInput.OnPeriodicDamage(sim, &spell.SpellCast, &hitEffect.SpellEffect, hitEffect.Damage)
	}

	hitEffect.DotInput.tickIndex++
}

// This should be called after the final tick of the dot, or when the dot is cancelled.
func (hitEffect *SpellHitEffect) onDotComplete(sim *Simulation, spellCast *SpellCast) {
	// Clean up the dot object.
	hitEffect.DotInput.finalTickTime = 0

	if hitEffect.DotInput.DebuffID != 0 {
		hitEffect.Target.AddAuraUptime(hitEffect.DotInput.DebuffID, spellCast.ActionID, sim.CurrentTime-hitEffect.DotInput.startTime)
	}
}

func (spellEffect *SpellEffect) String() string {
	outcomeStr := spellEffect.Outcome.String()
	if !spellEffect.Landed() {
		return outcomeStr
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s for %0.3f damage", outcomeStr, spellEffect.Damage)
	return sb.String()
}

func (spellEffect *SpellEffect) DotResultString() string {
	return "tick " + spellEffect.String()
}

// Return value is (newDamage, resistMultiplier)
func calculateResists(sim *Simulation, damage float64, spellEffect *SpellEffect) float64 {
	// Average Resistance (AR) = (Target's Resistance / (Caster's Level * 5)) * 0.75
	// P(x) = 50% - 250%*|x - AR| <- where X is %resisted
	// Using these stats:
	//    13.6% chance of
	//  FUTURE: handle boss resists for fights/classes that are actually impacted by that.
	resVal := sim.RandomFloat("DirectSpell Resist")
	if resVal > 0.18 { // 13% chance for 25% resist, 4% for 50%, 1% for 75%
		// No partial resist.
		return damage
	}

	var multiplier float64
	if resVal < 0.01 {
		spellEffect.Outcome |= OutcomePartial3_4
		multiplier = 0.25
	} else if resVal < 0.05 {
		spellEffect.Outcome |= OutcomePartial2_4
		multiplier = 0.5
	} else {
		spellEffect.Outcome |= OutcomePartial1_4
		multiplier = 0.75
	}

	return damage * multiplier
}
