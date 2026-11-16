#!/usr/bin/env python3
"""Read-only Warlock candidate proof; strict mode requires actual publication."""
import argparse
from functools import lru_cache
from collections import Counter
import hashlib
import importlib
import json
from pathlib import Path
import re
import subprocess

from warlock_contract_oracle import (CONTRACTS, SPLITS, TOME_CANTRIPS, TOME_RITUALS,
                                     FIEND_SPELLS, NAMED_SPELLS, body, level_fixture)
from warlock_shared_oracle import SPELLS, CASTING, EXISTING_OWNERS
from warlock_familiar_oracle import NORMAL, SPECIAL, UNDEAD, FAMILIAR, SOURCE_PAGES
from warlock_gear_oracle import GEAR
from warlock_contract_oracle import COMMON
from warlock_contract_oracle import TRANSITION_WITNESSES, CONSUMER_EVIDENCE

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / 'docs/investigations/issue-331-warlock-scope.json'
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
BASE = 'e7cd1ba8318ba296edd4fc5b6b84715ca11016f2'
NAMES = ['warlock', 'fiend', 'invocations_01', 'invocations_02', 'invocations_03']


@lru_cache(maxsize=5)
def accepted_catalogue(path):
    # Required starting-base provenance is immutable; metadata closure supplies
    # no runtime. Caching avoids repeating Git subprocesses for contrary probes.
    return json.loads(subprocess.check_output(['git', 'show', BASE + ':' + path], cwd=ROOT, text=True))


def exact(a, b):
    if type(a) is not type(b):
        return False
    if isinstance(a, dict):
        return a.keys() == b.keys() and all(exact(a[k], b[k]) for k in a)
    if isinstance(a, list):
        return len(a) == len(b) and all(exact(x, y) for x, y in zip(a, b))
    return a == b


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, 'PDF pin'
    result = {}
    for name in NAMES:
        mod = importlib.import_module('verify_srd_' + name)
        records = mod.source_records(pdf) if name in ('warlock', 'invocations_02') else mod.source_records()
        result[name] = records
    assert [len(x) for x in result.values()] == [71, 17, 41, 49, 13]
    return result


def spell_rows(pdf, pages):
    result, level = [], None
    for page, x in pages:
        text = subprocess.check_output(['pdftotext', '-f', str(page), '-l', str(page),
            '-x', str(x), '-y', '0', '-W', '310', '-H', '792', '-layout', str(pdf), '-'], text=True)
        for line in text.splitlines():
            h = re.search(r'(?:Level (\d) \w+ Spells|Cantrips \(Level (0))', line)
            if h:
                level = int(h[1] or h[2])
            m = re.match(r'\s*(.+?)\s{2,}(Abjuration|Conjuration|Divination|Enchantment|Evocation|Illusion|Necromancy|Transmutation)\s+([CMR, —]+)\s*$', line)
            if m:
                result.append(dict(name=m[1].strip(), level=level, page=page, school=m[2], special=m[3].strip()))
    return result


def source_sections(pdf, inventory, pages):
    """Bounded local PDF columns; headings come from the accepted finite index."""
    names = {r['name'] for r in json.loads((ROOT / inventory).read_text())['records']}
    sections, current = {}, None
    for page in pages:
        for x in (0, 310):
            text = subprocess.check_output(['pdftotext', '-f', str(page), '-l', str(page),
                '-x', str(x), '-y', '0', '-W', '310', '-H', '792', '-layout', str(pdf), '-'], text=True, timeout=10)
            for line in text.splitlines():
                clean = line.strip()
                if clean in names:
                    current = clean
                    sections.setdefault(current, dict(page=page, lines=[]))
                if current:
                    sections[current]['lines'].append(clean)
    return sections


