#!/usr/bin/env python3
"""Independent SRD 5.2.1 pp66–67 metadata/PDF semantic proof; offline only."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
C = 'srd-5.2.1:character/'
M = 'srd-5.2.1:mechanic/sorcerer/metamagic/'
P = 'srd-5.2.1:planning/metamagic/'
NOTICE = 'Ten Metamagic identities on PDF pp66–67 adapted into unavailable option metadata. Conditions, alternatives, costs and exceptions are opaque keys and numbers; graph and keys are Edda additions. Canonical Sorcerer is composed externally once. No executable support or spell definitions imported.'


def source_records(facts=None):
    f = facts or dict(costs=[1,1,1,1,2,2,1,1,1,1], minimum=1, range_min=5, range_factor=2, touch=30, duration_min=1, duration_factor=2, duration_max=24, targets=1, die=20, effective=1)
    # Independently transcribed semantic clauses; complete-record equality also
    # rejects invented recovery/scaling and changed canonical operands.
    rows = [
        ('careful-spell', ['when-cast-forces-other-creatures-saving-throw', 'choose-up-to-charisma-modifier-creatures', 'chosen-auto-succeed-save', 'chosen-no-damage-only-if-normally-half-on-success'], {'minimum-creatures':f['minimum']}, []),
        ('distant-spell', ['when-cast', 'choose-one-range-alternative'], {'minimum-range-feet':f['range_min'], 'range-multiplier':f['range_factor'], 'touch-range-feet':f['touch']}, ['range-at-least-minimum-double', 'range-touch-becomes-fixed-feet']),
        ('empowered-spell', ['when-roll-spell-damage', 'reroll-up-to-charisma-modifier-damage-dice', 'must-use-new-rolls', 'may-use-after-different-metamagic-during-casting'], {'minimum-dice':f['minimum']}, []),
        ('extended-spell', ['when-cast-duration-at-least-minimum', 'double-duration-up-to-maximum', 'if-concentration-advantage-all-saves-to-maintain-it'], {'minimum-duration-minutes':f['duration_min'], 'duration-multiplier':f['duration_factor'], 'maximum-duration-hours':f['duration_max']}, []),
        ('heightened-spell', ['when-cast-forces-creature-saving-throw', 'one-spell-target-disadvantage-all-saves-against-spell'], {'target-count':f['targets']}, []),
        ('quickened-spell', ['when-cast-action-casting-time', 'change-to-bonus-action-for-this-casting', 'cannot-if-already-cast-level-one-plus-spell-current-turn', 'cannot-cast-level-one-plus-spell-after-use-this-turn'], {}, []),
        ('seeking-spell', ['when-spell-attack-roll-misses', 'reroll-attack-d20', 'must-use-new-roll', 'may-use-after-different-metamagic-during-casting'], {'die-sides':f['die']}, []),
        ('subtle-spell', ['when-cast', 'omit-verbal-somatic-material-components', 'retain-consumed-material-components', 'retain-material-components-with-specified-cost'], {}, []),
        ('transmuted-spell', ['when-cast-dealing-listed-damage-type', 'change-that-type-to-one-other-listed-type'], {}, ['acid', 'cold', 'fire', 'lightning', 'poison', 'thunder']),
        ('twinned-spell', ['when-cast-spell-upcastable-with-higher-slot-to-target-additional-creature', 'increase-effective-spell-level-not-slot-level'], {'effective-level-increase':f['effective']}, []),
    ]
    out = []
    for i, (slug, clauses, nums, options) in enumerate(rows):
        def rec(kind, suffix, name, parent, keys, numbers=None):
            r = dict(id=C+kind+'/sorcerer/metamagic/'+suffix, kind=kind, name=name,
                     sources=[dict(inventory_id=P+slug, page=66 if i==0 else 67)], parent_id=parent,
                     level=0, choice_refs=[], resource_refs=[], prerequisite_refs=[], required_mechanics=[M+slug+'/'+k for k in keys], select=0, option_keys=[], capacity=0)
            if numbers:
                r['parameters'] = {M+slug+'/'+k:v for k,v in numbers.items()}
                r['required_mechanics'] += list(r['parameters'])
            out.append(r)
            return r
        name = slug.replace('-', ' ').title()
        nums = {'sorcery-point-cost':f['costs'][i], **nums}
        r = rec('feature', slug, name, C+'choice/sorcerer/metamagic', clauses, nums)
        q = rec('prerequisite', slug+'-owners', name+' canonical owners', r['id'], ['chosen-option-spend-canonical-points-modify-own-cast-spell'])
        refs = ['choice/sorcerer/metamagic', 'resource/sorcerer/sorcery-points', 'feature/sorcerer/spellcasting', 'choice/sorcerer/prepared-spells']
        if slug == 'twinned-spell':
            refs += ['resource/sorcerer/spell-slot-'+str(n) for n in range(1,10)]
        q['requirement'] = dict(key=q['required_mechanics'][0], record_refs=[C+x for x in refs], parameters={})
        r['prerequisite_refs'].append(q['id'])
        if options:
            q = rec('choice', slug+'-alternatives', name+' alternatives', r['id'], ['choose-alternative'])
            q['select'] = 1
            q['option_keys'] = [M+slug+'/'+x for x in options]
            r['choice_refs'].append(q['id'])
    return out


def compare(records, facts=None):
    assert records == source_records(facts), 'Metamagic semantic graph differs from pp66–67'


def verify(pdf, candidate=None):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, 'PDF pin'
    # Reading order: p66 right column, p67 left then right. No spell list.
    def column(page, x):
        raw = subprocess.check_output(['pdftotext', '-f', str(page), '-l', str(page), '-x', str(x), '-y', '0', '-W', '307', '-H', '792', '-layout', str(pdf), '-'], text=True)
        return ' '.join(re.sub(r'-\n\s*', '', raw.replace('\u00ad', '')).split())
    text = column(66,305).split('Careful Spell',1)[1] + column(67,0) + column(67,305).split('Sorcerer Spell List',1)[0]
    costs = [int(x) for x in re.findall(r'Cost: (\d+) Sorcery Point', text)]
    def num(pattern):
        m = re.search(pattern, text)
        assert m, pattern
        return int(m[1])
    facts = dict(costs=costs, minimum=1, range_min=num(r'range of at least (\d+)'), range_factor=2, touch=num(r'range (\d+) feet'), duration_min=1, duration_factor=2, duration_max=num(r'maximum duration of (\d+) hours'), targets=1, die=num(r'reroll the d(\d+)'), effective=num(r'effective level by (\d+)'))
    # Clauses whose source numbers are words and the nonnumeric predicates.
    for clause in [
        'up to your Charisma modifier (minimum of one creature)', 'automatically succeeds on its saving throw', 'takes no damage if it would normally take half damage on a successful save',
        'double the spell’s range', 'range of Touch', 'When you roll damage for a spell', 'up to your Charisma modifier (minimum of one)', 'must use the new rolls',
        'duration of 1 minute or longer', 'double its duration', 'Advantage on any saving throw you make to maintain that Concentration',
        'one target of the spell Disadvantage on saves against the spell', 'casting time of an action', 'Bonus Action for this casting', 'already cast a level 1+ spell on the current turn', 'nor can you cast a level 1+ spell on this turn after modifying',
        'attack roll for a spell and miss', 'must use the new roll', 'without any Verbal, Somatic, or Material components', 'consumed by the spell', 'cost specified in the spell',
        'one of the other listed types: Acid, Cold, Fire, Lightning, Poison, Thunder', 'higher-level spell slot to target an additional creature',
    ]:
        assert clause in text, clause
    assert text.count('even if you’ve already used a different Metamagic option during the casting of the spell') == 2
    batch = candidate or json.loads((ROOT/'internal/rules/srd521/testdata/character-metamagic.json').read_text())
    inventory = json.loads((ROOT/'docs/investigations/issue-74-character-inventory.json').read_text())
    expected_source = dict(inventory['source'], changes=NOTICE)
    assert batch.keys() == {'schema_version','source','records'}
    assert batch['schema_version'] == 1 and batch['source'] == expected_source
    # Check canonical owners independently, then compose each catalogue once.
    import verify_srd_sorcerer as sorcerer
    canonical = json.loads(sorcerer.DATA.read_text())
    sorcerer.compare(canonical['records'], sorcerer.source_records(pdf))
    manifest = json.loads((ROOT/'internal/rules/srd521/testdata/character-metamagic-composition.json').read_text())
    assert manifest == {'catalogues':['character-sorcerer.json','character-metamagic.json']}
    combined = canonical['records'] + batch['records']
    ids = {r['id']:r for r in combined}
    assert len(ids) == len(combined), 'duplicate canonical composition'
    for r in batch['records']:
        if 'requirement' in r:
            assert all(x in ids for x in r['requirement']['record_refs'])
    compare(batch['records'], facts)
    inv = {r['id']:r for r in inventory['records']}
    for r in batch['records']:
        assert inv[r['sources'][0]['inventory_id']]['source_page'] == r['sources'][0]['page']
    assert len({r['sources'][0]['inventory_id'] for r in batch['records']}) == 10
    print('Metamagic: ten identities, costs, conditions, alternatives, exceptions and canonical operands match pinned PDF')


if __name__ == '__main__':
    if sys.argv[1] == '--records':
        compare(json.loads(Path(sys.argv[2]).read_text()))
    elif sys.argv[1] == '--controls':
        controls = json.loads(Path(sys.argv[2]).read_text())
        for records in controls:
            try:
                compare(records)
            except AssertionError:
                continue
            raise AssertionError('contrary graph accepted')
        print(f'{len(controls)} contrary graphs rejected')
    else:
        verify(Path(sys.argv[1]), json.loads(Path(sys.argv[2]).read_text()) if len(sys.argv)>2 else None)
