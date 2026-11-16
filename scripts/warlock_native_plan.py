"""Build the finite #331 publication proposal without native writes.

Dependency and containment graphs are separate. Foreign proposed outputs remain
named source-bound inputs; this file cannot certify their publication.
"""
import copy
from collections import Counter
import json

from warlock_contract_oracle import COMMON

PROJECT = 'PVT_kwDODjtmGM4BrFJ2'
CORE = [61, 64, 65, 66, 67, 353, 774, 780, 781, 782]
CONDITIONS = [487, 488, 504, 727, 777, 778, 960, 961, 962, 963, 964]
CAST_SEAMS = ['cast-components', 'cast-cost-privacy', 'cast-long-ritual', 'effect-time-span-stack']


def build(m, sorcerer, baseline):
    tasks = copy.deepcopy(m['tasks'] + m['shared_effect_tasks'] + m['creature_tasks'] + m['gear_tasks'])
    rows = {t['key']: t for t in tasks}
    own = set(rows)
    gates = []
    def gate(key, title, parent, depends=(), prerequisites=(), scope=''):
        t = dict(key=key, title=title, type='Task', size='M', project=PROJECT,
                 parent=parent, depends=sorted(set(depends)), prerequisites=sorted(set(prerequisites)),
                 issue=None, role='qualification',
                 body=f'Decision #331; one qualification PR; Task Size M.\n\n{scope}\n\n'
                      'Own NO runtime, class grant, spell effect or production admission. '
                      'Require each finite named output; source/planning closure never substitutes. '
                      'Preserve title/body/parent/Task/Size/source boundary and all accepted prerequisites. '
                      'Only ROOT-approved source-bound monotonic native prerequisite extensions with semantic scope, '
                      'manifest/head/receipt may precede decision acceptance; never accept by hash alone.\n\n' + COMMON + '\n')
        gates.append(t)
        rows[key] = t
        return t

    # One new edge per old primitive, then finite factoring below100 consumers.
    impl = sorted(t['key'] for t in tasks if t.get('role') != 'qualification' and t['key'] != 'hp')
    gate('execution-output', 'Qualify finite Warlock dice target transaction and action outputs', 331,
         prerequisites=CORE, scope='Exact primitive owners ' + ', '.join(map(str, CORE)) +
         '. Heroic774 is ANY owned known typed die, not Bardic512; each future resolver registers its exact die/phase adapter, '
         'unknown/unregistered/foreign/NPC-private dice unavailable, no blanket known-outcome requirement. '
         '780/781/782 WHOLE Dash/Disengage/Dodge outputs, not inferred NPC Unarmed779. '
         'Consumers are EXACTLY the three execution-cohort outputs below, covering this finite leaf roster: ' + ', '.join(impl) + '.')
    for start in range(0, len(impl), 90):
        key = 'execution-cohort-' + str(start // 90 + 1)
        consumers = impl[start:start+90]
        gate(key, 'Qualify Warlock execution cohort ' + str(start // 90 + 1), 331,
             depends=['execution-output', 'condition-output'], scope='Only these finite consumers use this output: ' + ', '.join(consumers) + '.')
        for consumer in consumers:
            rows[consumer]['depends'].append(key)

    gate('condition-output', 'Qualify finite Warlock bound condition outputs', 331,
         prerequisites=CONDITIONS, scope='Consume WHOLE487/488/504/727/777/778/960/961/962/963/964 condition, immunity, source binding and ending APIs. '
         'Finite consumers are the named Warlock/Fiend/Invocation and spell/creature bodies in this decision. '
         'No generic condition implementation or authority is added.')
    gate('casting-output', 'Qualify finite Warlock consumption of four shared casting seams', 331,
         depends=['330:' + x for x in CAST_SEAMS], scope='Exact external Sorcerer-owned finite245-source seams: ' +
         ', '.join(CAST_SEAMS) + '. Their actual qualified outputs are required, provisional contracts do not qualify. '
         'Scope ONLY the117 Warlock consumed identities and exact feature modes; no273/277 planning execution and no generic any-spell inference.')
    gate('provenance-output', 'Qualify finite Warlock source grant provenance', 331,
         prerequisites=[430], scope='WHOLE430 provenance for creation/later-entry/each20level/Invocation/Tome/Fiend/Blade grants; '
         'retain distinct source instances, removal affects only own grants, replay grants once. No grant implementation here.')

    def deps(key, *values):
        rows[key]['depends'] = sorted(set(rows[key]['depends']) | set(values))
    for t in tasks:
        if t.get('spell') or t['key'] in ('pact-magic', 'pact-cross-casting', 'contact-patron', 'chain-cast-mode', 'tome-finite-preparation', 'depths-cast-mode') or t.get('part') == 'invisibility':
            deps(t['key'], 'casting-output')
        if t['key'].startswith('level-') or t['key'] in ('creation', 'later-entry', 'core-traits', 'skills', 'equipment', 'eldritch-invocations', 'preparation', 'blade-proficiency-focus', 'tome-finite-preparation', 'fiend-spells'):
            deps(t['key'], 'provenance-output')
    # Structural source dependencies are real outputs, never consuming class gates.
    deps('pact-magic', 'warlock-list', 'preparation', 'slots', 'pact-cross-casting')
    deps('magical-cunning', 'slots')
    deps('eldritch-master', 'magical-cunning')
    rows['slots']['prerequisites'] += [427, 428]
    rows['magical-cunning']['prerequisites'] += [428]
    rows['hp']['prerequisites'] += [774]
    for t in tasks:
        if '1/Day' in t['fixture'] or '2/Day' in t['fixture']:
            t['prerequisites'].append(428)
            t['body'] += '\nX/Day resets only on THIS CREATURE completed Long Rest via428; caster rest alone and calendar/dawn are insufficient.\n'
    for key in ('ability-score-improvement', 'epic-boon'):
        rows[key]['prerequisites'] += list(range(384, 402))
    rows['lessons-of-the-first-ones']['prerequisites'] += [384, 385, 386, 387, 388, 389]
    rows['spell-befuddlement-ending']['prerequisites'] += [585, 621, 623]
    rows['spell-contact-other-plane-failed-contact']['prerequisites'] += [428, 621, 623]
    rows['creature-mummy-rot-curse']['prerequisites'] += [428, 625, 937]
    deps('equipment', *(t['key'] for t in m['gear_tasks'] if t['key'] != 'gear-glass-bottle-profile'))
    deps('gear-dagger', 'gear-dagger-sickle-light')
    deps('gear-sickle', 'gear-dagger-sickle-light')
    rows['gear-dagger-sickle-light']['prerequisites'] += [491, 394]
    deps('gear-lamp-oil-fuel', 'gear-tinderbox')
    deps('spell-true-polymorph-creature-form', 'creature-owl-bounded-stat-actions')
    deps('spell-true-polymorph-object-creature', 'creature-owl-bounded-stat-actions', 'gear-glass-bottle-profile')
    deps('spell-true-polymorph-creature-object', 'gear-glass-bottle-profile')

    # Exact whole-owner reuse, with qualified existing effects preferred over
    # provisional new effects. No name match is a qualification claim.
    spell_owners = {}
    for entry in m['spell_overlap_audit']:
        name = entry['name']
        existing = [str(x['issue']) for x in entry['candidate_owners']]
        own_effects = [t['key'] for t in tasks if t.get('spell') == name]
        foreign = ['330:' + t['key'] for t in sorcerer['shared_split_proposal'] if t.get('spell') == name]
        if name == 'Find Familiar':
            owners = [t['key'] for t in tasks if t.get('creature') in m['familiar_source_roster']['normal'] + m['familiar_source_roster']['chain'] + ['Find Familiar']]
        else:
            owners = existing or own_effects or foreign
        assert owners, ('unassigned consumed spell effect', name)
        spell_owners[name] = owners

    def spell(key, *names):
        deps(key, *(owner for name in names for owner in spell_owners[name]))
    for key, names in {
        'armor-of-shadows':['Mage Armor'], 'ascendant-step':['Levitate'], 'fiendish-vigor':['False Life'],
        'mask-of-many-faces':['Disguise Self'], 'master-of-myriad-forms':['Alter Self'],
        'misty-visions':['Silent Image'], 'one-with-shadows':['Invisibility'], 'otherworldly-leap':['Jump'],
        'visions-of-distant-realms':['Arcane Eye'], 'whispers-of-the-grave':['Speak with Dead'],
        'contact-patron':['Contact Other Plane'], 'depths-cast-mode':['Water Breathing'],
        'chain-cast-mode':['Find Familiar'],
        'chain-form-choice':['Find Familiar'], 'chain-attack-exchange':['Find Familiar'],
        'investment-command':['Find Familiar'],
    }.items():
        spell(key, *names)
    for t in tasks:
        if t.get('part') == 'invisibility':
            spell(t['key'], 'Invisibility')
        if t['key'].endswith('zombie-rise'):
            deps(t['key'], '330:zombie-stat-actions', '330:zombie-undead-fortitude')
    for name in m['familiar_source_roster']['undead']:
        deps('spell-create-undead-upcast', *(t['key'] for t in tasks if t.get('creature') == name))
    for group, parent in [('shared_effect_tasks', 277), ('creature_tasks', 277), ('gear_tasks', 278)]:
        cohorts = sorted({t['parent'] for t in m[group]})
        for key in cohorts:
            members = [t['key'] for t in m[group] if t['parent'] == key]
            gate(key, 'Qualify Warlock ' + key.replace('-', ' '), parent, depends=members,
                 scope='Exact source-bound leaf roster: ' + ', '.join(members) + '.')
    for key in ('features-cohort', 'levels-cohort'):
        members = [t['key'] for t in m['tasks'] if t['parent'] == key]
        gate(key, 'Qualify Warlock ' + key.replace('-', ' '), 331, depends=members,
             scope='Exact class leaf/qualification roster: ' + ', '.join(members) + '.')
    names = m['spells']['consumed']
    reuse = []
    for start in range(0, len(names), 20):
        batch = names[start:start+20]
        members = sorted({owner for name in batch for owner in spell_owners[name]})
        # Split qualification inputs if a familiar cohort makes a batch exceed100.
        for sub in range(0, len(members), 80):
            key = f'consumed-output-{start//20+1}-{sub//80+1}'
            gate(key, 'Qualify Warlock consumed source cohort ' + key.removeprefix('consumed-output-'), 331,
                 depends=members[sub:sub+80], scope='Finite spell identities: ' + ', '.join(batch) +
                 '. Exact mandatory outputs in this slice: ' + ', '.join(members[sub:sub+80]) + '.')
            reuse.append(key)
    # Every shared leaf has both required paths even where its source effect
    # is only a nested lifecycle/statblock part rather than the spell's title.
    all_shared_cohorts = sorted({t['parent'] for t in m['shared_effect_tasks'] + m['creature_tasks'] + m['gear_tasks']})
    gate('shared-output', 'Qualify the finite Warlock consumed shared execution outputs', 331,
         depends=reuse + all_shared_cohorts, scope='Source boundary EXACT117 identities: ' + ', '.join(names) +
         '. Includes all72printed spells, Fiend/namedInvocation effects,27Tomecantrips/11rituals,24CR0normalforms/7Chainforms '
         'and4CreateUndead forms; starting-gear/object scopes EXACTLY: ' + ', '.join(t['key'] for t in m['gear_tasks']) + '. Unknown adjacent statblock/item/world effects stay unavailable. '
         'This is ONE of exactlyTHREE class-SHARED outputs for330/331/332 aggregate, not class admission.')
    gate('class-gate', 'Qualify exact Warlock feature level and shared output paths', 331,
         depends=['features-cohort', 'levels-cohort', 'shared-output'], prerequisites=[331],
         scope='Every exact Warlock feature/option/level/shared output is mandatory, including each repeated grant. '
         'Decision331 may close while implementation edges remain mandatory; closure alone never admits Warlock. '
         'This qualification blocks337, never directly279 and never consumed by shared outputs.')

    # Native bodies carry full contextual source operands, distinct from titles.
    for t in tasks:
        contexts = [dict(id=c['record']['id'], keys=sorted(k for k, owner in c['key_owners'].items() if owner == t['key']))
                    for c in m['record_matrix'] if t['key'] in c['key_owners'].values()]
        t['owned_contexts'] = contexts
        t['source_records'] = [c['record'] for c in m['record_matrix'] if t['key'] in c['key_owners'].values()]
        t['project'] = PROJECT
        t['native_body'] = t['body'] + '\nExact owned contextual operands:\n```json\n' + json.dumps(t['source_records'], indent=2) + '\n```\n'
        envelope = (m['consumed_source_envelopes'].get(t.get('spell')) or
                    m['creature_source_evidence'].get(t.get('creature')))
        if envelope:
            t['native_body'] += '\nComplete pinned source context (not permission for adjacent behavior):\n```json\n' + json.dumps(envelope, indent=2) + '\n```\n'
        t['depends'] = sorted(set(t['depends']))
        t['prerequisites'] = sorted(set(t['prerequisites']))

    assert len(rows) == len(tasks) + len(gates)
    return dict(tasks=tasks, gates=gates, spell_owners=spell_owners,
                status='PENDING ROOT semantic/protocol/combined-cap approval; no native writes',
                native_roots={'337':['class-gate'], '279':[], '280':[]},
                aggregate_input='330:shared-aggregate consumes331:shared-output',
                preserved_old_native_nodes=len(baseline))


def projection(plan, baseline, sorcerer, wizard=None):
    graph = {str(x['number']): [str(n) for n in x['blocked_by']] for x in baseline}
    parents = {str(x['number']): str(x['parent']) for x in baseline if x['parent'] is not None}
    own = {t['key'] for t in plan['tasks'] + plan['gates']}
    foreign = {t['key'] for t in sorcerer['tasks'] + sorcerer['shared_split_proposal'] + sorcerer['gates']}
    def ref(key, prefix):
        if isinstance(key, int) or str(key).isdigit():
            return str(key)
        if ':' in key:
            return key
        if key == 'warlock-shared-output':
            return '331:shared-output'
        if key == 'wizard-shared-output':
            return '332:shared-output'
        return prefix + ':' + key
    packs = [('331', plan['tasks'] + plan['gates']),
             ('330', sorcerer['tasks'] + sorcerer['shared_split_proposal'] + sorcerer['gates'])]
    if wizard is not None:
        packs.append(('332', wizard['tasks'] + wizard['gates']))
    for prefix, tasks in packs:
        for t in tasks:
            ident = ref(t['key'], prefix)
            assert ident not in graph, ('duplicate proposal identity', ident)
            graph[ident] = sorted(set([str(n) for n in t.get('prerequisites', [])] + [ref(k, prefix) for k in t['depends']]))
            parents[ident] = ref(t['parent'], prefix)
    # Wizard audit supplies this finite input's concrete structure. No invented
    # zero-edge node can serve as a final certificate.
    unresolved = sorted({d for ds in graph.values() for d in ds if d not in graph})
    graph['279'].append('330:shared-aggregate')
    graph['280'].append('330:shared-aggregate')
    graph['337'] += ['330:class-gate', '331:class-gate']
    if wizard is not None:
        graph['337'].append('332:class-gate')
    outgoing = Counter(d for ds in graph.values() for d in ds)
    children = Counter(parents.values())
    return dict(graph=graph, parents=parents, unresolved_foreign=unresolved,
                max_incoming=max(map(len, graph.values())), max_outgoing=max(outgoing.values()),
                max_children=max(children.values()), root279=len(graph['279']), root280=len(graph['280']),
                root337=len(graph['337']), children283=children['283'], outgoing429=outgoing['429'],
                own_nodes=len(own), foreign330_nodes=len(foreign),
                foreign332_nodes=len(wizard['tasks'] + wizard['gates']) if wizard is not None else None,
                final_certificate=False)