def validate_source_envelopes(m, spells, creatures):
    assert exact(m['consumed_source_envelopes'], {n: spells[n] for n in m['spells']['consumed']}), 'actual117 complete source envelopes'
    assert exact(m['creature_source_evidence'], {n: creatures[n] for n in list(NORMAL) + list(SPECIAL) + list(UNDEAD)}), 'actual35 complete statblock envelopes'
    actual = {n for n, s in creatures.items() if any('Beast,' in line for line in s['lines'][:3]) and any('CR 0 (' in line for line in s['lines'])}
    assert actual == set(NORMAL), 'all actual pinned CR0 Beasts, not just named examples'


def verify_gear_source(pdf):
    clauses = {
        89: ['you don’t add your ability modi-', 'unless that modi-', 'Finesse'],
        90: ['You can’t attack a target beyond the long range.', 'you can draw that weapon as part of the attack.'],
        91: ['Dagger', '1d4 Piercing', 'Thrown (Range 20/60)', 'Sickle', '1d4 Slashing'],
        92: ['Light Armor (1 Minute to Don or Doff)', '11 + Dex modifier', 'Leather Armor'],
        96: ['Orb', '20 GP', 'accurate nonfiction Book about its topic', '+5 bonus to Intelligence', 'holds up to 30 pounds within 1 cubic'],
        98: ['about 500 pages', '15-', 'additional 30 feet', 'extra 5 Fire damage', 'only', 'once per turn'],
        99: ['Back-', 'pack, Book, Ink, Ink Pen, Lamp, 10 flasks of Oil, 10', 'sheets of Parchment, and Tinderbox.', '6 hours', '250 hand-'],
        100: ['exposed fuel—takes a Bonus Action.', 'other fire takes 1 minute.'],
        178: ['13', 'Crystal, glass, ice', 'Tiny (bottle, lock)', '2 (1d4)', 'Poison and Psychic'],
    }
    for page, anchors in clauses.items():
        text = subprocess.check_output(['pdftotext', '-f', str(page), '-l', str(page), '-layout', str(pdf), '-'], text=True, timeout=10)
        views = [text, ' '.join(text.split())]
        for x in (0, 310):
            column = subprocess.check_output(['pdftotext', '-f', str(page), '-l', str(page), '-x', str(x), '-y', '0', '-W', '310', '-H', '792', '-layout', str(pdf), '-'], text=True, timeout=10)
            views += [column, ' '.join(column.split())]
        for anchor in anchors:
            assert any(anchor in view for view in views), ('actual gear source predicate', page, anchor)


