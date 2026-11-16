#!/usr/bin/env python3
"""Standalone pinned PDF -> actual #320 metadata comparator (no Go/network)."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[1]
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
C = 'srd-5.2.1:character/'
M = 'srd-5.2.1:mechanic/multiclass/'
NAMES = 'barbarian bard cleric druid fighter monk paladin ranger rogue sorcerer warlock wizard'.split()
PAGES = [28, 31, 36, 41, 47, 49, 53, 57, 61, 64, 70, 77]


def extract(pdf, page, x=0, width=612):
    raw = subprocess.check_output(['pdftotext', '-f', str(page), '-l', str(page),
                                   '-x', str(x), '-y', '0', '-W', str(width),
                                   '-H', '792', '-layout', str(pdf), '-'], text=True)
    return raw


def normalized(raw):
    return ' '.join(re.sub(r'-\n\s*', '', raw.replace('\u00ad', '')).split())


def source_expectations(pdf):
    """Read ability alternatives, restricted proficiency text and all 180 cells.

    Requirement clauses below are a separate semantic transcription of pp24–26;
    source assertions bind each group to its actual PDF paragraph, not a digest.
    """
    classes = []
    for name, page in zip(NAMES, PAGES):
        columns = [normalized(extract(pdf, page, x, 305)) for x in (0, 305)]
        text = ' '.join(columns)
        ability = re.search(r'Primary Ability (.*?) Hit Point Die', text)
        assert ability, (name, 'primary ability')
        primary = [[a.lower() for a in group.split(' and ')]
                   for group in ability[1].split(' or ')]
        block = next(col.split('As a Multiclass Character', 1)[1]
                     for col in columns if 'As a Multiclass Character' in col)
        block = block.split('• Gain the '+name.title()+'’s level 1 features')[0]
        # Translate only the listed multiclass subset; do not import first-class traits.
        profs = []
        if 'Martial weapons' in block:
            profs.append('martial-weapons')
        if 'one skill' in block:
            profs.append('any-skill' if name == 'bard' else 'class-skill')
        if 'Musical Instrument' in block:
            profs.append('musical-instrument')
        if 'Thieves’ Tools' in block:
            profs.append('thieves-tools')
        if 'Light' in block:
            profs.append('light-armor')
        if 'Medium' in block:
            profs.append('medium-armor')
        if 'Shields' in block:
            profs.append('shields')
        owners = [C+'class/'+name]
        if name in ('bard', 'ranger', 'rogue'):
            owners.append(C+'choice/'+name+'/multiclass-skill')
        if name == 'bard':
            owners.append(C+'choice/bard/multiclass-instrument')
        classes.append(dict(class_ref=C+'class/'+name,
                            source=dict(inventory_id='srd-5.2.1:planning/classes/'+name, page=page),
                            minimum_score=13, primary_any=primary,
                            proficiency_owner_refs=owners, proficiencies=profs))
    pages = {p: normalized(extract(pdf, p, 0, 305))+' '+normalized(extract(pdf, p, 305, 307))
             for p in (24, 25, 26)}
    assert 'at least 13' in pages[24]
    rules = []
    all_classes = ['class/'+n for n in NAMES]
    spellcasting = ['feature/'+n+'/spellcasting' for n in
                    'bard cleric druid paladin ranger sorcerer wizard'.split()]

    def rule(key, first, last, clauses, refs, nums, anchors):
        for anchor in anchors:
            assert anchor in ' '.join(pages[p] for p in range(first, last+1)), (key, anchor)
        rules.append(dict(key=M+key, first_page=first, last_page=last,
                          clauses=clauses.split(';'), record_refs=[C+r for r in refs], parameters=nums))

    rule('prerequisites', 24, 25,
         'new-and-every-current-class;each-primary-ability-at-least-minimum;alternatives-preserved', [], {},
         ['new class and your current classes', 'at least 13'])
    rule('experience-and-proficiency', 25, 25,
         'use-total-character-level;reuse-global-advancement', ['advancement/'+str(i) for i in range(1, 21)], {},
         ['Experience Point cost', 'total character level', 'Proficiency Bonus', 'Character Advancement table'])
    rule('hit-points-and-dice', 25, 25,
         'new-class-hp-as-levels-after-first;first-level-hp-only-at-total-level-one;sum-class-hit-dice;same-die-type-may-pool;different-die-types-tracked-separately',
         all_classes, {'first-level-hp-total-level': 1},
         ['levels after 1', 'total character level is 1', 'same die type', 'track them separately'])
    rule('proficiencies', 25, 25,
         'first-level-in-noninitial-class;only-listed-starting-proficiencies;no-other-core-traits', [], {},
         ['first level in a class other than your initial class', 'only some'])
    rule('class-features', 25, 25,
         'new-class-level-features;check-class-specific-multiclass-rules', all_classes, {},
         ['features for that level', 'each of your classes’ descriptions'])
    rule('armor-class', 25, 25,
         'multiple-calculations;benefit-from-only-one-at-a-time;monk-unarmored-defense-or-sorcerer-draconic-resilience-example',
         ['feature/monk/unarmored-defense'], {'simultaneous-calculations': 1},
         ['only one at a time', 'Draconic Resilience', 'choose only one'])
    rule('extra-attack', 25, 25,
         'extra-attack-from-multiple-classes-does-not-stack;maximum-two-unless-feature-explicitly-says-more;fighter-two-extra-attacks-exception;thirsting-blade-pact-weapon-no-additional-attacks-with-extra-attack',
         ['feature/fighter/extra-attack', 'feature/fighter/two-extra-attacks', 'feature/warlock/eldritch-invocations'],
         {'ordinary-maximum-attacks': 2}, ['features don’t stack', 'Two Extra Attacks', 'Thirsting Blade', 'pact weapon'])
    rule('spellcasting-scope', 25, 25,
         'combined-and-individual-class-levels;shared-rules-only-with-spellcasting-from-more-than-one-class;one-spellcasting-class-use-own-class-rules',
         spellcasting, {'minimum-spellcasting-classes': 2},
         ['more than one class', 'only one class', 'follow the rules for that class'])
    rule('prepared-spells', 25, 26,
         'prepare-per-class-as-single-class;each-prepared-spell-associated-with-one-class;use-that-class-spellcasting-ability;higher-slots-do-not-unlock-higher-preparations',
         spellcasting, {}, ['each class individually', 'associated with one of your classes',
                           'spellcasting ability of that class', 'can’t prepare any level 3 spells'])
    rule('cantrips', 25, 25,
         'power-scaling-total-character-level;unless-spell-says-otherwise', [], {},
         ['increase is based on your total character level', 'unless the spell says otherwise'])
    rule('spell-slots', 25, 26,
         'sum-full-caster-levels;add-half-paladin-and-ranger-levels-rounded-up;look-up-total-in-shared-table;cast-appropriate-level-spells-from-any-owned-spellcasting-class;higher-slots-only-cast-prepared-lower-level-spells;enhanced-effect-only-if-spell-provides-it',
         spellcasting, {'full-caster-divisor': 1, 'paladin-divisor': 2, 'ranger-divisor': 2, 'round-up': 1},
         ['Bard, Cleric, Druid, Sorcerer, and Wizard', 'Half your levels (round up)',
          'Paladin and Ranger', 'only to cast your lower-level spells', 'has an enhanced effect'])
    rule('pact-magic', 26, 26,
         'requires-warlock-pact-magic-and-spellcasting;pact-slots-cast-prepared-spellcasting-class-spells;spellcasting-slots-cast-prepared-warlock-spells;pact-levels-excluded-from-shared-slot-sum',
         ['feature/warlock/pact-magic']+spellcasting, {},
         ['Pact Magic feature from the Warlock class and the Spellcasting feature',
          'from Pact Magic to cast spells', 'from the Spellcasting feature to cast Warlock spells'])
    rows = []
    for line in extract(pdf, 26, 0, 305).splitlines():
        cells = line.split()
        if len(cells) == 10 and cells[0].isdigit() and all(re.fullmatch(r'\d+|—', s) for s in cells):
            rows.append(dict(level=int(cells[0]), slots=[0 if s == '—' else int(s) for s in cells[1:]]))
    assert [r['level'] for r in rows] == list(range(1, 21))
    return classes, rules, rows


def compare(candidate, expected):
    classes, rules, rows = expected
    assert set(candidate) == {'schema_version', 'source', 'scope', 'catalogues', 'classes', 'requirements', 'slot_rows', 'slot_terms'}
    assert candidate['schema_version'] == 1
    assert candidate['scope'] == dict(issue=320, section='multiclassing', first_page=24, last_page=26)
    assert candidate['catalogues'] == ['character-'+n+'.json' for n in NAMES]+['character-advancement.json']
    assert candidate['classes'] == classes, 'class prerequisite/proficiency/source/ownership mismatch'
    assert candidate['requirements'] == rules, 'requirement operands/clauses/parameters/source mismatch'
    assert candidate['slot_terms'] == [dict(class_ref=C+'class/'+n,
        level_divisor=2 if n in ('paladin', 'ranger') else 1,
        rounding='up' if n in ('paladin', 'ranger') else 'none')
        for n in 'bard cleric druid paladin ranger sorcerer wizard'.split()], 'slot formula operands'
    assert candidate['slot_rows'] == rows, 'actual PDF table cells mismatch'
    source = json.loads((ROOT/'internal/rules/srd521/testdata/character-advancement.json').read_text())['source']
    source['changes'] = 'Multiclass section and table adapted to non-executable metadata. Edda IDs, scope bindings and unresolved requirements added; canonical class and advancement owners referenced.'
    assert candidate['source'] == source and source['sha256'] == PIN, 'source pin/attribution/adaptation'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('pdf', type=Path)
    parser.add_argument('candidate', nargs='?', type=Path, default=ROOT/'internal/rules/srd521/testdata/character-multiclass.json')
    args = parser.parse_args()
    assert hashlib.sha256(args.pdf.read_bytes()).hexdigest() == PIN, 'PDF SHA256'
    expected = source_expectations(args.pdf)
    compare(json.loads(args.candidate.read_text()), expected)
    print('PASS: all 12 classes, 180 actual PDF cells, all requirement fields and canonical source bindings')


if __name__ == '__main__':
    main()
