"""Deterministic contracts from readable source expectations, not candidate declarations."""
import json
from sorcerer_contract_oracle import (
    CONTRACTS, DEPENDENCIES, GRANTS, COMMON, SUMMON_DRAGON, execution_owner, level_fixture,
)
from sorcerer_spell_oracle import (
    SPELLS, SEAMS, SOURCE_PROTOCOLS, SOURCE_PAGES, EXTRA_DEPENDENCIES,
)


def spell_key(name, part):
    return 'spell-' + name.lower().replace('’', '').replace('/', '-').replace(' ', '-') + '-' + part.lower()


def class_contracts(records):
    tasks = []
    for key, (title, fixture, pre) in CONTRACTS.items():
        tasks.append(dict(key=key, title=title, fixture=fixture, required_outputs=list(pre),
                          prerequisites=list(pre), depends=list(DEPENDENCIES.get(key, []))))
    for n in range(1, 21):
        deps = ['hp', 'predicates', 'later-entry', 'slots', 'multiclass-slots', 'preparation', 'cantrips'] + GRANTS.get(n, [])
        if n == 1:
            deps.append('creation')
        if n >= 3:
            deps.append('draconic-hp')
        tasks.append(dict(key='level-' + str(n), title='Apply Sorcerer class level ' + str(n),
                          fixture=level_fixture(n), required_outputs=[285], prerequisites=[285], depends=sorted(set(deps))))
    for t in tasks:
        if t['key'].startswith('level-'):
            t['prerequisites'].remove(285)
            t['depends']=sorted(set(t['depends']+['advancement-output-gateway']))
        t.update(parent=330, issue=None, type='Task', size='M')
        t['owned_contexts'] = [dict(id=r['id'], keys=[k for k in r['required_mechanics'] + r['option_keys'] if execution_owner(r, k) == t['key']]) for r in records]
        t['owned_contexts'] = [c for c in t['owned_contexts'] if c['keys']]
        ids = {c['id'] for c in t['owned_contexts']}
        t['source_records'] = [r for r in records if r['id'] in ids]
    return tasks