def source_owner(record, key, by):
    if record['kind'] == 'level':
        return 'level-' + str(record['level'])
    if key.endswith('/hit-die-sides'):
        return 'hp'
    if '/multiclass/' in key:
        return 'later-entry'
    ident = record['id']
    if record['kind'] == 'class':
        return 'level-1' if '/level-1/' in key else 'core-traits'
    if ident.endswith('/warlock/skills'):
        return 'skills'
    if ident.endswith('/warlock/starting-equipment'):
        return 'equipment'
    if '/arcanum-' in ident:
        return 'arcanum-' + re.search(r'/arcanum-(\d)', ident)[1]
    if '/pact-slots' in ident:
        return 'slots'
    if any('/' + s in ident for s in ('cantrips', 'prepared-spells', 'cantrip-eligibility', 'prepared-spell-eligibility')):
        return 'preparation'
    if record['kind'] == 'subclass':
        return 'warlock-subclass'
    if record['kind'] == 'feature':
        slug = ident.rsplit('/', 1)[-1]
        routes = {
            'pact-of-the-blade': [('ends', 'blade-ending'), ('disappear', 'blade-ending'), ('minute', 'blade-ending'), ('distance', 'blade-ending'), ('proficiency', 'blade-proficiency-focus'), ('focus', 'blade-proficiency-focus'), ('attack', 'blade-combat'), ('damage', 'blade-combat'), ('charisma', 'blade-combat'), ('conjure', 'blade-bond'), ('bond', 'blade-bond')],
            'pact-of-the-chain': [('forgo', 'chain-attack-exchange'), ('reaction', 'chain-attack-exchange'), ('attack', 'chain-attack-exchange'), ('forms', 'chain-form-choice'), ('monsters/', 'chain-form-choice'), ('familiar', 'chain-cast-mode'), ('spell', 'chain-cast-mode')],
            'pact-of-the-tome': [('focus', 'tome-focus'), ('cantrip', 'tome-finite-preparation'), ('ritual', 'tome-finite-preparation'), ('spell', 'tome-finite-preparation'), ('prepared', 'tome-finite-preparation'), ('book', 'tome-conjuration-ending'), ('conjure', 'tome-conjuration-ending')],
            'gaze-of-two-minds': [('cast', 'gaze-cast-origin'), ('sixty', 'gaze-cast-origin'), ('60', 'gaze-cast-origin'), ('range', 'gaze-cast-origin'), ('sense', 'gaze-connection'), ('connection', 'gaze-connection'), ('touch', 'gaze-connection'), ('maintain', 'gaze-connection')],
            'gift-of-the-depths': [('swim', 'depths-passive'), ('underwater', 'depths-passive'), ('water-breathing', 'depths-cast-mode'), ('slot', 'depths-cast-mode'), ('rest', 'depths-cast-mode')],
            'gift-of-the-protectors': [('zero', 'protectors-rescue'), ('hit-point', 'protectors-rescue'), ('trigger', 'protectors-rescue'), ('rest', 'protectors-rescue'), ('killed', 'protectors-rescue'), ('name', 'protectors-inscription'), ('page', 'protectors-inscription'), ('erase', 'protectors-inscription')],
            'investment-of-the-chain-master': [('reaction', 'investment-resistance'), ('resistance', 'investment-resistance'), ('save', 'investment-save-dc'), ('damage', 'investment-damage'), ('attack', 'investment-command'), ('command', 'investment-command'), ('speed', 'investment-movement'), ('movement', 'investment-movement')],
            'hurl-through-hell': [('damage', 'hurl-psychic'), ('fiend', 'hurl-psychic'), ('return', 'hurl-transit-ending'), ('incapacitated', 'hurl-transit-ending'), ('space', 'hurl-transit-ending'), ('disappear', 'hurl-transit-ending'), ('slot', 'hurl-admission-cost'), ('save', 'hurl-admission-cost'), ('use', 'hurl-admission-cost'), ('turn', 'hurl-admission-cost')],
        }
        # Match only the mechanic suffix; the feature name itself must not
        # accidentally match a semantic predicate (e.g. Fiend Hurl damage).
        suffix = key.split(slug + '/', 1)[-1]
        if slug == 'hurl-through-hell':
            suffix = key.split('/hurl/', 1)[-1] if '/hurl/' in key else ''
        for token, owner in routes.get(slug, []):
            if token in suffix:
                return owner
        return slug
    if record['parent_id'] in by:
        return source_owner(by[record['parent_id']], key, by)
    return 'creation'


def acyclic(graph):
    done, active = set(), set()
    def visit(n):
        assert n not in active, 'cycle'
        if n in done:
            return
        active.add(n)
        for d in graph.get(n, []):
            visit(d)
        active.remove(n)
        done.add(n)
    for n in graph:
        visit(n)


def caps(graph, parents):
    acyclic(graph)
    acyclic({k: [v] for k, v in parents.items()})
    assert all(len(v) == len(set(v)) and len(v) <= 100 for v in graph.values()), 'incoming cap'
    assert max(Counter(d for v in graph.values() for d in v).values(), default=0) <= 100, 'outgoing cap'
    assert max(Counter(parents.values()).values(), default=0) <= 100, 'children cap'


