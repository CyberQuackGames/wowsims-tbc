import { Player } from '../proto/api.js';
import { PlayerOptions } from '../proto/api.js';
import { BalanceDruid, BalanceDruid_Agent as BalanceDruidAgent } from '../proto/druid.js';
import { ElementalShaman, ElementalShaman_Agent as ElementalShamanAgent } from '../proto/shaman.js';
import { ComputeStatsRequest } from '../proto/api.js';
import { IndividualSimRequest } from '../proto/api.js';
export function makeComputeStatsRequest(buffs, consumes, customStats, encounter, gear, race, agent, talents, classOptions) {
    return ComputeStatsRequest.create({
        player: Player.create({
            customStats: customStats.asArray(),
            equipment: gear.asSpec(),
            options: withSpecProto(PlayerOptions.create({
                consumes: consumes,
                race: race,
            }), agent, talents, classOptions),
        }),
        buffs: buffs,
    });
}
export function makeIndividualSimRequest(buffs, consumes, customStats, encounter, gear, race, agent, talents, classOptions, iterations, debug) {
    return IndividualSimRequest.create({
        player: Player.create({
            customStats: customStats.asArray(),
            equipment: gear.asSpec(),
            options: withSpecProto(PlayerOptions.create({
                consumes: consumes,
                race: race,
            }), agent, talents, classOptions),
        }),
        buffs: buffs,
        encounter: encounter,
        iterations: iterations,
        gcdMin: 0.75,
        debug: debug,
    });
}
// Returns a copy of playerOptions, with the class field set.
function withSpecProto(playerOptions, agent, talents, specOptions) {
    const copy = PlayerOptions.clone(playerOptions);
    if (BalanceDruidAgent.is(agent)) {
        copy.spec = {
            oneofKind: 'balanceDruid',
            balanceDruid: BalanceDruid.create({
                agent: agent,
                talents: talents,
                options: specOptions,
            }),
        };
    }
    else if (ElementalShamanAgent.is(agent)) {
        copy.spec = {
            oneofKind: 'elementalShaman',
            elementalShaman: ElementalShaman.create({
                agent: agent,
                talents: talents,
                options: specOptions,
            }),
        };
    }
    else {
        throw new Error('Unrecognized talents with options: ' + PlayerOptions.toJsonString(playerOptions));
    }
    return copy;
}