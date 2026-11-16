#!/usr/bin/env python3
"""Independent p74 invocation semantic comparator. No Go, network or runtime."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
C = 'srd-5.2.1:character/'
M = 'srd-5.2.1:mechanic/warlock/invocations/'
NOTICE = 'Four Invocation identities on PDF p74 adapted into unavailable metadata with exact prerequisites, nonrepeatability, spell-slot exceptions, pact-weapon Attack action timing and Truesight range. Graph and mechanic keys are Edda additions. Canonical Warlock is external; Pact of the Blade and spells have separate owners. No executable support imported.'


def source_records(levels=(5, 9, 7, 15), attacks=2, distance=30):
    out = []
    names = ('Thirsting Blade', 'Visions of Distant Realms', 'Whispers of the Grave', 'Witch Sight')
    for name, level in zip(names, levels):
        slug = name.lower().replace(' ', '-')
        def record(kind, suffix, label, parent, keys):
            r = dict(id=C+kind+'/warlock/invocations/'+suffix, kind=kind, name=label,
                     sources=[dict(inventory_id='srd-5.2.1:planning/invocations/'+slug, page=74)],
                     parent_id=parent, level=0, choice_refs=[], resource_refs=[], prerequisite_refs=[],
                     required_mechanics=[M+k for k in keys], select=0, option_keys=[], capacity=0)
            out.append(r)
            return r
        f = record('feature', slug, name, C+'choice/warlock/invocations', [slug, 'not-repeatable'])
        def predicate(suffix, key, operands, numbers):
            r = record('prerequisite', slug+'-'+suffix, name+' '+suffix, f['id'], [key])
            r['requirement'] = dict(key=M+key, record_refs=[C+x for x in operands], parameters=numbers)
            f['prerequisite_refs'].append(r['id'])
        predicate('level', 'warlock-minimum-level', ['class/warlock'], {'minimum-level':level})
        predicate('selection', 'requires-canonical-invocation-selection', ['choice/warlock/invocations'], {})
        if slug == 'thirsting-blade':
            predicate('pact', 'requires-pact-of-the-blade-invocation', ['feature/warlock/invocations/pact-of-the-blade'], {})
            f['required_mechanics'] += [M+x for x in ('thirsting-blade/extra-attack-for-pact-weapon-only', 'thirsting-blade/when-take-attack-action-on-your-turn', 'thirsting-blade/attacks-instead-of-once')]
            f['parameters'] = {M+'thirsting-blade/attacks-instead-of-once':attacks}
        elif slug in ('visions-of-distant-realms', 'whispers-of-the-grave'):
            spell = 'arcane-eye' if slug == 'visions-of-distant-realms' else 'speak-with-dead'
            f['required_mechanics'] += ['srd-5.2.1:mechanic/spells/'+spell, M+slug+'/can-cast', M+slug+'/without-expending-spell-slot']
        else:
            f['required_mechanics'] += [M+'witch-sight/truesight', M+'witch-sight/range-feet']
            f['parameters'] = {M+'witch-sight/range-feet':distance}
    return out


def compare(records, **facts):
    expected = source_records(**facts)
    assert records == expected, 'p74 complete semantic graph mismatch'


def verify(pdf, candidate=None):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, 'PDF pin'
    raw = subprocess.check_output(['pdftotext', '-f', '74', '-l', '74', '-x', '305', '-y', '0', '-W', '307', '-H', '792', '-layout', str(pdf), '-'], text=True)
    text = ' '.join(re.sub(r'-\n\s*', '', raw.replace('\u00ad', '')).split())
    levels = []
    for name in ('Thirsting Blade', 'Visions of Distant Realms', 'Whispers of the Grave', 'Witch Sight'):
        match = re.search(re.escape(name)+r' Prerequisite: Level (\d+)\+ Warlock', text)
        assert match, name
        levels.append(int(match[1]))
    clauses = ('Pact of the Blade Invocation', 'Extra Attack feature for your pact weapon only',
               'attack twice with the weapon instead of once', 'take the Attack action on your turn',
               'cast Arcane Eye without expending a spell slot',
               'cast Speak with Dead without expending a spell slot',
               'You have Truesight with a range of 30 feet.')
    assert all(clause in text for clause in clauses), [c for c in clauses if c not in text]
    section = text[text.index('Thirsting Blade'):text.index('Warlock Spell List')]
    assert 'Repeatable' not in section
    distance = int(re.search(r'Truesight with a range of (\d+) feet', section)[1])
    actual = json.loads((candidate or ROOT/'internal/rules/srd521/testdata/character-invocations-03.json').read_text())
    inventory = json.loads((ROOT/'docs/investigations/issue-74-character-inventory.json').read_text())
    source = dict(inventory['source'], changes=NOTICE)
    assert set(actual) == {'schema_version', 'source', 'records'}
    assert actual['schema_version'] == 1 and actual['source'] == source
    compare(actual['records'], levels=levels, attacks=2, distance=distance)
    expected = {r['id']:r['source_page'] for r in inventory['records'] if r['batch_key'] == 'options-invocations-03'}
    identities = [r for r in actual['records'] if r['kind'] == 'feature']
    assert len(identities) == len(expected) == 4
    assert {s['inventory_id']:s['page'] for r in identities for s in r['sources']} == expected
    print('p74: four identities and complete metadata match pinned PDF facts')


if __name__ == '__main__':
    if len(sys.argv) == 3 and sys.argv[1] == '--records':
        compare(json.loads(Path(sys.argv[2]).read_text()))
        print('valid p74 graph accepted')
    elif len(sys.argv) == 3 and sys.argv[1] == '--controls':
        controls = json.loads(Path(sys.argv[2]).read_text())
        for records in controls:
            try:
                compare(records)
            except AssertionError:
                continue
            raise AssertionError('contrary graph accepted')
        print(f'{len(controls)} contrary graphs rejected')
    elif len(sys.argv) in (2, 3):
        verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) == 3 else None)
    else:
        raise SystemExit('usage: verify_srd_invocations_03.py PDF [CATALOGUE] | --records JSON | --controls JSON')
