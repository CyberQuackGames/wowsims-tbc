package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// ActiveMeleeAbility doesn't support cast times, so we wrap it in a SimpleCast.
func (hunter *Hunter) newSteadyShotCastTemplate(sim *core.Simulation) core.SimpleCast {
	template := core.SimpleCast{
		Cast: core.Cast{
			ActionID:     core.ActionID{SpellID: 34120},
			Character:    hunter.GetCharacter(),
			BaseManaCost: 110,
			ManaCost:     110,
			CastTime:     time.Second * 1,
			GCD:          core.GCDDefault,
		},
		DisableMetrics: true,
	}

	template.ManaCost *= 1 - 0.02*float64(hunter.Talents.Efficiency)

	return template
}

func (hunter *Hunter) newSteadyShotAbilityTemplate(sim *core.Simulation) core.MeleeAbilityTemplate {
	ama := core.ActiveMeleeAbility{
		MeleeAbility: core.MeleeAbility{
			ActionID:       core.ActionID{SpellID: 34120},
			Character:      &hunter.Character,
			SpellSchool:    stats.AttackPower,
			IgnoreCost:     true,
			CritMultiplier: hunter.critMultiplier(true, sim.GetPrimaryTarget()),
		},
		Effect: core.AbilityHitEffect{
			AbilityEffect: core.AbilityEffect{
				DamageMultiplier:       1,
				StaticDamageMultiplier: 1,
				ThreatMultiplier:       1,
			},
			WeaponInput: core.WeaponDamageInput{
				IsRanged: true,
				CalculateDamage: func(attackPower float64, bonusWeaponDamage float64) float64 {
					return attackPower*0.2 +
						hunter.AutoAttacks.Ranged.BaseDamage(sim)*2.8/hunter.AutoAttacks.Ranged.SwingSpeed +
						150
				},
			},
		},
	}

	if ItemSetRiftStalker.CharacterHasSetBonus(&hunter.Character, 4) {
		ama.Effect.BonusCritRating += 5 * core.MeleeCritRatingPerCritChance
	}
	if ItemSetGronnstalker.CharacterHasSetBonus(&hunter.Character, 4) {
		ama.Effect.DamageMultiplier *= 1.1
	}

	return core.NewMeleeAbilityTemplate(ama)
}

func (hunter *Hunter) NewSteadyShot(sim *core.Simulation, target *core.Target) core.SimpleCast {
	hunter.steadyShotCast = hunter.steadyShotCastTemplate

	// Set dynamic fields, i.e. the stuff we couldn't precompute.
	hunter.steadyShotCast.OnCastComplete = func(sim *core.Simulation, cast *core.Cast) {
		ss := &hunter.steadyShotAbility
		hunter.steadyShotAbilityTemplate.Apply(ss)
		ss.Effect.Target = target
		ss.Attack(sim)
	}

	hunter.steadyShotCast.Init(sim)
	return hunter.steadyShotCast
}
