package paladin

import (
	"github.com/wowsims/tbc/sim/core"
)

func init() {
	core.AddItemSet(&ItemSetJusticarBattlegear)
	core.AddItemSet(&ItemSetCrystalforgeBattlegear)
}

var ItemSetJusticarBattlegear = core.ItemSet{
	Name: "Justicar Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// sim/debuffs.go handles this (and paladin/judgement.go)
		},
		4: func(agent core.Agent) {
			// TODO: if we ever implemented judgement of command, add bonus from 4p
		},
	},
}

var ItemSetCrystalforgeBattlegear = core.ItemSet{
	Name: "Crystalforge Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// judgement.go
		},
		4: func(agent core.Agent) {
			// TODO: if we implement healing, this heals party.
		},
	},
}

var ItemSetLightbringerBattlegear = core.ItemSet{
	Name: "Lightbringer Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.RegisterAura(core.Aura{
				Label:    "Lightbringer Battlegear 2pc",
				Duration: core.NeverExpires,
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					aura.Activate(sim)
				},
				OnSpellHit: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
					if !spellEffect.ProcMask.Matches(core.ProcMaskMelee) {
						return
					}
					if sim.RandomFloat("lightbringer 2pc") > 0.2 {
						return
					}
					character.AddMana(sim, 50, core.ActionID{SpellID: 38428}, true)
				},
			})
		},
		4: func(agent core.Agent) {
			// TODO: if we implemented hammer of wrath.. this ups dmg
		},
	},
}

// Librams implemented in seals.go and judgement.go

// TODO: once we have judgement of command.. https://tbc.wowhead.com/item=33503/libram-of-divine-judgement
