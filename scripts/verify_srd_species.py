#!/usr/bin/env python3
"""Offline SRD 5.2.1 species comparator. No network or runtime registration."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
ROOT = Path(__file__).resolve().parents[1]
M = 'srd-5.2.1:mechanic/'


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, 'PDF pin'
    pages = {}
    for page in (84, 85, 86):
        columns = [subprocess.check_output(['pdftotext', '-f', str(page), '-l', str(page),
                   '-x', str(x), '-y', '0', '-W', '310', '-H', '792', '-layout', str(pdf), '-'], text=True)
                   for x in (0, 306)]
        pages[page] = re.sub(r'\s+', ' ', ' '.join(columns)).replace('\u200b', '').replace('’', "'")
    c = json.loads((ROOT / 'internal/rules/srd521/testdata/character-species.json').read_text())
    assert c['source']['sha256'] == PIN and c['source']['license'] == 'CC-BY-4.0'
    origins = {r['name'].lower(): r for r in c['records'] if r['kind'] == 'origin'}
    expected = dict(dragonborn=84, dwarf=84, elf=84, gnome=85, goliath=85,
                    halfling=86, human=86, orc=86, tiefling=86)
    assert set(origins) == set(expected)
    inv = json.loads((ROOT / 'docs/investigations/issue-74-character-inventory.json').read_text())
    assert {r['id']: r['source_page'] for r in inv['records'] if r.get('catalogue_issue') == 307} == {
        'srd-5.2.1:planning/species/' + name: page for name, page in expected.items()}
    for name, page in expected.items():
        assert origins[name]['sources'] == [dict(inventory_id='srd-5.2.1:planning/species/' + name, page=page)]
        assert name.title() in pages[page]
    traits = {name: {t['key'].removeprefix(M + 'species/' + name + '/'): t
                     for t in r['species']['traits']} for name, r in origins.items()}
    # Independent source-number transcription, including explicit zero operands.
    numbers = {
        'dragonborn': {'darkvision': {'range-feet': 60}, 'breath-shape/cone': {'length-feet': 15},
            'breath-shape/line': {'length-feet': 30, 'width-feet': 5},
            'breath-weapon': {'dc-base': 8, 'damage-dice': 1, 'damage-die': 10},
            'draconic-flight': {'minimum-character-level': 5, 'duration-minutes': 10}},
        'dwarf': {'darkvision': {'range-feet': 120}, 'dwarven-toughness': {'initial-increase': 1, 'increase-per-character-level': 1},
            'stonecunning': {'range-feet': 60, 'duration-minutes': 10}},
        'elf': {'darkvision': {'range-feet': 60}, 'drow-darkvision': {'range-feet': 120},
            'wood-elf-speed': {'speed-feet': 35}, 'trance': {'long-rest-hours': 4}},
        'gnome': {'darkvision': {'range-feet': 60}, 'clockwork-device': {'creation-minutes': 10,
            'armor-class': 5, 'hit-points': 1, 'maximum-devices': 3, 'duration-hours': 8}},
        'goliath': {'giant-ancestry/cloud': {'range-feet': 30}, 'giant-ancestry/fire': {'damage-dice': 1, 'damage-die': 10},
            'giant-ancestry/frost': {'damage-dice': 1, 'damage-die': 6, 'speed-reduction-feet': 10},
            'giant-ancestry/stone': {'reduction-dice': 1, 'reduction-die': 12},
            'giant-ancestry/storm': {'range-feet': 60, 'damage-dice': 1, 'damage-die': 8},
            'large-form': {'minimum-character-level': 5, 'duration-minutes': 10, 'speed-increase-feet': 10},
            'powerful-build': {'carrying-size-increase': 1}},
        'halfling': {'luck': {'trigger-d20-roll': 1}}, 'human': {},
        'orc': {'darkvision': {'range-feet': 120}, 'relentless-endurance': {'trigger-hit-points': 0, 'replacement-hit-points': 1}},
        'tiefling': {'darkvision': {'range-feet': 60}},
    }
    tables = {
        'elf': {'drow': ('dancing-lights', 'faerie-fire', 'darkness'),
                'high-elf': ('prestidigitation', 'detect-magic', 'misty-step'),
                'wood-elf': ('druidcraft', 'longstrider', 'pass-without-trace')},
        'tiefling': {'abyssal': ('poison-spray', 'ray-of-sickness', 'hold-person'),
                    'chthonic': ('chill-touch', 'false-life', 'ray-of-enfeeblement'),
                    'infernal': ('fire-bolt', 'hellish-rebuke', 'darkness')},
    }
    for name, table in tables.items():
        for lineage, spells in table.items():
            for level, spell in zip((1, 3, 5), spells):
                numbers[name]['spell/' + lineage + '/' + spell] = {'minimum-character-level': level}
                if level > 1:
                    numbers[name]['spell/' + lineage + '/' + spell]['free-casts'] = 1
                assert M + 'spells/' + spell in traits[name]['spell/' + lineage + '/' + spell]['required_mechanics']
                label = spell.replace('-', ' ').lower()
                assert label in pages[85 if name == 'elf' else 86].lower(), label
    for spell in ('minor-illusion', 'speak-with-animals', 'mending', 'prestidigitation'):
        numbers['gnome']['spell/' + spell] = {'minimum-character-level': 1}
    numbers['tiefling']['spell/thaumaturgy'] = {'minimum-character-level': 1}
    for name, ts in traits.items():
        actual = {slug: {k.removeprefix(M + 'species/' + name + '/' + slug + '/'): v
                         for k, v in t['parameters'].items()} for slug, t in ts.items() if t['parameters']}
        assert actual == numbers[name], (name, 'complete numeric metadata')
    records = {r['id']: r for r in c['records']}
    choices = {r['id'].split('/choice/')[1]: r for r in c['records'] if r['kind'] == 'choice'}
    ancestors = dict(black='acid', blue='lightning', brass='fire', bronze='lightning', copper='acid',
                     gold='fire', green='poison', red='fire', silver='cold', white='cold')
    assert choices['dragonborn/draconic-ancestry']['option_keys'] == [M + 'dragon/' + a for a in ancestors]
    for ancestry, damage in ancestors.items():
        t = traits['dragonborn']['ancestry/' + ancestry]
        prefix = M + 'species/dragonborn/ancestry/' + ancestry + '/'
        assert t['required_mechanics'] == [prefix + 'breath-damage/' + damage,
            prefix + 'resistance/' + damage, prefix + 'appearance-matches-ancestry']
        predicate = records[t['prerequisite_refs'][0]]['requirement']
        assert predicate == dict(key=M + 'species/option-selected', record_refs=[choices['dragonborn/draconic-ancestry']['id']],
                                 option_key=M + 'dragon/' + ancestry, parameters={})
        assert re.search(ancestry.title() + r'\s+' + damage.title(), pages[84])
    option_sets = {
        'dragonborn/breath-shape': ['shape/cone', 'shape/line'],
        'elf/elven-lineage': ['lineage/drow', 'lineage/high-elf', 'lineage/wood-elf'],
        'elf/keen-senses': ['skills/insight', 'skills/perception', 'skills/survival'],
        'gnome/gnomish-lineage': ['lineage/forest-gnome', 'lineage/rock-gnome'],
        'goliath/giant-ancestry': ['giant/' + g for g in ('cloud', 'fire', 'frost', 'hill', 'stone', 'storm')],
        'tiefling/fiendish-legacy': ['lineage/abyssal', 'lineage/chthonic', 'lineage/infernal'],
        'human/size': ['size/medium', 'size/small'], 'tiefling/size': ['size/medium', 'size/small'],
        'gnome/device-ignite-extinguish': ['spells/prestidigitation/ignite', 'spells/prestidigitation/extinguish'],
    }
    for name in ('elf', 'gnome', 'tiefling'):
        option_sets[name + '/spellcasting-ability'] = ['abilities/intelligence', 'abilities/wisdom', 'abilities/charisma']
    for slug, options in option_sets.items():
        assert choices[slug]['select'] == 1 and choices[slug]['option_keys'] == [M + k for k in options], slug
    dynamic = [r for r in c['records'] if r.get('capacity_mechanic')]
    assert len(dynamic) == 5
    assert all(r['capacity'] == 0 and r['capacity_mechanic'] == M + 'character/proficiency-bonus' for r in dynamic)
    breath = traits['dragonborn']['breath-weapon']
    assert breath['character_levels'] == [dict(level=n, parameters={M + 'species/dragonborn/breath-weapon/damage-dice': d})
                                           for n, d in ((5, 2), (11, 3), (17, 4))]
    assert all(not t['character_levels'] for name, ts in traits.items() for slug, t in ts.items()
               if (name, slug) != ('dragonborn', 'breath-weapon'))
    # Edition-specific clauses anchor the opaque condition keys and continuations.
    anchors = {
        84: ['15-foot Cone', '30-foot Line that is 5 feet wide', 'character levels 5 (2d10), 11 (3d10), and 17 (4d10)',
             '8 plus your Constitution modifier and Proficiency Bonus', 'or have the Incapacitated condition',
             'on a stone surface or touching a stone sur', 'whenever you gain a level'],
        85: ['Insight, Perception, or Survival', 'Long Rest in 4 hours', 'Wisdom, and Charisma saving throws',
             'three such devices', '8 hours', 'ignites or extinguishes fire', 'the device doesn’t do both',
             'Large or smaller creature', 'until the start of your next turn', 'Add your Constitution modifier'],
        86: ['Storm', '1d8 Thunder damage', 'big enough space', 'Advantage on Strength checks',
             'Short or Long Rest', '0 Hit Points but not killed outright', 'you must use the new roll',
             'Origin feat of your choice', 'Skilled is recommended'],
    }
    for page, fragments in anchors.items():
        for fragment in fragments:
            assert fragment.replace('’', "'") in pages[page], (page, fragment)
    print('PASS: pinned PDF, nine identities/pages, complete numeric values, lineage spell tables, character-level rows and source conditions')


if __name__ == '__main__':
    if len(sys.argv) != 2:
        raise SystemExit('usage: verify_srd_species.py /path/to/SRD_CC_v5.2.1.pdf')
    verify(Path(sys.argv[1]))
