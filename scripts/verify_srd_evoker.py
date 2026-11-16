#!/usr/bin/env python3
"""Independent offline p82 semantic metadata comparison; no Go or network."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
C = 'srd-5.2.1:character/'
M = 'srd-5.2.1:mechanic/wizard/evoker/'
P = 'srd-5.2.1:planning/'


def source_records(facts=None):
    # Independent transcription of the right column of pinned PDF p82.
    f = facts or dict(savant=3, potent=3, sculpt=6, empowered=10, over=14,
                     initial=2, maximum=2, later=1, half=2, base=1, rolls=1,
                     slot_min=1, slot_max=5, dice=2, sides=12, increase=1)
    out = []
    def add(kind, slug, name, parent, source, keys=(), nums=None):
        r = dict(id=C+kind+'/wizard/evoker/'+slug, kind=kind, name=name,
                 sources=[dict(inventory_id=P+source, page=82)], parent_id=parent,
                 level=0, choice_refs=[], resource_refs=[], prerequisite_refs=[],
                 required_mechanics=[M+k for k in keys], select=0, option_keys=[], capacity=0)
        if nums:
            r['parameters'] = {M+k:v for k,v in nums.items()}
            r['required_mechanics'] += list(r['parameters'])
        out.append(r)
        return r
    def pred(owner, suffix, key, refs, nums=None):
        r = add('prerequisite', owner['id'].split('/')[-1]+'-'+suffix,
                owner['name']+' '+suffix, owner['id'], owner['sources'][0]['inventory_id'][len(P):], [key])
        r['requirement'] = dict(key=M+key, record_refs=[C+x for x in refs], parameters=nums or {})
        owner['prerequisite_refs'].append(r['id'])
    sub = add('subclass', 'identity', 'Evoker', C+'class/wizard', 'subclasses/wizard-evoker', ['subclass'])
    pred(sub, 'level', 'class-level', ['class/wizard'], {'minimum-level':f['savant']})
    pred(sub, 'selection', 'requires-canonical-evoker-subclass-selection', ['choice/wizard/wizard-subclass'])
    def feature(slug, name, level, keys=(), nums=None):
        r = add('feature', slug, name, sub['id'], f'subclass-features/wizard-level-{level}-{slug}', keys, nums)
        pred(r, 'level', 'class-level', ['class/wizard'], {'minimum-level':level})
        return r
    def choice(owner, slug, name, count, options):
        r = add('choice', slug, name, owner['id'], owner['sources'][0]['inventory_id'][len(P):], ['choose/'+slug])
        r['select'] = count
        r['option_keys'] = [M+x for x in options]
        owner['choice_refs'].append(r['id'])
        return r
    savant = feature('evocation-savant', 'Evocation Savant', f['savant'], ['savant/add-chosen-spells-to-spellbook-for-free'])
    initial = choice(savant, 'savant-initial', 'Initial Evocation spells', f['initial'], ['savant/wizard-evocation-spells-no-higher-than-level-two'])
    pred(initial, 'book', 'savant/initial-wizard-evocation-spells-maximum-level', ['choice/wizard/spellbook-initial'], {'maximum-spell-level':f['maximum']})
    later = choice(savant, 'savant-new-slots', 'New spell-slot level Evocation spell', f['later'], ['savant/wizard-evocation-spell-at-a-level-with-slots'])
    pred(later, 'trigger', 'savant/whenever-gain-access-to-new-slot-level-in-this-class', ['feature/wizard/spellcasting', 'choice/wizard/spellbook-level-gain', 'class/wizard'])
    potent = feature('potent-cantrip', 'Potent Cantrip', f['potent'], ['potent/when-cast-damaging-cantrip-at-creature', 'potent/target-takes-half-cantrip-damage-if-any', 'potent/no-additional-cantrip-effect'], {'potent/damage-divisor':f['half']})
    choice(potent, 'potent-trigger', 'Alternative Potent Cantrip triggers', 1, ['potent/miss-with-attack-roll', 'potent/target-succeeds-on-save-against-cantrip'])
    sculpt = feature('sculpt-spells', 'Sculpt Spells', f['sculpt'], ['sculpt/when-cast-evocation-affecting-other-creatures-you-can-see', 'sculpt/may-choose-number-of-them-equal-to-base-plus-spell-level', 'sculpt/chosen-creatures-automatically-succeed-on-saves-against-spell', 'sculpt/no-damage-only-if-normally-half-on-successful-save'], {'sculpt/creature-count-base':f['base']})
    empowered = feature('empowered-evocation', 'Empowered Evocation', f['empowered'], ['empowered/whenever-cast-wizard-evocation-spell', 'empowered/may-add-intelligence-modifier-to-one-damage-roll-of-that-spell'], {'empowered/damage-roll-count':f['rolls']})
    pred(empowered, 'ability', 'empowered/canonical-intelligence-spellcasting-ability', ['feature/wizard/spellcasting'])
    over = feature('overchannel', 'Overchannel', f['over'], ['overchannel/when-cast-damaging-wizard-spell-with-qualifying-slot', 'overchannel/may-deal-maximum-spell-damage-on-turn-you-cast-it', 'overchannel/first-use-no-adverse-effect', 'overchannel/reuse-before-finish-long-rest-necrotic-self-damage', 'overchannel/initial-reuse-dice-per-spell-slot-level', 'overchannel/immediately-after-casting', 'overchannel/ignores-resistance-and-immunity', 'overchannel/each-further-use-before-finish-long-rest-increases-damage-per-spell-level'], {'overchannel/minimum-slot-level':f['slot_min'], 'overchannel/maximum-slot-level':f['slot_max'], 'overchannel/initial-reuse-dice':f['dice'], 'overchannel/die-sides':f['sides'], 'overchannel/further-use-dice-increase':f['increase']})
    pred(over, 'slots', 'overchannel/uses-canonical-wizard-slot-owner-not-prepared-count-or-character-level', ['feature/wizard/spellcasting', 'class/wizard'])
    return out


def compare(records, facts=None):
    expected = source_records(facts)
    assert len(records) == len(expected), 'p82 record count'
    for got, want in zip(records, expected):
        assert got == want, f"p82 semantic mismatch {want['id']}: {got} != {want}"


def verify(pdf, candidate=None):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, 'PDF pin'
    raw = subprocess.check_output(['pdftotext', '-f', '82', '-l', '82', '-x', '305', '-y', '0', '-W', '307', '-H', '792', '-layout', str(pdf), '-'], text=True)
    text = ' '.join(re.sub(r'-\n\s*', '', raw.replace('\u00ad', '')).split())
    def number(pattern):
        match = re.search(pattern, text)
        assert match, pattern
        return int(match[1])
    facts = {key:number(r'Level (\d+): '+name) for key,name in [('savant','Evocation Savant'), ('potent','Potent Cantrip'), ('sculpt','Sculpt Spells'), ('empowered','Empowered Evocation'), ('over','Overchannel')]}
    facts.update(maximum=number(r'no higher than level (\d+)'), slot_min=number(r'spell slot of levels (\d+)–5'), slot_max=number(r'spell slot of levels 1–(\d+)'), dice=number(r'take (\d+)d12 Necrotic'), sides=number(r'take 2d(\d+) Necrotic'), increase=number(r'increases by (\d+)d12'), base=number(r'equal to (\d+) plus'))
    # Written numbers and the complete qualifying clauses are independently
    # mapped to graph fields; comparison below checks every actual record field.
    clauses = [
        'Choose two Wizard spells from the Evocation school',
        'add them to your spellbook for free',
        'whenever you gain access to a new level of spell slots in this class',
        'add one Wizard spell from the Evocation school to your spellbook for free',
        'chosen spell must be of a level for which you have spell slots',
        'When you cast a cantrip at a creature and you miss with the attack roll or the target succeeds on a saving throw against the cantrip',
        'target takes half the cantrip’s damage (if any) but suffers no additional effect from the cantrip',
        'When you cast an Evocation spell that affects other creatures that you can see',
        'you can choose a number of them equal to 1 plus the spell’s level',
        'chosen creatures automatically succeed on their saving throws against the spell',
        'take no damage if they would normally take half damage on a successful save',
        'Whenever you cast a Wizard spell from the Evocation school',
        'you can add your Intelligence modifier to one damage roll of that spell',
        'When you cast a Wizard spell with a spell slot of levels 1–5 that deals damage',
        'you can deal maximum damage with that spell on the turn you cast it',
        'first time you do so, you suffer no adverse effect',
        'use this feature again before you finish a Long Rest',
        'take 2d12 Necrotic damage for each level of the spell slot immediately after you cast it',
        'This damage ignores Resistance and Immunity',
        'Each time you use this feature again before finishing a Long Rest',
        'Necrotic damage per spell level increases by 1d12',
    ]
    assert all(x in text for x in clauses), [x for x in clauses if x not in text]
    facts.update(initial=2, later=1, half=2, rolls=1)
    path = candidate or ROOT/'internal/rules/srd521/testdata/character-evoker.json'
    actual = json.loads(path.read_text())
    inventory = json.loads((ROOT/'docs/investigations/issue-74-character-inventory.json').read_text())
    source = inventory['source'].copy()
    source['changes'] = 'Evoker and five feature identities on PDF p82 adapted into unavailable offline metadata with spell choices, conditions, operands, numbers, timing and escalating costs. Mechanic keys and graph are Edda additions. Canonical Wizard is composed externally; adjacent spells remain opaque. No executable support imported.'
    assert set(actual) == {'schema_version', 'source', 'records'}
    assert actual['schema_version'] == 1 and actual['source'] == source
    compare(actual['records'], facts)
    expected = {r['id']:r['source_page'] for r in inventory['records'] if r['batch_key'] == 'subclass-wizard'}
    identities = [r for r in actual['records'] if r['kind'] in ('subclass', 'feature')]
    assert len(identities) == len(expected) == 6
    assert {s['inventory_id']:s['page'] for r in identities for s in r['sources']} == expected
    print(f'p82: six identities, {len(actual["records"])} complete records match independent PDF facts')


if __name__ == '__main__':
    if len(sys.argv) == 3 and sys.argv[1] == '--records':
        compare(json.loads(Path(sys.argv[2]).read_text()))
        print('valid p82 graph accepted')
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
        raise SystemExit('usage: verify_srd_evoker.py PDF [CATALOGUE] | --controls JSON')