def bounded_pages(fetch, endpoint, limit=16):
    assert endpoint.endswith('per_page=100') and 'page=' not in endpoint.replace('per_page=', ''), 'exact query'
    out, seen, payloads = [], set(), set()
    for page in range(1, limit + 1):
        data = fetch(endpoint + '&page=' + str(page))
        assert isinstance(data, list) and len(data) <= 100, 'page shape'
        if not data:
            return out
        payload = json.dumps(data, sort_keys=True)
        assert payload not in payloads, 'repeated page'
        payloads.add(payload)
        for row in data:
            assert type(row['number']) is int and row['number'] not in seen, 'duplicate/invalid identity'
            seen.add(row['number'])
            out.append(row)
        if len(data) < 100:
            return out
    raise AssertionError('call cap')


def prefix_check(actual, prefix, append, marker):
    assert actual.startswith(prefix + append), 'immutable prefix and exact append'
    assert actual.count(marker) == 1, 'append occurrence'


def expected_tasks(records):
    by = {r['id']: r for r in records}
    tasks = []
    def add(key, fixture, depends=(), role='implementation', parent='features-cohort'):
        tasks.append(dict(key=key, title='Implement Warlock ' + key.replace('-', ' ') if role == 'implementation' else 'Qualify Warlock ' + key.replace('-', ' '),
            fixture=fixture, body=body(key, fixture), type='Task', size='M', parent=parent,
            issue=None, prerequisites=[968] if key == 'hp' else [430] if key == 'creation' else [],
            depends=sorted(set(depends)), role=role))
    for key, fixture in CONTRACTS.items():
        depends = list(SPLITS.get(key, {}))
        if key == 'creation':
            depends += ['core-traits', 'skills', 'equipment', 'hp', 'pact-magic', 'eldritch-invocations']
        add(key, fixture, depends, 'qualification' if key in SPLITS else 'implementation')
    for parent, leaves in SPLITS.items():
        for key, fixture in leaves.items():
            add(key, fixture)
    for level in range(1, 21):
        row = by['srd-5.2.1:character/level/warlock/' + str(level)]
        deps = ['hp', 'later-entry', 'preparation', 'eldritch-invocations', 'slots', 'warlock-list', 'pact-cross-casting'] + [r.rsplit('/', 1)[-1] for r in row['progression']['feature_refs']]
        if level in (3, 6, 10, 14):
            deps += ['warlock-subclass']
        if level == 1:
            deps += ['creation', 'later-entry']
        if level >= 3:
            deps += ['dark-ones-blessing', 'fiend-spells']
        if level >= 6:
            deps += ['dark-ones-own-luck']
        if level >= 10:
            deps += ['fiendish-resilience']
        if level >= 14:
            deps += ['hurl-through-hell']
        add('level-' + str(level), level_fixture(level, row), deps, parent='levels-cohort')
    return tasks