def shared_contracts():
    result = []
    for name, parts in sorted(SPELLS.items()):
        first = next(iter(parts))
        for part, (fixture, pre) in parts.items():
            deps = [] if part == first else [spell_key(name, first)]
            deps += [spell_key(n, p) for n, p in EXTRA_DEPENDENCIES.get((name, part), [])]
            if (name, part) in [('Flesh to Stone', 'petrification-ending'), ('Prismatic Spray', 'indigo')]:
                deps.append('condition-petrified')
            if (name, part) == ('Finger of Death', 'zombie-creation-control'):
                deps += ['zombie-stat-actions', 'zombie-undead-fortitude']
            result.append(dict(key=spell_key(name, part), spell=name, part=part, fixture=fixture,
                               admission_context=parts[first][0], prerequisites=list(pre), required_outputs=list(pre),
                               depends=sorted(set(deps)), source_page=SOURCE_PAGES[name], title='Resolve ' + name + ' ' + part.replace('-', ' ')))
    for part, fixture in SUMMON_DRAGON.items():
        pre = {'initiative-verbal-command-and-default-Dodge': [782], 'Draconic-Spirit-defenses-senses': [487, 488, 727], 'zeroHP-and-spell-ending': [487]}.get(part, [])
        deps = [] if part == 'manifest-and-slot-scaling' else [spell_key('Summon Dragon', 'manifest-and-slot-scaling')]
        if part == 'Rend-and-Multiattack':
            deps.append(spell_key('Summon Dragon', 'Breath-Weapon'))
        if part == 'zeroHP-and-spell-ending':
            deps.append(spell_key('Summon Dragon', 'shared-resistance'))
        result.append(dict(key=spell_key('Summon Dragon', part), spell='Summon Dragon', part=part,
                           fixture=fixture, admission_context=SUMMON_DRAGON['manifest-and-slot-scaling'],
                           prerequisites=list(pre), required_outputs=list(pre), depends=sorted(set(deps)),
                           source_page=166, title='Resolve Summon Dragon ' + part.replace('-', ' ')))
    for key, (fixture, pre) in (SEAMS | SOURCE_PROTOCOLS).items():
        result.append(dict(key=key, spell=None, part=key, fixture=fixture, admission_context=fixture,
                           prerequisites=list(pre), required_outputs=list(pre), depends=[],
                           source_page=186 if key == 'condition-petrified' else 343 if key.startswith('zombie-') else 104,
                           title='Resolve bounded ' + key.replace('-', ' ')))
    for i, t in enumerate(result):
        t['depends'].append('execution-gateway-' + str(i // 80 + 1))
        if t['spell']:
            t['depends'].append('casting-gateway-' + str(i // 80 + 1))
        t['depends'] = sorted(set(t['depends']))
        t.update(parent='spell-cohort-' + str(i // 20 + 1), issue=None, type='Task', size='M', status='PENDING root-approved native publication')
    return result


def qualify_incapacitated(tasks):
    consumers = []
    for t in tasks:
        if 487 in t['prerequisites']:
            t['prerequisites'].remove(487)
            t['depends'] = sorted(set(t['depends'] + ['incapacitated-output-gateway']))
            consumers.append(t['key'])
    return sorted(consumers)


def shared_body(t):
    from sorcerer_contract_oracle import FINITE_CONSUMED
    union=sorted(set.union(*(set(v) for v in FINITE_CONSUMED.values())))
    boundary='\nExact finite source identities: '+json.dumps(union,ensure_ascii=False)+'\n' if t['key'] in SOURCE_PROTOCOLS else ''
    return ('Decision #330; one bounded implementation PR; Task Size M.\n\n'
            'Exact independent source execution fixture and contrary controls:\n' + t['fixture']
            + '\n\nAdmission/initial-cast context supplied by its separately owned resolver (not copied execution):\n'
            + t['admission_context'] + '\n\n' + COMMON
            + '\n\nIndependently violate every stated operand, target, action, resource, timing and source predicate; use deterministic positive and negative fixtures. The source fixture above, not the generic requirement, defines the actual behavior.\n'
            + '\nExact consumed old execution/data outputs: ' + str(t['required_outputs'])
            + '\nNative old prerequisites: ' + str(t['prerequisites'])
            + '\nExact bounded owned prerequisites: ' + str(t['depends'])
            + '\nNative parent: ' + str(t['parent']) + boundary
            + '\nQualification: this leaf -> its finite spell cohort -> Sorcerer shared output -> exactly-three-output aggregate ->279/280; also shared output -> own class gate ->337. World-cohort is additionally mandatory through shared-output; do not mutate old183 prerequisites. Planning185/274/273/277 closures never execute effects. Unknown modes/variants remain unavailable; source/header membership conflicts never grant class admission.\n')


def class_body(t):
    return ('Decision #330; one implementation PR; Task Size M.\n\n'
            'Exact execution scope and positive/independently violated negative controls:\n' + t['fixture']
            + '\n\n' + COMMON + '\n\nExact native prerequisite numbers: ' + str(t['prerequisites'])
            + '; planned owned prerequisites: ' + str(t['depends'])
            + '.\n\nOwned source contexts:\n```json\n' + json.dumps(t['owned_contexts'], indent=2) + '\n```\n')


def bind_effect_routes(tasks):
    """Spell references never acquire execution from a preparation grant."""
    from sorcerer_spell_oracle import REUSED_OUTPUTS
    route = {name: [spell_key(name, p) for p in parts] for name, parts in SPELLS.items()}
    route['Summon Dragon'] = [spell_key('Summon Dragon', p) for p in SUMMON_DRAGON]
    route.update(REUSED_OUTPUTS)
    for t in tasks:
        if t['key'] == 'spellcasting':
            t['depends'] = sorted(set(t['depends'] + list(SOURCE_PROTOCOLS)))
        if t['key'] == 'epic-boon':
            t['depends'] = sorted(set(t['depends'] + ['spellcasting']))
        if t['key'] == 'draconic-spells':
            names = ['Alter Self', 'Chromatic Orb', 'Command', 'Dragon’s Breath', 'Fear', 'Fly', 'Arcane Eye', 'Charm Monster', 'Legend Lore', 'Summon Dragon']
            t['depends'] = sorted(set(t['depends'] + [v for name in names for v in route[name] if isinstance(v, str)]))
            t['prerequisites'] = sorted(set(t['prerequisites'] + [v for name in names for v in route[name] if type(v) is int]))
            t['required_outputs'] = list(t['prerequisites'])
    return route


def qualification_gates(tasks, shared, incapacitated_consumers):
    from sorcerer_spell_oracle import REUSED_OUTPUTS
    gates = []

    def gate(key, title, parent, pre=(), deps=(), consumers=()):
        gates.append(dict(key=key, title=title, parent=parent, prerequisites=list(pre), depends=list(deps),
                          consumers=list(consumers), role='qualification-only', issue=None, type='Task', size='M', status='PENDING root approval'))

    gate('advancement-output-gateway', 'Qualify complete advancement table outputs for finite Sorcerer applications', 330, [285], consumers=['level-'+str(n) for n in range(1,21)])
    gate('incapacitated-output-gateway', 'Qualify exact Incapacitated outputs for finite Sorcerer consumers', 330, [487], consumers=incapacitated_consumers)
    for batch in range(1, 4):
        cs = [t['key'] for i, t in enumerate(shared) if i // 80 + 1 == batch]
        gate('execution-gateway-' + str(batch), 'Qualify finite Sorcerer execution protocol cohort ' + str(batch), 330, [61, 64, 65, 66, 67, 273, 353], consumers=cs)
        cs = [t['key'] for i, t in enumerate(shared) if t['spell'] and i // 80 + 1 == batch]
        gate('casting-gateway-' + str(batch), 'Qualify finite source casting outputs cohort ' + str(batch), 330, deps=list(SOURCE_PROTOCOLS), consumers=cs)
    spellcohorts = []
    for i in range(0, len(shared), 20):
        k = 'spell-cohort-' + str(i // 20 + 1)
        spellcohorts.append(k)
        gate(k, 'Qualify finite Sorcerer shared spell cohort ' + str(i // 20 + 1), 277, deps=[t['key'] for t in shared[i:i + 20]])
    reused = sorted({n for ns in REUSED_OUTPUTS.values() for n in ns})
    reusecohorts = []
    for i in range(0, len(reused), 40):
        k = 'reuse-cohort-' + str(i // 40 + 1)
        reusecohorts.append(k)
        gate(k, 'Qualify finite reused Sorcerer shared outputs cohort ' + str(i // 40 + 1), 330, reused[i:i + 40])
    gate('feature-cohort', 'Qualify finite Sorcerer feature outputs', 330, deps=[t['key'] for t in tasks if not t['key'].startswith('level-')])
    gate('level-cohort', 'Qualify all twenty Sorcerer applications', 330, deps=[t['key'] for t in tasks if t['key'].startswith('level-')])
    gate('shared-output', 'Qualify the finite Sorcerer consumed shared execution outputs', 330, deps=spellcohorts + reusecohorts + ['world-cohort'])
    gate('class-gate', 'Qualify exact Sorcerer class and level execution', 330, [330], ['feature-cohort', 'level-cohort', 'shared-output'])
    world_names = {'Animate Objects', 'Arcane Eye', 'Arcane Hand', 'Blink', 'Clairvoyance', 'Creation', 'Demiplane', 'Dimension Door', 'Earthquake', 'Etherealness', 'Finger of Death', 'Gate', 'Move Earth', 'Plane Shift', 'Prismatic Spray', 'Reverse Gravity', 'Summon Dragon', 'Telekinesis', 'Teleport', 'Teleportation Circle'}
    world = [t['key'] for t in shared if t['spell'] in world_names or t['key'].startswith('zombie-')]
    gate('world-cohort', 'Qualify finite Sorcerer consumed world outputs', 330, deps=world)
    gate('shared-aggregate', 'Qualify the finite Sorcerer Warlock Wizard shared execution outputs', 191, deps=['shared-output', 'warlock-shared-output', 'wizard-shared-output'])
    return gates


def gate_body(t):
    from sorcerer_contract_oracle import FINITE_CONSUMED
    boundary=''
    if t['key']=='advancement-output-gateway':
        boundary='\nSource23-26 COMPLETE20row285table,XP/livePB selected by POST-transition sumofadmittedclasses,neverclassrow. Fighter4+Sorcerer1 total5=6500/PB3;Fighter4/Sorcerer1->2 total6=14000/PB3;class1/2 featuresandlaterD6/fixed4HPseparate. Rejectbool/decimal/zero,total21,nonadjacent,unknowncombination,firstcharacterHPonlaterentry. ThisqualifiesdataoperandoutputsONLY,noXP/HPtransactionorclassgrant;all20consumersretainownHP->968->429.\n'
    if t['key']=='shared-output':
        boundary='\nImmutable142Sorcerer consumed source identities: '+json.dumps(FINITE_CONSUMED['sorcerer'],ensure_ascii=False)+'\nSorcerer MindSpike68/149 andPhantasmalForce68/151 disputed LIST admissions remain unavailable. Explicit independently qualified feature routes are separate; header never grants.\n'
    if t['key']=='shared-aggregate':
        boundary='\nImmutable three finite source domains (142/117/217),union245: '+json.dumps(FINITE_CONSUMED,ensure_ascii=False)+'\nExactly three SHARED heads, never consuming class qualification gates. Future heads preserve title/body/parent/TaskSize/source obligations and old prerequisites; ONLYroot-approved source-bound monotonic extensions recorded with manifest/head/native receipt may be wired BEFOREdecision acceptance. Every future head needs complete genuine nonempty inputs before native publication; incomplete inputs or planning closure cannot qualify. Preserve all old279/280edges; this aggregate alone adds one relation to each.\n'
    return ('Decision #330; one qualification PR; Task Size M.\n\n'
            + t['title'] + '. This owns QUALIFICATION ONLY, no runtime, spell effect, class grant, HP die or production entitlement.\n'
            + 'Exact parent: ' + str(t['parent']) + '\nExact old prerequisites: ' + str(t['prerequisites'])
            + '\nExact finite owned prerequisites: ' + str(t['depends'])
            + '\nExact finite consumers: ' + str(t['consumers']) + boundary
            + '\nQualify actual source-defined outputs of every listed input; data/catalogue/planning closure is insufficient. '
              'Source predicates/costs/targets/timing and deterministic positive/independently violated negative fixtures must match the independent readable oracle. '
              'Require prepare/confirm/retry/reload/concurrency/rollback in isolated PostgreSQL and shared web/TUI availability. '
              'Protocol prerequisites are build/qualification coupling, never blanket active-encounter/caster/concentration eligibility for ordinary feat or passive selection. '
              'Unknown variants and unqualified actual outputs remain unavailable. Preserve exact old source bodies/parents/prerequisites and all old qualification edges. '
              'Read back exact native body/title/parent/blocked-by/Task/ProjectSize and all three100caps plus separate dependency/containment DAGs. '
              'Every shared leaf must reach shared-output->three-caster aggregate->279/280 and shared-output->class-gate->337; world-cohort is a mandatory shared-output prerequisite; old183 is unchanged. '
              'No consuming class gate may block a shared head. Retain pinned SRD5.2.1 CC-BY4 attribution and adaptation notices.\n')


def complete_plan(records):
    tasks, shared = class_contracts(records), shared_contracts()
    routes = bind_effect_routes(tasks)
    consumers = qualify_incapacitated(tasks + shared)
    gates = qualification_gates(tasks, shared, consumers)
    for t in tasks:
        t['body'] = class_body(t)
    for t in shared:
        t['body'] = shared_body(t)
    for t in gates:
        t['body'] = gate_body(t)
    return tasks, shared, gates, routes


def transition_fixtures(root):
    from pathlib import Path
    base=Path(root)/'internal/rules/srd521/testdata'
    advancement=json.loads((base/'character-advancement.json').read_text())['records']
    rows={r['level']:r['advancement'] for r in advancement}
    result=[]
    for n in range(1,21):
        result.append(dict(key='single-class-'+str(n),before={} if n==1 else {'sorcerer':n-1},after={'sorcerer':n},sorcerer_row=n,total_row=n,xp=rows[n]['experience_points'],pb=rows[n]['proficiency_bonus'],hp_mode='first-character-d6-max6' if n==1 else 'later-d6-or-fixed4'))
    for before,after,n,total in [({'fighter':4},{'fighter':4,'sorcerer':1},1,5),({'fighter':4,'sorcerer':1},{'fighter':4,'sorcerer':2},2,6)]:
        result.append(dict(key='unequal-class-'+str(n),before=before,after=after,sorcerer_row=n,total_row=total,xp=rows[total]['experience_points'],pb=rows[total]['proficiency_bonus'],hp_mode='later-d6-or-fixed4'))
    return result
