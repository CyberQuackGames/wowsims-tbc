package hunter

import (
	"log"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

func init() {
	core.AddItemEffect(30448, ApplyTalonOfAlar)
	core.AddItemEffect(30892, ApplyBeasttamersShoulders)
	core.AddItemEffect(32336, ApplyBlackBowOfTheBetrayer)
	core.AddItemEffect(32487, ApplyAshtongueTalismanOfSwiftness)

	core.AddItemSet(&ItemSetBeastLord)
	core.AddItemSet(&ItemSetDemonStalker)
	core.AddItemSet(&ItemSetRiftStalker)
	core.AddItemSet(&ItemSetGronnstalker)
}

var ItemSetBeastLord = core.ItemSet{
	Name: "Beast Lord Armor",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
		},
		4: func(agent core.Agent) {
			// Handled in kill_command.go
		},
	},
}

var ItemSetDemonStalker = core.ItemSet{
	Name: "Demon Stalker Armor",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
		},
		4: func(agent core.Agent) {
			// Handled in multi_shot.go
		},
	},
}

var ItemSetRiftStalker = core.ItemSet{
	Name: "Rift Stalker Armor",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
		},
		4: func(agent core.Agent) {
			// Handled in steady_shot.go
		},
	},
}

var ItemSetGronnstalker = core.ItemSet{
	Name: "Gronnstalker's Armor",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// Handled in rotation.go
		},
		4: func(agent core.Agent) {
			// Handled in steady_shot.go
		},
	},
}

func ApplyTalonOfAlar(agent core.Agent) {
	hunterAgent, ok := agent.(Agent)
	if !ok {
		log.Fatalf("Non-hunter attempted to activate hunter item effect.")
	}
	hunter := hunterAgent.GetHunter()

	procAura := hunter.GetOrRegisterAura(core.Aura{
		Label:    "Talon of Alar Proc",
		ActionID: core.ActionID{ItemID: 30448},
		// Add 1 in case we use arcane shot exactly off CD.
		Duration: time.Second*6 + 1,
	})

	hunter.TalonOfAlarAura = hunter.GetOrRegisterAura(core.Aura{
		Label:    "Talon of Alar",
		Duration: core.NeverExpires,
		OnSpellHit: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
			if spell.SameAction(ArcaneShotActionID) {
				procAura.Activate(sim)
			}
		},
	})

	hunter.AddPermanentAura(func(sim *core.Simulation) *core.Aura {
		return hunter.TalonOfAlarAura
	})
}

func (hunter *Hunter) talonOfAlarDamageMod(baseDamageConfig core.BaseDamageConfig) core.BaseDamageConfig {
	if hunter.HasTrinketEquipped(30448) {
		return core.WrapBaseDamageConfig(baseDamageConfig, func(oldCalculator core.BaseDamageCalculator) core.BaseDamageCalculator {
			return func(sim *core.Simulation, hitEffect *core.SpellEffect, spell *core.Spell) float64 {
				normalDamage := oldCalculator(sim, hitEffect, spell)
				if hunter.TalonOfAlarAura != nil && hunter.TalonOfAlarAura.IsActive() {
					return normalDamage + 40
				} else {
					return normalDamage
				}
			}
		})
	} else {
		return baseDamageConfig
	}
}

func ApplyBeasttamersShoulders(agent core.Agent) {
	hunterAgent, ok := agent.(Agent)
	if !ok {
		log.Fatalf("Non-hunter attempted to activate hunter item effect.")
	}
	hunter := hunterAgent.GetHunter()

	hunter.pet.PseudoStats.DamageDealtMultiplier *= 1.03
	hunter.pet.AddStat(stats.MeleeCrit, core.MeleeCritRatingPerCritChance*2)
}

func ApplyBlackBowOfTheBetrayer(agent core.Agent) {
	character := agent.GetCharacter()
	const manaGain = 8.0
	character.AddPermanentAura(func(sim *core.Simulation) *core.Aura {
		return character.GetOrRegisterAura(core.Aura{
			Label: "Black Bow of the Betrayer",
			OnSpellHit: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
				if !spellEffect.Landed() || !spellEffect.ProcMask.Matches(core.ProcMaskRanged) {
					return
				}
				character.AddMana(sim, manaGain, core.ActionID{SpellID: 46939}, false)
			},
		})
	})
}

func ApplyAshtongueTalismanOfSwiftness(agent core.Agent) {
	character := agent.GetCharacter()
	character.AddPermanentAura(func(sim *core.Simulation) *core.Aura {
		procAura := character.NewTemporaryStatsAura("Ashtongue Talisman Proc", core.ActionID{ItemID: 32487}, stats.Stats{stats.AttackPower: 275, stats.RangedAttackPower: 275}, time.Second*8)
		const procChance = 0.15

		return character.GetOrRegisterAura(core.Aura{
			Label: "Ashtongue Talisman",
			OnSpellHit: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
				if !spell.SameAction(SteadyShotActionID) {
					return
				}
				if sim.RandomFloat("Ashtongue Talisman of Swiftness") > procChance {
					return
				}
				procAura.Activate(sim)
			},
		})
	})
}
