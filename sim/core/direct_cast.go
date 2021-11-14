package core

import (
	"github.com/wowsims/tbc/sim/core/stats"
)

// A direct spell is one that does a single instance of damage once casting is
// complete, i.e. shadowbolt or fire blast.
// Note that some spell casts can have more than 1 DirectSpellEffect, e.g.
// Chain Lightning.
//
// This struct holds additional inputs beyond what a SpellEffect already contains,
// which are necessary for a direct spell damage calculation.
type DirectDamageSpellInput struct {
	MinBaseDamage float64
	MaxBaseDamage float64

	// Increase in damage per point of spell power.
	SpellCoefficient float64
}

func (spellEffect *SpellEffect) calculateDirectDamage(sim *Simulation, spellCast *SpellCast, ddInput *DirectDamageSpellInput) {
	baseDamage := ddInput.MinBaseDamage + sim.RandomFloat("DirectSpell Base Damage")*(ddInput.MaxBaseDamage-ddInput.MinBaseDamage)

	totalSpellPower := spellCast.Character.GetStat(stats.SpellPower) + spellCast.Character.GetStat(spellCast.SpellSchool) + spellEffect.BonusSpellPower
	damageFromSpellPower := (totalSpellPower * ddInput.SpellCoefficient)

	damage := baseDamage + damageFromSpellPower

	damage *= spellEffect.DamageMultiplier

	crit := (spellCast.Character.GetStat(stats.SpellCrit) + spellEffect.BonusSpellCritRating) / (SpellCritRatingPerCritChance * 100)
	if spellCast.GuaranteedCrit || sim.RandomFloat("DirectSpell Crit") < crit {
		spellEffect.Crit = true
		damage *= spellCast.CritMultiplier
	}

	// Average Resistance (AR) = (Target's Resistance / (Caster's Level * 5)) * 0.75
	// P(x) = 50% - 250%*|x - AR| <- where X is %resisted
	// Using these stats:
	//    13.6% chance of
	//  FUTURE: handle boss resists for fights/classes that are actually impacted by that.
	resVal := sim.RandomFloat("DirectSpell Resist")
	if resVal < 0.18 { // 13% chance for 25% resist, 4% for 50%, 1% for 75%
		if resVal < 0.01 {
			spellEffect.PartialResist_3_4 = true
			damage *= .25
		} else if resVal < 0.05 {
			spellEffect.PartialResist_2_4 = true
			damage *= .5
		} else {
			spellEffect.PartialResist_1_4 = true
			damage *= .75
		}
	}

	spellEffect.Damage = damage
}

type DirectDamageSpellEffect struct {
	SpellEffect
	DirectDamageSpellInput
}

func (ddEffect *DirectDamageSpellEffect) apply(sim *Simulation, spellCast *SpellCast) {
	ddEffect.SpellEffect.beforeCalculations(sim, spellCast)

	if ddEffect.Hit {
		ddEffect.SpellEffect.calculateDirectDamage(sim, spellCast, &ddEffect.DirectDamageSpellInput)
	}

	ddEffect.SpellEffect.afterCalculations(sim, spellCast)
}

type SingleTargetDirectDamageSpell struct {
	// Embedded spell cast.
	SpellCast

	// Individual direct damage effect of this spell.
	Effect DirectDamageSpellEffect
}

func (spell *SingleTargetDirectDamageSpell) Init(sim *Simulation) {
	spell.SpellCast.init(sim)
}

func (spell *SingleTargetDirectDamageSpell) Act(sim *Simulation) bool {
	return spell.startCasting(sim, func(sim *Simulation, cast *Cast) {
		spell.Effect.apply(sim, &spell.SpellCast)
		sim.MetricsAggregator.AddSpellEffects(&spell.SpellCast)
	})
}

type SingleTargetDirectDamageSpellTemplate struct {
	template SingleTargetDirectDamageSpell
}

func (template *SingleTargetDirectDamageSpellTemplate) Apply(newAction *SingleTargetDirectDamageSpell) {
	*newAction = template.template
}

// Takes in a cast template and returns a template, so you don't need to keep track of which things to allocate yourself.
func NewSingleTargetDirectDamageSpellTemplate(spellTemplate SingleTargetDirectDamageSpell) SingleTargetDirectDamageSpellTemplate {
	return SingleTargetDirectDamageSpellTemplate{
		template: spellTemplate,
	}
}

type MultiTargetDirectDamageSpell struct {
	// Embedded spell cast.
	SpellCast

	// Individual direct damage effects of this spell.
	// For most spells this will only have 1 element, but for multi-damage spells
	// like Arcane Explosion of Chain Lightning this will have multiple elements.
	Effects []DirectDamageSpellEffect
}

func (spell *MultiTargetDirectDamageSpell) Init(sim *Simulation) {
	spell.SpellCast.init(sim)
}

func (spell *MultiTargetDirectDamageSpell) Act(sim *Simulation) bool {
	return spell.startCasting(sim, func(sim *Simulation, cast *Cast) {
		for effectIdx := range spell.Effects {
			effect := &spell.Effects[effectIdx]
			effect.apply(sim, &spell.SpellCast)
		}

		sim.MetricsAggregator.AddSpellEffects(&spell.SpellCast)
	})
}

type MultiTargetDirectDamageSpellTemplate struct {
	template MultiTargetDirectDamageSpell
	effects  []DirectDamageSpellEffect
}

func (template *MultiTargetDirectDamageSpellTemplate) Apply(newAction *MultiTargetDirectDamageSpell) {
	*newAction = template.template
	newAction.Effects = template.effects
	copy(newAction.Effects, template.template.Effects)
}

// Takes in a cast template and returns a template, so you don't need to keep track of which things to allocate yourself.
func NewMultiTargetDirectDamageSpellTemplate(spellTemplate MultiTargetDirectDamageSpell) MultiTargetDirectDamageSpellTemplate {
	return MultiTargetDirectDamageSpellTemplate{
		template: spellTemplate,
		effects: make([]DirectDamageSpellEffect, len(spellTemplate.Effects)),
	}
}