#!/usr/bin/env python3
"""Independent offline p64 semantic metadata comparison; no Go or network."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
C = 'srd-5.2.1:character/'
M = 'srd-5.2.1:mechanic/rogue/thief/'
P = 'srd-5.2.1:planning/'


def source_records(facts=None):
    # Independently transcribed source clauses. PDF mode supplies extracted numbers.
    f = facts or dict(fast=3, work=3, sneak=9, device=13, reflex=17,
                     cost=1, sides=6, attune=4, charge_die=6, charge_success=6,
                     reliable=1, dc=10, turns=2, round=1, initiative=10)
    out = []
    def add(kind, slug, name, parent, source, keys=(), nums=None):
        r = dict(id=C+kind+'/rogue/thief/'+slug, kind=kind, name=name,
                 sources=[dict(inventory_id=P+source, page=64)], parent_id=parent,
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
    sub = add('subclass', 'identity', 'Thief', C+'class/rogue', 'subclasses/rogue-thief', ['subclass'])
    pred(sub, 'level', 'class-level', ['class/rogue'], {'minimum-level':f['fast']})
    pred(sub, 'selection', 'requires-canonical-thief-subclass-selection', ['choice/rogue/rogue-subclass'])
    def feature(slug, name, level, keys=(), nums=None):
        r = add('feature', slug, name, sub['id'], f'subclass-features/rogue-level-{level}-{slug}', keys, nums)
        pred(r, 'level', 'class-level', ['class/rogue'], {'minimum-level':level})
        return r
    def choice(owner, slug, name, keys):
        r = add('choice', slug, name, owner['id'], owner['sources'][0]['inventory_id'][len(P):], ['choose-one/'+slug])
        r['select'] = 1
        r['option_keys'] = [M+k for k in keys]
        owner['choice_refs'].append(r['id'])
        return r
    fast = feature('fast-hands', 'Fast Hands', f['fast'], ['fast-hands/bonus-action'])
    action = choice(fast, 'fast-hands', 'Fast Hands options', ['fast-hands/sleight-of-hand', 'fast-hands/use-an-object'])
    hand = choice(action, 'sleight-of-hand', 'Sleight of Hand targets', ['sleight-of-hand/pick-lock-with-thieves-tools', 'sleight-of-hand/disarm-trap-with-thieves-tools', 'sleight-of-hand/pick-pocket'])
    pred(hand, 'check', 'sleight-of-hand/dexterity-sleight-of-hand-check', ['choice/rogue/thief/sleight-of-hand'])
    obj = choice(action, 'use-an-object', 'Use an Object actions', ['use-an-object/utilize-action', 'use-an-object/magic-action-use-magic-item-requiring-that-action'])
    pred(obj, 'selection', 'fast-hands/conditional-on-use-an-object-option', ['choice/rogue/thief/fast-hands'])
    pred(hand, 'selection', 'fast-hands/conditional-on-sleight-of-hand-option', ['choice/rogue/thief/fast-hands'])
    work = feature('second-story-work', 'Second-Story Work', f['work'], ['climber/climb-speed-equals-speed', 'jumper/may-use-dexterity-instead-of-strength-for-jump-distance'])
    sneak = feature('supreme-sneak', 'Supreme Sneak', f['sneak'], ['supreme-sneak/add-cunning-strike-option'])
    strike = choice(sneak, 'stealth-attack', 'Stealth Attack', ['stealth-attack/retain-hide-invisible-if-end-turn-behind-required-cover'])
    pred(strike, 'strike', 'stealth-attack/extends-canonical-cunning-strike-pay-sneak-attack-dice-before-roll', ['choice/rogue/cunning-strike', 'feature/rogue/sneak-attack'], {'cost-dice':f['cost'], 'die-sides':f['sides']})
    pred(strike, 'hide', 'stealth-attack/requires-hide-actions-invisible-condition-on-self', ['choice/rogue/cunning-action'])
    cover = choice(strike, 'stealth-cover', 'Stealth Attack end-turn cover', ['stealth-attack/end-turn-behind-three-quarters-cover', 'stealth-attack/end-turn-behind-total-cover'])
    pred(cover, 'timing', 'stealth-attack/this-attack-does-not-end-that-condition-only-if-end-turn-behind-cover', ['choice/rogue/thief/stealth-attack'])
    device = feature('use-magic-device', 'Use Magic Device', f['device'], [
        'attunement/up-to-maximum-magic-items-at-once',
        'charges/whenever-using-magic-item-property-that-expends-charges-roll-die',
        'charges/on-success-use-property-without-expending-charges',
        'scrolls/any-spell-scroll-intelligence-spellcasting-ability'],
        {'attunement/maximum-items':f['attune'], 'charges/dice-count':1, 'charges/die-sides':f['charge_die'], 'charges/success-roll':f['charge_success']})
    scroll = choice(device, 'scrolls', 'Spell Scroll level cases', ['scrolls/cantrip-or-level-one-cast-reliably', 'scrolls/higher-level-requires-intelligence-arcana-check-dc-base-plus-spell-level'])
    pred(scroll, 'level', 'scrolls/reliable-through-level-otherwise-higher-level-check', ['choice/rogue/thief/scrolls'], {'reliable-maximum-spell-level':f['reliable'], 'check-dc-base':f['dc']})
    result = choice(scroll, 'scroll-check', 'Higher-level scroll check outcomes', ['scrolls/success-cast-spell-from-scroll', 'scrolls/failure-scroll-disintegrates'])
    pred(result, 'condition', 'scrolls/outcomes-only-for-higher-level-scroll-check', ['choice/rogue/thief/scrolls'])
    feature('thiefs-reflexes', 'Thief’s Reflexes', f['reflex'], ['reflexes/may-take-two-turns-first-round-any-combat', 'reflexes/first-turn-normal-initiative', 'reflexes/second-turn-own-initiative-minus-offset'], {'reflexes/turn-count':f['turns'], 'reflexes/combat-round':f['round'], 'reflexes/second-turn-initiative-offset':f['initiative']})
    return out


def compare(records, facts=None):
    expected = source_records(facts)
    assert len(records) == len(expected), 'p64 record count'
    for got, want in zip(records, expected):
        assert got == want, f"p64 semantic mismatch {want['id']}: {got} != {want}"


def verify(pdf, candidate=None):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, 'PDF pin'
    raw = subprocess.check_output(['pdftotext', '-f', '64', '-l', '64', '-x', '0', '-y', '0', '-W', '305', '-H', '792', '-layout', str(pdf), '-'], text=True)
    text = ' '.join(re.sub(r'-\n\s*', '', raw.replace('\u00ad', '')).split())
    def number(pattern):
        match = re.search(pattern, text)
        assert match, pattern
        return int(match[1])
    facts = {k:number(r'Level (\d+): '+heading) for k,heading in [('fast','Fast Hands'), ('work','Second-Story Work'), ('sneak','Supreme Sneak'), ('device','Use Magic Device'), ('reflex','Thief’s Reflexes')]}
    facts.update(cost=number(r'Stealth Attack \(Cost: (\d+)d6\)'), sides=number(r'Stealth Attack \(Cost: 1d(\d+)\)'), charge_die=number(r'charges, roll 1d(\d+)'), charge_success=number(r'On a roll of (\d+)'), reliable=number(r'cantrip or a level (\d+) spell'), dc=number(r'DC (\d+) plus'), initiative=number(r'Initiative minus (\d+)'))
    for key, phrase, word, value in [('attune','up to four magic items at once','four',4), ('turns','two turns during the first round','two',2), ('round','two turns during the first round','first',1)]:
        assert phrase in text and word in phrase
        facts[key] = value
    clauses = ['Rogue Subclass: Thief', 'As a Bonus Action', 'one of the following', 'Dexterity (Sleight of Hand) check', 'pick a lock or disarm a trap with Thieves’ Tools or to pick a pocket', 'Take the Utilize action, or take the Magic action to use a magic item that requires that action', 'Climb Speed equal to your Speed', 'jump distance using your Dexterity rather than your Strength', 'following Cunning Strike option', 'Hide action’s Invisible condition', 'this attack doesn’t end that condition on you', 'end the turn behind Three-Quarters Cover or Total Cover', 'Whenever you use a magic item property that expends charges', 'use the property without expending the charges', 'any Spell Scroll', 'Intelligence as your spellcasting ability', 'cast it reliably', 'higher-level spell', 'first succeed on an Intelligence (Arcana) check', 'plus the spell’s level', 'On a successful check, you cast the spell from the scroll', 'On a failed check, the scroll disintegrates', 'first round of any combat', 'first turn at your normal Initiative', 'second turn at your Initiative minus']
    assert all(x in text for x in clauses), [x for x in clauses if x not in text]
    path = candidate or ROOT/'internal/rules/srd521/testdata/character-thief.json'
    actual = json.loads(path.read_text())
    baseline = json.loads((ROOT/'internal/rules/srd521/testdata/character-rogue.json').read_text())
    assert set(actual) == {'schema_version', 'source', 'records'}
    baseline['source']['changes'] = 'Thief subclass and five feature identities on PDF p64 adapted into unavailable offline metadata with nested choices, prerequisites, operands, costs, timing and limits. Mechanic keys and graph are Edda additions. Canonical Rogue is composed externally; spell and item definitions remain opaque. No executable support imported.'
    assert actual['schema_version'] == 1 and actual['source'] == baseline['source']
    compare(actual['records'], facts)
    inv = json.loads((ROOT/'docs/investigations/issue-74-character-inventory.json').read_text())
    expected = {r['id']:r['source_page'] for r in inv['records'] if r['batch_key'] == 'subclass-rogue'}
    identities = [r for r in actual['records'] if r['kind'] in ('subclass', 'feature')]
    assert len(identities) == len(expected) == 6
    assert {s['inventory_id']:s['page'] for r in identities for s in r['sources']} == expected
    print(f'p64: six identities, {len(actual["records"])} complete records match independent PDF facts')


if __name__ == '__main__':
    if len(sys.argv) == 3 and sys.argv[1] == '--records':
        compare(json.loads(Path(sys.argv[2]).read_text()))
        print('valid p64 graph accepted')
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
        raise SystemExit('usage: verify_srd_thief.py PDF [CATALOGUE] | --controls JSON')