def expected_shared_tasks():
    result = []
    for name, (page, parts) in SPELLS.items():
        for part, fixture in parts.items():
            result.append(dict(key='spell-' + name.lower().replace(' ', '-') + '-' + part,
                spell=name, source_page=page, part=part, fixture=fixture,
                title=f'Resolve {name} {part}',
                body=f'Decision #331; one shared effect implementation PR; Task Size M.\n\nPDF{page}: {fixture}\n\n{CASTING}\n\n{COMMON}\n',
                parent='shared-effects-cohort-' + str(len(result) // 20 + 1),
                type='Task', size='M', issue=None, prerequisites=[], depends=[],
                publication='PENDING complete protocol/overlap/cap approval'))
    return result


def expected_creature_tasks():
    result = []
    def add(name, part, fixture, page):
        result.append(dict(key='creature-' + name.lower().replace(' ', '-') + '-' + part,
            creature=name, part=part, source_page=page, fixture=fixture,
            body=f'Decision #331; one bounded shared implementation PR; Task Size M.\n\nPDF{page}: {fixture}\n\n{COMMON}\n',
            title=f'Resolve {name} {part}', type='Task', size='M',
            parent='shared-creature-cohort-' + str(len(result) // 20 + 1),
            issue=None, prerequisites=[], depends=[], publication='PENDING protocol/whole-owner/cap approval'))
    for name, fixture in NORMAL.items():
        add(name, 'bounded-stat-actions', fixture, SOURCE_PAGES[name])
    for name, parts in {**SPECIAL, **UNDEAD}.items():
        for part, fixture in parts.items():
            add(name, part, fixture, SOURCE_PAGES[name])
    for part, fixture in FAMILIAR.items():
        add('Find Familiar' if part != 'octopus-ink' else 'Octopus', part, fixture, 130 if part != 'octopus-ink' else 357)
    return result


def expected_gear_tasks():
    result = []
    for key, (page, name, fixture) in GEAR.items():
        result.append(dict(key='gear-' + key, item=name, source_page=page,
            fixture=fixture, title='Resolve bounded ' + name,
            body=f'Decision #331; one bounded shared item implementation PR; Task Size M.\n\nPDF{page}: {fixture}\n\n{COMMON}\n',
            parent='shared-gear-cohort', type='Task', size='M', issue=None,
            prerequisites=[], depends=[], publication='PENDING whole-owner/protocol/cap approval'))
    return result


def validate_candidate(m, expected):
    assert type(m['decision']) is int and m['decision'] == 331 and m['base'] == BASE
    assert exact(m['consumer_evidence'], CONSUMER_EVIDENCE), 'independent eight-consumer evidence'
    assert exact(m['transition_witnesses'], TRANSITION_WITNESSES), 'independent unequal class/total/HP numeric witnesses'
    advancement = json.loads((ROOT / 'internal/rules/srd521/testdata/character-advancement.json').read_text())['records']
    assert [r['level'] for r in advancement] == list(range(1, 21)), 'complete canonical285 table'
    for w in m['transition_witnesses']:
        assert sum(w['after'].values()) == w['total_row']
        source = advancement[w['total_row'] - 1]['advancement']
        assert source['experience_points'] == w['experience_points'] and source['proficiency_bonus'] == w['proficiency_bonus']
    records = [r for rs in expected.values() for r in rs]
    assert exact(m['records'], records), 'actual candidate source records'
    for name, rs in expected.items():
        relative = 'internal/rules/srd521/testdata/character-' + name.replace('_', '-') + '.json'
        path = ROOT / relative
        catalogue = json.loads(path.read_text())
        assert exact(catalogue, accepted_catalogue(relative)), 'complete immutable catalogue metadata/schema/source'
        assert exact(catalogue['records'], rs), 'actual catalogue source'
    expected_source = accepted_catalogue('internal/rules/srd521/testdata/character-warlock.json')['source']
    assert exact(m['source'], expected_source) and m['source']['sha256'] == PIN and m['source']['license'] == 'CC-BY-4.0', 'source pin/attribution/adaptation'
    by = {r['id']: r for r in records}
    assert len(by) == 191 and len(m['record_matrix']) == 191
    matrix = []
    for r in records:
        keys = list(dict.fromkeys(r['required_mechanics'] + r['option_keys']))
        matrix.append(dict(record=r, key_owners={k: source_owner(r, k, by) for k in keys},
                           status='unavailable', evidence=m['consumer_evidence']))
    assert exact(m['record_matrix'], matrix), 'complete contextual owner matrix'
    tasks = {t['key']: t for t in m['tasks']}
    expected_rows = {t['key']: t for t in expected_tasks(records)}
    expected_keys = expected_rows.keys()
    assert tasks.keys() == expected_keys, 'complete finite class task roster'
    for key, t in tasks.items():
        assert exact(t, expected_rows[key]), ('independent complete Task contract', key)
        fixture = expected_rows[key]['fixture']
        if key.startswith('level-'):
            level = int(key[6:])
            row = by['srd-5.2.1:character/level/warlock/' + str(level)]
            fixture = level_fixture(level, row)
            assert 'hp' in t['depends'] and 'preparation' in t['depends'] and 'eldritch-invocations' in t['depends'], 'every granting path'
            features = {x.rsplit('/', 1)[-1] for x in row['progression']['feature_refs']}
            assert features <= set(t['depends']), 'all repeated feature grants'
            if level in (3, 6, 10, 14):
                assert 'warlock-subclass' in t['depends'], 'subclass path'
        assert t['fixture'] == fixture and t['body'] == body(key, fixture), ('complete candidate body', key)
        assert t['type'] == 'Task' and t['size'] == 'M' and t['parent'] in ('features-cohort', 'levels-cohort')
        assert t['prerequisites'] == ([968] if key == 'hp' else [430] if key == 'creation' else []), 'owner prerequisites'
    caps({k: v['depends'] for k, v in tasks.items()}, {k: v['parent'] for k, v in tasks.items()})
    assert len(m['spells']['base']) == 72 and len(m['spells']['consumed']) == 117
    assert exact(m['shared_effect_tasks'], expected_shared_tasks()), 'all70 bounded effect bodies/operands'
    assert exact(m['creature_tasks'], expected_creature_tasks()), 'all70 creature/control bodies/operands'
    assert exact(m['gear_tasks'], expected_gear_tasks()), 'all named starting gear/object profiles'
    assert exact(m['familiar_source_roster'], dict(normal=list(NORMAL), chain=list(SPECIAL), undead=list(UNDEAD))), 'finite source form roster'
    assert m['source_conflicts']['Sorcerer–Mind Spike'].startswith('p68 absent; p149 header includes; list admission unavailable.'), 'no automatic header admission'
    assert [x['name'] for x in m['spell_overlap_audit']] == m['spells']['consumed']
    actual_owners = {x['name']: tuple(sorted(t['issue'] for t in x['candidate_owners'])) for x in m['spell_overlap_audit'] if x['candidate_owners']}
    assert all(type(t['issue']) is int for x in m['spell_overlap_audit'] for t in x['candidate_owners']), 'typed existing owner identities'
    assert actual_owners == EXISTING_OWNERS, 'audited whole-owner roster, never name-only matches'
    assert m['publication']['state'] in ('PENDING', 'PUBLISHED'), 'publication state is checked independently against native readback'


def validate_publication(m, receipt, live=False):
    assert live, 'strict proof requires live native readback, never forged published flags'
    import warlock_native_proof as native
    assert native.ACCEPTED_POST_SORCERER_PREFIXES is not None, 'ROOT final prefix release pending'
    assert receipt and receipt.get('registry'), 'PENDING native publication'
    from warlock_approved_native import load_frozen, load_renderer, load_milestone_contract, validate_registry, validate_milestone
    _, frozen, baseline, expected, _, _ = load_frozen()
    load_renderer()
    load_milestone_contract()
    delivery = receipt['delivery']
    raw = Path(delivery['registry_path']).read_bytes()
    registry = validate_milestone(raw, delivery, expected, baseline, receipt['allocation_delivery'],
        Path(delivery['verified_receipt_path']).read_bytes(), Path(receipt['allocation_delivery']['registry_path']).read_bytes())
    assert exact(registry, receipt['registry']), 'retained actual registry matches immutable delivered bytes'
    plan = json.loads(native.PROPOSAL.read_text())
    native.validate_proposal(m, plan, baseline, frozen[330])
    numbers = validate_registry(receipt['registry'], expected, baseline)
    snapshot = native.read_live(numbers, baseline['project'])
    return native.validate_native(m, plan, receipt, snapshot)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--pdf', type=Path, required=True)
    mode = p.add_mutually_exclusive_group()
    mode.add_argument('--source-class-checkpoint', action='store_true')
    mode.add_argument('--allocated-registry', type=Path, help='ROOT-provided complete registry; allocation proof only')
    mode.add_argument('--wired-registry', type=Path, help='ROOT-provided VERIFIED registry; no final prefix seal')
    p.add_argument('--milestone', type=Path, help='ROOT-delivered immutable milestone sidecar')
    p.add_argument('--allocation-evidence', type=Path, help='Retained own allocation readback for VERIFIED identity continuity')
    p.add_argument('--live', action='store_true', help='Read actual native issues/Task/Project Size after ROOT release')
    args = p.parse_args()
    m = json.loads(MANIFEST.read_text())
    expected = source_records(args.pdf)
    validate_candidate(m, expected)
    base = spell_rows(args.pdf, [(75, 0), (75, 310), (76, 0)])
    assert exact(m['spells']['base'], base), 'direct base list'
    allrows = spell_rows(args.pdf, [(p, x) for p in range(30, 84) for x in (0, 310)])
    assert sorted({r['name'] for r in allrows if r['level'] == 0}) == sorted(TOME_CANTRIPS)
    assert sorted({r['name'] for r in allrows if r['level'] == 1 and 'R' in r['special']}) == sorted(TOME_RITUALS)
    consumed = sorted({r['name'] for r in base} | set(TOME_CANTRIPS + TOME_RITUALS + FIEND_SPELLS + NAMED_SPELLS))
    assert m['spells']['consumed'] == consumed, 'full consumed union'
    spells = source_sections(args.pdf, 'docs/investigations/issue-80-spell-item-inventory.json', range(104, 175))
    creatures = source_sections(args.pdf, 'docs/investigations/issue-82-world-inventory.json', range(254, 365))
    validate_source_envelopes(m, spells, creatures)
    verify_gear_source(args.pdf)
    skeleton_attacks = subprocess.check_output(['pdftotext', '-f', '326', '-l', '326', '-x', '0', '-y', '0', '-W', '310', '-H', '792', '-layout', str(args.pdf), '-'], text=True)
    for clause in ('Shortsword. Melee Attack Roll: +5, reach 5 ft.', 'Shortbow. Ranged Attack Roll: +5, range 80/320 ft.', '(1d6 + 3) Piercing damage.'):
        assert clause in skeleton_attacks, 'actual Skeleton326attack clause'
    if args.allocated_registry or args.wired_registry:
        from warlock_native_proof import prove_registry
        assert args.live, 'native milestone proof requires --live actual collection'
        assert args.milestone is not None, 'ROOT-delivered immutable milestone required; no polling'
        output = ROOT / ('docs/investigations/issue-331-allocation-readback.json' if args.allocated_registry else 'docs/investigations/issue-331-native-readback.json')
        assert not output.exists(), 'retain delivered evidence; do not overwrite an existing milestone'
        delivery = json.loads(args.milestone.read_text())
        registry_path = args.allocated_registry or args.wired_registry
        assert registry_path.resolve() == Path(delivery['registry_path']).resolve(), 'use delivered immutable registry_path'
        registry_raw = registry_path.read_bytes()
        prior, receipt_raw, prior_raw = None, None, None
        if args.wired_registry:
            assert args.allocation_evidence is not None, 'retained allocation evidence required'
            prior = json.loads(args.allocation_evidence.read_text())['delivery']
            prior_raw = Path(prior['registry_path']).read_bytes()
            receipt_raw = Path(delivery['verified_receipt_path']).read_bytes()
        result = prove_registry(m, registry_raw, delivery, 'allocation' if args.allocated_registry else 'wiring', prior, receipt_raw, prior_raw)
        with output.open('x', encoding='utf-8') as evidence:
            evidence.write(json.dumps(result, indent=2) + '\n')
        print(json.dumps(result['result'], sort_keys=True))
    elif not args.source_class_checkpoint:
        path = ROOT / 'docs/investigations/issue-331-native-readback.json'
        receipt = json.loads(path.read_text()) if path.exists() else None
        print(json.dumps(validate_publication(m, receipt, live=args.live), sort_keys=True))
    print('PASS source/class contracts only:191 records/588contexts/28Invocations/20rows/117spell identities; actual native proof is separate')


if __name__ == '__main__':
    main()
